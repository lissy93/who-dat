package whois

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
)

var update = flag.Bool("update", false, "update golden files")

// goldenCases maps a frozen WHOIS fixture to the domain it represents.
var goldenCases = map[string]domain.Name{
	"google.it.txt":  {ASCII: "google.it", Unicode: "google.it", TLD: "it"},
	"jprs.co.jp.txt": {ASCII: "jprs.co.jp", Unicode: "jprs.co.jp", TLD: "co.jp"},
}

func TestMapWhoisGolden(t *testing.T) {
	for file, n := range goldenCases {
		t.Run(file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := mapWhois(n, string(raw))
			if err != nil {
				t.Fatalf("mapWhois: %v", err)
			}
			got.Meta.FetchedAt = time.Time{}
			got.Raw, got.RawContentType = nil, ""

			out, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			golden := filepath.Join("testdata", strings.TrimSuffix(file, ".txt")+".golden.json")
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
