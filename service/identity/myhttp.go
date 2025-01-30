package identity

import "net/http"

type myHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func newMockHttpClient(interDo func(req *http.Request) (*http.Response, error)) myHttpClient {
	return &mockHttpClient{
		interDo: interDo,
	}
}

type mockHttpClient struct {
	interDo func(req *http.Request) (*http.Response, error)
}

func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.interDo(req)
}
