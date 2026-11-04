package rdap

import (
	"cmp"
	"encoding/json"
	"strings"

	"github.com/lissy93/who-dat/internal/model"
)

// vcard holds the jCard fields we care about. jCard (RFC 7095) encodes a vCard as
// ["vcard", [ [name, params, type, value], ... ]]; we pull out the common properties.
type vcard struct {
	fn    string
	org   string
	email string
	tel   string
	adr   []string // structured address: pobox, ext, street, locality, region, code, country
	cc    string   // ISO country code (RFC 8605); registrars often leave the name empty
}

// UnmarshalJSON parses a jCard array. It is deliberately lenient: malformed cards yield an
// empty vcard rather than failing the whole lookup.
func (v *vcard) UnmarshalJSON(b []byte) error {
	var outer []json.RawMessage
	if err := json.Unmarshal(b, &outer); err != nil || len(outer) < 2 {
		return nil
	}
	var props [][]json.RawMessage
	if err := json.Unmarshal(outer[1], &props); err != nil {
		return nil
	}
	for _, p := range props {
		if len(p) < 4 {
			continue
		}
		var name string
		_ = json.Unmarshal(p[0], &name)
		switch strings.ToLower(name) {
		case "fn":
			v.fn = jcardString(p[3])
		case "org":
			if v.org == "" {
				v.org = jcardString(p[3])
			}
		case "email":
			if v.email == "" {
				v.email = jcardString(p[3])
			}
		case "tel":
			if v.tel == "" {
				v.tel = jcardString(p[3])
			}
		case "adr":
			var params struct {
				CC string `json:"cc"`
			}
			_ = json.Unmarshal(p[1], &params)
			v.adr, v.cc = jcardStrings(p[3]), params.CC
		}
	}
	return nil
}

// name returns the display name, falling back to org when fn is absent (e.g. SWITCH).
func (v *vcard) name() string {
	if v.fn != "" {
		return v.fn
	}
	return v.org
}

// address maps the jCard ADR components onto the model address.
func (v *vcard) address() model.Address {
	at := func(i int) *string {
		if i < len(v.adr) {
			return disclosed(v.adr[i])
		}
		return nil
	}
	return model.Address{
		Street:     at(2),
		City:       at(3),
		State:      at(4),
		PostalCode: at(5),
		Country:    cmp.Or(disclosed(v.cc), at(6)),
	}
}

// jcardString reads a jCard value that may be a plain string or an array of strings.
func jcardString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(strings.Join(jcardStrings(raw), " "))
}

// jcardStrings flattens a jCard value (string, or array possibly containing sub-arrays)
// into a slice of strings, preserving component positions.
func jcardStrings(raw json.RawMessage) []string {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		var s string
		if json.Unmarshal(raw, &s) == nil {
			return []string{strings.TrimSpace(s)}
		}
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		var s string
		if json.Unmarshal(e, &s) == nil {
			out = append(out, strings.TrimSpace(s))
			continue
		}
		out = append(out, strings.TrimSpace(strings.Join(jcardStrings(e), " ")))
	}
	return out
}
