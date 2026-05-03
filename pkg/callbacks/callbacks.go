// Package callbacks provides a client for CXAS callback resource management.
package callbacks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/common"
)

const apiVersion = "v1beta"

// ErrNotSupported is returned for operations that cannot be ported from Python (e.g., exec-based callback execution).
var ErrNotSupported = errors.New("not supported in Go SDK: use external callback testing")

// Callback represents a CXAS callback resource.
type Callback struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Handler     map[string]interface{} `json:"handler,omitempty"`
}

// Client provides CRUD access to CXAS Callbacks.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Callbacks client scoped to a specific App.
func NewClient(ctx context.Context, appName string, cfg auth.Config, opts ...clientOption) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		appName: appName,
		http:    httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL: httpclient.BaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s/%s/%s", c.baseURL, apiVersion, path)
}

// ListCallbacks returns all callbacks in the App.
func (c *Client) ListCallbacks(ctx context.Context) ([]Callback, error) {
	return common.Paginate(func(token string) ([]Callback, string, error) {
		u := c.url(c.appName + "/callbacks")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Callbacks     []Callback `json:"callbacks"`
			NextPageToken string     `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Callbacks, resp.NextPageToken, nil
	})
}

// GetCallback fetches a single callback by resource name.
func (c *Client) GetCallback(ctx context.Context, name string) (*Callback, error) {
	var cb Callback
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &cb); err != nil {
		return nil, err
	}
	return &cb, nil
}

// CreateCallback creates a new callback.
func (c *Client) CreateCallback(ctx context.Context, cb Callback, callbackID string) (*Callback, error) {
	u := c.url(c.appName + "/callbacks")
	if callbackID != "" {
		u += "?callbackId=" + url.QueryEscape(callbackID)
	}
	var created Callback
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, cb, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateCallback patches a callback with the specified fields.
func (c *Client) UpdateCallback(ctx context.Context, name string, fields map[string]interface{}) (*Callback, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var cb Callback
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &cb); err != nil {
		return nil, err
	}
	return &cb, nil
}

// DeleteCallback deletes the callback.
func (c *Client) DeleteCallback(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}

// ExecuteCallback is not supported in the Go SDK — Python's exec()-based execution
// cannot be ported. Use external callback testing instead.
func (c *Client) ExecuteCallback(_ context.Context, _ string, _ map[string]interface{}) error {
	return ErrNotSupported
}
