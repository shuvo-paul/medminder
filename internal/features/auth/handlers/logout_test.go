package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogout_Successful(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()

	mockRepo.On("DeleteUserRefreshTokens", mock.Anything, userID).Return(nil)

	handler := handlers.LogoutHandler(mockRepo)

	err := handler(context.Background(), &handlers.LogoutInput{
		UserID: userID,
	})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogout_DBError(t *testing.T) {
	mockRepo := new(MockRefreshTokenRepository)
	userID := uuid.New()

	mockRepo.On("DeleteUserRefreshTokens", mock.Anything, userID).Return(errors.New("database error"))

	handler := handlers.LogoutHandler(mockRepo)

	err := handler(context.Background(), &handlers.LogoutInput{
		UserID: userID,
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, handlers.ErrLogoutFailed))
	mockRepo.AssertExpectations(t)
}
