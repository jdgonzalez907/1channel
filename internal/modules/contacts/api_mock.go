package contacts

import (
	"context"

	"uuid"

	"github.com/stretchr/testify/mock"
)

type MockContactsAPI struct {
	ContactsAPI
	mock.Mock
}

func (m *MockContactsAPI) GetOrCreateContactIDByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error) {
	args := m.Called(ctx, externalContactID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockContactsAPI) FindExternalContactIDByContactID(ctx context.Context, contactID uuid.UUID) (string, error) {
	args := m.Called(ctx, contactID)
	return args.Get(0).(string), args.Error(1)
}
