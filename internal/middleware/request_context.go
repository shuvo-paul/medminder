// Package middleware provides HTTP middleware used across the application.
package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// ctxKey is an unexported type used for context keys to prevent collisions.
type ctxKey string

const (
	// ctxKeyIP is the context key for the client's IP address.
	ctxKeyIP ctxKey = "ip_address"
	// ctxKeyUserAgent is the context key for the client's User-Agent header.
	ctxKeyUserAgent ctxKey = "user_agent"
)

// IPFromContext extracts the client IP address from the request context.
func IPFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyIP).(string); ok {
		return v
	}
	return ""
}

// UserAgentFromContext extracts the User-Agent header from the request context.
func UserAgentFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyUserAgent).(string); ok {
		return v
	}
	return ""
}

// RequestInfo is a Chi-compatible middleware that extracts the client's IP
// address and User-Agent header and stores them in the request context.
//
// IP extraction uses X-Forwarded-For (trusted proxy header) with fallback
// to RemoteAddr. The stored values can be retrieved via IPFromContext and
// UserAgentFromContext helpers.
func RequestInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		userAgent := r.Header.Get("User-Agent")

		ctx := context.WithValue(r.Context(), ctxKeyIP, ip)
		ctx = context.WithValue(ctx, ctxKeyUserAgent, userAgent)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractIP extracts the client's IP address from the request, preferring
// the X-Forwarded-For header when behind a reverse proxy.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy deployments)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain a comma-separated list; take the first (origin) IP
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (alternative proxy header)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr, stripping the port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
