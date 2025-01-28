package aws

import "github.com/aws/aws-sdk-go/service/sesv2"

func WithMockStore(store awsTemplateStore) awsTemplateStoreOpt {
	return func(a *awsTplImpl) {
		a.awsTemplateStore = store
	}
}

type mockStoreOpt func(*awsMockStore)

func WithMockCreate(create func(*sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error)) mockStoreOpt {
	return func(a *awsMockStore) {
		a.create = create
	}
}

func WithMockDelete(delete func(*sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error)) mockStoreOpt {
	return func(a *awsMockStore) {
		a.deletre = delete
	}
}

func WithMockList(list func(*sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error)) mockStoreOpt {
	return func(a *awsMockStore) {
		a.list = list
	}
}

func WithMockUpdate(update func(*sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error)) mockStoreOpt {
	return func(a *awsMockStore) {
		a.update = update
	}
}

func WithMockGet(get func(*sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error)) mockStoreOpt {
	return func(a *awsMockStore) {
		a.get = get
	}
}

func NewMockStore(opts ...mockStoreOpt) awsTemplateStore {
	store := &awsMockStore{}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

type awsMockStore struct {
	create  func(*sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error)
	deletre func(*sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error)
	list    func(*sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error)
	update  func(*sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error)
	get     func(*sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error)
}

func (a *awsMockStore) CreateEmailTemplate(input *sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error) {
	return a.create(input)
}

func (a *awsMockStore) DeleteEmailTemplate(input *sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error) {
	return a.deletre(input)
}

func (a *awsMockStore) ListEmailTemplates(input *sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error) {
	return a.list(input)
}

func (a *awsMockStore) UpdateEmailTemplate(input *sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error) {
	return a.update(input)
}

func (a *awsMockStore) GetEmailTemplate(input *sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error) {
	return a.get(input)
}

func NewMockSender(send func(*sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error)) awsSender {
	return &mockAwsSender{send: send}
}

type mockAwsSender struct {
	send func(*sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error)
}

func (a *mockAwsSender) SendEmail(input *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error) {
	return a.send(input)
}
