package agents

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

// Client provides CRUD access to CXAS Agents within an App.
type Client struct {
	appName string
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates an Agents client scoped to a specific App.
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

// ListAgents returns all agents in the App.
func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	return common.Paginate(func(token string) ([]Agent, string, error) {
		u := c.url(c.appName + "/agents")
		if token != "" {
			u += "?pageToken=" + url.QueryEscape(token)
		}
		var resp struct {
			Agents        []Agent `json:"agents"`
			NextPageToken string  `json:"nextPageToken"`
		}
		if err := httpclient.DoJSON(ctx, c.http, "GET", u, nil, &resp); err != nil {
			return nil, "", err
		}
		return resp.Agents, resp.NextPageToken, nil
	})
}

// GetAgent fetches a single Agent by resource name.
func (c *Client) GetAgent(ctx context.Context, agentName string) (*Agent, error) {
	var agent Agent
	if err := httpclient.DoJSON(ctx, c.http, "GET", c.url(agentName), nil, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetAgentByDisplayName returns the unique Agent with the given display name.
func (c *Client) GetAgentByDisplayName(ctx context.Context, displayName string) (*Agent, error) {
	list, err := c.ListAgents(ctx)
	if err != nil {
		return nil, err
	}
	var match *Agent
	for i := range list {
		if list[i].DisplayName == displayName {
			if match != nil {
				return nil, fmt.Errorf("found multiple agents with displayName %q", displayName)
			}
			match = &list[i]
		}
	}
	return match, nil
}

// CreateAgent creates a new Agent. Workflow agents use REST POST (not gRPC).
func (c *Client) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error) {
	body := map[string]interface{}{
		"displayName": req.DisplayName,
	}
	if req.Extra != nil {
		for k, v := range req.Extra {
			body[k] = v
		}
	}

	switch req.AgentType {
	case AgentTypeLLM:
		body["llmAgent"] = map[string]interface{}{}
		if req.Model != "" || req.Instruction != "" {
			body["instruction"] = req.Instruction
			if req.Model != "" {
				body["modelSettings"] = map[string]interface{}{"model": req.Model}
			}
		}
	case AgentTypeDFCX:
		body["remoteDialogflowAgent"] = map[string]interface{}{
			"agent": req.DFCXAgentResource,
		}
	case AgentTypeWorkflow:
		wf := map[string]interface{}{}
		if req.WorkflowConfig != nil {
			wf = req.WorkflowConfig
		}
		body["workflowAgent"] = wf
	}

	u := c.url(c.appName+"/agents")
	if req.AgentID != "" {
		u += "?agentId=" + url.QueryEscape(req.AgentID)
	}
	var agent Agent
	if err := httpclient.DoJSON(ctx, c.http, "POST", u, body, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// UpdateAgent patches an Agent with the specified fields.
func (c *Client) UpdateAgent(ctx context.Context, agentName string, fields map[string]interface{}) (*Agent, error) {
	keys := common.MapKeys(fields)
	u := c.url(agentName) + "?updateMask=" + url.QueryEscape(common.FieldMask(keys...))
	var agent Agent
	if err := httpclient.DoJSON(ctx, c.http, "PATCH", u, fields, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// DeleteAgent deletes the Agent with the given resource name.
func (c *Client) DeleteAgent(ctx context.Context, agentName string) error {
	return httpclient.DoJSON(ctx, c.http, "DELETE", c.url(agentName), nil, nil)
}
