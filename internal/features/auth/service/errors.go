package service

import (
	"errors"
	"time"
)

const (
	BcryptCost         = 12
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLogoutFailed       = errors.New("logout failed")
)
