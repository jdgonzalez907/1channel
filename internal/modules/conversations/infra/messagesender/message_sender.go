package messagesender

import (
	"context"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
)

type StubSender struct{}

func NewStubSender() *StubSender {
	return &StubSender{}
}

func (s *StubSender) Send(_ context.Context, _ string, _ *domain.Message) (string, error) {
	return uuid.NewV7().String(), nil
}
