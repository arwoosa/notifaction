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

func WithUrl(url string) apiSenderOpt {
	return func(s *smtp) {
		s.url = url
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
	parsedUrl, err := url.Parse(s.url)
	if err != nil {
		return nil, fmt.Errorf("invalid smtp url: %w", err)
	}
	portNum, err := strconv.Atoi(parsedUrl.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	pass, _ := parsedUrl.User.Password()
	s.dialer = gomail.NewDialer(
		parsedUrl.Hostname(),
		portNum,
		parsedUrl.User.Username(),
		pass,
	)

	return s, nil
}

type smtp struct {
	tpl    mail.Template
	url    string
	from   string
	dialer *gomail.Dialer
}

func (s *smtp) Send(notify *service.Notification) (string, error) {
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

	// Send the email
	if err := s.dialer.DialAndSend(msg); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return "", nil
}
