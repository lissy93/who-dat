package rdap

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/model"
)

var update = flag.Bool("update", false, "update golden files")

// goldenCases maps a frozen RDAP fixture to the domain it represents.
var goldenCases = map[string]domain.Name{
	"example.com.json": {ASCII: "example.com", Unicode: "example.com", TLD: "com"},
	"vaduz.li.json":    {ASCII: "vaduz.li", Unicode: "vaduz.li", TLD: "li"}, // SWITCH: bare event dates, org-only registrar vcard
}

func TestParseEventDate(t *testing.T) {
	cases := map[string]string{
		"2021-03-09T14:30:00Z":      "2021-03-09T14:30:00Z",
		"2021-03-09T14:30:00+02:00": "2021-03-09T12:30:00Z",
		"2021-03-09T14:30:00":       "2021-03-09T14:30:00Z",
		"1997-06-24":                "1997-06-24T00:00:00Z",
		" 1997-06-24 ":              "1997-06-24T00:00:00Z",
		"24.06.1997":                "",
		"":                          "",
	}
	for in, want := range cases {
		got := ""
		if t2 := parseEventDate(in); !t2.IsZero() {
			got = t2.UTC().Format(time.RFC3339)
		}
		if got != want {
			t.Errorf("parseEventDate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisclosed(t *testing.T) {
	cases := map[string]bool{
		"REDACTED FOR PRIVACY":       false,
		"Redacted | EU Data Subject": false,
		"[REDACTED]":                 false,
		"Unredacted Labs":            true,
		"Jane Doe":                   true,
	}
	for in, want := range cases {
		if got := disclosed(in) != nil; got != want {
			t.Errorf("disclosed(%q) kept = %v, want %v", in, got, want)
		}
	}
}

// Gandi labels its RDAP server as "about"; the website is the other link.
func TestRegistrarURLSkipsRDAPLinks(t *testing.T) {
	r := model.New("gandi.net", "gandi.net", "net")
	mapRegistrar(r, rdapEntity{Links: []rdapLink{
		{Rel: "service", Href: "https://www.gandi.net/"},
		{Rel: "about", Type: "application/rdap+json", Href: "https://rdap.gandi.net/"},
	}})
	if got := str(r.Registrar.URL); got != "https://www.gandi.net/" {
		t.Errorf("registrar.url = %q, want the website", got)
	}
}

func TestUpdatedIgnoresDatabaseTimestamp(t *testing.T) {
	r := model.New("nic.ye", "nic.ye", "ye")
	mapEvents(r, []rdapEvent{{Action: "last update of RDAP database", Date: "2026-11-03T18:27:51Z"}})
	if r.Dates.Updated != nil {
		t.Errorf("dates.updated = %v, want null", r.Dates.Updated)
	}
}

func TestMapDomainGolden(t *testing.T) {
	for file, n := range goldenCases {
		t.Run(file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := mapDomain(n, "https://rdap.example/", raw)
			if err != nil {
				t.Fatalf("mapDomain: %v", err)
			}
			// Zero the volatile/non-serialized fields so the golden is deterministic.
			got.Meta.FetchedAt = time.Time{}
			got.Raw, got.RawContentType = nil, ""

			out, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			golden := filepath.Join("testdata", strings.TrimSuffix(file, ".json")+".golden.json")
			if *update {
				if err := os.WriteFile(golden, out, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run -update first): %v", err)
			}
			if !bytes.Equal(bytes.TrimSpace(out), bytes.TrimSpace(want)) {
				t.Errorf("mapped result mismatch for %s\n--- got ---\n%s", file, out)
			}
		})
	}
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
