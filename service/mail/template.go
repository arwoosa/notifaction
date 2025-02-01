package mail

import "github.com/arwoosa/notifaction/service/mail/dao"

type Template interface {
	Apply(tplfile string) error
	List(nextToken string) (*dao.ListTemplateResponse, error)
	Delete(name string) error
	// Detail(name string) (string, string, string, string, error)
}

type TemplateStore interface {
	IsTemplateExist(name string) (bool, error)
	UpdateTemplate(tpl *dao.Template) error
	CreateTpl(tpl *dao.Template) error
	Delete(name string) error
	List(token string) (*dao.ListTemplateResponse, error)
}
