package mail

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/arwoosa/notifaction/service"
	"github.com/go-gomail/gomail"

	"github.com/spf13/viper"
)

type Server interface {
	service.Sender
	Run(ctx context.Context)
}

type bodyOpt func(*gomail.Message)

func WithHtml(body string) bodyOpt {
	return func(m *gomail.Message) {
		m.SetBody("text/html", body)
	}
}

func WithText(body string) bodyOpt {
	return func(m *gomail.Message) {
		m.SetBody("text/plain", body)
	}
}

type mail struct {
	ch chan *gomail.Message

	host     string
	user     string
	password string
	port     int
}

func NewServer() (Server, error) {
	smtpUrl := viper.GetString("mail.url")
	if smtpUrl == "" {
		return nil, errors.New("mail.url is empty")
	}

	fmt.Println(smtpUrl)

	myurl, err := url.Parse(smtpUrl)
	if err != nil {
		fmt.Println("dddddd")
		return nil, err
	}

	fmt.Println(*myurl)
	if myurl.Scheme != "smtp" {
		return nil, errors.New("invalid smtp url")
	}
	if myurl.Host == "" {
		return nil, errors.New("empty host")
	}

	portNum, err := strconv.Atoi(myurl.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	pass, _ := myurl.User.Password()

	return &mail{
		ch:       make(chan *gomail.Message),
		host:     myurl.Hostname(),
		user:     myurl.User.Username(),
		password: pass,
		port:     portNum,
	}, nil
}

func (m *mail) Send(notify *service.Notification) (string, error) {

	// msg := gomail.NewMessage()
	// msg.SetHeader("From", "developer@arwork.tw")
	// msg.SetHeader("To", to)
	// // msg.SetHeader("To", "bob@example.com", "cora@example.com")
	// // msg.SetAddressHeader("Cc", "dan@example.com", "Dan")
	// msg.SetHeader("Subject", subject)

	// for _, body := range bodies {
	// 	body(msg)
	// }
	// m.ch <- msg
	return "", nil
}

func (m *mail) Run(ctx context.Context) {
	go func(host, user, password string, port int) {
		d := gomail.NewDialer(host, port, user, password)

		var s gomail.SendCloser
		var err error
		open := false
		for {
			select {
			case msg, ok := <-m.ch:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}
				if err := gomail.Send(s, msg); err != nil {
					log.Print(err)
				}
			// Close the connection to the SMTP server if no email was sent in
			// the last 30 seconds.
			case <-time.After(30 * time.Second):
				if open {
					if err := s.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}(m.host, m.user, m.password, m.port)
	<-ctx.Done()
	// Use the channel in your program to send emails.

	// Close the channel to stop the mail daemon.
	close(m.ch)
}
