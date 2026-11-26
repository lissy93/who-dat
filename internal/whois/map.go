package whois

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	whoisparser "github.com/likexian/whois-parser"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/model"
	"github.com/lissy93/who-dat/internal/srcerr"
)

const rawContentType = "text/plain; charset=utf-8"

// mapWhois parses raw WHOIS text and maps it to the canonical model. A "not found" parse
// result is a successful "not registered" answer, not an error.
func mapWhois(n domain.Name, raw string) (*model.Result, error) {
	r := model.New(n.ASCII, n.ASCII, n.TLD)
	r.Meta = model.Meta{Source: model.SourceWhois, FetchedAt: time.Now().UTC()}
	r.Raw = []byte(raw)
	r.RawContentType = rawContentType
	if n.IsIDN() {
		r.DomainUnicode = model.Str(n.Unicode)
	}

	info, err := whoisparser.Parse(raw)
	if err != nil {
		switch {
		case errors.Is(err, whoisparser.ErrNotFoundDomain),
			errors.Is(err, whoisparser.ErrReservedDomain),
			errors.Is(err, whoisparser.ErrPremiumDomain),
			errors.Is(err, whoisparser.ErrBlockedDomain):
			r.IsRegistered = false
			return r, nil
		case errors.Is(err, whoisparser.ErrDomainDataInvalid) &&
			slices.ContainsFunc(notFoundPhrases, func(p string) bool { return strings.Contains(raw, p) }):
			r.IsRegistered = false
			return r, nil
		case errors.Is(err, whoisparser.ErrDomainLimitExceed):
			return nil, &srcerr.RateLimited{Err: err}
		default:
			return nil, fmt.Errorf("%w: whois parse: %v", srcerr.ErrUpstream, err)
		}
	}

	r.IsRegistered = true
	mapDomain(r, info.Domain)
	mapRegistrar(r, info.Registrar)
	r.Contacts.Registrant = contact(info.Registrant)
	r.Contacts.Admin = contact(info.Administrative)
	r.Contacts.Tech = contact(info.Technical)
	r.Contacts.Billing = contact(info.Billing)
	return r, nil
}

// notFoundPhrases are "not registered" replies the parser doesn't know about.
var notFoundPhrases = []string{
	"No information was found matching that query.", // ZARC (.co.za, .net.za, .org.za, .web.za)
}

func mapDomain(r *model.Result, d *whoisparser.Domain) {
	if d == nil {
		return
	}
	r.ID = model.Str(d.ID)
	r.Status = model.NormalizeStatuses(d.Status)
	r.DNSSEC.Signed = d.DNSSec
	r.Dates.Created = timePtr(d.CreatedDateInTime)
	r.Dates.Updated = timePtr(d.UpdatedDateInTime)
	r.Dates.Expires = timePtr(d.ExpirationDateInTime)
	r.Registrar.WhoisServer = model.Str(d.WhoisServer)
	for _, ns := range d.NameServers {
		r.Nameservers = append(r.Nameservers, model.Nameserver{
			Name: ns,
			IPv4: []string{},
			IPv6: []string{},
		})
	}
}

// timePtr converts a whois-parser nullable time to the model's UTC pointer form.
func timePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return model.Time(*t)
}

func mapRegistrar(r *model.Result, c *whoisparser.Contact) {
	if c == nil {
		return
	}
	name := c.Name
	if name == "" {
		name = c.Organization
	}
	r.Registrar.Name = model.Str(name)
	r.Registrar.URL = model.Str(c.ReferralURL)
	r.Registrar.AbuseEmail = model.Str(c.Email)
	r.Registrar.AbusePhone = model.Str(c.Phone)
}

func contact(c *whoisparser.Contact) model.Contact {
	if c == nil {
		return model.Contact{Redacted: true}
	}
	out := model.Contact{
		Name:         model.Str(c.Name),
		Organization: model.Str(c.Organization),
		Email:        model.Str(c.Email),
		Phone:        model.Str(c.Phone),
		Address: model.Address{
			Street:     model.Str(c.Street),
			City:       model.Str(c.City),
			State:      model.Str(c.Province),
			PostalCode: model.Str(c.PostalCode),
			Country:    model.Str(c.Country),
		},
	}
	out.Redacted = out.Name == nil && out.Organization == nil && out.Email == nil && out.Phone == nil
	return out
}
