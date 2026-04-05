package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/shuvo-paul/medminder/internal/common/database/migrations"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/router"
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

	migrator, err := database.NewMigratorWithFS(migrations.FS,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	if err != nil {
		log.Error("failed to create migrator", log.F("error", err.Error()))
		return
	}

	if err := migrator.Up(); err != nil {
		log.Error("failed to run migrations", log.F("error", err.Error()))
		return
	}

	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Error("failed to create sub filesystem", log.F("error", err.Error()))
		return
	}

	dbConn, err := database.Connect(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", log.F("error", err.Error()))
		return
	}
	defer dbConn.Close()

	r, err := router.New(distFS, dbConn, cfg)
	if err != nil {
		log.Error("failed to create router", log.F("error", err.Error()))
		return
	}

	log.Info("starting server", log.F("port", cfg.AppPort), log.F("env", cfg.AppEnv))
	addr := fmt.Sprintf(":%d", cfg.AppPort)
	if err := http.ListenAndServe(addr, r); err != nil {
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
