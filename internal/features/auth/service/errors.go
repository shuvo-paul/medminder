package service

import (
	"errors"
	"fmt"

	commonerrors "github.com/shuvo-paul/medminder/internal/common/errors"
)

var (
	ErrEmailExists           = fmt.Errorf("%w: %w", errors.New("email already exists"), commonerrors.ErrConflict)
	ErrInvalidCredentials    = fmt.Errorf("%w: %w", errors.New("invalid credentials"), commonerrors.ErrUnauthorized)
	ErrLogoutFailed          = fmt.Errorf("%w: %w", errors.New("logout failed"), commonerrors.ErrInternal)
	ErrRateLimitExceeded     = fmt.Errorf("%w: %w", errors.New("rate limit exceeded"), commonerrors.ErrTooManyRequests)
	ErrEmailNotFound         = fmt.Errorf("%w: %w", errors.New("email not found"), commonerrors.ErrNotFound)
	ErrOAuthProviderError    = errors.New("oauth provider error")
	ErrOAuthProviderFailed   = fmt.Errorf("%w: %w", ErrOAuthProviderError, commonerrors.ErrInternal)
	ErrWrongPassword         = fmt.Errorf("%w: %w", errors.New("current password is incorrect"), commonerrors.ErrForbidden)
	ErrNoPasswordSet         = fmt.Errorf("%w: %w", errors.New("user has no password set"), commonerrors.ErrBadRequest)
	ErrAccountWillBeLocked   = fmt.Errorf("%w: %w", errors.New("account will be locked out"), commonerrors.ErrForbidden)
	ErrProviderAlreadyLinked = fmt.Errorf("%w: %w", errors.New("provider already linked"), commonerrors.ErrConflict)
	ErrUserNotFound          = fmt.Errorf("%w: %w", errors.New("user not found"), commonerrors.ErrNotFound)
	ErrSamePassword          = fmt.Errorf("%w: %w", errors.New("new password must differ from current password"), commonerrors.ErrBadRequest)
)
