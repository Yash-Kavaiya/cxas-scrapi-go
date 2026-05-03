// Package guardrails provides a client for CXAS guardrail resource management.
package guardrails

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/common"
)

const apiVersion = "v1beta"

// Guardrail represents a CXAS guardrail resource.
type Guardrail struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// Client provides CRUD access to CXAS Guardrails.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Guardrails client scoped to a specific App.
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

// ListGuardrails returns all guardrails in the App.
func (c *Client) ListGuardrails(ctx context.Context) ([]Guardrail, error) {
	return common.Paginate(func(token string) ([]Guardrail, string, error) {
		u := c.url(c.appName + "/guardrails")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Guardrails    []Guardrail `json:"guardrails"`
			NextPageToken string      `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Guardrails, resp.NextPageToken, nil
	})
}

// GetGuardrail fetches a single guardrail by resource name.
func (c *Client) GetGuardrail(ctx context.Context, name string) (*Guardrail, error) {
	var g Guardrail
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// CreateGuardrail creates a new guardrail.
func (c *Client) CreateGuardrail(ctx context.Context, g Guardrail, guardrailID string) (*Guardrail, error) {
	u := c.url(c.appName + "/guardrails")
	if guardrailID != "" {
		u += "?guardrailId=" + url.QueryEscape(guardrailID)
	}
	var created Guardrail
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, g, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateGuardrail patches a guardrail with the specified fields.
func (c *Client) UpdateGuardrail(ctx context.Context, name string, fields map[string]interface{}) (*Guardrail, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var g Guardrail
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// DeleteGuardrail deletes the guardrail.
func (c *Client) DeleteGuardrail(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}
