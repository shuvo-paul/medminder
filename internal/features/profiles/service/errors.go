package service

import (
	"errors"
	"fmt"

	commonerrors "github.com/shuvo-paul/medminder/internal/common/errors"
)

var (
	ErrProfileNotFound    = fmt.Errorf("%w: %w", errors.New("profile not found"), commonerrors.ErrNotFound)
	ErrScheduleNotFound   = fmt.Errorf("%w: %w", errors.New("dose schedule not found"), commonerrors.ErrNotFound)
	ErrUnauthorizedAccess = fmt.Errorf("%w: %w", errors.New("unauthorized access to profile"), commonerrors.ErrForbidden)
	ErrInvalidTimezone    = fmt.Errorf("%w: %w", errors.New("invalid timezone"), commonerrors.ErrBadRequest)
	ErrScheduleNameExists = fmt.Errorf("%w: %w", errors.New("dose schedule name already exists for this profile"), commonerrors.ErrConflict)
)
