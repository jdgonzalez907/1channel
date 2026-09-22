package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockMessageSender struct {
	MessageSender
	mock.Mock
}

func (m *MockMessageSender) Send(ctx context.Context, to string, message *Message) (string, error) {
	args := m.Called(ctx, to, message)
	return args.Get(0).(string), args.Error(1)
}
