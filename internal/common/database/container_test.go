package database_test

import (
	"testing"

	"github.com/shuvo-paul/medminder/internal/common/database/testutil"
	"github.com/stretchr/testify/require"
)

func TestContainerExists(t *testing.T) {
	tc := testutil.SetupPostgresContainer(t)
	defer tc.Teardown(t)

	db := tc.Connect(t)
	defer db.Close()

	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err, "should execute simple query")
	require.Equal(t, 1, result, "should return 1")
}

func TestContainerWithMigrations(t *testing.T) {
	tc := testutil.SetupPostgresContainer(t)
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
