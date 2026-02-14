package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	log.Printf("MedMinder server listening on :%s (db=%s:%s/%s sslmode=%s)", port, dbCfg.host, dbCfg.port, dbCfg.name, dbCfg.sslmode)
	if err := router.Run(":" + port); err != nil {
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
