package rdap

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAddressCountry(t *testing.T) {
	tests := map[string]struct{ params, name, want string }{
		"code only":  {`{"cc":"US"}`, "", "US"},
		"code array": {`{"cc":["US"]}`, "", "US"},
		"name only":  {`{}`, "Finland", "Finland"},
		"both":       {`{"cc":"IS"}`, "Iceland", "IS"},
		"neither":    {`{"type":"work"}`, "", ""},
	}
	for name, tt := range tests {
		card := fmt.Sprintf(`["vcard",[["adr",%s,"text",["","","1 Main St","Springfield","","","%s"]]]]`, tt.params, tt.name)
		var v vcard
		if err := json.Unmarshal([]byte(card), &v); err != nil {
			t.Fatal(err)
		}
		if got := str(v.address().Country); got != tt.want {
			t.Errorf("%s: country = %q, want %q", name, got, tt.want)
		}
	}
}
