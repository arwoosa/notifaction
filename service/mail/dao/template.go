package dao

import (
	"fmt"

	"github.com/arwoosa/notifaction/service"
)

type tplContent struct {
	Subject string
	Body    struct {
		Plaint string
		Html   string
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
	// TODO: complete
	fmt.Println("template event:", a.Event)
	fmt.Println("template lang:", a.Lang)
	fmt.Println("subject:", a.Subject)
	fmt.Println("body (HTML):\n", a.Body.Html)
	fmt.Println("body (Plaint):\n", a.Body.Plaint)
	return nil
}
