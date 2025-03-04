package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/94peter/microservice/apitool"
	apiErr "github.com/94peter/microservice/apitool/err"
	"github.com/arwoosa/notifaction/router/request"
	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/identity"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetApis(t *testing.T) {
	tests := []struct {
		name    string
		want    []apitool.GinAPI
		prefunc func()
	}{
		{
			name: "test GetApis",
			want: []apitool.GinAPI{
				&notification{},
				&health{},
			},
		},
		{
			name: "test GetApis with test api",
			want: []apitool.GinAPI{
				&notification{},
				&health{},
				&test{},
			},
			prefunc: func() {
				viper.Set("api.test", true)
			},
		},
	}

	for _, tt := range tests {
		viper.Reset()
		if tt.prefunc != nil {
			tt.prefunc()
		}
		t.Run(tt.name, func(t *testing.T) {
			got := GetApis()
			if len(got) != len(tt.want) {
				t.Errorf("GetApis() = %v, want %v", got, tt.want)
			}
			for i, api := range got {
				if fmt.Sprintf("%T", api) != fmt.Sprintf("%T", tt.want[i]) {
					t.Errorf("GetApis() = %v, want %v", api, tt.want[i])
				}
			}
		})
	}
}

func TestAliveHandler(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "returns 200 status code",
			statusCode: http.StatusOK,
			body:       `{"message":"I am alive"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			aliveHandler(c)

			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, w.Code)
			}

			if w.Body.String() != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, w.Body.String())
			}
		})
	}
}

func TestGetHealthHandlers(t *testing.T) {
	m := &health{}

	handlers := m.GetHandlers()

	// Test that the function returns two handlers
	if len(handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(handlers))
	}

	// Test that the first handler has the correct path and method
	if handlers[0].Path != "/health/alive" || handlers[0].Method != "GET" {
		t.Errorf("expected first handler to have path '/health/alive' and method 'GET', got path '%s' and method '%s'", handlers[0].Path, handlers[0].Method)
	}

	// Test that the second handler has the correct path and method
	if handlers[1].Path != "/health/ready" || handlers[1].Method != "GET" {
		t.Errorf("expected second handler to have path '/health/ready' and method 'GET', got path '%s' and method '%s'", handlers[1].Path, handlers[1].Method)
	}
}

func TestGetNotificationHandlers(t *testing.T) {
	m := &notification{}

	handlers := m.GetHandlers()

	// Test that the function returns two handlers
	if len(handlers) != 1 {
		t.Errorf("expected 1 handlers, got %d", len(handlers))
	}

	// Test that the first handler has the correct path and method
	if handlers[0].Path != "/notification" || handlers[0].Method != "POST" {
		t.Errorf("expected first handler to have path '/notification' and method 'POST', got path '%s' and method '%s'", handlers[0].Path, handlers[0].Method)
	}
}

func TestReadyHandler(t *testing.T) {
	tests := []struct {
		name                 string
		newIdentityErr       error
		newTemplateErr       error
		healthFunc           func() (bool, error)
		listTplFunc          func(nextToken string) (*dao.ListTemplateResponse, error)
		expectedStatus       int
		expectedResponseBody string
	}{
		{
			name:                 "new identity service not ready with error",
			newIdentityErr:       errors.New("new identity error"),
			expectedStatus:       500,
			expectedResponseBody: `{"error":"identity service is not ready: new identity error"}`,
		},
		{
			name:                 "identity service not ready with error",
			healthFunc:           func() (bool, error) { return false, errors.New("identity error") },
			expectedStatus:       500,
			expectedResponseBody: `{"error":"identity service is not ready: identity error"}`,
		},
		{
			name:                 "identity service not ready without error",
			healthFunc:           func() (bool, error) { return false, nil },
			expectedStatus:       500,
			expectedResponseBody: `{"error":"identity service is not ready"}`,
		},
		{
			name:       "email service not ready with error",
			healthFunc: func() (bool, error) { return true, nil },
			listTplFunc: func(nextToken string) (*dao.ListTemplateResponse, error) {
				return nil, errors.New("email error")
			},
			expectedStatus:       500,
			expectedResponseBody: `{"error":"email service is not ready: email error"}`,
		},
		{
			name:                 "new email service error",
			newTemplateErr:       errors.New("new email error"),
			expectedStatus:       500,
			expectedResponseBody: `{"error":"email service is not ready: new email error"}`,
		},
		{
			name:       "both services ready",
			healthFunc: func() (bool, error) { return true, nil },
			listTplFunc: func(nextToken string) (*dao.ListTemplateResponse, error) {
				return nil, nil
			},
			expectedStatus:       200,
			expectedResponseBody: `{"message":"service notification is ready"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			health := &health{}
			health.SetErrorHandler(func(c *gin.Context, err error) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			})
			identity.SetNewException(test.newIdentityErr)
			identity.SetMockHealthFunc(test.healthFunc)
			factory.SetMockList(test.listTplFunc)
			factory.SetMockNewTemplateException(test.newTemplateErr)
			health.readyHandler(c)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.JSONEq(t, test.expectedResponseBody, w.Body.String())
		})
	}
}

func TestCreateNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                string
		requestBody         *request.CreateNotification
		mockSenderException error
		mockSender          func(t *testing.T, msg *service.Notification) (messageId string, err error)
		mockNewIdentityErr  error
		mockSubToInfo       func(from string, to []string) (*identity.ClassificationLang, error)
		statusCode          int
	}{
		{
			name:        "bind error",
			requestBody: nil,
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "invalid request body",
			requestBody: &request.CreateNotification{
				To: []string{"invalid"},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "NewApiSender error",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockSenderException: errors.New("new sender error"),
			statusCode:          http.StatusInternalServerError,
		},
		{
			name: "NewIdentity error",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockNewIdentityErr: errors.New("new identity error"),
			statusCode:         http.StatusInternalServerError,
		},
		{
			name: "SubToInfo error",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return nil, errors.New("sub to info error")
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "send error",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": {{Sub: "valid"}}}),
				), nil
			},
			mockSender: func(t *testing.T, msg *service.Notification) (messageId string, err error) {
				return "", errors.New("send error")
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "partial send error",
			requestBody: &request.CreateNotification{
				To:    []string{"valid", "fail"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				tos := make([]*service.Info, len(to))
				for i, t := range to {
					tos[i] = &service.Info{Sub: t}
				}
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": tos}),
				), nil
			},
			mockSender: func(t *testing.T, msg *service.Notification) (messageId string, err error) {
				if msg.SendTo[0].Sub == "fail" {
					return "", errors.New("send error")
				}
				return "", nil
			},
			statusCode: http.StatusPartialContent,
		},
		{
			name: "successful send",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": {{Sub: "valid"}}}),
				), nil
			},
			statusCode: http.StatusAccepted,
		},
	}

	for _, test := range tests {
		defer func() {
			identity.ResetMock()
			factory.ResetMockSender()
			factory.ResetMockTemplate()
		}()
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var requestData *bytes.Buffer
			if test.requestBody != nil {
				data, _ := json.Marshal(test.requestBody)
				requestData = bytes.NewBuffer(data)
			} else {
				requestData = bytes.NewBuffer([]byte{})
			}

			c.Request, _ = http.NewRequest("POST", "/notification", requestData)
			c.Request.Header.Set("Content-Type", "application/json")

			factory.SetMockSender(test.mockSender)
			factory.SetMockNewSenderException(test.mockSenderException)

			identity.SetNewException(test.mockNewIdentityErr)
			identity.SetMockSubToInfoFunc(test.mockSubToInfo)

			notification := &notification{}
			notification.SetErrorHandler(func(c *gin.Context, err error) {
				if apiErr, ok := err.(apiErr.ApiError); ok {
					c.JSON(apiErr.GetStatus(), gin.H{
						"error": apiErr.Error(),
					})
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			})
			notification.createNotification(c)

			assert.Equal(t, test.statusCode, w.Code)
		})
	}
}

func TestGetHandlers(t *testing.T) {
	m := &test{}

	handlers := m.GetHandlers()

	// Test that the function returns a non-empty slice of handlers
	if len(handlers) == 0 {
		t.Errorf("expected at least one handler, got 0")
	}

	// Test that the returned handler has the correct path, method, and handler function
	if handlers[0].Path != "/test/header2post" || handlers[0].Method != "POST" {
		t.Errorf("expected handler with path '/test/header2post', method 'POST', and handler 'testHeader2Post', got %+v", handlers[0])
	}
}

func TestTestHeader2Post(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
		header     string
	}{
		{
			name:       "success",
			body:       "Hello, World!",
			statusCode: http.StatusOK,
			header:     "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:       "invalid body",
			body:       "",
			statusCode: http.StatusInternalServerError,
			header:     "",
		},
	}
	const testPath = "/test/header2post"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.body != "" {
				c.Request = httptest.NewRequest("POST", testPath, strings.NewReader(tt.body))
			} else {
				c.Request = httptest.NewRequest("POST", testPath, errReader(0))
			}

			m := &test{}
			m.SetErrorHandler(func(c *gin.Context, err error) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			})
			m.testHeader2Post(c)

			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, w.Code)
			}

			if tt.header != "" {
				if headerValue := w.Header().Get("X-Notify"); headerValue != tt.header {
					t.Errorf("expected header value %q, got %q", tt.header, headerValue)
				}
			}
		})
	}
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestCreateNotificationWithForwardedHeaders(t *testing.T) {

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestBody      *request.CreateNotification
		header           http.Header
		forwardedHeaders string
		mockSubToInfo    func(from string, to []string) (*identity.ClassificationLang, error)
		mockSender       func(t *testing.T, msg *service.Notification) (messageId string, err error)
		statusCode       int
	}{
		{
			name: "not set header forward",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			header: http.Header{
				"X-Forwarded-Host": []string{"localhost"},
			},
			forwardedHeaders: "",
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": {{Sub: "valid"}}}),
				), nil
			},
			mockSender: func(t *testing.T, msg *service.Notification) (messageId string, err error) {
				assert.Equal(t, "", msg.Data["X-Forwarded-Host"])
				return "", nil
			},
			statusCode: http.StatusAccepted,
		},
		{
			name: "header not exist",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			forwardedHeaders: "X-Forwarded-Not-Exist",
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": {{Sub: "valid"}}}),
				), nil
			},
			mockSender: func(t *testing.T, msg *service.Notification) (messageId string, err error) {
				assert.Equal(t, "missing header: X-Forwarded-Not-Exist", msg.Data["X-Forwarded-Not-Exist"])
				return "", nil
			},
			statusCode: http.StatusAccepted,
		},
		{
			name: "header exist",
			requestBody: &request.CreateNotification{
				To:    []string{"valid"},
				From:  "fff",
				Event: "event",
				Data:  map[string]string{},
			},
			header: http.Header{
				"X-Forwarded-Host": []string{"localhost"},
			},
			forwardedHeaders: "X-Forwarded-Host",
			mockSubToInfo: func(from string, to []string) (*identity.ClassificationLang, error) {
				return identity.NewClassificationLang(
					identity.WithClassificationLangKeys([]string{"en"}),
					identity.WithClassificationLangFrom(&service.Info{Sub: "valid"}),
					identity.WithClassificationLangFromLang("en"),
					identity.WithClassificationLang(map[string][]*service.Info{"en": {{Sub: "valid"}}}),
				), nil
			},
			mockSender: func(t *testing.T, msg *service.Notification) (messageId string, err error) {
				assert.Equal(t, "localhost", msg.Data["X-Forwarded-Host"])
				return "", nil
			},
			statusCode: http.StatusAccepted,
		},
	}
	for _, test := range tests {
		defer func() {
			identity.ResetMock()
			factory.ResetMockSender()
			factory.ResetMockTemplate()
		}()
		viper.Set("mail.header2data", test.forwardedHeaders)
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			var requestData *bytes.Buffer
			if test.requestBody != nil {
				data, _ := json.Marshal(test.requestBody)
				requestData = bytes.NewBuffer(data)
			} else {
				requestData = bytes.NewBuffer([]byte{})
			}

			c.Request, _ = http.NewRequest("POST", "/notification", requestData)
			c.Request.Header.Set("Content-Type", "application/json")
			for k, v := range test.header {
				c.Request.Header.Set(k, v[0])
			}
			factory.SetMockSender(test.mockSender, factory.WithMockSenderT(t))
			identity.SetMockSubToInfoFunc(test.mockSubToInfo)

			notification := newNotification()
			notification.SetErrorHandler(func(c *gin.Context, err error) {
				if apiErr, ok := err.(apiErr.ApiError); ok {
					c.JSON(apiErr.GetStatus(), gin.H{
						"error": apiErr.Error(),
					})
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			})
			notification.createNotification(c)

			assert.Equal(t, test.statusCode, w.Code)
		})
	}
}
