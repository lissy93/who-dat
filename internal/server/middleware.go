package server

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lissy93/who-dat/internal/config"
)

// middleware is a standard HTTP wrapper.
type middleware func(http.Handler) http.Handler

// chain applies middlewares so the first listed runs outermost.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// recoverer turns a panic into a 500 so one bad request can't crash the server.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "err", rec, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, codeInternal, "internal error", "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// logger logs each request once, escalating the level for server errors.
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		level := slog.LevelInfo
		if sw.status >= 500 {
			level = slog.LevelError
		} else if sw.status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "request",
			"method", r.Method, "path", r.URL.Path, "status", sw.status,
			"duration", time.Since(start).String())
	})
}

// securityHeaders sets conservative, CDN-friendly security headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		next.ServeHTTP(w, r)
	})
}

// cors allows access from any origin (CORS is intentionally disabled per spec).
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// auth enforces a valid key only when the API is private (AUTH_KEY set). In public mode
// it is a no-op; keys still grant a rate-limit bypass via rateLimit.
func auth(cfg *config.Config) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.AuthEnabled() {
				next.ServeHTTP(w, r)
				return
			}
			if !cfg.ValidKey(tokenFrom(r)) {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "valid API key required", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimit applies a per-IP token bucket. perMinute <= 0 disables it. Requests that
// carry a valid API key bypass the limit, so legitimate developers are never throttled.
func rateLimit(perMinute, burst int, privileged func(*http.Request) bool) middleware {
	if perMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	lim := newLimiter(float64(perMinute)/60, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privileged != nil && privileged(r) {
				next.ServeHTTP(w, r)
				return
			}
			if ok, retry := lim.allow(clientIP(r)); !ok {
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				writeError(w, http.StatusTooManyRequests, codeRateLimited, "rate limit exceeded", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// tokenFrom extracts an API key from the Authorization (Bearer) or X-API-Key header.
func tokenFrom(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

// statusWriter captures the response status for logging.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusWriter) WriteHeader(code int) {
	if !s.wroteHeader {
		s.status = code
		s.wroteHeader = true
		s.ResponseWriter.WriteHeader(code)
	}
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.WriteHeader(http.StatusOK)
	}
	return s.ResponseWriter.Write(b)
}

// limiter is a simple per-key token bucket.
type limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64 // tokens per second
	burst   float64
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newLimiter(ratePerSec float64, burst int) *limiter {
	if burst < 1 {
		burst = 1 // refills clamp to the burst, so a zero bucket never admits a second request
	}
	l := &limiter{buckets: make(map[string]*bucket), rate: ratePerSec, burst: float64(burst)}
	go l.sweep()
	return l
}

// allow reports whether a request from key is permitted, and if not, the Retry-After
// seconds until the next token.
func (l *limiter) allow(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: l.burst - 1, last: now}
		return true, 0
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	retry := int((1-b.tokens)/l.rate) + 1
	return false, retry
}

func (l *limiter) sweep() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		l.mu.Lock()
		for k, b := range l.buckets {
			if now.Sub(b.last) > 10*time.Minute {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}

// clientIP extracts the caller IP, trusting common proxy headers (Vercel sets these).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
