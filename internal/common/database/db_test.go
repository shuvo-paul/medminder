package database_test

import (
	"testing"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/common/database"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tc := SetupPostgresContainer(t)
	defer tc.Teardown(t)

	db := tc.Connect(t)
	defer db.Close()

	assert.NotNil(t, db, "db should not be nil")

	err := db.Close()
	assert.NoError(t, err, "should close without error")
}

func TestConnect_InvalidConfig(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     5432,
		User:     "invalid",
		Password: "invalid",
		Name:     "invalid",
		SSLMode:  false,
	}

	db, err := database.Connect(cfg)
	assert.Error(t, err, "should fail with invalid config")
	assert.Nil(t, db, "db should be nil")
}
