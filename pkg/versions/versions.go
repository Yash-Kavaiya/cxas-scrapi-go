// Package versions provides a client for CXAS app version snapshots.
package versions

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

// Version represents a CXAS app version snapshot.
type Version struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	CreateTime  string `json:"createTime,omitempty"`
}

// Client provides CRUD access to CXAS Versions.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Versions client scoped to a specific App.
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

// ListVersions returns all version snapshots for the App.
func (c *Client) ListVersions(ctx context.Context) ([]Version, error) {
	return common.Paginate(func(token string) ([]Version, string, error) {
		u := c.url(c.appName + "/versions")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Versions      []Version `json:"versions"`
			NextPageToken string    `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Versions, resp.NextPageToken, nil
	})
}

// GetVersion fetches a single version by resource name.
func (c *Client) GetVersion(ctx context.Context, name string) (*Version, error) {
	var v Version
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// CreateVersion creates a new version snapshot of the App.
func (c *Client) CreateVersion(ctx context.Context, v Version, versionID string) (*Version, error) {
	u := c.url(c.appName + "/versions")
	if versionID != "" {
		u += "?versionId=" + url.QueryEscape(versionID)
	}
	var created Version
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, v, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteVersion deletes the version snapshot.
func (c *Client) DeleteVersion(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}
