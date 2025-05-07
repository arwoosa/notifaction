package aws

import (
	"fmt"

	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sesv2"
)

type awsTemplateStore interface {
	CreateEmailTemplate(*sesv2.CreateEmailTemplateInput) (*sesv2.CreateEmailTemplateOutput, error)
	DeleteEmailTemplate(*sesv2.DeleteEmailTemplateInput) (*sesv2.DeleteEmailTemplateOutput, error)
	ListEmailTemplates(*sesv2.ListEmailTemplatesInput) (*sesv2.ListEmailTemplatesOutput, error)
	UpdateEmailTemplate(*sesv2.UpdateEmailTemplateInput) (*sesv2.UpdateEmailTemplateOutput, error)
	GetEmailTemplate(*sesv2.GetEmailTemplateInput) (*sesv2.GetEmailTemplateOutput, error)
}

type awsTemplateStoreOpt func(*awsTplImpl)

func WithSesSession(ses *sesv2.SESV2) awsTemplateStoreOpt {
	return func(a *awsTplImpl) {
		a.awsTemplateStore = ses
	}
}

func NewTemplateStore(opts ...awsTemplateStoreOpt) (mail.TemplateStore, error) {
	awsTplImpl := &awsTplImpl{}

	for _, opt := range opts {
		opt(awsTplImpl)
	}

	if awsTplImpl.awsTemplateStore == nil {
		sess, err := newAwsSession()
		if err != nil {
			return nil, fmt.Errorf("new aws session fail: %w", err)
		}
		awsTplImpl.awsTemplateStore = sesv2.New(sess)
	}

	return awsTplImpl, nil
}

type awsTplImpl struct {
	awsTemplateStore
}

func (a *awsTplImpl) CreateTpl(tpl *dao.Template) error {
	_, err := a.CreateEmailTemplate(&sesv2.CreateEmailTemplateInput{
		TemplateName: aws.String(tpl.GetName()),
		TemplateContent: &sesv2.EmailTemplateContent{
			Subject: aws.String(tpl.Subject),
			Html:    aws.String(tpl.Body.Html),
			Text:    aws.String(tpl.Body.Plaint),
		},
	})
	return err
}

func (a *awsTplImpl) UpdateTemplate(tpl *dao.Template) error {
	_, err := a.UpdateEmailTemplate(&sesv2.UpdateEmailTemplateInput{
		TemplateName: aws.String(tpl.GetName()),
		TemplateContent: &sesv2.EmailTemplateContent{
			Subject: aws.String(tpl.Subject),
			Html:    aws.String(tpl.Body.Html),
			Text:    aws.String(tpl.Body.Plaint),
		},
	})
	return err
}

func (a *awsTplImpl) IsTemplateExist(name string) (bool, error) {
	tpl, err := a.GetEmailTemplate(&sesv2.GetEmailTemplateInput{
		TemplateName: aws.String(name),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == sesv2.ErrCodeNotFoundException {
				return false, nil
			}
		}
		return false, err
	}
	return (*tpl.TemplateName == name), nil
}

func (a *awsTplImpl) List(token string) (*dao.ListTemplateResponse, error) {
	var nextToken *string
	if token != "" {
		nextToken = aws.String(token)
	}

	tpls, err := a.ListEmailTemplates(&sesv2.ListEmailTemplatesInput{
		NextToken: nextToken,
		PageSize:  aws.Int64(100),
	})
	if err != nil {
		return nil, err
	}
	results := make([]*dao.ListTemplate, len(tpls.TemplatesMetadata))
	for i, data := range tpls.TemplatesMetadata {
		results[i] = &dao.ListTemplate{
			Name:       *data.TemplateName,
			CreateTime: *data.CreatedTimestamp,
		}
	}
	return &dao.ListTemplateResponse{
		NextToken: tpls.NextToken,
		Templates: results,
	}, nil
}

func (a *awsTplImpl) Delete(name string) error {
	_, err := a.DeleteEmailTemplate(&sesv2.DeleteEmailTemplateInput{
		TemplateName: aws.String(name),
	})
	return err
}

func (a *awsTplImpl) Detail(name string) (*dao.DetailTemplateResponse, error) {
	tpl, err := a.GetEmailTemplate(&sesv2.GetEmailTemplateInput{
		TemplateName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	resp := &dao.DetailTemplateResponse{}
	resp.Title = *tpl.TemplateName
	resp.Subject = *tpl.TemplateContent.Subject
	resp.Body.Plaint = *tpl.TemplateContent.Text
	resp.Body.Html = *tpl.TemplateContent.Html

	return resp, nil
}
