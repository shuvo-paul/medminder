package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expected      Config
		expectedError bool
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  false,
				},
			},
			expectedError: false,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"APP_PORT":    "3000",
				"DB_HOST":     "db.example.com",
				"DB_PORT":     "5433",
				"DB_USER":     "admin",
				"DB_PASSWORD": "secret",
				"DB_NAME":     "testdb",
				"DB_SSLMODE":  "true",
			},
			expected: Config{
				AppPort: 3000,
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5433,
					User:     "admin",
					Password: "secret",
					Name:     "testdb",
					SSLMode:  true,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid APP_PORT",
			envVars: map[string]string{
				"APP_PORT": "not-a-number",
			},
			expectedError: true,
		},
		{
			name: "invalid DB_PORT",
			envVars: map[string]string{
				"DB_PORT": "not-a-number",
			},
			expectedError: true,
		},
		{
			name: "SSLMode false",
			envVars: map[string]string{
				"DB_SSLMODE": "false",
			},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  false,
				},
			},
			expectedError: false,
		},
		{
			name: "SSLMode true",
			envVars: map[string]string{
				"DB_SSLMODE": "true",
			},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  true,
				},
			},
			expectedError: false,
		},
		{
			name: "SSLMode 1",
			envVars: map[string]string{
				"DB_SSLMODE": "1",
			},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  true,
				},
			},
			expectedError: false,
		},
		{
			name: "SSLMode 0",
			envVars: map[string]string{
				"DB_SSLMODE": "0",
			},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  false,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid DB_SSLMODE",
			envVars: map[string]string{
				"DB_SSLMODE": "invalid",
			},
			expectedError: true,
		},
		{
			name: "negative APP_PORT",
			envVars: map[string]string{
				"APP_PORT": "-1",
			},
			expectedError: true,
		},
		{
			name: "negative DB_PORT",
			envVars: map[string]string{
				"DB_PORT": "-5432",
			},
			expectedError: true,
		},
		{
			name: "APP_PORT out of range",
			envVars: map[string]string{
				"APP_PORT": "70000",
			},
			expectedError: true,
		},
		{
			name: "empty string env vars use defaults",
			envVars: map[string]string{
				"APP_PORT":    "",
				"DB_HOST":     "",
				"DB_PORT":     "",
				"DB_USER":     "",
				"DB_PASSWORD": "",
				"DB_NAME":     "",
				"DB_SSLMODE":  "",
			},
			expected: Config{
				AppPort: 8080,
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "medminder",
					Password: "medminder",
					Name:     "medminder",
					SSLMode:  false,
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			got, err := Load()

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
