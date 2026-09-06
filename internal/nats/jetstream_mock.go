package nats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/mock"
)

type MockJetStream struct {
	jetstream.JetStream
	mock.Mock
}

func (m *MockJetStream) Publish(_ context.Context, subject string, data []byte, _ ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	args := m.Called(subject, data)
	return args.Get(0).(*jetstream.PubAck), args.Error(1)
}

func (m *MockJetStream) CreateOrUpdateStream(_ context.Context, cfg jetstream.StreamConfig) (jetstream.Stream, error) {
	args := m.Called(cfg)
	return args.Get(0).(jetstream.Stream), args.Error(1)
}

func (m *MockJetStream) CreateOrUpdateConsumer(_ context.Context, stream string, cfg jetstream.ConsumerConfig) (jetstream.Consumer, error) {
	args := m.Called(stream, cfg)
	return args.Get(0).(jetstream.Consumer), args.Error(1)
}

type MockStream struct {
	jetstream.Stream
	mock.Mock
}

type MockConsumer struct {
	jetstream.Consumer
	mock.Mock
}

func (m *MockConsumer) Consume(handler jetstream.MessageHandler, _ ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error) {
	args := m.Called(handler)
	return args.Get(0).(jetstream.ConsumeContext), args.Error(1)
}

type MockConsumeContext struct {
	jetstream.ConsumeContext
	mock.Mock
}

func (m *MockConsumeContext) Stop() {
	m.Called()
}

type MockMsg struct {
	jetstream.Msg
	mock.Mock
}

func (m *MockMsg) Subject() string {
	return m.Called().String(0)
}

func (m *MockMsg) Data() []byte {
	return m.Called().Get(0).([]byte)
}

func (m *MockMsg) Metadata() (*jetstream.MsgMetadata, error) {
	args := m.Called()
	return args.Get(0).(*jetstream.MsgMetadata), args.Error(1)
}

func (m *MockMsg) Ack() error {
	return m.Called().Error(0)
}

func (m *MockMsg) Nak() error {
	return m.Called().Error(0)
}
