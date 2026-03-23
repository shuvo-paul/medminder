package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/shuvo-paul/medminder/internal/common/database/migrations"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type TestContainer struct {
	Container *postgres.PostgresContainer
	Config    config.DatabaseConfig
	ConnStr   string
}

func SetupPostgresContainer(t *testing.T) *TestContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("medminder"),
		postgres.WithUsername("medminder"),
		postgres.WithPassword("medminder"),
	)
	require.NoError(t, err, "should start postgres container")

	connStr, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err, "should get connection string")

	cfg, err := parseDSN(connStr)
	require.NoError(t, err, "should parse DSN")

	return &TestContainer{
		Container: pgContainer,
		Config:    cfg,
		ConnStr:   connStr,
	}
}

func parseDSN(connStr string) (config.DatabaseConfig, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return config.DatabaseConfig{}, fmt.Errorf("failed to parse DSN: %w", err)
	}

	port := 5432
	if p := u.Port(); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	dbName := u.Path
	if dbName != "" && dbName[0] == '/' {
		dbName = dbName[1:]
	}

	password, hasPassword := u.User.Password()
	if !hasPassword {
		password = ""
	}

	return config.DatabaseConfig{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		Name:     dbName,
		SSLMode:  false,
	}, nil
}

func (tc *TestContainer) Connect(t *testing.T) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			require.NoError(t, ctx.Err(), "timeout connecting to database")
			return nil
		default:
		}

		db, err = database.Connect(tc.Config)
		if err == nil {
			return db
		}
		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	require.NoError(t, err, "should connect to database")
	return db
}

func (tc *TestContainer) NewMigrator(t *testing.T) *database.Migrator {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var m *database.Migrator
	var err error

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			require.NoError(t, ctx.Err(), "timeout creating migrator")
			return nil
		default:
		}

		m, err = database.NewMigratorWithFS(
			migrations.FS,
			tc.Config.Host,
			tc.Config.Port,
			tc.Config.User,
			tc.Config.Password,
			tc.Config.Name,
			tc.Config.SSLMode,
		)
		if err == nil {
			return m
		}
		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	require.NoError(t, err, "should create migrator")
	return m
}

func (tc *TestContainer) Teardown(t *testing.T) {
	t.Helper()

	if tc.Container != nil {
		err := tc.Container.Terminate(context.Background())
		if err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

func TestContainerExists(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	db := tc.Connect(t)
	defer db.Close()

	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err, "should execute simple query")
	require.Equal(t, 1, result, "should return 1")
}

func TestContainerWithMigrations(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	err := m.Up()
	require.NoError(t, err, "should run migrations up")

	var count int
	db := tc.Connect(t)
	defer db.Close()

	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	require.NoError(t, err, "should query schema_migrations")
}
