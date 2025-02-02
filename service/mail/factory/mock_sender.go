package factory

import (
	"fmt"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
)

var mockSendor mail.ApiSender

func ResetMockSender() {
	mockSendor = nil
}

func SetMockSender(sendFunc func(msg *service.Notification) (messageId string, err error)) {
	mock := getMockSender()
	mock.sendFunc = sendFunc
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
	fmt.Println("newMockSender", mockSendor)
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
	newException error
	sendFunc     func(msg *service.Notification) (messageId string, err error)
}

func (m *mockSenderImpl) Send(msg *service.Notification) (messageId string, err error) {
	if m.sendFunc != nil {
		return m.sendFunc(msg)
	}
	return "", nil
}
