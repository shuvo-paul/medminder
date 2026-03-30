package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/shuvo-paul/medminder/internal/common/database/migrations"
	"github.com/shuvo-paul/medminder/internal/common/log"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up, down, steps")
	steps := flag.Int("steps", 1, "number of steps for 'steps' direction (negative for down)")
	version := flag.Int("version", -1, "version to force migrations to")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log.Configure(log.Config{
		Env:   log.Environment(cfg.AppEnv),
		Level: log.InfoLevel,
	})

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
		os.Exit(1)
	}

	switch *direction {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Error("failed to run migrations up", log.F("error", err.Error()))
			os.Exit(1)
		}
		log.Info("migrations up completed successfully")

	case "down":
		if err := migrator.Down(); err != nil {
			log.Error("failed to run migrations down", log.F("error", err.Error()))
			os.Exit(1)
		}
		log.Info("migrations down completed successfully")

	case "steps":
		if err := migrator.Steps(*steps); err != nil {
			log.Error("failed to run migration steps", log.F("error", err.Error()), log.F("steps", *steps))
			os.Exit(1)
		}
		log.Info("migration steps completed", log.F("steps", *steps))

	case "force":
		if *version < 0 {
			fmt.Fprintf(os.Stderr, "version must be >= 0 for force\n")
			os.Exit(1)
		}
		if err := migrator.Force(*version); err != nil {
			log.Error("failed to force migration version", log.F("error", err.Error()), log.F("version", *version))
			os.Exit(1)
		}
		log.Info("migration forced", log.F("version", *version))

	default:
		fmt.Fprintf(os.Stderr, "unknown direction: %s (use: up, down, steps, force)\n", *direction)
		os.Exit(1)
	}
}
