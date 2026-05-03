package tools

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

// Client provides CRUD access to CXAS Tools and Toolsets.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Tools client scoped to a specific App.
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

// ListTools returns all tools in the App.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	return common.Paginate(func(token string) ([]Tool, string, error) {
		u := c.url(c.appName + "/tools")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Tools         []Tool `json:"tools"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Tools, resp.NextPageToken, nil
	})
}

// GetTool fetches a single tool by resource name.
func (c *Client) GetTool(ctx context.Context, name string) (*Tool, error) {
	var t Tool
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTool creates a new tool.
func (c *Client) CreateTool(ctx context.Context, t Tool, toolID string) (*Tool, error) {
	u := c.url(c.appName + "/tools")
	if toolID != "" {
		u += "?toolId=" + url.QueryEscape(toolID)
	}
	var created Tool
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, t, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateTool patches a tool with the specified fields.
func (c *Client) UpdateTool(ctx context.Context, name string, fields map[string]interface{}) (*Tool, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var t Tool
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// DeleteTool deletes the tool.
func (c *Client) DeleteTool(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}

// ListToolsets returns all toolsets in the App.
func (c *Client) ListToolsets(ctx context.Context) ([]Toolset, error) {
	return common.Paginate(func(token string) ([]Toolset, string, error) {
		u := c.url(c.appName + "/toolsets")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Toolsets      []Toolset `json:"toolsets"`
			NextPageToken string    `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Toolsets, resp.NextPageToken, nil
	})
}

// ExecuteTool invokes a tool directly via the REST endpoint.
// Uses "tool" field for atomic tools, "toolsetTool" for toolset-bound tools.
func (c *Client) ExecuteTool(ctx context.Context, req ExecuteToolRequest) (*ExecuteToolResponse, error) {
	body := map[string]interface{}{
		"input": req.Input,
	}
	if req.ToolsetTool != nil {
		body["toolsetTool"] = map[string]interface{}{
			"toolset": req.ToolsetTool.Toolset,
			"toolId":  req.ToolsetTool.ToolID,
		}
	} else if req.Tool != "" {
		body["tool"] = req.Tool
	}

	u := c.url(req.AppName + ":executeTool")
	var resp ExecuteToolResponse
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
