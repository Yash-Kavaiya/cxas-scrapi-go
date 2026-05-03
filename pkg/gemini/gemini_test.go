package gemini_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func fakeVertexResponse(text string) map[string]interface{} {
	return map[string]interface{}{
		"candidates": []interface{}{
			map[string]interface{}{
				"content": map[string]interface{}{
					"parts": []interface{}{
						map[string]interface{}{"text": text},
					},
				},
			},
		},
	}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*gemini.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := auth.Config{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-tok"})}
	client, err := gemini.NewClient(context.Background(), "proj", "us-central1", cfg,
		gemini.WithBaseURL(srv.URL),
		gemini.WithMaxConcurrent(5),
	)
	require.NoError(t, err)
	return client, srv
}

func TestGenerate_ReturnsText(t *testing.T) {
	client, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fakeVertexResponse("Hello from Gemini"))
	})
	defer srv.Close()

	resp, err := client.Generate(context.Background(), gemini.GenerateRequest{Prompt: "Say hello"})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Gemini", resp.Text)
}

func TestGenerate_SendsSystemPrompt(t *testing.T) {
	var gotBody map[string]interface{}
	client, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(fakeVertexResponse("ok"))
	})
	defer srv.Close()

	client.Generate(context.Background(), gemini.GenerateRequest{
		Prompt:       "hello",
		SystemPrompt: "You are helpful",
	})
	_, hasSystem := gotBody["system_instruction"]
	assert.True(t, hasSystem, "system_instruction should be in payload")
}

func TestGenerate_APIError(t *testing.T) {
	client, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"internal"}`))
	})
	defer srv.Close()

	_, err := client.Generate(context.Background(), gemini.GenerateRequest{Prompt: "test"})
	assert.Error(t, err)
}

func TestDefaultModel(t *testing.T) {
	assert.Equal(t, "gemini-2.0-flash", gemini.DefaultModel)
}

func TestGenerateWithRetry_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	client, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(fakeVertexResponse("result"))
	})
	defer srv.Close()

	resp, err := client.GenerateWithRetry(context.Background(), gemini.GenerateRequest{Prompt: "test"}, 3)
	require.NoError(t, err)
	assert.Equal(t, "result", resp.Text)
	assert.Equal(t, 1, calls)
}
