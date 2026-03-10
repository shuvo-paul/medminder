package errors_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shuvo-paul/medminder/internal/common/errors"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ErrNotFound returns 404",
			err:        errors.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrUnauthorized returns 401",
			err:        errors.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ErrForbidden returns 403",
			err:        errors.ErrForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "ErrConflict returns 409",
			err:        errors.ErrConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ErrBadRequest returns 400",
			err:        errors.ErrBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrInternal returns 500",
			err:        errors.ErrInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unknown error returns 500",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped ErrNotFound returns 404",
			err:        fmt.Errorf("getting user: %w", errors.ErrNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped ErrBadRequest returns 400",
			err:        fmt.Errorf("validating: %w", errors.ErrBadRequest),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "double wrapped ErrConflict returns 409",
			err:        fmt.Errorf("service: %w", fmt.Errorf("repo: %w", errors.ErrConflict)),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "nil error returns 500",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.StatusCode(tt.err)
			assert.Equal(t, tt.wantStatus, got)
		})
	}
}
