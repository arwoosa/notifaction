package identity

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/arwoosa/notifaction/service"
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
		{
			name: "invalid url",
			httpClient: &mockHttpClient{
				interDo: func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusInternalServerError, Body: http.NoBody}, nil
				},
			},
			heathUri:       "http://dddd 	fdadf",
			expectedResult: false,
			expectedError:  errors.New(`parse "http://dddd \tfdadf/admin/health/ready": net/url: invalid control character in URL`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("identity.url", test.heathUri)
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
			if test.expectedError == nil && err != nil {
				t.Errorf("expected error is nil, but get err: %v", err)
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

func TestSubToInfo(t *testing.T) {
	viper.Reset()
	tests := []struct {
		name         string
		url          string
		from         string
		to           []string
		opts         []option
		want         *ClassificationLang
		wantErr      bool
		errorContain string
	}{
		{
			name:    "empty to slice",
			url:     "https://example.com",
			from:    "from",
			to:      []string{},
			wantErr: false,
			want:    nil,
		},
		{
			name: "non-empty to slice with valid data",
			url:  "https://example.com",
			from: "3",
			to:   []string{"1", "2"},
			opts: []option{
				WithHttpClient(newMockHttpClient(func(req *http.Request) (*http.Response, error) {
					stringReader := strings.NewReader(`[
						{"id": "1", "state": "active", "traits": {
							"name": "To1 Name",
							"email": "to1@example.com",
							"language": "en"
						}},
						{"id": "2", "state": "active", "traits": {
							"name": "To2 Name",
							"email": "to2@example.com",
							"language": "en"
						}},
						{"id": "3", "state": "active", "traits": {
							"name": "from3 Name",
							"email": "frome3@example.com",
							"language": "en"
						}}
						]`)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(stringReader),
					}, nil
				}))},
			want: &ClassificationLang{
				keys: []string{"en"},
				data: map[string][]*service.Info{
					"en": {
						{
							Sub:    "1",
							Name:   "To1 Name",
							Email:  "to1@example.com",
							Enable: true,
						},
						{
							Sub:    "2",
							Name:   "To2 Name",
							Email:  "to2@example.com",
							Enable: true,
						},
					},
				},
				From: &service.Info{
					Sub:    "3",
					Name:   "from3 Name",
					Email:  "frome3@example.com",
					Enable: true,
				},
				FromLang: "en",
			},
			wantErr: false,
		},
		{
			name: "no body response",
			url:  "https://example.com",
			from: "3",
			to:   []string{"1", "2"},
			opts: []option{
				WithHttpClient(newMockHttpClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       http.NoBody,
					}, nil
				}))},
			want:         nil,
			wantErr:      true,
			errorContain: "failed to decode response",
		},
		{
			name: "do with error",
			url:  "https://example.com",
			from: "3",
			to:   []string{"1", "2"},
			opts: []option{
				WithHttpClient(newMockHttpClient(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("any error")
				}))},
			want:         nil,
			wantErr:      true,
			errorContain: "failed to send request",
		},
		{
			name: "invalid url",
			url:  "https://exam	ple.com",
			from: "3",
			to:   []string{"1", "2"},
			opts: []option{
				WithHttpClient(newMockHttpClient(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("any error")
				}))},
			want:         nil,
			wantErr:      true,
			errorContain: "failed to create request",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			viper.Set("identity.url", test.url)
			i, err := NewIdentity(test.opts...)
			if err != nil {
				t.Errorf("NewIdentity() error = %v", err)
				return
			}
			classLang, err := i.SubToInfo(test.from, test.to)
			if (err != nil) != test.wantErr {
				t.Errorf("SubToInfo() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr && test.errorContain != "" && !strings.Contains(err.Error(), test.errorContain) {
				t.Errorf("errors should contain: %s", test.errorContain)
				return
			}
			if classLang == nil {
				return
			}
			if !test.want.isEqual(classLang) {
				t.Errorf("SubToInfo() got = %v, want %v", classLang, test.want)
			}
		})
	}
}

func TestIsEqual(t *testing.T) {
	tests := []struct {
		name     string
		c        *ClassificationLang
		cc       *ClassificationLang
		expected bool
	}{
		{
			name: "identical",
			c: &ClassificationLang{
				keys:     []string{"key1", "key2"},
				data:     map[string][]*service.Info{"key1": {{Sub: "sub1"}}, "key2": {{Sub: "sub2"}}},
				From:     &service.Info{Sub: "from"},
				FromLang: "lang",
			},
			cc: &ClassificationLang{
				keys:     []string{"key1", "key2"},
				data:     map[string][]*service.Info{"key1": {{Sub: "sub1"}}, "key2": {{Sub: "sub2"}}},
				From:     &service.Info{Sub: "from"},
				FromLang: "lang",
			},
			expected: true,
		},
		{
			name: "different lengths",
			c: &ClassificationLang{
				keys: []string{"key1", "key2"},
			},
			cc: &ClassificationLang{
				keys: []string{"key1"},
			},
			expected: false,
		},
		{
			name: "missing key in cc.data",
			c: &ClassificationLang{
				keys: []string{"key1", "key2"},
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}},
			},
			cc: &ClassificationLang{
				keys: []string{"key1", "key2"},
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}},
			},
			expected: false,
		},
		{
			name: "different FromLang",
			c: &ClassificationLang{
				FromLang: "lang1",
			},
			cc: &ClassificationLang{
				FromLang: "lang2",
			},
			expected: false,
		},
		{
			name: "different From",
			c: &ClassificationLang{
				From: &service.Info{Sub: "from1"},
			},
			cc: &ClassificationLang{
				From: &service.Info{Sub: "from2"},
			},
			expected: false,
		},
		{
			name: "different data lengths",
			c: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}},
			},
			cc: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}, {Sub: "sub2"}}},
			},
			expected: false,
		},
		{
			name: "different map lengths",
			c: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}},
			},
			cc: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}, "key2": {{Sub: "sub2"}}},
			},
			expected: false,
		},
		{
			name: "different data values",
			c: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub1"}}},
			},
			cc: &ClassificationLang{
				data: map[string][]*service.Info{"key1": {{Sub: "sub2"}}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.isEqual(tt.cc); got != tt.expected {
				t.Errorf("isEqual() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetLangs(t *testing.T) {
	tests := []struct {
		name     string
		cl       *ClassificationLang
		expected []string
	}{
		{
			name:     "empty keys",
			cl:       &ClassificationLang{},
			expected: []string{},
		},
		{
			name:     "single key",
			cl:       &ClassificationLang{keys: []string{"lang1"}},
			expected: []string{"lang1"},
		},
		{
			name:     "multiple keys",
			cl:       &ClassificationLang{keys: []string{"lang1", "lang2", "lang3"}},
			expected: []string{"lang1", "lang2", "lang3"},
		},
		{
			name:     "nil ClassificationLang",
			cl:       nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.cl.GetLangs()
			if !slices.Equal(actual, tt.expected) {
				t.Errorf("GetLangs() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestGetInfos(t *testing.T) {
	tests := []struct {
		name     string
		cl       *ClassificationLang
		lang     string
		expected []*service.Info
	}{
		{
			name: "existing language",
			cl: &ClassificationLang{
				data: map[string][]*service.Info{
					"en": {{Sub: "sub1"}, {Sub: "sub2"}},
				},
			},
			lang: "en",
			expected: []*service.Info{
				{Sub: "sub1"},
				{Sub: "sub2"},
			},
		},
		{
			name: "non-existing language",
			cl: &ClassificationLang{
				data: map[string][]*service.Info{
					"en": {{Sub: "sub1"}, {Sub: "sub2"}},
				},
			},
			lang:     "fr",
			expected: nil,
		},
		{
			name:     "nil ClassificationLang",
			cl:       nil,
			lang:     "en",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.cl.GetInfos(tt.lang)
			if len(actual) != len(tt.expected) {
				t.Errorf("GetInfos() = %v, want %v", actual, tt.expected)
				return
			}
			for i, v := range tt.expected {
				if !reflect.DeepEqual(actual[i], v) {
					t.Errorf("GetInfos() = %v, want %v", actual, tt.expected)
					return
				}
			}
		})
	}
}
