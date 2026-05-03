package apps

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/common"
)

const apiVersion = "v1beta"

// Client provides CRUD access to CXAS Apps.
type Client struct {
	projectID string
	location  string
	parent    string
	http      *http.Client
	baseURL   string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates an Apps client.
func NewClient(ctx context.Context, projectID, location string, cfg auth.Config, opts ...clientOption) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		projectID: projectID,
		location:  location,
		parent:    fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		http:      httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL:   httpclient.BaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func (c *Client) url(path string) string {
	// path may already start with "projects/…" or be an absolute resource name
	if strings.HasPrefix(path, "http") {
		return path
	}
	return fmt.Sprintf("%s/%s/%s", c.baseURL, apiVersion, path)
}

// ListApps returns all Apps in the project/location.
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	return common.Paginate(func(token string) ([]App, string, error) {
		u := c.url(c.parent + "/apps")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Apps          []App  `json:"apps"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Apps, resp.NextPageToken, nil
	})
}

// GetApp fetches a single App by resource name.
func (c *Client) GetApp(ctx context.Context, appName string) (*App, error) {
	var app App
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(appName), nil, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// GetAppByDisplayName returns the unique App with the given display name, or nil if not found.
func (c *Client) GetAppByDisplayName(ctx context.Context, displayName string) (*App, error) {
	list, err := c.ListApps(ctx)
	if err != nil {
		return nil, err
	}
	var match *App
	for i := range list {
		if list[i].DisplayName == displayName {
			if match != nil {
				return nil, fmt.Errorf("found multiple apps with displayName %q", displayName)
			}
			match = &list[i]
		}
	}
	return match, nil
}

// GetAppsMap returns a map of name→displayName (or reversed when reverse=true).
func (c *Client) GetAppsMap(ctx context.Context, reverse bool) (map[string]string, error) {
	list, err := c.ListApps(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(list))
	for _, a := range list {
		if reverse {
			m[a.DisplayName] = a.Name
		} else {
			m[a.Name] = a.DisplayName
		}
	}
	return m, nil
}

// CreateApp creates a new App and waits for the LRO to complete.
func (c *Client) CreateApp(ctx context.Context, req CreateAppRequest) (*App, error) {
	body := map[string]interface{}{
		"displayName": req.DisplayName,
		"description": req.Description,
	}
	if req.RootAgent != "" {
		body["rootAgent"] = req.RootAgent
	}
	u := c.url(c.parent+"/apps") + "?appId=" + url.QueryEscape(req.AppID)
	var opResp struct {
		Name string `json:"name"`
	}
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &opResp); err != nil {
		return nil, err
	}
	var app App
	if err := common.WaitForOperation(ctx, c.http, c.baseURL, opResp.Name, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApp patches an App with the specified fields and returns the updated resource.
func (c *Client) UpdateApp(ctx context.Context, appName string, fields map[string]interface{}) (*App, error) {
	keys := common.MapKeys(fields)
	u := c.url(appName) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var app App
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// DeleteApp deletes the App with the given resource name.
func (c *Client) DeleteApp(ctx context.Context, appName string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(appName), nil, nil)
}

// ExportApp triggers an export LRO.
func (c *Client) ExportApp(ctx context.Context, req ExportAppRequest) ([]byte, error) {
	body := map[string]interface{}{}
	if req.GCSUri != "" {
		body["gcsDestination"] = map[string]interface{}{"uri": req.GCSUri}
	}
	if req.ExportFormat != "" {
		body["exportFormat"] = req.ExportFormat
	}
	u := c.url(req.AppName + ":export")
	var opResp struct {
		Name string `json:"name"`
	}
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &opResp); err != nil {
		return nil, err
	}
	var result struct {
		AppContent []byte `json:"appContent"`
	}
	if err := common.WaitForOperation(ctx, c.http, c.baseURL, opResp.Name, &result); err != nil {
		return nil, err
	}
	return result.AppContent, nil
}

// ImportApp imports content into an existing App.
func (c *Client) ImportApp(ctx context.Context, req ImportAppRequest) (*App, error) {
	body := map[string]interface{}{}
	if len(req.AppContent) > 0 {
		body["appContent"] = req.AppContent
	}
	if req.GCSUri != "" {
		body["gcsSource"] = map[string]interface{}{"uri": req.GCSUri}
	}
	if req.ConflictStrategy != "" {
		body["conflictStrategy"] = req.ConflictStrategy
	}
	u := c.url(req.AppName + ":import")
	var opResp struct {
		Name string `json:"name"`
	}
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &opResp); err != nil {
		return nil, err
	}
	var app App
	if err := common.WaitForOperation(ctx, c.http, c.baseURL, opResp.Name, &app); err != nil {
		return nil, err
	}
	return &app, nil
}
