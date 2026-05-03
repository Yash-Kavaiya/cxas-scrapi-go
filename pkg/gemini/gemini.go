package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"golang.org/x/sync/semaphore"
)

const (
	// DefaultModel is the default Gemini model. Note: the Python SDK used
	// "gemini-3.1-flash-lite-preview" which is invalid; corrected here.
	DefaultModel         = "gemini-2.0-flash"
	defaultMaxConcurrent = 2
)

// Client wraps the Vertex AI generateContent REST API with rate limiting.
type Client struct {
	projectID  string
	location   string
	model      string
	httpClient *http.Client
	baseURL    string
	sem        *semaphore.Weighted
}

// Option configures a Client.
type Option func(*Client)

// WithModel sets the default model name.
func WithModel(model string) Option {
	return func(c *Client) { c.model = model }
}

// WithMaxConcurrent sets the maximum number of simultaneous in-flight requests.
func WithMaxConcurrent(n int) Option {
	return func(c *Client) { c.sem = semaphore.NewWeighted(int64(n)) }
}

// WithBaseURL overrides the Vertex AI endpoint (used in tests).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// NewClient creates a new Gemini client.
func NewClient(ctx context.Context, projectID, location string, cfg auth.Config, opts ...Option) (*Client, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		projectID:  projectID,
		location:   location,
		model:      DefaultModel,
		httpClient: httpclient.New(httpclient.Options{TokenSource: ts}),
		baseURL:    fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1", location),
		sem:        semaphore.NewWeighted(defaultMaxConcurrent),
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Generate calls the Gemini API synchronously and returns the generated text.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model := req.ModelName
	if model == "" {
		model = c.model
	}

	payload := c.buildPayload(req)
	url := fmt.Sprintf("%s/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		c.baseURL, c.projectID, c.location, model)

	var resp map[string]interface{}
	if err := httpclient.DoJSON(ctx, c.httpClient, "POST", url, payload, &resp); err != nil {
		return nil, err
	}
	return extractResponse(resp)
}

// GenerateWithRetry calls Generate with exponential backoff on quota errors (429/RESOURCE_EXHAUSTED).
// maxRetries=0 means use default of 5.
func (c *Client) GenerateWithRetry(ctx context.Context, req GenerateRequest, maxRetries int) (*GenerateResponse, error) {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	baseDelay := 10 * time.Second

	if err := c.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer c.sem.Release(1)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isQuotaError(err) {
			return nil, err
		}
		delay := time.Duration(float64(baseDelay) * pow(1.5, float64(attempt)))
		jitter := time.Duration(rand.Int63n(int64(delay / 4)))
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay + jitter):
		}
	}
	return nil, fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
}

func (c *Client) buildPayload(req GenerateRequest) map[string]interface{} {
	contents := []map[string]interface{}{
		{"role": "user", "parts": []map[string]interface{}{{"text": req.Prompt}}},
	}
	payload := map[string]interface{}{"contents": contents}

	if req.SystemPrompt != "" {
		payload["system_instruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{{"text": req.SystemPrompt}},
		}
	}
	if req.ResponseSchema != nil {
		payload["generation_config"] = map[string]interface{}{
			"response_mime_type": "application/json",
			"response_schema":    req.ResponseSchema,
		}
	}
	return payload
}

func extractResponse(resp map[string]interface{}) (*GenerateResponse, error) {
	candidates, ok := resp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}
	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid candidate format")
	}
	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no content in candidate")
	}
	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return nil, fmt.Errorf("no parts in content")
	}
	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid part format")
	}
	text, _ := part["text"].(string)
	return &GenerateResponse{Text: text, Raw: resp}, nil
}

func isQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "429") || contains(msg, "RESOURCE_EXHAUSTED") || contains(msg, "quota")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// EvaluateExpectations uses Gemini to judge whether a conversation trace meets expectations.
// Returns a JSON string with evaluation results.
func (c *Client) EvaluateExpectations(ctx context.Context, trace string, expectations []string) (string, error) {
	expectationsJSON, _ := json.Marshal(expectations)
	prompt := fmt.Sprintf("Evaluate whether the following conversation trace meets the expectations.\n\nTrace:\n%s\n\nExpectations:\n%s\n\nFor each expectation, respond with JSON: [{\"expectation\": \"...\", \"status\": \"MET|NOT_MET\", \"justification\": \"...\"}]",
		trace, string(expectationsJSON))

	resp, err := c.GenerateWithRetry(ctx, GenerateRequest{Prompt: prompt}, 3)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

