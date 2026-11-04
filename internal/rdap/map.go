package rdap

import (
	"encoding/json"
	"fmt"
	"mime"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/model"
	"github.com/lissy93/who-dat/internal/srcerr"
)

// rdapDomain is the subset of an RFC 9083 domain object that we map.
type rdapDomain struct {
	Handle      laxString        `json:"handle"`
	LDHName     string           `json:"ldhName"`
	UnicodeName string           `json:"unicodeName"`
	Port43      string           `json:"port43"`
	Status      []string         `json:"status"`
	Events      []rdapEvent      `json:"events"`
	Entities    []rdapEntity     `json:"entities"`
	Nameservers []rdapNameserver `json:"nameservers"`
	SecureDNS   *rdapSecureDNS   `json:"secureDNS"`
}

type rdapEvent struct {
	Action string `json:"eventAction"`
	Date   string `json:"eventDate"`
}

type rdapEntity struct {
	Handle    laxString      `json:"handle"`
	Roles     []string       `json:"roles"`
	PublicIDs []rdapPublicID `json:"publicIds"`
	VCard     vcard          `json:"vcardArray"`
	Entities  []rdapEntity   `json:"entities"`
	Links     []rdapLink     `json:"links"`
	URL       string         `json:"url"` // nonstandard, but SWITCH puts the registrar URL here
}

// TWNIC pads entities arrays with bare 404s; skip anything that is not an object.
func (e *rdapEntity) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] != '{' {
		return nil
	}
	type plain rdapEntity
	return json.Unmarshal(b, (*plain)(e))
}

type rdapPublicID struct {
	Type       string    `json:"type"`
	Identifier laxString `json:"identifier"`
}

// laxString accepts both "1234" and 1234; some alpha-grade registries can't decide.
type laxString string

func (s *laxString) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = laxString(v)
	} else if string(b) != "null" {
		*s = laxString(b)
	}
	return nil
}

type rdapLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type"`
}

// isRDAP reports whether the link points at an RDAP document rather than a web page.
func (l rdapLink) isRDAP() bool {
	t, _, err := mime.ParseMediaType(l.Type)
	return err == nil && t == contentType
}

type rdapNameserver struct {
	LDHName     string `json:"ldhName"`
	IPAddresses struct {
		V4 []string `json:"v4"`
		V6 []string `json:"v6"`
	} `json:"ipAddresses"`
}

type rdapSecureDNS struct {
	DelegationSigned bool `json:"delegationSigned"`
	DSData           []struct {
		KeyTag     int    `json:"keyTag"`
		Algorithm  int    `json:"algorithm"`
		DigestType int    `json:"digestType"`
		Digest     string `json:"digest"`
	} `json:"dsData"`
}

// mapDomain decodes an RDAP domain payload and maps it to the canonical model.
func mapDomain(n domain.Name, server string, raw []byte) (*model.Result, error) {
	var d rdapDomain
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("%w: rdap decode: %v", srcerr.ErrUpstream, err)
	}

	r := model.New(n.ASCII, n.ASCII, n.TLD)
	r.IsRegistered = true
	r.ID = model.Str(string(d.Handle))
	r.Status = model.NormalizeStatuses(d.Status)
	r.Meta = model.Meta{Source: model.SourceRDAP, Server: model.Str(server), FetchedAt: time.Now().UTC()}
	r.Raw = raw
	r.RawContentType = contentType

	if n.IsIDN() {
		r.DomainUnicode = model.Str(n.Unicode)
	} else if d.UnicodeName != "" && !strings.EqualFold(d.UnicodeName, n.ASCII) {
		r.DomainUnicode = model.Str(d.UnicodeName)
	}

	mapEvents(r, d.Events)
	mapNameservers(r, d.Nameservers)
	mapSecureDNS(r, d.SecureDNS)
	r.Registrar.WhoisServer = model.Str(d.Port43)
	mapEntities(r, d.Entities)

	return r, nil
}

// eventDateLayouts covers what registries actually emit: RFC 9083 mandates RFC 3339,
// but some (e.g. SWITCH for .ch/.li) send zoneless timestamps or bare dates.
var eventDateLayouts = []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}

// parseEventDate parses an RDAP event date leniently; zero time means unparseable.
func parseEventDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range eventDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func mapEvents(r *model.Result, events []rdapEvent) {
	for _, e := range events {
		t := parseEventDate(e.Date)
		if t.IsZero() {
			continue
		}
		switch strings.ToLower(e.Action) {
		case "registration":
			r.Dates.Created = model.Time(t)
		// Not "last update of RDAP database": that is only when the record was served.
		case "last changed":
			r.Dates.Updated = model.Time(t)
		case "expiration":
			r.Dates.Expires = model.Time(t)
		}
	}
}

func mapNameservers(r *model.Result, nss []rdapNameserver) {
	for _, ns := range nss {
		if ns.LDHName == "" {
			continue
		}
		v4, v6 := ns.IPAddresses.V4, ns.IPAddresses.V6
		if v4 == nil {
			v4 = []string{}
		}
		if v6 == nil {
			v6 = []string{}
		}
		r.Nameservers = append(r.Nameservers, model.Nameserver{
			Name: strings.ToLower(ns.LDHName),
			IPv4: v4,
			IPv6: v6,
		})
	}
}

func mapSecureDNS(r *model.Result, s *rdapSecureDNS) {
	if s == nil {
		return
	}
	r.DNSSEC.Signed = s.DelegationSigned
	for _, ds := range s.DSData {
		r.DNSSEC.DSData = append(r.DNSSEC.DSData, model.DSData{
			KeyTag:     ds.KeyTag,
			Algorithm:  ds.Algorithm,
			DigestType: ds.DigestType,
			Digest:     ds.Digest,
		})
	}
}

func mapEntities(r *model.Result, entities []rdapEntity) {
	for _, e := range entities {
		for _, role := range e.Roles {
			switch strings.ToLower(role) {
			case "registrar":
				mapRegistrar(r, e)
			case "registrant":
				r.Contacts.Registrant = contactFromEntity(e)
			case "administrative":
				r.Contacts.Admin = contactFromEntity(e)
			case "technical":
				r.Contacts.Tech = contactFromEntity(e)
			case "billing":
				r.Contacts.Billing = contactFromEntity(e)
			}
		}
	}
}

func mapRegistrar(r *model.Result, e rdapEntity) {
	r.Registrar.Name = model.Str(e.VCard.name())
	for _, id := range e.PublicIDs {
		if strings.Contains(strings.ToLower(id.Type), "iana") {
			r.Registrar.IANAID = model.Str(string(id.Identifier))
		}
	}
	for _, l := range e.Links {
		if l.isRDAP() {
			continue // an RDAP endpoint, not the website
		}
		if strings.EqualFold(l.Rel, "about") || (r.Registrar.URL == nil && strings.HasPrefix(l.Href, "http")) {
			r.Registrar.URL = model.Str(l.Href)
		}
	}
	if r.Registrar.URL == nil && strings.HasPrefix(e.URL, "http") {
		r.Registrar.URL = model.Str(e.URL)
	}
	for _, sub := range e.Entities {
		for _, role := range sub.Roles {
			if strings.EqualFold(role, "abuse") {
				r.Registrar.AbuseEmail = model.Str(sub.VCard.email)
				r.Registrar.AbusePhone = model.Str(sub.VCard.tel)
			}
		}
	}
}

func contactFromEntity(e rdapEntity) model.Contact {
	c := model.Contact{
		Name:         disclosed(e.VCard.fn),
		Organization: disclosed(e.VCard.org),
		Email:        disclosed(e.VCard.email),
		Phone:        disclosed(e.VCard.tel),
		Address:      e.VCard.address(),
	}
	c.Redacted = c.Name == nil && c.Organization == nil && c.Email == nil && c.Phone == nil
	return c
}

// disclosed is model.Str, minus placeholders like "REDACTED FOR PRIVACY".
func disclosed(s string) *string {
	words := strings.FieldsFunc(strings.ToUpper(s), func(r rune) bool { return !unicode.IsLetter(r) })
	if slices.Contains(words, "REDACTED") {
		return nil
	}
	return model.Str(s)
}
