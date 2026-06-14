package service

import (
	"errors"
	"fmt"

	commonerrors "github.com/shuvo-paul/medminder/internal/common/errors"
)

var (
	ErrGuestTokenNotFound          = fmt.Errorf("%w: %w", errors.New("guest access token not found"), commonerrors.ErrNotFound)
	ErrGuestTokenExpired           = fmt.Errorf("%w: %w", errors.New("guest access token has expired"), commonerrors.ErrBadRequest)
	ErrGuestTokenInsufficientPerms = fmt.Errorf("%w: %w", errors.New("guest access token lacks required permission"), commonerrors.ErrForbidden)
	ErrInvalidPermissions          = fmt.Errorf("%w: %w", errors.New("invalid permissions"), commonerrors.ErrBadRequest)
)
