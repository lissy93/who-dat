// Package config loads runtime configuration from environment variables with sane
// defaults, so the same binary works locally, in Docker, and on Vercel.
package config

import (
	"crypto/subtle"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port          string        // HTTP listen port (binary only)
	AuthKey       string        // optional bearer token; when set the API is fully private
	APIKeys       []string      // optional keys that bypass the rate limit (does not lock the API)
	CacheTTL      time.Duration // how long lookups stay cached
	EnableCache   bool          // toggles the in-memory cache
	LookupTimeout time.Duration // per-request upstream timeout
	MaxDomains    int           // max domains per /multi request
	RatePerMinute int           // per-IP request budget; <= 0 disables rate limiting
	RateBurst     int           // per-IP burst allowance
	// CDNCacheControl is sent on successful lookups so a shared cache (Vercel's CDN)
	// serves repeats without re-invoking the function. Empty disables it.
	CDNCacheControl string

	// Contact form (/contact). Keys stay empty when unset; the page still renders
	// and POST /api/contact returns a JSON error pointing at the mailto fallback.
	TurnstileSiteKey   string // public, exposed by GET /api/contact/config
	TurnstileSecretKey string // server-side Turnstile siteverify
	ResendAPIKey       string // Resend bearer token
	ContactFrom        string // From header, e.g. `wi.fs.kg <onboarding@resend.dev>`
	// ContactRatePerMinute is the per-IP budget for POST /api/contact. <= 0 disables it.
	ContactRatePerMinute int
	ContactRateBurst     int
}

// DefaultContactFrom is the Resend sandbox sender used when CONTACT_FROM is unset.
// A verified domain can replace it in production.
const DefaultContactFrom = "wi.fs.kg <onboarding@resend.dev>"

// Load reads configuration from the environment.
func Load() *Config {
	// Strict per-IP limit on Vercel (public, abused); off by default elsewhere. Env overrides.
	rpm, burst := 0, 0
	if onVercel() {
		rpm, burst = 30, 10
	}
	return &Config{
		Port:                 env("PORT", "8080"),
		AuthKey:              os.Getenv("AUTH_KEY"),
		APIKeys:              splitKeys(os.Getenv("API_KEYS")),
		CacheTTL:             time.Duration(intEnv("CACHE_TTL_SECONDS", 3600)) * time.Second,
		EnableCache:          boolEnv("ENABLE_CACHE", true),
		LookupTimeout:        time.Duration(intEnv("LOOKUP_TIMEOUT_SECONDS", 5)) * time.Second,
		MaxDomains:           intEnv("MAX_DOMAINS", 10),
		RatePerMinute:        intEnv("RATE_PER_MINUTE", rpm),
		RateBurst:            intEnv("RATE_BURST", burst),
		CDNCacheControl:      cdnCacheControl(intEnv("CDN_CACHE_TTL_SECONDS", 3600), intEnv("CDN_CACHE_SWR_SECONDS", 86400)),
		TurnstileSiteKey:     os.Getenv("TURNSTILE_SITE_KEY"),
		TurnstileSecretKey:   os.Getenv("TURNSTILE_SECRET_KEY"),
		ResendAPIKey:         os.Getenv("RESEND_API_KEY"),
		ContactFrom:          env("CONTACT_FROM", DefaultContactFrom),
		ContactRatePerMinute: intEnv("CONTACT_RATE_PER_MINUTE", 5),
		ContactRateBurst:     intEnv("CONTACT_RATE_BURST", 3),
	}
}

// onVercel reports whether we are running on Vercel, which sets VERCEL=1.
func onVercel() bool { return os.Getenv("VERCEL") != "" }

// splitKeys parses a comma-separated key list, trimming blanks.
func splitKeys(raw string) []string {
	if raw == "" {
		return nil
	}
	var keys []string
	for _, k := range strings.Split(raw, ",") {
		if k = strings.TrimSpace(k); k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

// ValidKey reports whether token matches the private AuthKey or any configured API key,
// using a constant-time comparison. Empty tokens are never valid.
func (c *Config) ValidKey(token string) bool {
	if token == "" {
		return false
	}
	if c.AuthKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(c.AuthKey)) == 1 {
		return true
	}
	for _, k := range c.APIKeys {
		if subtle.ConstantTimeCompare([]byte(token), []byte(k)) == 1 {
			return true
		}
	}
	return false
}

// cdnCacheControl builds the shared-cache header; ttl <= 0 disables it. WHOIS records
// change rarely, so a long s-maxage lets the CDN absorb repeated lookups.
func cdnCacheControl(ttl, swr int) string {
	if ttl <= 0 {
		return ""
	}
	if swr < 0 {
		swr = 0
	}
	return fmt.Sprintf("public, s-maxage=%d, stale-while-revalidate=%d", ttl, swr)
}

// AuthEnabled reports whether bearer auth is configured.
func (c *Config) AuthEnabled() bool { return c.AuthKey != "" }

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func boolEnv(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
