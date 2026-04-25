package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth"
)

func New(distFS fs.FS, dbConn *sql.DB, cfg config.Config) (http.Handler, error) {
	router := chi.NewRouter()

	api := setupHuma(router)

	queries := db.New(dbConn)

	emailClient := email.NewEmailClient(cfg.Email)

	auth.RegisterRoutes(api, queries, cfg.JWT.Secret, emailClient, cfg.FrontendURL)

	registerHealthRoute(api)
	registerOpenAPIRoute(router, api)

	swHandler := newServiceWorkerHandler(distFS)
	router.Get("/sw.js", swHandler)
	router.Head("/sw.js", swHandler)

	router.Handle("/*", newSPAHandler(distFS))

	return router, nil
}

func setupHuma(router *chi.Mux) huma.API {
	humaConfig := huma.DefaultConfig("MedMinder API", "1.0.0")
	humaConfig.Info.Description = "Medication reminder application API"
	humaConfig.DocsRenderer = huma.DocsRendererScalar
	humaConfig.DocsPath = "/api/docs"
	return humachi.New(router, humaConfig)
}
