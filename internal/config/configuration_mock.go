package config

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockConfiguration struct {
	Configuration
	mock.Mock
}

func (m *MockConfiguration) AppName() string { return m.Called().String(0) }

func (m *MockConfiguration) HTTPPort() string { return m.Called().String(0) }

func (m *MockConfiguration) LogLevel() string { return m.Called().String(0) }

func (m *MockConfiguration) OneChannelSecret() string { return m.Called().String(0) }

func (m *MockConfiguration) MetaSecret() string { return m.Called().String(0) }

func (m *MockConfiguration) PostgresHost() string { return m.Called().String(0) }

func (m *MockConfiguration) PostgresPort() uint16 { return uint16(m.Called().Int(0)) }

func (m *MockConfiguration) PostgresDatabase() string { return m.Called().String(0) }

func (m *MockConfiguration) PostgresUsername() string { return m.Called().String(0) }

func (m *MockConfiguration) PostgresPassword() string { return m.Called().String(0) }

func (m *MockConfiguration) PostgresMaxConns() int32 { return int32(m.Called().Int(0)) }

func (m *MockConfiguration) PostgresMinConns() int32 { return int32(m.Called().Int(0)) }

func (m *MockConfiguration) PostgresMaxConnLifeTime() time.Duration {
	return m.Called().Get(0).(time.Duration)
}

func (m *MockConfiguration) PostgresMaxConnIdleTime() time.Duration {
	return m.Called().Get(0).(time.Duration)
}

func NewMockSecretsConfiguration(metaSecret, oneChannelSecret string) *MockConfiguration {
	m := &MockConfiguration{}
	m.On("MetaSecret").Return(metaSecret)
	m.On("OneChannelSecret").Return(oneChannelSecret)
	return m
}
