package service

import (
	"errors"
	"fmt"

	commonerrors "github.com/shuvo-paul/medminder/internal/common/errors"
)

var (
	ErrProfileNotFound            = fmt.Errorf("%w: %w", errors.New("profile not found"), commonerrors.ErrNotFound)
	ErrScheduleNotFound           = fmt.Errorf("%w: %w", errors.New("dose schedule not found"), commonerrors.ErrNotFound)
	ErrUnauthorizedAccess         = fmt.Errorf("%w: %w", errors.New("unauthorized access to profile"), commonerrors.ErrForbidden)
	ErrInvalidTimezone            = fmt.Errorf("%w: %w", errors.New("invalid timezone"), commonerrors.ErrBadRequest)
	ErrScheduleNameExists         = fmt.Errorf("%w: %w", errors.New("dose schedule name already exists for this profile"), commonerrors.ErrConflict)
	ErrInvitationNotFound         = fmt.Errorf("%w: %w", errors.New("invitation not found"), commonerrors.ErrNotFound)
	ErrInvitationExpired          = fmt.Errorf("%w: %w", errors.New("invitation has expired"), commonerrors.ErrBadRequest)
	ErrInvitationAlreadyProcessed = fmt.Errorf("%w: %w", errors.New("invitation has already been processed"), commonerrors.ErrBadRequest)
	ErrUserAlreadySharing         = fmt.Errorf("%w: %w", errors.New("user is already sharing this profile"), commonerrors.ErrConflict)
	ErrCannotShareWithSelf        = fmt.Errorf("%w: %w", errors.New("cannot share profile with yourself"), commonerrors.ErrBadRequest)
	ErrUserNotFound               = fmt.Errorf("%w: %w", errors.New("user not found"), commonerrors.ErrNotFound)
	ErrInvalidPermissions         = fmt.Errorf("%w: %w", errors.New("invalid permissions"), commonerrors.ErrBadRequest)
)
