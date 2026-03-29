package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/db"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

type HealthOutput struct {
	Body struct {
		Status    string `json:"status" doc:"Service status"`
		Timestamp string `json:"timestamp" doc:"Current server time in RFC3339 format"`
	}
}

func New(distFS fs.FS, cfg config.Config) (http.Handler, error) {
	router := chi.NewRouter()

	humaConfig := huma.DefaultConfig("MedMinder API", "1.0.0")
	humaConfig.Info.Description = "Medication reminder application API"
	humaConfig.DocsRenderer = huma.DocsRendererSwaggerUI
	humaConfig.DocsPath = "/api/docs"
	api := humachi.New(router, humaConfig)

	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/api/healthz",
		Summary:     "Health check",
		Tags:        []string{"system"},
	}, func(_ context.Context, _ *struct{}) (*HealthOutput, error) {
		resp := &HealthOutput{}
		resp.Body.Status = "ok"
		resp.Body.Timestamp = time.Now().UTC().Format(time.RFC3339)
		return resp, nil
	})

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		map[bool]string{true: "require", false: "disable"}[cfg.Database.SSLMode])

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("failed to open db connection", log.F("error", err.Error()))
		return nil, err
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Error("failed to ping db", log.F("error", err.Error()))
		return nil, err
	}

	queries := db.New(dbConn)
	tokenSvc := service.NewTokenService(cfg.JWT.Secret)

	type RegisterOutput struct {
		Body struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			User         struct {
				ID            uuid.UUID `json:"id"`
				Email         string    `json:"email"`
				DisplayName   string    `json:"display_name"`
				EmailVerified bool      `json:"email_verified"`
			} `json:"user"`
		}
	}

	repo := repository.NewUserRepository(queries)
	registerHandler := handlers.RegisterHandler(repo, tokenSvc)

	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/api/auth/register",
		Summary:     "Register a new user",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *handlers.RegisterInput) (*RegisterOutput, error) {
		resp, err := registerHandler(ctx, input)
		if err != nil {
			if errors.Is(err, handlers.ErrInvalidInput) {
				return nil, huma.Error400BadRequest("Invalid input", err)
			}
			if errors.Is(err, handlers.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already exists", err)
			}
			return nil, err
		}
		out := &RegisterOutput{}
		out.Body.AccessToken = resp.AccessToken
		out.Body.RefreshToken = resp.RefreshToken
		out.Body.User = resp.User
		return out, nil
	})

	router.Get("/api/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/openapi+json")
		json.NewEncoder(w).Encode(api.OpenAPI()) //nolint:errcheck
	})

	swHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Service-Worker-Allowed", "/")
		http.FileServerFS(distFS).ServeHTTP(w, r)
	}
	router.Get("/sw.js", swHandler)
	router.Head("/sw.js", swHandler)

	router.Handle("/*", spaHandler(distFS))

	return router, nil
}

func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := fsys.Open(path)
		if err != nil {
			indexFile, indexErr := fsys.Open("index.html")
			if indexErr != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
			defer indexFile.Close()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.Copy(w, indexFile) //nolint:errcheck
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
