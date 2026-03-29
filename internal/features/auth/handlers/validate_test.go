package handlers_test

import (
	"testing"

	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail_ValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"simple", "user@example.com"},
		{"with dot", "first.last@example.com"},
		{"with plus", "user+tag@example.com"},
		{"subdomain", "user@mail.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateEmail(tt.email)
			assert.NoError(t, err, "expected valid email to pass validation")
		})
	}
}

func TestValidateEmail_InvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"no at", "userexample.com"},
		{"no domain", "user@"},
		{"no local", "@example.com"},
		{"no tld", "user@example"},
		{"space in local", "user name@example.com"},
		{"double dots", "user..name@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateEmail(tt.email)
			assert.Error(t, err, "expected invalid email to fail validation")
		})
	}
}

func TestValidatePassword_ValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"minimum valid", "Password1"},
		{"mixed case", "MySecure123"},
		{"special chars", "Pass@word1"},
		{"long password", "VeryLongSecurePassword123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidatePassword(tt.password)
			assert.NoError(t, err, "expected valid password to pass validation")
		})
	}
}

func TestValidatePassword_InvalidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"empty", ""},
		{"too short", "Pass1"},
		{"no uppercase", "password1"},
		{"no lowercase", "PASSWORD1"},
		{"no number", "Password"},
		{"just letters", "Password"},
		{"just numbers", "12345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidatePassword(tt.password)
			assert.Error(t, err, "expected invalid password to fail validation")
		})
	}
}

func TestValidateDisplayName_Valid(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
	}{
		{"simple", "John"},
		{"with spaces", "John Doe"},
		{"max length", string(make([]byte, 100))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "max length" {
				for i := range tt.displayName {
					tt.displayName = tt.displayName[:i] + "a"
				}
			}
			err := handlers.ValidateDisplayName(tt.displayName)
			assert.NoError(t, err, "expected valid display name to pass validation")
		})
	}
}

func TestValidateDisplayName_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"too long", string(make([]byte, 101))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "too long" {
				tt.displayName = ""
				for i := 0; i < 101; i++ {
					tt.displayName += "a"
				}
			}
			err := handlers.ValidateDisplayName(tt.displayName)
			assert.Error(t, err, "expected invalid display name to fail validation")
		})
	}
}
