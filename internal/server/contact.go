package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lissy93/who-dat/internal/config"
)

// contactInbox is the only recipient of the contact form.
const contactInbox = "i@wi.fs.kg"

const (
	contactSubjectZH = "[wi.fs.kg] 来自联系页面"
	contactSubjectEN = "[wi.fs.kg] from the contact page"

	contactMaxName    = 80
	contactMaxEmail   = 200
	contactMinMessage = 5
	contactMaxMessage = 5000
	contactMaxPage    = 300
	contactMaxUA      = 400
)

// Contact error codes. The page maps these to localized copy.
const (
	codeContactBadJSON       = "BAD_JSON"
	codeContactEmail         = "INVALID_EMAIL"
	codeContactShort         = "MESSAGE_TOO_SHORT"
	codeContactLong          = "MESSAGE_TOO_LONG"
	codeContactRejected      = "REJECTED"
	codeContactTurnstileNeed = "TURNSTILE_REQUIRED"
	codeContactTurnstileFail = "TURNSTILE_FAILED"
	codeContactNotConfigured = "NOT_CONFIGURED"
	codeContactRate          = "RATE_LIMITED"
	codeContactSend          = "SEND_FAILED"
	codeContactMethod        = "METHOD_NOT_ALLOWED"
)

type contactRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Message   string `json:"message"`
	Website   string `json:"website"` // honeypot; real submissions leave this empty
	Turnstile string `json:"turnstile"`
	Page      string `json:"page"`
	Lang      string `json:"lang"`
}

type contactResult struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
}

type contactConfigBody struct {
	SiteKey string `json:"siteKey"`
	Inbox   string `json:"inbox"`
	Ready   bool   `json:"ready"`
}

type resendMail struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	ReplyTo string   `json:"reply_to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

type turnstileResult struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// handleContactConfig exposes the public Turnstile site key. Secrets stay on the server.
func (s *server) handleContactConfig(w http.ResponseWriter, r *http.Request) {
	writeContact(w, http.StatusOK, contactConfigBody{
		SiteKey: s.cfg.TurnstileSiteKey,
		Inbox:   contactInbox,
		Ready:   s.contactReady(),
	})
}

// handleContactMethod answers GET /api/contact so the path is not parsed as a domain.
func (s *server) handleContactMethod(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "POST, OPTIONS")
	writeContactCode(w, http.StatusMethodNotAllowed, codeContactMethod)
}

// handleContact accepts the contact form, verifies Turnstile, and relays via Resend.
func (s *server) handleContact(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 48<<10))
	if err := dec.Decode(&req); err != nil {
		writeContactCode(w, http.StatusBadRequest, codeContactBadJSON)
		return
	}

	// A filled honeypot is a bot. Reject it without calling Turnstile or Resend.
	if strings.TrimSpace(req.Website) != "" {
		writeContactCode(w, http.StatusBadRequest, codeContactRejected)
		return
	}

	name := cleanLine(req.Name, contactMaxName)
	email := cleanLine(req.Email, contactMaxEmail)
	message := cleanBlock(req.Message, contactMaxMessage+1) // +1 so over-long stays detectable
	if !validEmail(email) {
		writeContactCode(w, http.StatusBadRequest, codeContactEmail)
		return
	}
	msgRunes := utf8.RuneCountInString(message)
	if msgRunes < contactMinMessage {
		writeContactCode(w, http.StatusBadRequest, codeContactShort)
		return
	}
	if msgRunes > contactMaxMessage {
		writeContactCode(w, http.StatusBadRequest, codeContactLong)
		return
	}

	if s.contactLimit != nil {
		if ok, retry := s.contactLimit.allow(clientIP(r)); !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeContactCode(w, http.StatusTooManyRequests, codeContactRate)
			return
		}
	}

	if !s.contactReady() {
		writeContactCode(w, http.StatusServiceUnavailable, codeContactNotConfigured)
		return
	}

	token := strings.TrimSpace(req.Turnstile)
	if token == "" {
		writeContactCode(w, http.StatusBadRequest, codeContactTurnstileNeed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	ok, err := s.verifyTurnstile(ctx, token, clientIP(r))
	if err != nil {
		slog.Error("turnstile verify", "err", err)
		writeContactCode(w, http.StatusServiceUnavailable, codeContactTurnstileFail)
		return
	}
	if !ok {
		writeContactCode(w, http.StatusForbidden, codeContactTurnstileFail)
		return
	}

	lang := "zh"
	if strings.EqualFold(strings.TrimSpace(req.Lang), "en") {
		lang = "en"
	}
	subject := contactSubjectZH
	if lang == "en" {
		subject = contactSubjectEN
	}
	if name != "" {
		subject += " — " + name
	}

	mail := resendMail{
		From:    s.contactFrom(),
		To:      []string{contactInbox},
		ReplyTo: email,
		Subject: subject,
		Text:    contactBody(name, email, cleanLine(req.Page, contactMaxPage), lang, cleanLine(r.UserAgent(), contactMaxUA), message),
	}
	if strings.ContainsAny(mail.From, "\r\n") {
		slog.Error("contact from address rejected")
		writeContactCode(w, http.StatusServiceUnavailable, codeContactSend)
		return
	}
	if err := s.sendResend(ctx, mail); err != nil {
		slog.Error("contact send", "err", err)
		writeContactCode(w, http.StatusServiceUnavailable, codeContactSend)
		return
	}
	writeContact(w, http.StatusOK, contactResult{OK: true})
}

func (s *server) contactReady() bool {
	c := s.cfg
	return c.TurnstileSiteKey != "" && c.TurnstileSecretKey != "" && c.ResendAPIKey != ""
}

func (s *server) contactFrom() string {
	if s.cfg.ContactFrom != "" {
		return s.cfg.ContactFrom
	}
	return config.DefaultContactFrom
}

func (s *server) verifyTurnstile(ctx context.Context, token, remoteIP string) (bool, error) {
	form := url.Values{}
	form.Set("secret", s.cfg.TurnstileSecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.turnstileURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("turnstile status %d", resp.StatusCode)
	}
	var out turnstileResult
	if err := json.Unmarshal(body, &out); err != nil {
		return false, fmt.Errorf("turnstile decode: %w", err)
	}
	if !out.Success {
		slog.Warn("turnstile rejected contact", "codes", strings.Join(out.ErrorCodes, ","))
	}
	return out.Success, nil
}

func (s *server) sendResend(ctx context.Context, mail resendMail) error {
	payload, err := json.Marshal(mail)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.resendURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend status %d: %s", resp.StatusCode, truncate(string(body), 180))
	}
	return nil
}

func writeContact(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, v)
}

func writeContactCode(w http.ResponseWriter, status int, code string) {
	writeContact(w, status, contactResult{OK: false, Code: code, Error: contactMessage(code)})
}

func contactMessage(code string) string {
	switch code {
	case codeContactBadJSON:
		return "Could not read that message."
	case codeContactEmail:
		return "Email address looks wrong."
	case codeContactShort:
		return "Write a few more words."
	case codeContactLong:
		return "Message is too long."
	case codeContactRejected:
		return "Could not accept this message."
	case codeContactTurnstileNeed:
		return "Security check has not finished."
	case codeContactTurnstileFail:
		return "Security check failed. Write to " + contactInbox + " instead."
	case codeContactNotConfigured:
		return "Contact mail is not configured on this server. Write to " + contactInbox + " instead."
	case codeContactRate:
		return "Too many messages. Try again shortly, or write to " + contactInbox + "."
	case codeContactSend:
		return "Could not send the message. Write to " + contactInbox + " instead."
	case codeContactMethod:
		return "Method not allowed."
	default:
		return "Could not send the message."
	}
}

func contactBody(name, email, page, lang, ua, message string) string {
	var b strings.Builder
	b.WriteString("wi.fs.kg contact form\n\n")
	b.WriteString("Name: " + dash(name) + "\n")
	b.WriteString("Email: " + email + "\n")
	b.WriteString("Page: " + dash(page) + "\n")
	b.WriteString("Lang: " + lang + "\n")
	b.WriteString("User-Agent: " + dash(ua) + "\n\n")
	b.WriteString(message)
	b.WriteString("\n")
	return b.String()
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func validEmail(s string) bool {
	if s == "" || len(s) > contactMaxEmail || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	at := strings.LastIndex(s, "@")
	if at <= 0 || at >= len(s)-1 {
		return false
	}
	local, domain := s[:at], s[at+1:]
	if strings.ContainsAny(local, ",;:<>()") || strings.ContainsAny(domain, ",;:<>()") {
		return false
	}
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}

// cleanLine trims, strips control characters, and caps rune length.
func cleanLine(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)
	return clipRunes(s, maxRunes)
}

// cleanBlock is cleanLine that keeps newlines and tabs so a message can stay multi-line.
func cleanBlock(s string, maxRunes int) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)
	return clipRunes(s, maxRunes)
}

func clipRunes(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
