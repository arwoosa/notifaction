package aws

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/spf13/viper"
)

type awsSender interface {
	SendEmail(*sesv2.SendEmailInput) (*sesv2.SendEmailOutput, error)
}

type apiSenderOpt func(*awsApiSender)

func WithTemplateStore(store mail.TemplateStore) apiSenderOpt {
	return func(a *awsApiSender) {
		a.tplStore = store
	}
}

func WithAwsSender(sender awsSender) apiSenderOpt {
	return func(a *awsApiSender) {
		a.awsSender = sender
	}
}

func NewApiSender(opts ...apiSenderOpt) (mail.ApiSender, error) {
	sender := &awsApiSender{}
	for _, opt := range opts {
		opt(sender)
	}

	if sender.awsSender == nil {
		sess, err := newAwsSession()
		if err != nil {
			return nil, err
		}
		sesv2 := sesv2.New(sess)
		sender.awsSender = sesv2
		tplStore, err := NewTemplateStore(WithSesSession(sesv2))
		if err != nil {
			return nil, err
		}
		sender.tplStore = tplStore
	}

	if sender.tplStore == nil {
		return nil, errors.New("template store is nil")
	}

	from := viper.GetString("aws.ses.from")
	if from == "" {
		return nil, errors.New("aws.ses.from is empty")
	}
	sender.from = from
	return sender, nil
}

type awsApiSender struct {
	awsSender
	tplStore mail.TemplateStore
	from     string
}

func (a *awsApiSender) Send(notify *service.Notification) (string, error) {
	dataJson, err := json.Marshal(notify.Data)
	if err != nil {
		return "", errors.New("failed to marshal data")
	}
	addresses := make([]*string, len(notify.SendTo))
	for i, s := range notify.SendTo {
		addresses[i] = aws.String(s.Email)
	}
	tplName := notify.GetTemplateName()

	exist, err := a.tplStore.IsTemplateExist(tplName)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", errors.New("template does not exist: " + tplName)
	}
	output, err := a.SendEmail(&sesv2.SendEmailInput{
		Destination: &sesv2.Destination{
			ToAddresses: addresses,
		},
		FromEmailAddress: aws.String(a.from),
		Content: &sesv2.EmailContent{
			Template: &sesv2.Template{
				TemplateName: aws.String(tplName),
				TemplateData: aws.String(string(dataJson)),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}
	if output.MessageId == nil {
		return "", errors.New("failed to send email")
	}
	return *output.MessageId, nil
}
