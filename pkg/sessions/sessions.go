package sessions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/GoogleCloudPlatform/cxas-go/internal/resource"
)

// Client provides access to CXAS text-mode session execution.
type Client struct {
	http    *http.Client
	baseURL string
}

type clientOption func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) clientOption { return func(c *Client) { c.baseURL = u } }

// NewClient creates a Sessions client.
func NewClient(ctx context.Context, cfg auth.Config, opts ...clientOption) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		http:    httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL: httpclient.BaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s/%s/%s", c.baseURL, httpclient.APIVersion, path)
}

// Run executes a single text-mode session turn and returns the parsed output.
func (c *Client) Run(ctx context.Context, req RunSessionRequest) (*SessionOutput, error) {
	sessionName := resource.SessionName(req.AppName, req.SessionID)
	body := buildRunBody(req.Input)

	var raw map[string]interface{}
	if err := httpclient.DoJSON(ctx, c.http, "POST", c.url(sessionName+":run"), body, &raw); err != nil {
		return nil, fmt.Errorf("run session %s: %w", sessionName, err)
	}
	return parseSessionOutput(raw), nil
}

// buildRunBody constructs the REST request body from a SessionInput.
func buildRunBody(input SessionInput) map[string]interface{} {
	body := map[string]interface{}{}
	if input.Text != "" {
		body["text"] = input.Text
	}
	if input.Event != "" {
		body["event"] = input.Event
		if input.EventVars != nil {
			body["eventVars"] = input.EventVars
		}
	}
	if len(input.Blob) > 0 {
		body["blob"] = input.Blob
		if input.BlobMimeType != "" {
			body["blobMimeType"] = input.BlobMimeType
		}
	}
	if input.DTMF != "" {
		body["dtmf"] = input.DTMF
	}
	if len(input.ToolResponses) > 0 {
		body["toolResponses"] = input.ToolResponses
	}
	return body
}

// parseSessionOutput extracts structured data from the raw API response.
// Walks outputs[].diagnostic_info.messages[].chunks[] looking for text/tool/agent signals.
func parseSessionOutput(raw map[string]interface{}) *SessionOutput {
	out := &SessionOutput{Raw: raw}

	outputs, _ := raw["outputs"].([]interface{})
	for _, o := range outputs {
		om, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		di, _ := om["diagnosticInfo"].(map[string]interface{})
		messages, _ := di["messages"].([]interface{})
		for _, msg := range messages {
			mm, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}
			chunks, _ := mm["chunks"].([]interface{})
			for _, ch := range chunks {
				cm, ok := ch.(map[string]interface{})
				if !ok {
					continue
				}
				if t, ok := cm["text"].(string); ok && t != "" {
					out.Text += t
				}
				if tc, ok := cm["toolCall"].(map[string]interface{}); ok {
					toolInput, _ := tc["toolInput"].(map[string]interface{})
					out.ToolCalls = append(out.ToolCalls, ToolCall{
						ToolName:  fmt.Sprintf("%v", tc["toolName"]),
						ToolInput: toolInput,
					})
				}
				if at, ok := cm["agentTransfer"].(map[string]interface{}); ok {
					out.AgentTransfer, _ = at["targetAgent"].(string)
				}
				if se, ok := cm["sessionEnded"].(bool); ok && se {
					out.SessionEnded = true
				}
			}
		}
	}
	return out
}
