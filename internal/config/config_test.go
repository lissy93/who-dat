package config

import (
	"reflect"
	"testing"
)

func TestValidKey(t *testing.T) {
	c := &Config{AuthKey: "secret", APIKeys: []string{"k1", "k2"}}
	cases := []struct {
		token string
		want  bool
	}{
		{"secret", true},
		{"k1", true},
		{"k2", true},
		{"", false},
		{"nope", false},
		{"Secret", false}, // case-sensitive
	}
	for _, tc := range cases {
		if got := c.ValidKey(tc.token); got != tc.want {
			t.Errorf("ValidKey(%q) = %v, want %v", tc.token, got, tc.want)
		}
	}
	if (&Config{}).ValidKey("anything") {
		t.Error("config with no keys must reject every token")
	}
}

func TestSplitKeys(t *testing.T) {
	if got := splitKeys(""); got != nil {
		t.Errorf("splitKeys(%q) = %v, want nil", "", got)
	}
	got := splitKeys(" a , b ,, c ")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitKeys = %v, want %v", got, want)
	}
}

func TestCdnCacheControl(t *testing.T) {
	cases := []struct {
		ttl, swr int
		want     string
	}{
		{0, 10, ""},
		{-1, 10, ""},
		{3600, 86400, "public, s-maxage=3600, stale-while-revalidate=86400"},
		{60, -5, "public, s-maxage=60, stale-while-revalidate=0"},
	}
	for _, tc := range cases {
		if got := cdnCacheControl(tc.ttl, tc.swr); got != tc.want {
			t.Errorf("cdnCacheControl(%d, %d) = %q, want %q", tc.ttl, tc.swr, got, tc.want)
		}
	}
}

func TestLoadContactDefaults(t *testing.T) {
	t.Setenv("TURNSTILE_SITE_KEY", "")
	t.Setenv("TURNSTILE_SECRET_KEY", "")
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("CONTACT_FROM", "")
	t.Setenv("CONTACT_RATE_PER_MINUTE", "")
	t.Setenv("CONTACT_RATE_BURST", "")

	c := Load()
	if c.TurnstileSiteKey != "" || c.TurnstileSecretKey != "" || c.ResendAPIKey != "" {
		t.Fatal("unset contact secrets must stay empty")
	}
	if c.ContactFrom != DefaultContactFrom {
		t.Errorf("ContactFrom = %q, want %q", c.ContactFrom, DefaultContactFrom)
	}
	if c.ContactRatePerMinute != 5 || c.ContactRateBurst != 3 {
		t.Errorf("contact rate = %d/%d, want 5/3", c.ContactRatePerMinute, c.ContactRateBurst)
	}

	t.Setenv("TURNSTILE_SITE_KEY", "site_public")
	t.Setenv("TURNSTILE_SECRET_KEY", "secret_value")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("CONTACT_FROM", "Wi <hi@wi.fs.kg>")
	t.Setenv("CONTACT_RATE_PER_MINUTE", "0")
	c = Load()
	if c.TurnstileSiteKey != "site_public" || c.TurnstileSecretKey != "secret_value" || c.ResendAPIKey != "re_test" {
		t.Fatal("contact env was not loaded")
	}
	if c.ContactFrom != "Wi <hi@wi.fs.kg>" || c.ContactRatePerMinute != 0 {
		t.Errorf("override = from %q rate %d", c.ContactFrom, c.ContactRatePerMinute)
	}
}

func TestLoadRateDefaults(t *testing.T) {
	t.Run("on Vercel", func(t *testing.T) {
		t.Setenv("VERCEL", "1")
		if c := Load(); c.RatePerMinute != 30 || c.RateBurst != 10 {
			t.Errorf("got %d/%d, want 30/10", c.RatePerMinute, c.RateBurst)
		}
	})
	t.Run("off Vercel", func(t *testing.T) {
		t.Setenv("VERCEL", "")
		if c := Load(); c.RatePerMinute != 0 || c.RateBurst != 0 {
			t.Errorf("got %d/%d, want 0/0 (rate limiting off)", c.RatePerMinute, c.RateBurst)
		}
	})
	t.Run("env overrides Vercel default", func(t *testing.T) {
		t.Setenv("VERCEL", "1")
		t.Setenv("RATE_PER_MINUTE", "99")
		if c := Load(); c.RatePerMinute != 99 {
			t.Errorf("got %d, want 99", c.RatePerMinute)
		}
	})
}
