package service

import (
	"fmt"
)

type Notification struct {
	Event  string
	Lang   string
	Data   map[string]string
	From   *Info
	SendTo []*Info
}

func (n *Notification) GetTemplateName() string {
	return GetTemplateName(n.Event, n.Lang)
}

func GetTemplateName(event, lang string) string {
	return fmt.Sprintf("%s_%s", event, lang)
}

type Info struct {
	Sub    string
	Name   string
	Email  string
	Enable bool
}

type Sender interface {
	Send(*Notification) (messageId string, err error)
}

type Health interface {
	IsReady() (bool, error)
}
