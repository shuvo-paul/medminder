package repository

import "errors"

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenUsed     = errors.New("token already used")

	// ErrOAuthAccountNotFound is returned when an OAuth account does not exist in the database.
	ErrOAuthAccountNotFound = errors.New("oauth account not found")

	// ErrOAuthCodeNotFound is returned when an OAuth authorization code is not found,
	// already used, or expired. All three states return the same error to prevent timing oracles.
	ErrOAuthCodeNotFound = errors.New("oauth code not found")
)
