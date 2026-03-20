package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/log"
)

//go:embed all:web/dist
var webDist embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", log.F("error", err.Error()))
		return
	}

	configureLogger(cfg.AppEnv)

	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Error("failed to create sub filesystem", log.F("error", err.Error()))
		return
	}

	router := chi.NewRouter()

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})

	// Service worker — must be served with special headers for both GET and HEAD.
	swHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Service-Worker-Allowed", "/")
		http.FileServerFS(distFS).ServeHTTP(w, r)
	}
	router.Get("/sw.js", swHandler)
	router.Head("/sw.js", swHandler)

	// SPA — serve embedded static files at root with index.html fallback
	router.Handle("/*", spaHandler(distFS))

	log.Info("starting server", log.F("port", cfg.AppPort), log.F("env", cfg.AppEnv))
	addr := fmt.Sprintf(":%d", cfg.AppPort)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Error("server error", log.F("error", err.Error()))
	}
}

// spaHandler serves static files from fsys and falls back to index.html for
// paths that don't match a file (enabling SvelteKit client-side routing).
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := fsys.Open(path)
		if err != nil {
			// Not a static asset — serve index.html directly for SPA routing.
			// We read and copy the file instead of using FileServerFS to avoid
			// the 301 redirect that FileServer issues for index.html paths.
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

func configureLogger(appEnv string) {
	var logLevel log.Level
	var filePath string

	if appEnv == "production" {
		logLevel = log.InfoLevel
		filePath = "logs/app.log"
	} else {
		logLevel = log.DebugLevel
	}

	log.Configure(log.Config{
		Env:      log.Environment(appEnv),
		Level:    logLevel,
		FilePath: filePath,
	})
}
