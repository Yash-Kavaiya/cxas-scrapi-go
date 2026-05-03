// Package scorecards provides a client for CXAS scorecard management.
package scorecards

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
)

// Client provides access to CXAS scorecards.
type Client struct {
	appName    string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a Scorecards client.
func NewClient(ctx context.Context, appName string, cfg auth.Config) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		appName:    appName,
		httpClient: httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL:    "https://ces.googleapis.com",
	}, nil
}

// ListScorecards lists all scorecards for the app.
func (c *Client) ListScorecards(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1beta/%s/scorecards", c.baseURL, c.appName)
	var resp struct {
		Scorecards    []map[string]interface{} `json:"scorecards"`
		NextPageToken string                    `json:"nextPageToken"`
	}
	if err := httpclient.DoJSON(ctx, c.httpClient, "GET", url, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Scorecards, nil
}
