package httpclient

import (
	"io"
	"net/http"
	"strings"
)

// MockRoundTripper returns pre-configured responses keyed by "METHOD URL".
// Use NewMockResponse to build response values.
type MockRoundTripper struct {
	Responses map[string]*http.Response
	Calls     []string
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()
	m.Calls = append(m.Calls, key)
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}
	return NewMockResponse(404, `{"error":"no mock for `+key+`"}`), nil
}

// NewMockResponse builds a mock *http.Response with the given status code and JSON body.
func NewMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
