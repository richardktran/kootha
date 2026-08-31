package cors

import (
	"net/http"

	"github.com/richardktran/realtime-quiz/pkg/env"
)

const defaultOrigins = "http://localhost:3000,http://127.0.0.1:3000"

// AllowedOrigins is the configured CORS allow-list (CORS_ORIGIN, comma-separated).
func AllowedOrigins() []string {
	return env.CSV("CORS_ORIGIN", defaultOrigins)
}

func allowedSet() map[string]struct{} {
	set := make(map[string]struct{})
	for _, o := range AllowedOrigins() {
		set[o] = struct{}{}
	}
	return set
}

func pickOrigin(r *http.Request, allowed map[string]struct{}) string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return ""
	}
	if _, ok := allowed["*"]; ok {
		return origin
	}
	if _, ok := allowed[origin]; ok {
		return origin
	}
	return ""
}

// Middleware sets CORS headers from CORS_ORIGIN. Local default is localhost:3000.
func Middleware(next http.Handler) http.Handler {
	allowed := allowedSet()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := pickOrigin(r, allowed); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Upgrade, Connection")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OriginAllowed reports whether the WebSocket Origin header is permitted.
func OriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	allowed := allowedSet()
	if _, ok := allowed["*"]; ok {
		return true
	}
	_, ok := allowed[origin]
	return ok
}

// HeaderValue returns a single origin for static headers (first configured origin).
func HeaderValue() string {
	origins := AllowedOrigins()
	if len(origins) == 0 {
		return "http://localhost:3000"
	}
	return origins[0]
}
