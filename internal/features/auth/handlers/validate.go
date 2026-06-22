package handlers

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters with 1 uppercase, 1 lowercase, and 1 number")
	ErrInvalidDisplayName = errors.New("display name must be 1-100 characters")
)

func ValidateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	if len(email) > 255 {
		return ErrInvalidEmail
	}
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return ErrInvalidEmail
	}
	domain := strings.ToLower(parts[1])
	if !strings.Contains(domain, ".") || domain == "" || strings.HasPrefix(domain, ".") {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrInvalidPassword
	}

	return nil
}

func ValidateDisplayName(name string) error {
	if name == "" {
		return ErrInvalidDisplayName
	}
	trimmed := name
	for i := len(trimmed) - 1; i >= 0 && (trimmed[i] == ' ' || trimmed[i] == '\t'); i-- {
		trimmed = trimmed[:i]
	}
	if len(trimmed) == 0 {
		return ErrInvalidDisplayName
	}
	if len(name) > 100 {
		return ErrInvalidDisplayName
	}
	return nil
}
