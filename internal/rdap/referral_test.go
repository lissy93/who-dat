package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/model"
)

const registrarURL = "https://rdap.markmonitor.com/rdap/domain/GOOGLE.COM"

func TestReferral(t *testing.T) {
	const self = "https://registry.example/domain/x.com"
	links := func(rel, typ, href string) []byte {
		return fmt.Appendf(nil, `{"links":[{"rel":%q,"type":%q,"href":%q}]}`, rel, typ, href)
	}
	tests := map[string]struct {
		raw  []byte
		want string
	}{
		"registrar":    {links("related", "application/rdap+json", "https://rar.example/x.com"), "https://rar.example/x.com"},
		"type params":  {links("related", "application/rdap+json; charset=utf-8", "https://rar.example/x.com"), "https://rar.example/x.com"},
		"untyped":      {links("related", "", "https://rar.example/x.com"), "https://rar.example/x.com"},
		"html page":    {links("related", "text/html", "https://rar.example/"), ""},
		"plain http":   {links("related", "application/rdap+json", "http://rar.example/x.com"), ""},
		"back to self": {links("related", "application/rdap+json", self), ""},
		"no links":     {[]byte(`{}`), ""},
	}
	for name, tt := range tests {
		if got := referral(tt.raw, self); got != tt.want {
			t.Errorf("%s: referral = %q, want %q", name, got, tt.want)
		}
	}
}

func TestLookupMergesRegistrar(t *testing.T) {
	r := lookupGoogle(t, fixture(t, "google.com.registrar.json"))

	checks := []struct{ field, got, want string }{
		// The registry keeps the registration facts...
		{"registrar.name", str(r.Registrar.Name), "MarkMonitor Inc."},
		{"dates.expires", stamp(r.Dates.Expires), "2028-09-14T04:00:00Z"},
		// ...and the registrar adds what only it knows.
		{"dates.updated", stamp(r.Dates.Updated), "2024-08-02T02:17:33Z"},
		{"registrar.whoisServer", str(r.Registrar.WhoisServer), "whois.markmonitor.com"},
		{"registrant.state", str(r.Contacts.Registrant.Address.State), "CA"},
		{"registrant.country", str(r.Contacts.Registrant.Address.Country), "US"},
		{"registrant.name", str(r.Contacts.Registrant.Name), ""}, // was "REDACTED REGISTRANT"
		{"meta.referral", str(r.Meta.Referral), registrarURL},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.field, c.got, c.want)
		}
	}

	var raw rawPair
	if err := json.Unmarshal(r.Raw, &raw); err != nil || raw.Registry == nil || raw.Registrar == nil {
		t.Errorf("raw should hold both documents, got %.80s", r.Raw)
	}
}

func TestLookupIgnoresBadReferral(t *testing.T) {
	registrar := map[string]string{
		"registrar down": "",
		"not json":       "<html>down for maintenance</html>",
		"wrong domain":   `{"objectClassName":"domain","ldhName":"example.com"}`,
	}
	for name, body := range registrar {
		t.Run(name, func(t *testing.T) {
			r := lookupGoogle(t, body)
			if r.Meta.Referral != nil || r.RawContentType != contentType {
				t.Errorf("merged a bad referral: meta.referral = %q", str(r.Meta.Referral))
			}
			if got := stamp(r.Dates.Updated); got != "2019-09-09T15:39:04Z" {
				t.Errorf("dates.updated = %s, want the registry's", got)
			}
		})
	}
}

func TestTakeContact(t *testing.T) {
	nothing := model.Contact{Redacted: true}
	stateOnly := model.Contact{Redacted: true, Address: model.Address{State: model.Str("CA")}}
	stateAndCountry := model.Contact{Redacted: true, Address: model.Address{State: model.Str("CA"), Country: model.Str("US")}}
	registryCo := model.Contact{Organization: model.Str("Registry Co")}
	registrarCo := model.Contact{Organization: model.Str("Registrar Co")}

	tests := []struct {
		name                      string
		registry, registrar, want model.Contact
	}{
		{"registrar sent nothing", registryCo, nothing, registryCo},
		{"registrar redacts more", registryCo, stateOnly, registryCo},
		{"registrar adds a little", nothing, stateOnly, stateOnly},
		{"registrar has less address", stateAndCountry, stateOnly, stateAndCountry},
		{"both disclose", registryCo, registrarCo, registrarCo},
	}
	for _, tt := range tests {
		got := tt.registry
		takeContact(&got, tt.registrar)
		if got != tt.want {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

// lookupGoogle replays Verisign's google.com record. An empty registrar means it's down.
func lookupGoogle(t *testing.T, registrar string) *model.Result {
	t.Helper()
	upstream := fakeUpstream{
		bootstrapURL: `{"services":[[["com"],["https://rdap.verisign.com/com/v1/"]]]}`,
		"https://rdap.verisign.com/com/v1/domain/google.com": fixture(t, "google.com.json"),
	}
	if registrar != "" {
		upstream[registrarURL] = registrar
	}
	c := NewClient(&http.Client{Transport: upstream})
	r, err := c.Lookup(context.Background(), domain.Name{ASCII: "google.com", Unicode: "google.com", TLD: "com"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// fakeUpstream serves a canned body per URL, and 503 for anything else.
type fakeUpstream map[string]string

func (f fakeUpstream) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	if body, ok := f[req.URL.String()]; ok {
		rec.WriteString(body)
	} else {
		rec.WriteHeader(http.StatusServiceUnavailable)
	}
	return rec.Result(), nil
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func stamp(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
