package aws

import (
	"errors"
	"testing"
	"time"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewAwsSession(t *testing.T) {
	// Test case: Successful creation of AWS session
	viper.Set("aws.ses.region", "us-west-2")
	viper.Set("aws.ses.credentails.filename", "test_credentials")
	viper.Set("aws.ses.credentails.profile", "default")
	sess, err := newAwsSession()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if sess == nil {
		t.Errorf("Expected non-nil session, got nil")
	}

	// Test case: Error in creating AWS session, empty key
	viper.Set("aws.ses.region", "us-west-2")
	viper.Set("aws.ses.credentails.filename", "test_credentials_not_exist")
	viper.Set("aws.ses.credentails.profile", "empty_key")
	sess, err = newAwsSession()
	if err == nil {
		t.Errorf("Expected non-nil error, got nil")
	}
	if sess != nil {
		t.Errorf("Expected nil session, got %v", sess)
	}

	// Test case: Error in getting AWS region
	viper.Set("aws.ses.region", "")
	sess, err = newAwsSession()
	if err == nil {
		t.Errorf("Expected non-nil error, got nil")
	}
	assert.ErrorContains(t, err, "aws.ses.region is empty")
	if sess != nil {
		t.Errorf("Expected nil session, got %v", sess)
	}

	// Test case: Error in getting AWS credentials filename
	viper.Set("aws.ses.region", "us-west-2")
	viper.Set("aws.ses.credentails.filename", "")
	sess, err = newAwsSession()
	if err == nil {
		t.Errorf("Expected non-nil error, got nil")
	}
	assert.ErrorContains(t, err, "aws.ses.credentails.filename is empty")
	if sess != nil {
		t.Errorf("Expected nil session, got %v", sess)
	}

	// Test case: Error in getting AWS credentials profile
	viper.Set("aws.ses.credentails.filename", "credentials.txt")
	viper.Set("aws.ses.credentails.profile", "")
	sess, err = newAwsSession()
	if err == nil {
		t.Errorf("Expected non-nil error, got nil")
	}
	if sess != nil {
		t.Errorf("Expected nil session, got %v", sess)
	}
	assert.ErrorContains(t, err, "aws.ses.credentails.profile is empty")

}

func TestNewTemplateStore(t *testing.T) {
	viper.Set("aws.ses.region", "us-west-2")
	viper.Set("aws.ses.credentails.filename", "test_credentials")
	viper.Set("aws.ses.credentails.profile", "default")
	tests := []struct {
		name    string
		opts    []awsTemplateStoreOpt
		wantErr bool
		prefunc func()
	}{
		{
			name: "no options",
			prefunc: func() {
				viper.Set("aws.ses.region", "us-west-2")
				viper.Set("aws.ses.credentails.filename", "test_credentials")
				viper.Set("aws.ses.credentails.profile", "default")
			},
		},
		{
			name: "with SES session option",
			opts: []awsTemplateStoreOpt{WithSesSession(&sesv2.SESV2{})},
		},
		{
			name:    "invalid session",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			if tt.prefunc != nil {
				tt.prefunc()
			}
			store, err := NewTemplateStore(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, store)
		})
	}
}

func TestDelete(t *testing.T) {
	// Test case: Template exists
	t.Run("Template exists", func(t *testing.T) {
		mockStore := NewMockStore(WithMockDelete(func(input *sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error) {
			return &sesv2.DeleteEmailTemplateOutput{}, nil
		}))
		awsTplImpl := &awsTplImpl{
			awsTemplateStore: mockStore,
		}

		err := awsTplImpl.Delete("template_name")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test case: Template does not exist
	t.Run("Template does not exist", func(t *testing.T) {
		mockStore := NewMockStore(WithMockDelete(func(input *sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error) {
			return nil, awserr.New(sesv2.ErrCodeNotFoundException, "template not found", nil)
		}))
		awsTplImpl := &awsTplImpl{
			awsTemplateStore: mockStore,
		}
		err := awsTplImpl.Delete("non_existent_template")

		if err == nil {
			t.Error("Expected an error, got nil")
		}
		assert.ErrorContains(t, err, sesv2.ErrCodeNotFoundException)
	})
}

func TestAwsTplImpl_List(t *testing.T) {
	// Mock awsTemplateStore
	tests := []struct {
		name     string
		token    string
		wantErr  bool
		listFunc func(input *sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error)
		wantTpl  []string
	}{
		{
			name:    "Test List with empty token",
			token:   "",
			wantErr: false,
			listFunc: func(input *sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error) {
				return &sesv2.ListEmailTemplatesOutput{
					TemplatesMetadata: []*sesv2.EmailTemplateMetadata{
						{TemplateName: aws.String("template1"), CreatedTimestamp: aws.Time(time.Now())},
						{TemplateName: aws.String("template2"), CreatedTimestamp: aws.Time(time.Now())},
					},
				}, nil
			},
			wantTpl: []string{"template1", "template2"},
		},
		{
			name:    "Test List with non-empty token",
			token:   "token",
			wantErr: false,
			listFunc: func(input *sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error) {
				return &sesv2.ListEmailTemplatesOutput{
					TemplatesMetadata: []*sesv2.EmailTemplateMetadata{
						{TemplateName: aws.String("template3"), CreatedTimestamp: aws.Time(time.Now())},
						{TemplateName: aws.String("template4"), CreatedTimestamp: aws.Time(time.Now())},
					},
				}, nil
			},
			wantTpl: []string{"template3", "template4"},
		},
		{
			name:    "Test List with error from ListEmailTemplates",
			token:   "",
			wantErr: true,
			listFunc: func(input *sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error) {
				return nil, errors.New("error from ListEmailTemplates")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockStore(WithMockList(tt.listFunc))
			tplImpl, err := NewTemplateStore(WithMockStore(mockStore))
			if err != nil {
				t.Errorf("NewTemplateStore() error = %v", err)
				return
			}

			resp, err := tplImpl.List(tt.token)
			if tt.wantErr {
				_, err = tplImpl.List(tt.token)
				assert.Error(t, err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("awsTplImpl.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tpsNames := make([]string, len(resp.Templates))
			for i, tpl := range resp.Templates {
				tpsNames[i] = tpl.Name
			}
			assert.Equal(t, tt.wantTpl, tpsNames)
		})
	}
}

func TestAwsTplImpl_Create(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    bool
		createTpl  *dao.Template
		createFunc func(input *sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error)
		wantTpl    []string
	}{
		{
			name:      "Test Create with empty token",
			wantErr:   false,
			createTpl: dao.NewTemplate("templateName", "zh-TW", "subject", "bodyPlaint", "bodyHtml"),
			createFunc: func(input *sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error) {
				assert.Equal(t, "templateName_zh-TW", *input.TemplateName)
				assert.Equal(t, "subject", *input.TemplateContent.Subject)
				assert.Equal(t, "bodyPlaint", *input.TemplateContent.Text)
				assert.Equal(t, "bodyHtml", *input.TemplateContent.Html)
				return &sesv2.CreateEmailTemplateOutput{}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockStore(WithMockCreate(tt.createFunc))
			tplImpl, err := NewTemplateStore(WithMockStore(mockStore))
			if err != nil {
				t.Errorf("NewTemplateStore() error = %v", err)
				return
			}

			err = tplImpl.CreateTpl(tt.createTpl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("awsTplImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAwsTplImpl_Update(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    bool
		updateTpl  *dao.Template
		updateFunc func(input *sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error)
		wantTpl    []string
	}{
		{
			name:      "Test Create with empty token",
			wantErr:   false,
			updateTpl: dao.NewTemplate("templateName", "zh-TW", "subject", "bodyPlaint", "bodyHtml"),
			updateFunc: func(input *sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error) {
				assert.Equal(t, "templateName_zh-TW", *input.TemplateName)
				assert.Equal(t, "subject", *input.TemplateContent.Subject)
				assert.Equal(t, "bodyPlaint", *input.TemplateContent.Text)
				assert.Equal(t, "bodyHtml", *input.TemplateContent.Html)
				return &sesv2.UpdateEmailTemplateOutput{}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := NewMockStore(WithMockUpdate(tt.updateFunc))
			tplImpl, err := NewTemplateStore(WithMockStore(mockStore))
			if err != nil {
				t.Errorf("NewTemplateStore() error = %v", err)
				return
			}

			err = tplImpl.UpdateTemplate(tt.updateTpl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("awsTplImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAwsTplImpl_IsTemplateExist(t *testing.T) {

	// Test case: Template exists
	mockStore := NewMockStore(WithMockGet(func(input *sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error) {
		return &sesv2.GetEmailTemplateOutput{
			TemplateName: aws.String("template_name"),
		}, nil
	}))
	tplImpl, err := NewTemplateStore(WithMockStore(mockStore))
	if err != nil {
		t.Errorf("NewTemplateStore() error = %v", err)
		return
	}
	exists, err := tplImpl.IsTemplateExist("template_name")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test case: Template does not exist
	mockStore = NewMockStore(WithMockGet(func(input *sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error) {
		return nil, awserr.New(sesv2.ErrCodeNotFoundException, "template not found", nil)
	}))
	tplImpl, err = NewTemplateStore(WithMockStore(mockStore))
	if err != nil {
		t.Errorf("NewTemplateStore() error = %v", err)
		return
	}
	exists, err = tplImpl.IsTemplateExist("non_existent_template")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestNewApiSender(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		sessErr error
		prefunc func()
	}{
		{
			name:    "Successful creation",
			wantErr: false,
			sessErr: nil,
			prefunc: func() {
				viper.Set("aws.ses.region", "us-west-2")
				viper.Set("aws.ses.credentails.filename", "test_credentials")
				viper.Set("aws.ses.credentails.profile", "default")
				viper.Set("aws.ses.from", "test@example.com")
			},
		},
		{
			name:    "Error in creating AWS session",
			wantErr: true,
			sessErr: errors.New("session error"),
			prefunc: func() {
				viper.Set("aws.ses.from", "test@example.com")
			},
		},
		{
			name:    "Empty 'aws.ses.from' configuration",
			wantErr: true,
			sessErr: nil,
			prefunc: func() {
				viper.Set("aws.ses.region", "us-west-2")
				viper.Set("aws.ses.credentails.filename", "test_credentials")
				viper.Set("aws.ses.credentails.profile", "default")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock viper configuration
			viper.Reset()
			if tt.prefunc != nil {
				tt.prefunc()
			}

			// Call NewApiSender function
			sender, err := NewApiSender()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApiSender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check sender
			if tt.wantErr {
				assert.Nil(t, sender)
			} else {
				assert.NotNil(t, sender)
				assert.IsType(t, &awsApiSender{}, sender)
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name          string
		opts          []apiSenderOpt
		notify        *service.Notification
		wantErr       bool
		expectedMsgId string
	}{
		{
			name: "valid notification",
			opts: []apiSenderOpt{
				WithTemplateStore(mail.NewMockTemplateStore(
					mail.WithIsTemplateExist(func(name string) (bool, error) {
						return true, nil
					}),
				)),
				WithAwsSender(
					NewMockSender(func(input *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error) {
						return &sesv2.SendEmailOutput{
							MessageId: aws.String("1234"),
						}, nil
					}),
				),
			},
			notify: &service.Notification{
				Data: map[string]string{"key": "value"},
				SendTo: []*service.Info{
					{
						Sub:    "test-subject",
						Name:   "test-name",
						Email:  "sendto@example.com",
						Enable: true,
					},
				},
				Event: "test-event",
				Lang:  "zh-TW",
				From:  &service.Info{Email: "from@example.com"},
			},
			wantErr:       false,
			expectedMsgId: "1234",
		},
		{
			name: "tempate not found notification",
			opts: []apiSenderOpt{
				WithTemplateStore(mail.NewMockTemplateStore(
					mail.WithIsTemplateExist(func(name string) (bool, error) {
						return false, nil
					}),
				)),
				WithAwsSender(
					NewMockSender(func(input *sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error) {
						return &sesv2.SendEmailOutput{
							MessageId: aws.String("1234"),
						}, nil
					}),
				),
			},
			notify: &service.Notification{
				Data: map[string]string{"key": "value"},
				SendTo: []*service.Info{
					{
						Sub:    "test-subject",
						Name:   "test-name",
						Email:  "sendto@example.com",
						Enable: true,
					},
				},
				Event: "test-event",
				Lang:  "zh-TW",
				From:  &service.Info{Email: "from@example.com"},
			},
			wantErr:       true,
			expectedMsgId: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock template store
			viper.Set("aws.ses.from", tt.notify.From.Email)
			sender, err := NewApiSender(tt.opts...)
			if err != nil {
				t.Errorf("NewApiSender() error = %v", err)
				return
			}
			assert.NotNil(t, sender)
			msgId, err := sender.Send(tt.notify)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedMsgId, msgId)
		})
	}
}
