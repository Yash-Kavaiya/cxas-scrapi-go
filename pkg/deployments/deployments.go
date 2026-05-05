// Package deployments provides a client for CXAS deployment management.
package deployments

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

// Deployment represents a CXAS deployment resource.
type Deployment struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	State       string                 `json:"state,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// Client provides CRUD access to CXAS Deployments.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Deployments client scoped to a specific App.
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

// ListDeployments returns all deployments for the App.
func (c *Client) ListDeployments(ctx context.Context) ([]Deployment, error) {
	return common.Paginate(func(token string) ([]Deployment, string, error) {
		u := c.url(c.appName + "/deployments")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Deployments   []Deployment `json:"deployments"`
			NextPageToken string       `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Deployments, resp.NextPageToken, nil
	})
}

// GetDeployment fetches a single deployment by resource name.
func (c *Client) GetDeployment(ctx context.Context, name string) (*Deployment, error) {
	var d Deployment
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// CreateDeployment creates a new deployment.
func (c *Client) CreateDeployment(ctx context.Context, d Deployment, deploymentID string) (*Deployment, error) {
	u := c.url(c.appName + "/deployments")
	if deploymentID != "" {
		u += "?deploymentId=" + url.QueryEscape(deploymentID)
	}
	var created Deployment
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, d, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateDeployment patches a deployment.
func (c *Client) UpdateDeployment(ctx context.Context, name string, fields map[string]interface{}) (*Deployment, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var d Deployment
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// DeleteDeployment deletes the deployment.
func (c *Client) DeleteDeployment(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}
