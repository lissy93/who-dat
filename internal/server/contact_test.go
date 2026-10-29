package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/lissy93/who-dat/internal/config"
	"github.com/lissy93/who-dat/internal/lookup"
	"github.com/lissy93/who-dat/internal/srcerr"
)

func contactApp(cfg *config.Config) *server {
	if cfg.LookupTimeout == 0 {
		cfg.LookupTimeout = 2_000_000_000
	}
	if cfg.MaxDomains == 0 {
		cfg.MaxDomains = 5
	}
	svc := lookup.NewService(&fakeSource{err: srcerr.ErrNoSource}, &fakeSource{err: srcerr.ErrNoSource}, nil)
	return newServer(cfg, svc)
}

func postContact(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.8:4242"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeContact(t *testing.T, rec *httptest.ResponseRecorder) contactResult {
	t.Helper()
	var out contactResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return out
}

func TestContactPageIsNotALookup(t *testing.T) {
	h := newTestServer(registered(), nil)

	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/contact", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", page.Code)
	}
	if ct := page.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("content type = %q", ct)
	}
	body := page.Body.String()
	for _, want := range []string{
		"有话直说",
		"一封信",
		contactInbox,
		"/contact.js",
		"/styles.css",
		"回到查询台",
		"数据不准",
		"name=\"website\"",
		"%5Bwi.fs.kg%5D",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("contact page missing %q", want)
		}
	}
	if strings.Contains(body, "0x4AAAAA") {
		t.Error("contact page must not hardcode a Turnstile site key")
	}

	js := httptest.NewRecorder()
	h.ServeHTTP(js, httptest.NewRequest(http.MethodGet, "/contact.js", nil))
	script := js.Body.String()
	if js.Code != http.StatusOK {
		t.Fatalf("contact.js status = %d", js.Code)
	}
	for _, want := range []string{"whois-ui-lang", "[wi.fs.kg] 来自联系页面", "RDAP 或 WHOIS", "\"/api/contact\""} {
		if !strings.Contains(script, want) {
			t.Errorf("contact.js missing %q", want)
		}
	}

	home := httptest.NewRecorder()
	h.ServeHTTP(home, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(home.Body.String(), `href="/contact"`) {
		t.Error("home page missing contact link")
	}
	css := httptest.NewRecorder()
	h.ServeHTTP(css, httptest.NewRequest(http.MethodGet, "/styles.css", nil))
	sheet := css.Body.String()
	for _, want := range []string{".btn-snap", ".page-contact", "--snap: #2ec4a8"} {
		if !strings.Contains(sheet, want) {
			t.Errorf("styles.css missing %q", want)
		}
	}

	// A real bare domain still resolves through the lookup route.
	lookupRec := httptest.NewRecorder()
	h.ServeHTTP(lookupRec, httptest.NewRequest(http.MethodGet, "/example.com", nil))
	if lookupRec.Code != http.StatusOK {
		t.Fatalf("bare lookup status = %d", lookupRec.Code)
	}
}

func TestContactConfigHidesSecrets(t *testing.T) {
	cfg := &config.Config{
		TurnstileSiteKey:   "site_public",
		TurnstileSecretKey: "turnstile_secret_value",
		ResendAPIKey:       "re_test_key",
		ContactFrom:        config.DefaultContactFrom,
	}
	h := contactApp(cfg).routes()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/contact/config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("cache-control = %q", cc)
	}
	var body contactConfigBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.SiteKey != "site_public" || body.Inbox != contactInbox || !body.Ready {
		t.Errorf("config = %+v", body)
	}
	raw := rec.Body.String()
	if strings.Contains(raw, "turnstile_secret_value") || strings.Contains(raw, "re_test_key") {
		t.Fatalf("config leaked a secret: %s", raw)
	}

	empty := contactApp(&config.Config{}).routes()
	rec = httptest.NewRecorder()
	empty.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/contact/config", nil))
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Ready || body.SiteKey != "" || body.Inbox != contactInbox {
		t.Errorf("unconfigured = %+v", body)
	}
}

func TestContactValidation(t *testing.T) {
	h := contactApp(&config.Config{}).routes()
	valid := `{"email":"ada@example.com","message":"WHOIS for example.com looks wrong","turnstile":"tok"}`

	tests := []struct {
		name   string
		body   string
		status int
		code   string
	}{
		{"bad json", "{", http.StatusBadRequest, codeContactBadJSON},
		{"email", `{"email":"not-an-email","message":"hello there"}`, http.StatusBadRequest, codeContactEmail},
		{"short", `{"email":"ada@example.com","message":"hi"}`, http.StatusBadRequest, codeContactShort},
		{"honeypot", `{"email":"ada@example.com","message":"hello there","website":"http://spam.test"}`, http.StatusBadRequest, codeContactRejected},
		{"not configured", valid, http.StatusServiceUnavailable, codeContactNotConfigured},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postContact(h, tt.body)
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.status, rec.Body)
			}
			got := decodeContact(t, rec)
			if got.OK || got.Code != tt.code || got.Error == "" {
				t.Fatalf("body = %+v, want code %s", got, tt.code)
			}
			if tt.code == codeContactNotConfigured && !strings.Contains(got.Error, contactInbox) {
				t.Errorf("error %q should point at the inbox", got.Error)
			}
		})
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/contact", nil))
	if rec.Code != http.StatusMethodNotAllowed || decodeContact(t, rec).Code != codeContactMethod {
		t.Fatalf("GET /api/contact = %d %s", rec.Code, rec.Body)
	}
}

func TestContactRateLimit(t *testing.T) {
	h := contactApp(&config.Config{ContactRatePerMinute: 60, ContactRateBurst: 1}).routes()
	valid := `{"email":"ada@example.com","message":"WHOIS for example.com looks wrong"}`

	// Invalid input does not spend the budget.
	if rec := postContact(h, `{"email":"nope","message":"hello there"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid = %d", rec.Code)
	}
	first := postContact(h, valid)
	if first.Code != http.StatusServiceUnavailable {
		t.Fatalf("first = %d %s", first.Code, first.Body)
	}
	second := postContact(h, valid)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second = %d, want 429 (%s)", second.Code, second.Body)
	}
	if decodeContact(t, second).Code != codeContactRate || second.Header().Get("Retry-After") == "" {
		t.Fatalf("rate body/header = %s / %q", second.Body, second.Header().Get("Retry-After"))
	}
}

func TestContactSendsMail(t *testing.T) {
	var mailed resendMail
	var mailCalls atomic.Int32
	var failSend atomic.Bool
	resend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failSend.Load() {
			http.Error(w, "nope", http.StatusForbidden)
			return
		}
		mailCalls.Add(1)
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&mailed); err != nil {
			t.Errorf("decode mail: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"email_1"}`)
	}))
	defer resend.Close()

	var sawSecret, sawToken string
	turnstile := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		sawSecret, sawToken = r.Form.Get("secret"), r.Form.Get("response")
		if sawToken == "bad" {
			_, _ = io.WriteString(w, `{"success":false,"error-codes":["invalid-input-response"]}`)
			return
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer turnstile.Close()

	cfg := &config.Config{
		TurnstileSiteKey:   "site_public",
		TurnstileSecretKey: "turnstile_secret_value",
		ResendAPIKey:       "re_test",
		ContactFrom:        "wi.fs.kg <onboarding@resend.dev>",
	}
	app := contactApp(cfg)
	app.turnstileURL = turnstile.URL
	app.resendURL = resend.URL
	h := app.routes()

	t.Run("honeypot skips upstream", func(t *testing.T) {
		before := mailCalls.Load()
		rec := postContact(h, `{"name":"Ada","email":"ada@example.com","message":"hello there","website":"http://spam.test","turnstile":"tok"}`)
		if rec.Code != http.StatusBadRequest || decodeContact(t, rec).Code != codeContactRejected {
			t.Fatalf("honeypot = %d %s", rec.Code, rec.Body)
		}
		if mailCalls.Load() != before {
			t.Fatal("honeypot must not send mail")
		}
	})

	t.Run("missing token", func(t *testing.T) {
		rec := postContact(h, `{"email":"ada@example.com","message":"hello there"}`)
		if rec.Code != http.StatusBadRequest || decodeContact(t, rec).Code != codeContactTurnstileNeed {
			t.Fatalf("missing token = %d %s", rec.Code, rec.Body)
		}
	})

	t.Run("turnstile reject", func(t *testing.T) {
		rec := postContact(h, `{"email":"ada@example.com","message":"hello there","turnstile":"bad"}`)
		if rec.Code != http.StatusForbidden || decodeContact(t, rec).Code != codeContactTurnstileFail {
			t.Fatalf("reject = %d %s", rec.Code, rec.Body)
		}
		if mailCalls.Load() != 0 {
			t.Fatal("failed turnstile must not send mail")
		}
	})

	t.Run("zh delivery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(
			`{"name":"Ada\nBcc: x","email":"ada@example.com","message":"WHOIS for example.com\nlooks wrong","website":"","turnstile":"tok","page":"/contact","lang":"zh"}`,
		))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "TestBrowser/1.0")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || !decodeContact(t, rec).OK {
			t.Fatalf("send = %d %s", rec.Code, rec.Body)
		}
		if sawSecret != "turnstile_secret_value" || sawToken != "tok" {
			t.Errorf("turnstile secret/token = %q / %q", sawSecret, sawToken)
		}
		if len(mailed.To) != 1 || mailed.To[0] != contactInbox {
			t.Errorf("to = %#v", mailed.To)
		}
		if mailed.ReplyTo != "ada@example.com" {
			t.Errorf("reply_to = %q", mailed.ReplyTo)
		}
		if mailed.From != cfg.ContactFrom {
			t.Errorf("from = %q", mailed.From)
		}
		if mailed.Subject != contactSubjectZH+" — AdaBcc: x" || strings.Contains(mailed.Subject, "\n") {
			t.Errorf("subject = %q", mailed.Subject)
		}
		if !strings.Contains(mailed.Text, "WHOIS for example.com\nlooks wrong") || !strings.Contains(mailed.Text, "TestBrowser/1.0") {
			t.Errorf("text = %q", mailed.Text)
		}
		if strings.Contains(mailed.Text, "spam.test") {
			t.Errorf("honeypot leaked into the body: %s", mailed.Text)
		}
	})

	t.Run("en subject", func(t *testing.T) {
		rec := postContact(h, `{"email":"ada@example.com","message":"hello there","turnstile":"tok","lang":"en"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("en send = %d %s", rec.Code, rec.Body)
		}
		if mailed.Subject != contactSubjectEN {
			t.Errorf("subject = %q", mailed.Subject)
		}
	})

	t.Run("resend failure", func(t *testing.T) {
		failSend.Store(true)
		rec := postContact(h, `{"email":"ada@example.com","message":"hello there","turnstile":"tok"}`)
		got := decodeContact(t, rec)
		if rec.Code != http.StatusServiceUnavailable || got.Code != codeContactSend || !strings.Contains(got.Error, contactInbox) {
			t.Fatalf("send failure = %d %+v", rec.Code, got)
		}
	})
}
