package identity

func SetMockHealthFunc(f func() (bool, error)) {
	mock := getMockIdentity()
	mock.healthFunc = f
	mockIdentity = mock
}

func getMockIdentity() *mockIdentityImpl {
	var mock *mockIdentityImpl
	if mockIdentity == nil {
		mock = &mockIdentityImpl{}
	} else {
		mock = mockIdentity.(*mockIdentityImpl)
	}
	return mock
}

func SetMockSubToInfoFunc(f func(from string, to []string) (*ClassificationLang, error)) {
	mock := getMockIdentity()
	mock.subToInfo = f
	mockIdentity = mock
}

func SetNewException(e error) {
	mock := getMockIdentity()
	mock.newException = e
	mockIdentity = mock
}

func newMockIdentity() (Identity, error) {
	if mockIdentity == nil {
		return nil, nil
	}
	mock := getMockIdentity()
	if mock.newException != nil {
		return nil, mock.newException
	}
	return mockIdentity, nil

}

func ResetMock() {
	mockIdentity = nil
}

var mockIdentity Identity

type mockIdentityImpl struct {
	newException error
	healthFunc   func() (bool, error)
	subToInfo    func(from string, to []string) (*ClassificationLang, error)
}

func (m *mockIdentityImpl) IsReady() (bool, error) {
	if m.healthFunc == nil {
		return true, nil
	}
	return m.healthFunc()
}

func (m *mockIdentityImpl) SubToInfo(from string, to []string) (*ClassificationLang, error) {
	if m.subToInfo != nil {
		return m.subToInfo(from, to)
	}
	return nil, nil
}
