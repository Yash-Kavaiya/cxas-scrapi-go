// Package evaluations provides a client for CXAS evaluation management.
package evaluations

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

// Evaluation represents a CXAS evaluation resource.
type Evaluation struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	State       string                 `json:"state,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
}

// EvaluationResult holds the results of a completed evaluation run.
type EvaluationResult struct {
	Name    string                   `json:"name"`
	Results []map[string]interface{} `json:"results,omitempty"`
	Metrics map[string]interface{}   `json:"metrics,omitempty"`
}

// Client provides CRUD access to CXAS Evaluations.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates an Evaluations client scoped to a specific App.
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

// ListEvaluations returns all evaluations in the App.
func (c *Client) ListEvaluations(ctx context.Context) ([]Evaluation, error) {
	return common.Paginate(func(token string) ([]Evaluation, string, error) {
		u := c.url(c.appName + "/evaluations")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Evaluations   []Evaluation `json:"evaluations"`
			NextPageToken string       `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Evaluations, resp.NextPageToken, nil
	})
}

// GetEvaluation fetches a single evaluation by resource name.
func (c *Client) GetEvaluation(ctx context.Context, name string) (*Evaluation, error) {
	var e Evaluation
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(name), nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// CreateEvaluation creates a new evaluation resource.
func (c *Client) CreateEvaluation(ctx context.Context, e Evaluation, evalID string) (*Evaluation, error) {
	u := c.url(c.appName + "/evaluations")
	if evalID != "" {
		u += "?evaluationId=" + url.QueryEscape(evalID)
	}
	var created Evaluation
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, e, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateEvaluation patches an evaluation with the specified fields.
func (c *Client) UpdateEvaluation(ctx context.Context, name string, fields map[string]interface{}) (*Evaluation, error) {
	keys := common.MapKeys(fields)
	u := c.url(name) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var e Evaluation
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// DeleteEvaluation deletes the evaluation.
func (c *Client) DeleteEvaluation(ctx context.Context, name string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(name), nil, nil)
}

// RunEvaluation triggers an evaluation run LRO and waits for completion.
func (c *Client) RunEvaluation(ctx context.Context, evalName string, params map[string]interface{}) (*EvaluationResult, error) {
	u := c.url(evalName + ":run")
	var opResp struct {
		Name string `json:"name"`
	}
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, params, &opResp); err != nil {
		return nil, err
	}
	var result EvaluationResult
	if err := common.WaitForOperation(ctx, c.http, c.baseURL, opResp.Name, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEvaluationResult fetches the result of a completed evaluation run.
func (c *Client) GetEvaluationResult(ctx context.Context, evalName string) (*EvaluationResult, error) {
	u := c.url(evalName + "/result")
	var result EvaluationResult
	if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportEvaluation triggers an export LRO for evaluation data.
func (c *Client) ExportEvaluation(ctx context.Context, evalName, gcsURI string) error {
	u := c.url(evalName + ":export")
	body := map[string]interface{}{}
	if gcsURI != "" {
		body["gcsDestination"] = map[string]interface{}{"uri": gcsURI}
	}
	var opResp struct {
		Name string `json:"name"`
	}
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &opResp); err != nil {
		return err
	}
	return common.WaitForOperation(ctx, c.http, c.baseURL, opResp.Name, nil)
}
