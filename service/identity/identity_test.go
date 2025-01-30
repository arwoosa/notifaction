package identity

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestIsReady(t *testing.T) {
	tests := []struct {
		name           string
		httpClient     myHttpClient
		heathUri       string
		expectedResult bool
		expectedError  error
	}{
		{
			name: "Successful GET request with 200 status code",
			httpClient: newMockHttpClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			}),
			heathUri:       "https://example.com/health",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "Failed GET request with error",
			httpClient: newMockHttpClient(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("mock error")
			}),
			heathUri:       "https://example.com/health",
			expectedResult: false,
			expectedError:  errors.New("mock error"),
		},
		{
			name: "GET request with non-200 status code",
			httpClient: &mockHttpClient{
				interDo: func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusInternalServerError, Body: http.NoBody}, nil
				},
			},
			heathUri:       "https://example.com/health",
			expectedResult: false,
			expectedError:  nil,
		},
	}
	viper.Set("identity.url", "https://example.com")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := NewIdentity(WithHttpClient(test.httpClient))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			result, err := i.IsReady()
			if result != test.expectedResult {
				t.Errorf("expected result %v, got %v", test.expectedResult, result)
			}
			if test.expectedError != nil && err.Error() != test.expectedError.Error() {
				t.Errorf("expected error %v, got %v", test.expectedError, err)
			}
		})
	}
}

func TestNewIdentity(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		opts        []option
		wantErr     bool
		wantTimeout time.Duration
	}{
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
		{
			name:        "valid url",
			url:         "https://example.com",
			wantTimeout: time.Second * 5,
		},
		{
			name:        "valid url with WithHttpClient option",
			url:         "https://example.com",
			opts:        []option{WithHttpClient(&http.Client{Timeout: time.Second * 10})},
			wantTimeout: time.Second * 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("identity.url", test.url)
			api, err := NewIdentity(test.opts...)
			if (err != nil) != test.wantErr {
				t.Errorf("NewIdentity() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && api == nil {
				t.Errorf("api should not be nil")
			}
			if test.wantErr {
				return
			}
			// api != nil check
			if !strings.HasPrefix(api.(*identityApi).heathUri, test.url) {
				t.Errorf("heathUri should start with %s", test.url)
			}
			if !strings.HasPrefix(api.(*identityApi).identityUri, test.url) {
				t.Errorf("identityUri should start with %s", test.url)
			}
			if api.(*identityApi).httpClient.(*http.Client).Timeout != test.wantTimeout {
				t.Errorf("timeout should be %s, got %s", test.wantTimeout, api.(*identityApi).httpClient.(*http.Client).Timeout)
			}
		})
	}
}
