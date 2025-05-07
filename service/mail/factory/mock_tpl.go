package factory

import (
	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/dao"
)

var mockTemplate mail.Template

func ResetMockTemplate() {
	mockTemplate = nil
}

func getMockTemplate() *mockTemplateImpl {
	var mock *mockTemplateImpl
	if mockTemplate == nil {
		mock = &mockTemplateImpl{}
	} else {
		mock = mockTemplate.(*mockTemplateImpl)
	}
	return mock
}

func SetMockList(listFunc func(nextToken string) (*dao.ListTemplateResponse, error)) {
	mock := getMockTemplate()
	mock.listFunc = listFunc
	mockTemplate = mock
}

func SetMockNewTemplateException(e error) {
	mock := getMockTemplate()
	mock.newException = e
	mockTemplate = mock
}

func newMockTemplate() (mail.Template, error) {
	if mockTemplate == nil {
		return nil, nil
	}
	mock := getMockTemplate()
	if mock.newException != nil {
		return nil, mock.newException
	}
	return mockTemplate, nil
}

type mockTemplateImpl struct {
	newException error
	listFunc     func(nextToken string) (*dao.ListTemplateResponse, error)
}

func (m *mockTemplateImpl) List(nextToken string) (*dao.ListTemplateResponse, error) {
	if m.listFunc == nil {
		return nil, nil
	}
	return m.listFunc(nextToken)
}

func (m *mockTemplateImpl) Apply(tplfile string) error {
	return nil
}

func (m *mockTemplateImpl) Delete(name string) error {
	return nil
}

func (m *mockTemplateImpl) Detail(name string) (*dao.DetailTemplateResponse, error) {
	return nil, nil
}
