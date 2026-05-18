package service

import (
	"errors"
	"fmt"

	commonerrors "github.com/shuvo-paul/medminder/internal/common/errors"
)

var (
	ErrEmailExists         = fmt.Errorf("%w: %w", errors.New("email already exists"), commonerrors.ErrConflict)
	ErrInvalidCredentials  = fmt.Errorf("%w: %w", errors.New("invalid credentials"), commonerrors.ErrUnauthorized)
	ErrLogoutFailed        = fmt.Errorf("%w: %w", errors.New("logout failed"), commonerrors.ErrInternal)
	ErrRateLimitExceeded   = fmt.Errorf("%w: %w", errors.New("rate limit exceeded"), commonerrors.ErrTooManyRequests)
	ErrEmailNotFound       = fmt.Errorf("%w: %w", errors.New("email not found"), commonerrors.ErrNotFound)
	ErrOAuthProviderError  = errors.New("oauth provider error")
	ErrOAuthProviderFailed = fmt.Errorf("%w: %w", ErrOAuthProviderError, commonerrors.ErrInternal)
)
