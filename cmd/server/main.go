package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", log.F("error", err.Error()))
		return
	}

	configureLogger(cfg.AppEnv)

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

	log.Info("starting server", log.F("port", cfg.AppPort), log.F("env", cfg.AppEnv))
	addr := fmt.Sprintf(":%d", cfg.AppPort)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Error("server error", log.F("error", err.Error()))
	}
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
