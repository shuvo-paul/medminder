package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) CreateProfile(ctx context.Context, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error) {
	args := m.Called(ctx, name, dateOfBirth, timezone)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetProfileByID(ctx context.Context, id uuid.UUID) (db.Profile, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockProfileRepository) ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]db.Profile, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Profile), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error) {
	args := m.Called(ctx, id, name, dateOfBirth, timezone)
	return args.Get(0).(db.Profile), args.Error(1)
}

func (m *MockProfileRepository) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockDoseScheduleRepository struct {
	mock.Mock
}

func (m *MockDoseScheduleRepository) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	args := m.Called(ctx, profileID, name, t)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error) {
	args := m.Called(ctx, profileID)
	return args.Get(0).([]db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	args := m.Called(ctx, id, name, t)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDoseScheduleRepository) DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func TestProfileService_Create(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			name        string
			dateOfBirth *time.Time
			timezone    string
			schedules   []service.DoseScheduleInput
		}
		expect func(*testing.T, *service.ProfileResult, error)
	}{
		{
			name: "CreateProfile_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileID := uuid.New()
				now := time.Now()
				profileRepo.On("CreateProfile", mock.Anything, "Test Profile", mock.Anything, "UTC").
					Return(db.Profile{
						ID:        profileID,
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				scheduleRepo.On("CreateDoseSchedule", mock.Anything, profileID, "Morning", mock.Anything).
					Return(db.DoseSchedule{
						ID:        uuid.New(),
						ProfileID: profileID,
						Name:      "Morning",
						Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			input: struct {
				name        string
				dateOfBirth *time.Time
				timezone    string
				schedules   []service.DoseScheduleInput
			}{
				name:        "Test Profile",
				dateOfBirth: nil,
				timezone:    "UTC",
				schedules: []service.DoseScheduleInput{
					{Name: "Morning", Time: "08:00"},
				},
			},
			expect: func(t *testing.T, result *service.ProfileResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Test Profile", result.Profile.Name)
				assert.Equal(t, "UTC", result.Profile.Timezone)
			},
		},
		{
			name: "CreateProfile_InvalidTimezone",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
			},
			input: struct {
				name        string
				dateOfBirth *time.Time
				timezone    string
				schedules   []service.DoseScheduleInput
			}{
				name:        "Test Profile",
				dateOfBirth: nil,
				timezone:    "Invalid/Timezone",
				schedules:   nil,
			},
			expect: func(t *testing.T, result *service.ProfileResult, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrInvalidTimezone))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewProfileService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.CreateProfile(context.Background(), tt.input.name, tt.input.dateOfBirth, tt.input.timezone, tt.input.schedules)
			tt.expect(t, result, err)
		})
	}
}

func TestProfileService_Get(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID uuid.UUID
			userID    uuid.UUID
		}
		expect func(*testing.T, *service.ProfileResult, error)
	}{
		{
			name: "GetProfile_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				now := time.Now()
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{
						ID:        uuid.New(),
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, mock.Anything).
					Return([]db.DoseSchedule{}, nil)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
			}{
				profileID: uuid.New(),
				userID:    uuid.UUID{1},
			},
			expect: func(t *testing.T, result *service.ProfileResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "GetProfile_NotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{}, sql.ErrNoRows)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
			}{
				profileID: uuid.New(),
				userID:    uuid.New(),
			},
			expect: func(t *testing.T, result *service.ProfileResult, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrProfileNotFound))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewProfileService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.GetProfile(context.Background(), tt.input.profileID, tt.input.userID)
			tt.expect(t, result, err)
		})
	}
}

func TestProfileService_List(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockDoseScheduleRepository)
		input  uuid.UUID
		expect func(*testing.T, []service.ProfileResult, error)
	}{
		{
			name: "ListProfiles_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				now := time.Now()
				profileRepo.On("ListProfilesByUser", mock.Anything, mock.Anything).
					Return([]db.Profile{
						{
							ID:        uuid.New(),
							Name:      "Profile 1",
							Timezone:  "UTC",
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        uuid.New(),
							Name:      "Profile 2",
							Timezone:  "UTC",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, mock.Anything).
					Return([]db.DoseSchedule{}, nil).Maybe()
			},
			input: uuid.New(),
			expect: func(t *testing.T, result []service.ProfileResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, result, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewProfileService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.ListProfiles(context.Background(), tt.input)
			tt.expect(t, result, err)
		})
	}
}

func TestProfileService_Update(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID uuid.UUID
			userID    uuid.UUID
			name      *string
			timezone  *string
		}
		expect func(*testing.T, *service.ProfileResult, error)
	}{
		{
			name: "UpdateProfile_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				now := time.Now()
				profileID := uuid.New()
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{
						ID:        profileID,
						Name:      "Old Name",
						Timezone:  "UTC",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				newName := "New Name"
				profileRepo.On("UpdateProfile", mock.Anything, mock.Anything, newName, mock.Anything, "UTC").
					Return(db.Profile{
						ID:        profileID,
						Name:      newName,
						Timezone:  "UTC",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, mock.Anything).
					Return([]db.DoseSchedule{}, nil)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
				name      *string
				timezone  *string
			}{
				profileID: uuid.New(),
				userID:    uuid.UUID{1},
				name:      func() *string { s := "New Name"; return &s }(),
				timezone:  func() *string { s := "UTC"; return &s }(),
			},
			expect: func(t *testing.T, result *service.ProfileResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewProfileService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.UpdateProfile(context.Background(), tt.input.profileID, tt.input.userID, tt.input.name, nil, tt.input.timezone)
			tt.expect(t, result, err)
		})
	}
}

func TestProfileService_Delete(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID uuid.UUID
			userID    uuid.UUID
		}
		expect func(*testing.T, error)
	}{
		{
			name: "DeleteProfile_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{
						ID:        uuid.New(),
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("DeleteDoseSchedulesByProfile", mock.Anything, mock.Anything).Return(nil)
				profileRepo.On("DeleteProfile", mock.Anything, mock.Anything).Return(nil)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
			}{
				profileID: uuid.New(),
				userID:    uuid.UUID{1},
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewProfileService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			err := svc.DeleteProfile(context.Background(), tt.input.profileID, tt.input.userID)
			tt.expect(t, err)
		})
	}
}

func TestDoseScheduleService_Create(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID uuid.UUID
			userID    uuid.UUID
			name      string
			timeStr   string
		}
		expect func(*testing.T, *service.DoseScheduleResult, error)
	}{
		{
			name: "CreateSchedule_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				var capturedProfileID uuid.UUID
				profileRepo.On("GetProfileByID", mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool {
					capturedProfileID = id
					return true
				})).Return(db.Profile{
					ID:        capturedProfileID,
					Name:      "Test Profile",
					Timezone:  "UTC",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
				scheduleRepo.On("CreateDoseSchedule", mock.Anything, mock.Anything, "Morning", mock.Anything).
					Return(db.DoseSchedule{
						ID:        uuid.New(),
						ProfileID: capturedProfileID,
						Name:      "Morning",
						Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
				name      string
				timeStr   string
			}{
				profileID: uuid.New(),
				userID:    uuid.UUID{1},
				name:      "Morning",
				timeStr:   "08:00",
			},
			expect: func(t *testing.T, result *service.DoseScheduleResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Morning", result.Schedule.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewDoseScheduleService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.CreateDoseSchedule(context.Background(), tt.input.profileID, tt.input.userID, tt.input.name, tt.input.timeStr)
			tt.expect(t, result, err)
		})
	}
}

func TestDoseScheduleService_Get(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID  uuid.UUID
			scheduleID uuid.UUID
			userID     uuid.UUID
		}
		expect func(*testing.T, *service.DoseScheduleResult, error)
	}{
		{
			name: "GetSchedule_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileID := uuid.UUID{1}
				scheduleID := uuid.New()
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{
						ID:        profileID,
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("GetDoseScheduleByID", mock.Anything, mock.Anything).
					Return(db.DoseSchedule{
						ID:        scheduleID,
						ProfileID: profileID,
						Name:      "Morning",
						Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			input: struct {
				profileID  uuid.UUID
				scheduleID uuid.UUID
				userID     uuid.UUID
			}{
				profileID:  uuid.UUID{1},
				scheduleID: uuid.New(),
				userID:     uuid.UUID{1},
			},
			expect: func(t *testing.T, result *service.DoseScheduleResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "GetSchedule_NotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{
						ID:        uuid.New(),
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("GetDoseScheduleByID", mock.Anything, mock.Anything).
					Return(db.DoseSchedule{}, sql.ErrNoRows)
			},
			input: struct {
				profileID  uuid.UUID
				scheduleID uuid.UUID
				userID     uuid.UUID
			}{
				profileID:  uuid.New(),
				scheduleID: uuid.New(),
				userID:     uuid.UUID{1},
			},
			expect: func(t *testing.T, result *service.DoseScheduleResult, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrScheduleNotFound))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewDoseScheduleService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.GetDoseSchedule(context.Background(), tt.input.profileID, tt.input.scheduleID, tt.input.userID)
			tt.expect(t, result, err)
		})
	}
}

func TestDoseScheduleService_List(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID uuid.UUID
			userID    uuid.UUID
		}
		expect func(*testing.T, []service.DoseScheduleResult, error)
	}{
		{
			name: "ListSchedules_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{
						ID:        uuid.New(),
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, mock.Anything).
					Return([]db.DoseSchedule{
						{
							ID:        uuid.New(),
							ProfileID: uuid.New(),
							Name:      "Morning",
							Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil)
			},
			input: struct {
				profileID uuid.UUID
				userID    uuid.UUID
			}{
				profileID: uuid.New(),
				userID:    uuid.UUID{1},
			},
			expect: func(t *testing.T, result []service.DoseScheduleResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, result, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewDoseScheduleService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.ListDoseSchedules(context.Background(), tt.input.profileID, tt.input.userID)
			tt.expect(t, result, err)
		})
	}
}

func TestDoseScheduleService_Update(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID  uuid.UUID
			scheduleID uuid.UUID
			userID     uuid.UUID
			name       *string
			timeStr    *string
		}
		expect func(*testing.T, *service.DoseScheduleResult, error)
	}{
		{
			name: "UpdateSchedule_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileID := uuid.UUID{1}
				scheduleID := uuid.New()
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{
						ID:        profileID,
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("GetDoseScheduleByID", mock.Anything, mock.Anything).
					Return(db.DoseSchedule{
						ID:        scheduleID,
						ProfileID: profileID,
						Name:      "Morning",
						Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				newName := "Evening"
				scheduleRepo.On("UpdateDoseSchedule", mock.Anything, mock.Anything, newName, mock.Anything).
					Return(db.DoseSchedule{
						ID:        scheduleID,
						ProfileID: profileID,
						Name:      newName,
						Time:      time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			input: struct {
				profileID  uuid.UUID
				scheduleID uuid.UUID
				userID     uuid.UUID
				name       *string
				timeStr    *string
			}{
				profileID:  uuid.UUID{1},
				scheduleID: uuid.New(),
				userID:     uuid.UUID{1},
				name:       func() *string { s := "Evening"; return &s }(),
				timeStr:    nil,
			},
			expect: func(t *testing.T, result *service.DoseScheduleResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewDoseScheduleService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			result, err := svc.UpdateDoseSchedule(context.Background(), tt.input.profileID, tt.input.scheduleID, tt.input.userID, tt.input.name, tt.input.timeStr)
			tt.expect(t, result, err)
		})
	}
}

func TestDoseScheduleService_Delete(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository)
		input struct {
			profileID  uuid.UUID
			scheduleID uuid.UUID
			userID     uuid.UUID
		}
		expect func(*testing.T, error)
	}{
		{
			name: "DeleteSchedule_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository) {
				profileID := uuid.UUID{1}
				scheduleID := uuid.New()
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{
						ID:        profileID,
						Name:      "Test Profile",
						Timezone:  "UTC",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("GetDoseScheduleByID", mock.Anything, mock.Anything).
					Return(db.DoseSchedule{
						ID:        scheduleID,
						ProfileID: profileID,
						Name:      "Morning",
						Time:      time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				scheduleRepo.On("DeleteDoseSchedule", mock.Anything, mock.Anything).Return(nil)
			},
			input: struct {
				profileID  uuid.UUID
				scheduleID uuid.UUID
				userID     uuid.UUID
			}{
				profileID:  uuid.UUID{1},
				scheduleID: uuid.New(),
				userID:     uuid.UUID{1},
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			svc := service.NewDoseScheduleService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo)
			err := svc.DeleteDoseSchedule(context.Background(), tt.input.profileID, tt.input.scheduleID, tt.input.userID)
			tt.expect(t, err)
		})
	}
}
