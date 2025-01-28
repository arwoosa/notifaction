package mail

import (
	"github.com/arwoosa/notifaction/service"
)

type ApiSender interface {
	service.Sender
}
