package mail

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/arwoosa/notifaction/service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/spf13/viper"
)

type ApiSender interface {
	service.Sender
}

func NewApiSenderWithAws() (ApiSender, error) {
	sess, err := newAwsSession()
	if err != nil {
		return nil, err
	}
	from := viper.GetString("aws.ses.from")
	if from == "" {
		return nil, errors.New("aws.ses.from is empty")
	}
	return &awsApiSender{
		SESV2: sesv2.New(sess),
		from:  from,
	}, nil
}

type awsApiSender struct {
	*sesv2.SESV2
	from string
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
	tpl := newTemplateWithSes(a.SESV2)
	exist, err := tpl.isTemplateExist(tplName)
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
