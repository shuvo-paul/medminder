package main

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/shuvo-paul/medminder/internal/common/config"
	"github.com/shuvo-paul/medminder/internal/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDistFS() fs.FS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><html><body>test</body></html>")},
	}
}

func testConfig() config.Config {
	return config.Config{
		AppPort: 8080,
		AppEnv:  "test",
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "test",
			SSLMode:  false,
		},
	}
}

func TestHealthCheck_ReturnsOK(t *testing.T) {
	r, err := router.New(testDistFS(), &sql.DB{}, testConfig())
	require.NoError(t, err)
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body.Status)
	assert.NotEmpty(t, body.Timestamp)
}

func TestOpenAPISpec_IsServed(t *testing.T) {
	r, err := router.New(testDistFS(), &sql.DB{}, testConfig())
	require.NoError(t, err)
	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/openapi+json")

	var spec map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&spec))
	assert.Equal(t, "3.1.0", spec["openapi"])
}
