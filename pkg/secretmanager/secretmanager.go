// Package secretmanager provides utilities for Google Cloud Secret Manager.
package secretmanager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
)

const defaultBaseURL = "https://secretmanager.googleapis.com/v1"

// Client manages secrets in Google Cloud Secret Manager.
type Client struct {
	projectID  string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a Secret Manager client.
func NewClient(ctx context.Context, projectID string, cfg auth.Config) (*Client, error) {
	return NewClientWithBaseURL(ctx, projectID, cfg, defaultBaseURL)
}

// NewClientWithBaseURL creates a Secret Manager client with a custom base URL (for testing).
func NewClientWithBaseURL(ctx context.Context, projectID string, cfg auth.Config, baseURL string) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		projectID:  projectID,
		httpClient: httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL:    baseURL,
	}, nil
}

// GetOrCreateSecret retrieves an existing secret or creates it if not found.
// Uses get-then-create (O(1)) instead of list-then-scan (O(N)).
func (c *Client) GetOrCreateSecret(ctx context.Context, secretID string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s", c.projectID, secretID)
	url := fmt.Sprintf("%s/%s", c.baseURL, name)

	var secret map[string]interface{}
	err := httpclient.DoJSON(ctx, c.httpClient, "GET", url, nil, &secret)
	if err == nil {
		if n, ok := secret["name"].(string); ok {
			return n, nil
		}
		return name, nil
	}
	if !httpclient.IsNotFound(err) {
		return "", fmt.Errorf("get secret: %w", err)
	}

	createURL := fmt.Sprintf("%s/projects/%s/secrets?secretId=%s", c.baseURL, c.projectID, secretID)
	body := map[string]interface{}{
		"replication": map[string]interface{}{"automatic": map[string]interface{}{}},
	}
	var created map[string]interface{}
	if err := httpclient.DoJSON(ctx, c.httpClient, "POST", createURL, body, &created); err != nil {
		return "", fmt.Errorf("create secret: %w", err)
	}
	if n, ok := created["name"].(string); ok {
		return n, nil
	}
	return name, nil
}

// AddSecretVersion adds a new version with the given payload to the secret.
func (c *Client) AddSecretVersion(ctx context.Context, secretName string, payload []byte) error {
	url := fmt.Sprintf("%s/%s:addVersion", c.baseURL, secretName)
	body := map[string]interface{}{
		"payload": map[string]interface{}{"data": payload},
	}
	return httpclient.DoJSON(ctx, c.httpClient, "POST", url, body, nil)
}
