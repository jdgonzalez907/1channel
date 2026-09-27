package users

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/users/app"
	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

func TestNewUsersAPI(t *testing.T) {
	// Arrange
	findUC := &app.MockFindUserByID{}
	createUC := &app.MockCreateUser{}

	// Act
	api := NewUsersAPI(findUC, createUC)

	// Assert
	assert.NotNil(t, api)
}

func TestUsersAPI_FindUserByID(t *testing.T) {
	userID := uuid.NewV7()
	now := time.Now()
	user, err := domain.NewUser(userID, now)
	assert.NoError(t, err)
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockFindUserByID)
		id      uuid.UUID
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "success - returns user id",
			setup: func(t *testing.T, m *app.MockFindUserByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindUserByIDInput{ID: userID}).Return(user, nil).Once()
			},
			id:   userID,
			want: userID,
		},
		{
			name: "failure - propagates use case error",
			setup: func(t *testing.T, m *app.MockFindUserByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindUserByIDInput{ID: userID}).Return(nil, ucErr).Once()
			},
			id:      userID,
			want:    uuid.Nil(),
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockFindUserByID{}
			tt.setup(t, m)

			// Act
			got, err := NewUsersAPI(m, &app.MockCreateUser{}).FindUserByID(context.Background(), tt.id)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			m.AssertExpectations(t)
		})
	}
}

func TestUsersAPI_CreateUser(t *testing.T) {
	userID := uuid.NewV7()
	now := time.Now()
	user, err := domain.NewUser(userID, now)
	assert.NoError(t, err)
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockCreateUser)
		input   CreateUserInput
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "success - returns user id",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, app.CreateUserInput{ID: userID, CreatedAt: now}).Return(user, nil).Once()
			},
			input: CreateUserInput{ID: userID, CreatedAt: now},
			want:  userID,
		},
		{
			name: "failure - propagates use case error",
			setup: func(t *testing.T, m *app.MockCreateUser) {
				t.Helper()
				m.On("Execute", mock.Anything, app.CreateUserInput{ID: userID, CreatedAt: now}).Return(nil, ucErr).Once()
			},
			input:   CreateUserInput{ID: userID, CreatedAt: now},
			want:    uuid.Nil(),
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockCreateUser{}
			tt.setup(t, m)

			// Act
			got, err := NewUsersAPI(&app.MockFindUserByID{}, m).CreateUser(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, tt.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			m.AssertExpectations(t)
		})
	}
}
