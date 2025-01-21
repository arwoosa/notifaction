package factory

import (
	"errors"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
)

const (
	Sender_MAIL_AWS = "MAIL_AWS"
)

func NewSender(typ string) (service.Sender, error) {
	if typ == Sender_MAIL_AWS {
		return mail.NewApiSenderWithAws()
	}
	return nil, errors.New("invalid sender type")
}
