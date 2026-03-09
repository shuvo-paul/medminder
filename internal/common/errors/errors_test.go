package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	errpkg "github.com/shuvo-paul/medminder/internal/common/errors"
)

func TestStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ErrNotFound returns 404",
			err:        errpkg.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrUnauthorized returns 401",
			err:        errpkg.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "ErrForbidden returns 403",
			err:        errpkg.ErrForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "ErrConflict returns 409",
			err:        errpkg.ErrConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ErrBadRequest returns 400",
			err:        errpkg.ErrBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrInternal returns 500",
			err:        errpkg.ErrInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unknown error returns 500",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrapped ErrNotFound returns 404",
			err:        fmt.Errorf("getting user: %w", errpkg.ErrNotFound),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "wrapped ErrBadRequest returns 400",
			err:        fmt.Errorf("validating: %w", errpkg.ErrBadRequest),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "double wrapped ErrConflict returns 409",
			err:        fmt.Errorf("service: %w", fmt.Errorf("repo: %w", errpkg.ErrConflict)),
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
			got := errpkg.StatusCode(tt.err)
			assert.Equal(t, tt.wantStatus, got)
		})
	}
}
