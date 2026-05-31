package router

import (
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/common/ratelimit"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/audit/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth"
	"github.com/shuvo-paul/medminder/internal/middleware"
)

const (
	// authRate is the per-IP per-minute request limit for auth endpoints
	// (login, register, password reset, email verification).
	authRate = 5

	// apiRate is the per-IP per-minute request limit for general API endpoints.
	apiRate = 100

	// oauthRate is the per-IP per-minute request limit for OAuth endpoints
	// (initiate, callback, token exchange, link init).
	oauthRate = 100

	// rateLimitWindow is the sliding window duration for all rate limiters.
	rateLimitWindow = time.Minute
)

func New(distFS fs.FS, dbConn *sql.DB, cfg config.Config) (http.Handler, func(), error) {
	authLimiter := ratelimit.New(authRate, rateLimitWindow)
	apiLimiter := ratelimit.New(apiRate, rateLimitWindow)
	oauthLimiter := ratelimit.New(oauthRate, rateLimitWindow)

	router := chi.NewRouter()
	router.Use(chiMiddleware.RealIP)

	// Extract IP and User-Agent into request context (available to handlers via context).
	router.Use(middleware.RequestInfo)

	// Security headers on all API routes.
	router.Use(securityHeadersOnAPI)

	// CORS for cross-origin requests from the SPA.
	router.Use(middleware.CORSMiddleware([]string{cfg.FrontendURL}))

	// Selective rate limiting: auth gets strict, OAuth gets moderate,
	// general API gets moderate, static assets and infrastructure are exempt.
	router.Use(rateLimitByPath(authLimiter, oauthLimiter, apiLimiter))

	api := setupHuma(router)

	queries := db.New(dbConn)
	auditRepo := repository.NewAuditRepository(queries)

	emailClient := email.NewEmailClient(cfg.Email)
	email.StartEmailQueue(emailClient, 3, slog.Default())

	auth.RegisterRoutes(api, dbConn, queries, auditRepo, cfg.JWT.Secret, emailClient, cfg.FrontendURL)

	registerHealthRoute(api)
	registerOpenAPIRoute(router, api)

	swHandler := newServiceWorkerHandler(distFS)
	router.Get("/sw.js", swHandler)
	router.Head("/sw.js", swHandler)

	router.Handle("/*", newSPAHandler(distFS))

	cleanup := func() {
		email.StopEmailQueue()
		authLimiter.Stop()
		oauthLimiter.Stop()
		apiLimiter.Stop()
	}

	return router, cleanup, nil
}

// rateLimitByPath returns middleware that applies different rate limits based on
// the request path:
//
//   - /api/auth/oauth/* → moderate limit (100 req/min — OAuth flow)
//   - /api/auth/*        → strict limit (brute-force protection for login/register)
//   - /api/*              → moderate limit (general API usage)
//   - /api/healthz, /api/openapi.json, /api/docs/* → exempt (infrastructure)
//   - all other paths (SPA, service worker) → exempt (static assets)
func rateLimitByPath(authLimiter, oauthLimiter, apiLimiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Exempt: non-API routes (SPA static files, service worker).
			if !strings.HasPrefix(path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// Exempt: infrastructure endpoints (health check, OpenAPI spec, docs UI).
			if path == "/api/healthz" || path == "/api/openapi.json" || strings.HasPrefix(path, "/api/docs") {
				next.ServeHTTP(w, r)
				return
			}

			// OAuth routes: moderate limit (OAuth flow is user-initiated, not brute-force).
			if strings.HasPrefix(path, "/api/auth/oauth/") {
				oauthLimiter.Middleware(next).ServeHTTP(w, r)
				return
			}

			// Auth routes: strict limit.
			if strings.HasPrefix(path, "/api/auth/") {
				authLimiter.Middleware(next).ServeHTTP(w, r)
				return
			}

			// General API routes: moderate limit.
			apiLimiter.Middleware(next).ServeHTTP(w, r)
		})
	}
}

// securityHeadersOnAPI is a middleware that applies security headers only to
// API routes, leaving SPA static assets unmodified.
func securityHeadersOnAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			middleware.SecurityHeaders(next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupHuma(router *chi.Mux) huma.API {
	humaConfig := huma.DefaultConfig("MedMinder API", "1.0.0")
	humaConfig.Info.Description = "Medication reminder application API"
	humaConfig.DocsRenderer = huma.DocsRendererSwaggerUI
	humaConfig.DocsPath = "/api/docs"
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Enter your JWT token",
		},
	}
	return humachi.New(router, humaConfig)
}
