package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/shuvo-paul/medminder/internal/common/database/migrations"
	"github.com/shuvo-paul/medminder/internal/common/log"
)

func main() {
	direction := flag.String("direction", "up", "migration direction: up, down, steps, force, create")
	steps := flag.Int("steps", 0, "number of steps for 'steps' direction (negative for down)")
	version := flag.Int("version", -1, "version to force migrations to")
	name := flag.String("name", "", "name for migration (required for create)")
	migrationDir := flag.String("dir", "internal/common/database/migrations", "directory for migration files")
	flag.Parse()

	// Silently ignore missing .env — exported env vars take precedence.
	_ = godotenv.Load()

	if *direction == "create" {
		if *name == "" {
			fmt.Fprintf(os.Stderr, "name is required for create\n")
			os.Exit(1)
		}
		if err := createMigration(*migrationDir, *name); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create migration: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migration created:", *name)
		return
	}

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
		fmt.Fprintf(os.Stderr, "unknown direction: %s (use: up, down, steps, force, create)\n", *direction)
		os.Exit(1)
	}
}

func createMigration(dir, name string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var versions []int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			parts := strings.Split(e.Name(), "_")
			if len(parts) > 0 {
				v, err := strconv.Atoi(parts[0])
				if err == nil {
					versions = append(versions, v)
				}
			}
		}
	}

	nextVersion := 1
	if len(versions) > 0 {
		sort.Ints(versions)
		nextVersion = versions[len(versions)-1] + 1
	}

	versionStr := fmt.Sprintf("%06d", nextVersion)
	filename := fmt.Sprintf("%s_%s", versionStr, name)

	upPath := filepath.Join(dir, filename+".up.sql")
	downPath := filepath.Join(dir, filename+".down.sql")

	if err := os.WriteFile(upPath, []byte("-- migration: "+name+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create up migration: %w", err)
	}

	if err := os.WriteFile(downPath, []byte("-- migration: "+name+"\n\n"), 0644); err != nil {
		return fmt.Errorf("failed to create down migration: %w", err)
	}

	return nil
}
