package handlers_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, displayName, password string) (*service.RegisterResult, error) {
	args := m.Called(ctx, email, displayName, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterResult), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*service.LoginResult, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
