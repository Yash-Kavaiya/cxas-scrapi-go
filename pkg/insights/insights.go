// Package insights provides a client for the Contact Center Insights API.
package insights

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
)

const insightsBaseURL = "https://contactcenterinsights.googleapis.com/v1"

// Client provides access to Contact Center Insights.
type Client struct {
	projectID  string
	location   string
	httpClient *http.Client
}

// NewClient creates an Insights client.
func NewClient(ctx context.Context, projectID, location string, cfg auth.Config) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		projectID:  projectID,
		location:   location,
		httpClient: httpclient.New(httpclient.Options{TokenSource: ts}),
	}, nil
}

// ListConversations lists conversations in the given project.
func (c *Client) ListConversations(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/projects/%s/locations/%s/conversations", insightsBaseURL, c.projectID, c.location)
	var resp struct {
		Conversations []map[string]interface{} `json:"conversations"`
		NextPageToken string                    `json:"nextPageToken"`
	}
	if err := httpclient.DoJSON(ctx, c.httpClient, "GET", url, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Conversations, nil
}
