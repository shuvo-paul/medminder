package handlers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogout_Successful(t *testing.T) {
	mockSvc := new(MockAuthService)
	userID := uuid.New()

	mockSvc.On("Logout", mock.Anything, userID).Return(nil)

	handler := handlers.LogoutHandler(mockSvc)

	err := handler(context.Background(), &handlers.LogoutInput{
		UserID: userID,
	})

	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestLogout_DBError(t *testing.T) {
	mockSvc := new(MockAuthService)
	userID := uuid.New()

	mockSvc.On("Logout", mock.Anything, userID).Return(service.ErrLogoutFailed)

	handler := handlers.LogoutHandler(mockSvc)

	err := handler(context.Background(), &handlers.LogoutInput{
		UserID: userID,
	})

	assert.Error(t, err)
}

func TestLogout_NilUserID(t *testing.T) {
	mockSvc := new(MockAuthService)

	handler := handlers.LogoutHandler(mockSvc)

	err := handler(context.Background(), &handlers.LogoutInput{
		UserID: uuid.Nil,
	})

	assert.Error(t, err)
}