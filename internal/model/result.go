// Package model defines the canonical, source-independent shape returned for every
// domain lookup. The same keys are always present regardless of TLD or data source:
// null for unknown scalars, empty slices for unknown lists, dates in RFC3339 UTC.
package model

import (
	"strings"
	"time"
)

// Source identifies which protocol produced a Result.
const (
	SourceRDAP  = "rdap"
	SourceWhois = "whois"
)

// Result is the normalized response for a single domain lookup.
type Result struct {
	Query         string       `json:"query"`
	Domain        string       `json:"domain"`
	DomainUnicode *string      `json:"domainUnicode"`
	ID            *string      `json:"id"`
	TLD           string       `json:"tld"`
	IsRegistered  bool         `json:"isRegistered"`
	Registrar     Registrar    `json:"registrar"`
	Status        []string     `json:"status"`
	Nameservers   []Nameserver `json:"nameservers"`
	DNSSEC        DNSSEC       `json:"dnssec"`
	Dates         Dates        `json:"dates"`
	Contacts      Contacts     `json:"contacts"`
	Meta          Meta         `json:"meta"`

	// Raw and RawContentType carry the original upstream payload for ?raw=true. They are
	// never serialized as part of the normalized result.
	Raw            []byte `json:"-"`
	RawContentType string `json:"-"`
}

// Registrar holds the sponsoring registrar and its abuse contact.
type Registrar struct {
	Name        *string `json:"name"`
	IANAID      *string `json:"ianaId"`
	URL         *string `json:"url"`
	WhoisServer *string `json:"whoisServer"`
	AbuseEmail  *string `json:"abuseEmail"`
	AbusePhone  *string `json:"abusePhone"`
	Reseller    *string `json:"reseller"`
}

// Nameserver is a single delegated name server and any glue addresses.
type Nameserver struct {
	Name string   `json:"name"`
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// DNSSEC reports whether the zone is signed and lists any DS records.
type DNSSEC struct {
	Signed bool     `json:"signed"`
	DSData []DSData `json:"dsData"`
}

// DSData is a single delegation signer record.
type DSData struct {
	KeyTag     int    `json:"keyTag"`
	Algorithm  int    `json:"algorithm"`
	DigestType int    `json:"digestType"`
	Digest     string `json:"digest"`
}

// Dates carries the registration lifecycle timestamps.
type Dates struct {
	Created *time.Time `json:"created"`
	Updated *time.Time `json:"updated"`
	Expires *time.Time `json:"expires"`
}

// Contacts groups the four standard registration roles.
type Contacts struct {
	Registrant Contact `json:"registrant"`
	Admin      Contact `json:"admin"`
	Tech       Contact `json:"tech"`
	Billing    Contact `json:"billing"`
}

// Contact is a single registration contact. Redacted is true when the source
// withheld the personal data (the common case under GDPR/RDAP redaction).
type Contact struct {
	Name         *string `json:"name"`
	Organization *string `json:"organization"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      Address `json:"address"`
	Redacted     bool    `json:"redacted"`
}

// Address is a postal address with nullable parts.
type Address struct {
	Street     *string `json:"street"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postalCode"`
	Country    *string `json:"country"`
}

// Meta describes how the result was obtained.
type Meta struct {
	Source    string    `json:"source"`
	Server    *string   `json:"server"`
	Referral  *string   `json:"referral"` // registrar RDAP URL merged in, if any
	FetchedAt time.Time `json:"fetchedAt"`
	Cached    bool      `json:"cached"`
}

// New returns a Result with all list fields initialized to empty (non-nil) slices and
// every contact marked redacted, so a source only needs to fill what it actually has.
func New(query, domain, tld string) *Result {
	redacted := Contact{Redacted: true}
	return &Result{
		Query:       query,
		Domain:      domain,
		TLD:         tld,
		Status:      []string{},
		Nameservers: []Nameserver{},
		DNSSEC:      DNSSEC{DSData: []DSData{}},
		Contacts: Contacts{
			Registrant: redacted,
			Admin:      redacted,
			Tech:       redacted,
			Billing:    redacted,
		},
	}
}

// Str returns a pointer to s, or nil if s is empty after trimming. Sources use it so
// that missing scalars serialize as JSON null rather than "".
func Str(s string) *string {
	if t := strings.TrimSpace(s); t != "" {
		return &t
	}
	return nil
}

// Time returns a pointer to t in UTC, or nil if t is the zero value.
func Time(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	u := t.UTC()
	return &u
}
