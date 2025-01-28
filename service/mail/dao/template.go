package dao

import (
	"fmt"
	"time"

	"github.com/arwoosa/notifaction/service"
)

type tplContent struct {
	Subject string
	Body    struct {
		Plaint string
		Html   string
	}
}

func NewTemplate(event, lang, subject, bodyPlaint, bodyHtml string) *Template {
	return &Template{
		Event:      event,
		Lang:       lang,
		tplContent: tplContent{Subject: subject, Body: struct{ Plaint, Html string }{Plaint: bodyPlaint, Html: bodyHtml}},
	}
}

type Template struct {
	Event      string
	Lang       string
	tplContent `yaml:",inline"`
}

func (t *Template) GetName() string {
	return service.GetTemplateName(t.Event, t.Lang)
}

type ApplyTemplateInput struct {
	Template `yaml:",inline"`
}

func (a *ApplyTemplateInput) Validate() error {
	if a.Event == "" {
		return fmt.Errorf("event is required")
	}

	if a.Lang == "" {
		return fmt.Errorf("lang is required")
	}

	if a.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if a.Body.Plaint == "" && a.Body.Html == "" {
		return fmt.Errorf("body.plaint or body.html is required")
	}
	return nil
}

type ListTemplate struct {
	Name       string
	CreateTime time.Time
	UpdateTime *time.Time
}

type ListTemplateResponse struct {
	NextToken *string
	Templates []*ListTemplate
}
