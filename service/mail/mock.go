package mail

import "github.com/arwoosa/notifaction/service/mail/dao"

type templateStoreOpt func(*mockTemplateStore)

func WithIsTemplateExist(f func(name string) (bool, error)) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.isTemplateExist = f
	}
}

func WithUpdateTemplate(f func(tpl *dao.Template) error) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.updateTemplate = f
	}
}

func WithCreateTemplate(f func(tpl *dao.Template) error) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.createTemplate = f
	}
}

func WithDeleteTemplate(f func(name string) error) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.deleteTemplate = f
	}
}

func WithListTemplate(f func(token string) (*dao.ListTemplateResponse, error)) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.listTemplate = f
	}
}

func WithDetailTemplate(f func(name string) (*dao.DetailTemplateResponse, error)) templateStoreOpt {
	return func(m *mockTemplateStore) {
		m.detailTemplate = f
	}
}

func NewMockTemplateStore(opts ...templateStoreOpt) TemplateStore {
	m := &mockTemplateStore{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type mockTemplateStore struct {
	isTemplateExist func(name string) (bool, error)
	updateTemplate  func(tpl *dao.Template) error
	createTemplate  func(tpl *dao.Template) error
	deleteTemplate  func(name string) error
	listTemplate    func(token string) (*dao.ListTemplateResponse, error)
	detailTemplate  func(name string) (*dao.DetailTemplateResponse, error)
}

func (m *mockTemplateStore) IsTemplateExist(name string) (bool, error) {
	return m.isTemplateExist(name)
}

func (m *mockTemplateStore) UpdateTemplate(tpl *dao.Template) error {
	return m.updateTemplate(tpl)
}

func (m *mockTemplateStore) CreateTpl(tpl *dao.Template) error {
	return m.createTemplate(tpl)
}

func (m *mockTemplateStore) Delete(name string) error {
	return m.deleteTemplate(name)
}

func (m *mockTemplateStore) List(token string) (*dao.ListTemplateResponse, error) {
	return m.listTemplate(token)
}

func (m *mockTemplateStore) Detail(name string) (*dao.DetailTemplateResponse, error) {
	return m.detailTemplate(name)
}
