package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

type databaseConfig struct {
	host    string
	port    string
	user    string
	name    string
	sslmode string
}

func main() {
	port := getEnv("APP_PORT", "8080")
	dbCfg := loadDatabaseConfig()
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

	log.Printf("MedMinder server listening on :%s (db=%s:%s/%s sslmode=%s)", port, dbCfg.host, dbCfg.port, dbCfg.name, dbCfg.sslmode)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadDatabaseConfig() databaseConfig {
	return databaseConfig{
		host:    getEnv("DB_HOST", "localhost"),
		port:    getEnv("DB_PORT", "5432"),
		user:    getEnv("DB_USER", "medminder"),
		name:    getEnv("DB_NAME", "medminder"),
		sslmode: getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
