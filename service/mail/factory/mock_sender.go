package factory

import (
	"testing"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
)

var mockSendor mail.ApiSender

func ResetMockSender() {
	mockSendor = nil
}

type mockSenderOpt func(*mockSenderImpl)

func WithMockSenderT(t *testing.T) mockSenderOpt {
	return func(m *mockSenderImpl) {
		m.t = t
	}
}

func SetMockSender(sendFunc func(t *testing.T, msg *service.Notification) (messageId string, err error), opts ...mockSenderOpt) {
	mock := getMockSender()
	mock.sendFunc = sendFunc
	for _, opt := range opts {
		opt(mock)
	}
	mockSendor = mock
}

func SetMockNewSenderException(e error) {
	mock := getMockSender()
	mock.newException = e
	mockSendor = mock
}

func getMockSender() *mockSenderImpl {
	var mock *mockSenderImpl
	if mockSendor == nil {
		mock = &mockSenderImpl{}
	} else {
		mock = mockSendor.(*mockSenderImpl)
	}
	return mock
}

func newMockSender() (mail.ApiSender, error) {
	if mockSendor == nil {
		return nil, nil
	}
	mock := getMockSender()
	if mock.newException != nil {
		return nil, mock.newException
	}
	return mockSendor, nil
}

type mockSenderImpl struct {
	t            *testing.T
	newException error
	sendFunc     func(t *testing.T, msg *service.Notification) (messageId string, err error)
}

func (m *mockSenderImpl) Send(msg *service.Notification) (messageId string, err error) {
	if m.sendFunc != nil {
		return m.sendFunc(m.t, msg)
	}
	return "", nil
}
