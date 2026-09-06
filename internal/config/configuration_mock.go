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

func (m *MockConfiguration) NatsHost() string { return m.Called().String(0) }

func (m *MockConfiguration) NatsPort() int { return m.Called().Int(0) }

func (m *MockConfiguration) NatsUsername() string { return m.Called().String(0) }

func (m *MockConfiguration) NatsPassword() string { return m.Called().String(0) }

func (m *MockConfiguration) NatsTimeout() time.Duration { return m.Called().Get(0).(time.Duration) }

func (m *MockConfiguration) NatsMaxReconnects() int { return m.Called().Int(0) }

func (m *MockConfiguration) NatsReconnectWait() time.Duration {
	return m.Called().Get(0).(time.Duration)
}

func (m *MockConfiguration) NatsPingInterval() time.Duration {
	return m.Called().Get(0).(time.Duration)
}

func (m *MockConfiguration) NatsMaxPingsOut() int { return m.Called().Int(0) }

func NewMockSecretsConfiguration(metaSecret, oneChannelSecret string) *MockConfiguration {
	m := &MockConfiguration{}
	m.On("MetaSecret").Return(metaSecret)
	m.On("OneChannelSecret").Return(oneChannelSecret)
	return m
}

func NewMockNatsConfiguration(host string, port int, timeout time.Duration) *MockConfiguration {
	m := &MockConfiguration{}
	m.On("AppName").Return("1channel")
	m.On("NatsHost").Return(host)
	m.On("NatsPort").Return(port)
	m.On("NatsUsername").Return("dev_user")
	m.On("NatsPassword").Return("dev_pass")
	m.On("NatsTimeout").Return(timeout)
	m.On("NatsReconnectWait").Return(time.Second)
	m.On("NatsMaxReconnects").Return(-1)
	m.On("NatsPingInterval").Return(time.Minute)
	m.On("NatsMaxPingsOut").Return(2)
	return m
}
