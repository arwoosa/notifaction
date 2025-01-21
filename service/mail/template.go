package mail

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"gopkg.in/yaml.v2"
)

type Template interface {
	Apply(tplfile string) error
	List(nextToken string) (*listResponse, error)
	Delete(name string) error
	// Detail(name string) (string, string, string, string, error)
}

func NewTemplateWithAWS() (Template, error) {
	sess, err := newAwsSession()
	if err != nil {
		return nil, err
	}
	return &awsTplImpl{
		SESV2: sesv2.New(sess),
	}, nil
}

func newTemplateWithSes(ses *sesv2.SESV2) *awsTplImpl {
	return &awsTplImpl{
		SESV2: ses,
	}
}

type awsTplImpl struct {
	*sesv2.SESV2
}

func (a *awsTplImpl) Apply(file string) error {
	// check file exist
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("file %s does not exist: %w", file, err)
	}

	// read file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", file, err)
	}
	// yaml unmarshal
	var tplDao dao.ApplyTemplateInput
	if err := yaml.Unmarshal(data, &tplDao); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	// validate template dao
	if err := tplDao.Validate(); err != nil {
		return err
	}

	// check template exist
	name := tplDao.GetName()
	exist, err := a.isTemplateExist(name)
	if err != nil {
		return fmt.Errorf("failed to check template exist: %w", err)
	}
	if exist {
		// update template
		return a.updateTemplate(&tplDao.Template)
	}

	// aws create template
	return a.createTpl(&tplDao.Template)
}

func (a *awsTplImpl) createTpl(tpl *dao.Template) error {
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

func (a *awsTplImpl) updateTemplate(tpl *dao.Template) error {
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

func (a *awsTplImpl) isTemplateExist(name string) (bool, error) {
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

type listTemplate struct {
	Name       string
	CreateTime time.Time
	UpdateTime *time.Time
}

type listResponse struct {
	NextToken *string
	Templates []*listTemplate
}

func (a *awsTplImpl) List(token string) (*listResponse, error) {
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
	results := make([]*listTemplate, len(tpls.TemplatesMetadata))
	for i, data := range tpls.TemplatesMetadata {
		results[i] = &listTemplate{
			Name:       *data.TemplateName,
			CreateTime: *data.CreatedTimestamp,
		}
	}
	return &listResponse{
		NextToken: tpls.NextToken,
		Templates: results,
	}, nil
}

func (a *awsTplImpl) Delete(name string) error {
	exist, err := a.isTemplateExist(name)
	if err != nil {
		return fmt.Errorf("failed to check template exist: %w", err)
	}
	if !exist {
		// update template
		return errors.New("template does not exist")
	}
	_, err = a.DeleteEmailTemplate(&sesv2.DeleteEmailTemplateInput{
		TemplateName: aws.String(name),
	})
	return err
}
