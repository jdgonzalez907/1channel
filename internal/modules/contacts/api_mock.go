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

func (m *MockContactsAPI) GetOrCreateContactByExternalID(ctx context.Context, externalContactID string) (uuid.UUID, error) {
	args := m.Called(ctx, externalContactID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
