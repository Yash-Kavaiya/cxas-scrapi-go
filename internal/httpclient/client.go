// Package httpclient provides an instrumented HTTP client for the CXAS API.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	retryhttp "github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
)

const (
	// DefaultTimeout is applied to every request.
	DefaultTimeout = 60 * time.Second
	// SDKVersion is injected into the User-Agent header.
	SDKVersion = "1.0.0"
	// APIVersion is the CES REST API version prefix.
	APIVersion = "v1beta"
	// BaseURL is the default CES API base URL.
	BaseURL = "https://ces.googleapis.com"
)

// Options configures the HTTP client.
type Options struct {
	TokenSource  oauth2.TokenSource
	QuotaProject string
	Timeout      time.Duration
	MaxRetries   int
}

// New builds an authenticated, retrying *http.Client.
func New(opts Options) *http.Client {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	rc := retryhttp.NewClient()
	rc.RetryMax = maxRetries
	rc.Logger = nil
	base := rc.StandardClient()
	base.Timeout = timeout

	if opts.TokenSource != nil {
		base.Transport = &oauth2.Transport{
			Source: opts.TokenSource,
			Base:   &uaTransport{base: http.DefaultTransport, quotaProject: opts.QuotaProject},
		}
	} else {
		base.Transport = &uaTransport{base: http.DefaultTransport, quotaProject: opts.QuotaProject}
	}
	return base
}

type uaTransport struct {
	base         http.RoundTripper
	quotaProject string
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", "cxas-go/"+SDKVersion)
	if t.quotaProject != "" {
		req.Header.Set("x-goog-user-project", t.quotaProject)
	}
	return t.base.RoundTrip(req)
}

// APIError wraps a non-2xx response from the CES API.
type APIError struct {
	StatusCode int
	Body       string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("CES API error %d at %s: %s", e.StatusCode, e.URL, e.Body)
}

// IsNotFound returns true if err is an APIError with status 404.
func IsNotFound(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 404
}

// IsPermissionDenied returns true if err is an APIError with status 403.
func IsPermissionDenied(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.StatusCode == 403
}

// DoJSON performs an HTTP request, JSON-marshalling body if non-nil,
// and JSON-unmarshalling the response into out if non-nil.
// Returns *APIError for non-2xx responses.
func DoJSON(ctx context.Context, client *http.Client, method, url string, body, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(rawBody), URL: url}
	}
	if out != nil && len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, out); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}
