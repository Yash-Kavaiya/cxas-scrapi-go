// Package variables provides a client for CXAS variable declaration management.
package variables

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

// VariableDeclaration represents a CXAS variable declaration resource.
type VariableDeclaration struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	DataType    string `json:"dataType,omitempty"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
}

// Client provides CRUD access to CXAS VariableDeclarations.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Variables client scoped to a specific App.
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

// ListVariables returns all variable declarations in the App.
func (c *Client) ListVariables(ctx context.Context) ([]VariableDeclaration, error) {
	return common.Paginate(func(token string) ([]VariableDeclaration, string, error) {
		u := c.url(c.appName + "/variableDeclarations")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			VariableDeclarations []VariableDeclaration `json:"variableDeclarations"`
			NextPageToken        string                `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.VariableDeclarations, resp.NextPageToken, nil
	})
}

// GetVariable fetches a single variable declaration by resource name.
func (c *Client) GetVariable(ctx context.Context, name string) (*VariableDeclaration, error) {
	var v VariableDeclaration
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// CreateVariable creates a new variable declaration.
func (c *Client) CreateVariable(ctx context.Context, v VariableDeclaration, variableID string) (*VariableDeclaration, error) {
	u := c.url(c.appName + "/variableDeclarations")
	if variableID != "" {
		u += "?variableDeclarationId=" + url.QueryEscape(variableID)
	}
	var created VariableDeclaration
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, v, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateVariable patches a variable declaration.
func (c *Client) UpdateVariable(ctx context.Context, name string, fields map[string]interface{}) (*VariableDeclaration, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var v VariableDeclaration
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteVariable deletes the variable declaration.
func (c *Client) DeleteVariable(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}
