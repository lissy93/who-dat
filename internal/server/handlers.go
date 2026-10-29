package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/lissy93/who-dat/internal/domain"
	"github.com/lissy93/who-dat/internal/web"
)

// handleLookupPath serves the canonical GET /v1/whois/{domain}.
func (s *server) handleLookupPath(w http.ResponseWriter, r *http.Request) {
	s.lookup(w, r, r.PathValue("domain"))
}

// handleLookupBare serves the convenience GET /{domain}.
func (s *server) handleLookupBare(w http.ResponseWriter, r *http.Request) {
	s.lookup(w, r, strings.TrimPrefix(r.URL.Path, "/"))
}

// lookup parses the query, performs the lookup, and writes JSON (or the raw upstream
// payload when ?raw=true).
func (s *server) lookup(w http.ResponseWriter, r *http.Request, query string) {
	n, err := domain.Parse(query)
	if err != nil {
		writeDomainError(w, err, query)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.LookupTimeout)
	defer cancel()

	res, err := s.svc.Lookup(ctx, n)
	if err != nil {
		writeLookupError(w, err, query)
		return
	}
	res.Query = query

	s.setEdgeCache(w)
	if r.URL.Query().Get("raw") == "true" && len(res.Raw) > 0 {
		w.Header().Set("Content-Type", res.RawContentType)
		_, _ = w.Write(res.Raw)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// setEdgeCache lets a CDN cache this successful lookup. No-op when unconfigured.
func (s *server) setEdgeCache(w http.ResponseWriter) {
	if s.cfg.CDNCacheControl != "" {
		w.Header().Set("Cache-Control", s.cfg.CDNCacheControl)
	}
}

// handleMulti serves GET /multi?domains=a.com,b.com, returning per-domain outcomes.
func (s *server) handleMulti(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("domains")
	if raw == "" {
		writeError(w, http.StatusBadRequest, codeInvalidDomain, "missing domains parameter", "")
		return
	}
	queries := strings.Split(raw, ",")
	if len(queries) > s.cfg.MaxDomains {
		writeError(w, http.StatusBadRequest, codeInvalidDomain, "too many domains", raw)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.LookupTimeout)
	defer cancel()

	results := make([]any, 0, len(queries))
	cacheable := true // a transient lookup failure must not be cached
	for _, q := range queries {
		q = strings.TrimSpace(q)
		n, err := domain.Parse(q)
		if err != nil {
			results = append(results, multiError(q, "could not parse a registrable domain"))
			continue
		}
		res, err := s.svc.Lookup(ctx, n)
		if err != nil {
			_, _, d := lookupError(err, q)
			results = append(results, errorBody{Error: d})
			cacheable = false
			continue
		}
		res.Query = q
		results = append(results, res)
	}
	if cacheable {
		s.setEdgeCache(w)
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func multiError(query, message string) errorBody {
	return errorBody{Error: errorDetail{Code: codeInvalidDomain, Message: message, Query: query}}
}

// handleHealth is an unauthenticated liveness check.
func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleDocs serves the Scalar API reference.
func (s *server) handleDocs(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/docs.html"
	s.files.ServeHTTP(w, r)
}

// handleContactPage serves the contact form. Registered explicitly so the single
// path segment is not treated as a bare-domain lookup.
func (s *server) handleContactPage(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/contact.html"
	s.files.ServeHTTP(w, r)
}

// handleOpenAPI serves the OpenAPI specification.
func (s *server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/openapi.yaml"
	s.files.ServeHTTP(w, r)
}

// handleRoot serves the SPA and static assets, treating any other single path segment as
// a bare-domain lookup.
func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || isStaticFile(path) {
		s.files.ServeHTTP(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD, OPTIONS")
		writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "method not allowed", "")
		return
	}
	s.lookupAPI.ServeHTTP(w, r)
}

// isStaticFile reports whether path exists in the embedded asset set.
func isStaticFile(path string) bool {
	f, err := web.FS().Open(path)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
