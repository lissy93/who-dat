package rdap

import (
	"cmp"
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/model"
)

// referralTimeout caps the registrar lookup. It also gets at most half the time left.
const referralTimeout = 2 * time.Second

// rawPair is what ?raw=true returns after a referral: both upstream documents.
type rawPair struct {
	Registry  json.RawMessage `json:"registry"`
	Registrar json.RawMessage `json:"registrar"`
}

// followReferral merges the registrar's record into r, if there is one. Best effort.
func (c *Client) followReferral(ctx context.Context, n domain.Name, r *model.Result, self string) {
	href := referral(r.Raw, self)
	if href == "" {
		return
	}
	budget := referralTimeout
	if deadline, ok := ctx.Deadline(); ok {
		budget = min(budget, time.Until(deadline)/2)
	}
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	rr, err := c.get(ctx, n, href, href)
	if err == nil && rr.IsRegistered && describes(rr.Raw, n) {
		merge(r, rr)
	}
}

// referral returns the registrar URL from the registry's "related" link, or "" if none.
func referral(raw []byte, self string) string {
	var d struct {
		Links []rdapLink `json:"links"`
	}
	if json.Unmarshal(raw, &d) != nil {
		return ""
	}
	for _, l := range d.Links {
		// A missing type is fine, as describes checks the response anyway.
		if strings.EqualFold(l.Rel, "related") && (l.Type == "" || l.isRDAP()) &&
			strings.HasPrefix(l.Href, "https://") && !strings.EqualFold(l.Href, self) {
			return l.Href
		}
	}
	return ""
}

// describes reports whether raw is n's RDAP domain object, and not some other document.
func describes(raw []byte, n domain.Name) bool {
	var d struct {
		ObjectClassName string `json:"objectClassName"`
		LDHName         string `json:"ldhName"`
	}
	return json.Unmarshal(raw, &d) == nil &&
		strings.EqualFold(d.ObjectClassName, "domain") && strings.EqualFold(d.LDHName, n.ASCII)
}

// merge folds the registrar's record rr into the registry's r. Status and DNS stay as is.
func merge(r, rr *model.Result) {
	// The registry says who the registrar is, and the registrar knows its own details.
	registry, registrar := &r.Registrar, rr.Registrar
	registry.Name = cmp.Or(registry.Name, registrar.Name)
	registry.IANAID = cmp.Or(registry.IANAID, registrar.IANAID)
	registry.URL = cmp.Or(registrar.URL, registry.URL)
	registry.WhoisServer = cmp.Or(registrar.WhoisServer, registry.WhoisServer)
	registry.AbuseEmail = cmp.Or(registrar.AbuseEmail, registry.AbuseEmail)
	registry.AbusePhone = cmp.Or(registrar.AbusePhone, registry.AbusePhone)
	registry.Reseller = cmp.Or(registrar.Reseller, registry.Reseller)

	// The registrar collects contacts, so its copy is the current one.
	takeContact(&r.Contacts.Registrant, rr.Contacts.Registrant)
	takeContact(&r.Contacts.Admin, rr.Contacts.Admin)
	takeContact(&r.Contacts.Tech, rr.Contacts.Tech)
	takeContact(&r.Contacts.Billing, rr.Contacts.Billing)

	// Dates are the registry's, but either side may have the latest change.
	r.Dates.Created = cmp.Or(r.Dates.Created, rr.Dates.Created)
	r.Dates.Expires = cmp.Or(r.Dates.Expires, rr.Dates.Expires)
	if u := rr.Dates.Updated; u != nil && (r.Dates.Updated == nil || u.After(*r.Dates.Updated)) {
		r.Dates.Updated = u
	}

	r.Meta.Referral = rr.Meta.Server
	if raw, err := json.Marshal(rawPair{Registry: r.Raw, Registrar: rr.Raw}); err == nil {
		r.Raw, r.RawContentType = raw, "application/json"
	}
}

// takeContact uses the registrar's contact, unless the registry's discloses more.
func takeContact(dst *model.Contact, registrar model.Contact) {
	if registrar.Redacted && (!dst.Redacted || addressParts(registrar.Address) < addressParts(dst.Address)) {
		return
	}
	*dst = registrar
}

func addressParts(a model.Address) int {
	n := 0
	for _, p := range []*string{a.Street, a.City, a.State, a.PostalCode, a.Country} {
		if p != nil {
			n++
		}
	}
	return n
}
