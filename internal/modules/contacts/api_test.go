package contacts

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts/app"
	"github.com/jdgonzalez907/1channel/internal/modules/contacts/domain"
)

func TestNewContactsAPI(t *testing.T) {
	// Arrange
	getOrCreate := &app.MockGetOrCreateContactByExternalID{}
	findByID := &app.MockFindContactByID{}

	// Act
	api := NewContactsAPI(getOrCreate, findByID)

	// Assert
	assert.NotNil(t, api)
}

func TestContactsAPI_GetOrCreateContactIDByExternalID(t *testing.T) {
	contactID := uuid.NewV7()
	externalID := "5491112345678"
	now := time.Now()
	contact, err := domain.NewContact(contactID, externalID, now)
	assert.NoError(t, err)
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockGetOrCreateContactByExternalID)
		extID   string
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "success - returns contact id",
			setup: func(t *testing.T, m *app.MockGetOrCreateContactByExternalID) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.MatchedBy(func(in app.GetOrCreateContactByExternalIDInput) bool {
					return in.ExternalContactID == externalID && in.ContactID != uuid.Nil() && !in.CreatedAt.IsZero()
				})).Return(contact, nil).Once()
			},
			extID: externalID,
			want:  contactID,
		},
		{
			name: "failure - propagates use case error",
			setup: func(t *testing.T, m *app.MockGetOrCreateContactByExternalID) {
				t.Helper()
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, ucErr).Once()
			},
			extID:   externalID,
			want:    uuid.Nil(),
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockGetOrCreateContactByExternalID{}
			tt.setup(t, m)
			api := NewContactsAPI(m, &app.MockFindContactByID{})

			// Act
			got, err := api.GetOrCreateContactIDByExternalID(context.Background(), tt.extID)

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

func TestContactsAPI_FindExternalContactIDByContactID(t *testing.T) {
	contactID := uuid.NewV7()
	externalID := "5491112345678"
	now := time.Now()
	contact, err := domain.NewContact(contactID, externalID, now)
	assert.NoError(t, err)
	ucErr := errors.New("use case failure")

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *app.MockFindContactByID)
		id      uuid.UUID
		want    string
		wantErr error
	}{
		{
			name: "success - returns external id",
			setup: func(t *testing.T, m *app.MockFindContactByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindContactByIDInput{ContactID: contactID}).Return(contact, nil).Once()
			},
			id:   contactID,
			want: externalID,
		},
		{
			name: "failure - propagates use case error",
			setup: func(t *testing.T, m *app.MockFindContactByID) {
				t.Helper()
				m.On("Execute", mock.Anything, app.FindContactByIDInput{ContactID: contactID}).Return(nil, ucErr).Once()
			},
			id:      contactID,
			want:    "",
			wantErr: ucErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			m := &app.MockFindContactByID{}
			tt.setup(t, m)
			api := NewContactsAPI(&app.MockGetOrCreateContactByExternalID{}, m)

			// Act
			got, err := api.FindExternalContactIDByContactID(context.Background(), tt.id)

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
