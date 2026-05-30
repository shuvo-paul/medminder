package middleware

import "net/http"

// SecurityHeaders is a Chi-compatible middleware that sets security-related
// HTTP headers on all responses to protect against common web vulnerabilities.
//
// Headers set:
//   - Cache-Control: no-store (prevents caching sensitive auth responses)
//   - X-Content-Type-Options: nosniff (prevents MIME type sniffing)
//   - X-Frame-Options: DENY (prevents clickjacking)
//   - Referrer-Policy: strict-origin-when-cross-origin (limits referrer leakage)
//   - X-XSS-Protection: 0 (disabled — superseded by CSP, prevents false sense of security)
//   - Permissions-Policy: interest-cohort=() (disables FLoC tracking)
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Permissions-Policy", "interest-cohort=()")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware returns an HTTP middleware that adds CORS headers for
// cross-origin requests from the SPA development server or other trusted origins.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			for _, ao := range allowedOrigins {
				if ao == "*" || ao == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}

			// Handle preflight (only for allowed origins)
			if allowed && r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
