package smtp

import (
	"fmt"
	"io"
	"testing"

	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/go-gomail/gomail"
	"github.com/stretchr/testify/assert"
)

// MockTemplate is a mock implementation of the Template interface
type MockTemplate struct {
	mail.Template
	DetailFunc func(name string) (*dao.DetailTemplateResponse, error)
}

func (m *MockTemplate) Detail(name string) (*dao.DetailTemplateResponse, error) {
	if m.DetailFunc != nil {
		return m.DetailFunc(name)
	}
	return &dao.DetailTemplateResponse{
		Title:   "Test Title",
		Subject: "Test Subject",
		Body: struct {
			Plaint string
			Html   string
		}{
			Plaint: "Test Plain",
			Html:   "<p>Test HTML</p>",
		},
	}, nil
}

// MockDialer is a mock implementation of gomail.Dialer

func newMockSendCloser() gomail.SendCloser {
	return &mockSendCloser{}
}

type mockSendCloser struct {
	sendErr error
}

func (m *mockSendCloser) Close() error {
	return nil
}

func (m *mockSendCloser) Send(string, []string, io.WriterTo) error {
	return m.sendErr
}

func TestSend(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*smtp)
		notification *service.Notification
		wantErr      bool
		wantMessage  string
	}{
		{
			name: "successful send",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{}
				s.sendCloser = newMockSendCloser()
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event: "test_template",
				Lang:  "en",
				SendTo: []*service.Info{{
					Email: "test@example.com",
				}},
				Data: map[string]string{
					"NAME": "Test",
				},
			},
			wantErr:     false,
			wantMessage: "",
		},
		{
			name: "template detail error",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{
					DetailFunc: func(name string) (*dao.DetailTemplateResponse, error) {
						return nil, fmt.Errorf("template not found")
					},
				}
				s.sendCloser = newMockSendCloser()
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event: "test_template",
				Lang:  "en",
				SendTo: []*service.Info{{
					Email: "test@example.com",
				}},
			},
			wantErr:     true,
			wantMessage: "",
		},
		{
			name: "send error",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{}
				s.sendCloser = &mockSendCloser{
					sendErr: fmt.Errorf("send failed"),
				}
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event: "test_template",
				Lang:  "en",
				SendTo: []*service.Info{{
					Email: "test@example.com",
				}},
				Data: map[string]string{
					"NAME": "Test",
				},
			},
			wantErr:     true,
			wantMessage: "",
		},
		{
			name: "template variable replacement",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{
					DetailFunc: func(name string) (*dao.DetailTemplateResponse, error) {
						return &dao.DetailTemplateResponse{
							Subject: "Hello {{NAME}}",
							Body: struct {
								Plaint string
								Html   string
							}{
								Plaint: "Hello {{NAME}}",
								Html:   "<p>Hello {{NAME}}</p>",
							},
						}, nil
					},
				}
				s.sendCloser = newMockSendCloser()
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event: "test_template",
				Lang:  "en",
				SendTo: []*service.Info{{
					Email: "test@example.com",
				}},
				Data: map[string]string{
					"NAME": "John",
				},
			},
			wantErr:     false,
			wantMessage: "",
		},
		{
			name: "empty recipients",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{}
				s.sendCloser = newMockSendCloser()
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event:  "test_template",
				Lang:   "en",
				SendTo: []*service.Info{},
				Data: map[string]string{
					"NAME": "Test",
				},
			},
			wantErr:     true,
			wantMessage: "",
		},
		{
			name: "nil template",
			setup: func(s *smtp) {
				s.tpl = &MockTemplate{
					DetailFunc: func(name string) (*dao.DetailTemplateResponse, error) {
						return nil, fmt.Errorf("template not found")
					},
				}
				s.sendCloser = newMockSendCloser()
				s.from = "test@example.com"
			},
			notification: &service.Notification{
				Event: "test_template",
				Lang:  "en",
				SendTo: []*service.Info{{
					Email: "test@example.com",
				}},
			},
			wantErr:     true,
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &smtp{}
			tt.setup(s)
			msg, err := s.Send(tt.notification)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMessage, msg)
			}
		})
	}
}

func TestParseUrl(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid URL format",
			url:     "invalid-url",
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name:    "missing port",
			url:     "smtp://user:pass@smtp.example.com",
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name:    "invalid port number",
			url:     "smtp://user:pass@smtp.example.com:abc",
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name:    "missing hostname",
			url:     "smtp://user:pass@:587",
			wantErr: true,
			errMsg:  "failed to dial smtp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseUrl(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewApiSender(t *testing.T) {
	tests := []struct {
		name       string
		sendCloser gomail.SendCloser
		tpl        mail.Template
		from       string
		wantErr    bool
	}{
		{
			name:       "valid sendCloser",
			from:       "test@example.com",
			sendCloser: newMockSendCloser(),
			tpl:        &MockTemplate{},
			wantErr:    false,
		},
		{
			name:       "nil sendCloser",
			from:       "test@example.com",
			tpl:        &MockTemplate{},
			sendCloser: nil,
			wantErr:    true,
		},
		{
			name:       "invalid from",
			from:       "",
			tpl:        &MockTemplate{},
			sendCloser: newMockSendCloser(),
			wantErr:    true,
		},

		{
			name:       "nil template",
			from:       "test@example.com",
			sendCloser: newMockSendCloser(),
			tpl:        nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewApiSender(
				WithTemplate(tt.tpl),
				WithSendCloser(tt.sendCloser),
				WithFrom(tt.from),
			)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, sender)
		})
	}
}
