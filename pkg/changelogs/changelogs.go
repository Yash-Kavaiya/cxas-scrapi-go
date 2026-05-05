// Package changelogs provides a client for CXAS changelog management.
package changelogs

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

// Changelog represents a CXAS changelog entry.
type Changelog struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	CreateTime  string                 `json:"createTime,omitempty"`
	Changes     []map[string]interface{} `json:"changes,omitempty"`
}

// Client provides access to CXAS Changelogs.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Changelogs client scoped to a specific App.
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

// ListChangelogs returns all changelog entries for the App.
func (c *Client) ListChangelogs(ctx context.Context) ([]Changelog, error) {
	return common.Paginate(func(token string) ([]Changelog, string, error) {
		u := c.url(c.appName + "/changelogs")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Changelogs    []Changelog `json:"changelogs"`
			NextPageToken string      `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Changelogs, resp.NextPageToken, nil
	})
}

// GetChangelog fetches a single changelog entry by resource name.
func (c *Client) GetChangelog(ctx context.Context, name string) (*Changelog, error) {
	var cl Changelog
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}
