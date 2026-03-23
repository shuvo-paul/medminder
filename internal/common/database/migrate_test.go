package database_test

import (
	"testing"

	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/shuvo-paul/medminder/internal/common/database/migrations"
	"github.com/stretchr/testify/assert"
)

func TestNewMigrator(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	assert.NotNil(t, m, "migrator should not be nil")
}

func TestNewMigrator_InvalidDSN(t *testing.T) {
	m, err := database.NewMigratorWithFS(migrations.FS, "invalid", 5432, "invalid", "invalid", "invalid", false)
	assert.Error(t, err, "should fail with invalid DSN")
	assert.Nil(t, m, "migrator should be nil")
}

func TestMigrator_Up(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	err := m.Up()
	assert.NoError(t, err, "should run migrations up without error")
}

func TestMigrator_Down(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	err := m.Up()
	assert.NoError(t, err, "should run migrations up first")

	err = m.Down()
	assert.NoError(t, err, "should run migrations down without error")
}

func TestMigrator_Steps(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	err := m.Steps(1)
	assert.NoError(t, err, "should run 1 migration step without error")
}

func TestMigrator_Force(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	m := tc.NewMigrator(t)
	err := m.Force(0)
	assert.NoError(t, err, "should force migration version without error")
}
