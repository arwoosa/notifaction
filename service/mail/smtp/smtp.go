package smtp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
	"github.com/go-gomail/gomail"
)

type apiSenderOpt func(*smtp)

func WithTemplate(tpl mail.Template) apiSenderOpt {
	return func(s *smtp) {
		s.tpl = tpl
	}
}

func ParseUrl(myurl string) (gomail.SendCloser, error) {
	parsedUrl, err := url.Parse(myurl)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp url: %w", err)
	}
	portNum, err := strconv.Atoi(parsedUrl.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	pass, _ := parsedUrl.User.Password()

	dialer := gomail.NewDialer(
		parsedUrl.Hostname(),
		portNum,
		parsedUrl.User.Username(),
		pass,
	)
	sendCloser, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed to dial smtp: %w", err)
	}
	return sendCloser, nil
}

func WithSendCloser(sendCloser gomail.SendCloser) apiSenderOpt {
	return func(s *smtp) {
		s.sendCloser = sendCloser
	}
}

func WithFrom(from string) apiSenderOpt {
	return func(s *smtp) {
		s.from = from
	}
}

func NewApiSender(opts ...apiSenderOpt) (mail.ApiSender, error) {
	s := &smtp{}
	for _, opt := range opts {
		opt(s)
	}
	if s.sendCloser == nil {
		return nil, fmt.Errorf("sendCloser is required")
	}
	if s.from == "" {
		return nil, fmt.Errorf("from is required")
	}
	if s.tpl == nil {
		return nil, fmt.Errorf("template is required")
	}
	return s, nil
}

type smtp struct {
	tpl        mail.Template
	from       string
	sendCloser gomail.SendCloser
}

func (s *smtp) Send(notify *service.Notification) (string, error) {
	// Validate recipients
	if len(notify.SendTo) == 0 {
		return "", fmt.Errorf("no recipients specified")
	}

	// Create a new message
	msg := gomail.NewMessage()

	// Set sender
	msg.SetHeader("From", s.from)

	// Set recipients
	for _, to := range notify.SendTo {
		msg.SetHeader("To", to.Email)
	}

	// Get template name
	tplName := notify.GetTemplateName()

	// Get template content
	tplDetail, err := s.tpl.Detail(tplName)
	if err != nil {
		return "", fmt.Errorf("failed to get template detail: %w", err)
	}

	// replace template variables {{Variable}}
	for k, v := range notify.UpperKeyData() {
		tplDetail.Body.Html = strings.ReplaceAll(tplDetail.Body.Html, "{{"+k+"}}", v)
		tplDetail.Body.Plaint = strings.ReplaceAll(tplDetail.Body.Plaint, "{{"+k+"}}", v)
		tplDetail.Subject = strings.ReplaceAll(tplDetail.Subject, "{{"+k+"}}", v)
	}

	// Set subject
	msg.SetHeader("Subject", tplDetail.Subject)

	// Set body content
	msg.SetBody("text/html", tplDetail.Body.Html)
	msg.SetBody("text/plain", tplDetail.Body.Plaint)
	defer s.sendCloser.Close()
	if err := gomail.Send(s.sendCloser, msg); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return "", nil
}
