// Package server wires the HTTP routes, middleware, and static frontend into a single
// http.Handler used by both the standalone binary and the Vercel function.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/lissy93/who-dat/internal/cache"
	"github.com/lissy93/who-dat/internal/config"
	"github.com/lissy93/who-dat/internal/lookup"
	"github.com/lissy93/who-dat/internal/rdap"
	"github.com/lissy93/who-dat/internal/web"
	"github.com/lissy93/who-dat/internal/whois"
)

type server struct {
	cfg          *config.Config
	svc          *lookup.Service
	files        http.Handler
	lookupAPI    http.Handler // bare-domain lookup with API middleware applied
	httpClient   *http.Client
	turnstileURL string
	resendURL    string
	contactLimit *limiter // nil when contact rate limiting is off
}

// Build wires the HTTP client, RDAP and WHOIS sources, optional cache, and lookup service
// into the application handler. It is the single construction path shared by every deploy
// target.
func Build(cfg *config.Config) http.Handler {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     rdap.TLSConfig(),
			ForceAttemptHTTP2:   true, // custom TLS config turns h2 off; some registries demand it
		},
	}

	var c *cache.Cache
	if cfg.EnableCache {
		c = cache.New(cfg.CacheTTL)
	}
	svc := lookup.NewService(rdap.NewClient(httpClient), whois.NewClient(), c)
	return New(cfg, svc)
}

// New builds the application handler.
func New(cfg *config.Config, svc *lookup.Service) http.Handler {
	return newServer(cfg, svc).routes()
}

// newServer wires dependencies. Tests can point turnstileURL and resendURL at a
// local server before calling routes.
func newServer(cfg *config.Config, svc *lookup.Service) *server {
	s := &server{
		cfg:          cfg,
		svc:          svc,
		files:        http.FileServer(http.FS(web.FS())),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		turnstileURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		resendURL:    "https://api.resend.com/emails",
	}
	if cfg.ContactRatePerMinute > 0 {
		s.contactLimit = newLimiter(float64(cfg.ContactRatePerMinute)/60, cfg.ContactRateBurst)
	}
	return s
}

func (s *server) routes() http.Handler {
	cfg := s.cfg
	privileged := func(r *http.Request) bool { return cfg.ValidKey(tokenFrom(r)) }
	api := func(h http.HandlerFunc) http.Handler {
		return chain(h, rateLimit(cfg.RatePerMinute, cfg.RateBurst, privileged), auth(cfg))
	}
	s.lookupAPI = api(s.handleLookupBare)

	mux := http.NewServeMux()
	mux.Handle("GET /v1/whois/{domain}", api(s.handleLookupPath))
	mux.Handle("GET /multi", api(s.handleMulti))
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ping", s.handleHealth)
	mux.HandleFunc("GET /docs", s.handleDocs)
	mux.HandleFunc("GET /openapi.yaml", s.handleOpenAPI)
	mux.HandleFunc("GET /contact", s.handleContactPage)
	mux.HandleFunc("GET /api/contact/config", s.handleContactConfig)
	mux.HandleFunc("POST /api/contact", s.handleContact)
	mux.HandleFunc("GET /api/contact", s.handleContactMethod)
	mux.HandleFunc("/", s.handleRoot)

	return chain(mux, recoverer, logger, securityHeaders, cors)
}

// writeJSON encodes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}
