package sessions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestAuth() auth.Config {
	return auth.Config{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-tok"})}
}

func TestRunSession_TextInput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, ":run")

		resp := map[string]interface{}{
			"outputs": []interface{}{
				map[string]interface{}{
					"diagnosticInfo": map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"chunks": []interface{}{
									map[string]interface{}{"text": "Hello from agent"},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, err := sessions.NewClient(context.Background(), newTestAuth(), sessions.WithBaseURL(srv.URL))
	require.NoError(t, err)

	out, err := client.Run(context.Background(), sessions.RunSessionRequest{
		AppName:   "projects/p/locations/l/apps/a",
		SessionID: "sess-1",
		Input:     sessions.SessionInput{Text: "Hi"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from agent", out.Text)
}

func TestRunSession_ToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"outputs": []interface{}{
				map[string]interface{}{
					"diagnosticInfo": map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"chunks": []interface{}{
									map[string]interface{}{
										"toolCall": map[string]interface{}{
											"toolName":  "search_tool",
											"toolInput": map[string]interface{}{"query": "weather"},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := sessions.NewClient(context.Background(), newTestAuth(), sessions.WithBaseURL(srv.URL))
	out, err := client.Run(context.Background(), sessions.RunSessionRequest{
		AppName:   "projects/p/locations/l/apps/a",
		SessionID: "sess-1",
		Input:     sessions.SessionInput{Text: "Search for weather"},
	})
	require.NoError(t, err)
	require.Len(t, out.ToolCalls, 1)
	assert.Equal(t, "search_tool", out.ToolCalls[0].ToolName)
	assert.Equal(t, "weather", out.ToolCalls[0].ToolInput["query"])
}

func TestRunSession_SessionEnded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"outputs": []interface{}{
				map[string]interface{}{
					"diagnosticInfo": map[string]interface{}{
						"messages": []interface{}{
							map[string]interface{}{
								"chunks": []interface{}{
									map[string]interface{}{"sessionEnded": true},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := sessions.NewClient(context.Background(), newTestAuth(), sessions.WithBaseURL(srv.URL))
	out, err := client.Run(context.Background(), sessions.RunSessionRequest{
		AppName:   "projects/p/locations/l/apps/a",
		SessionID: "sess-1",
		Input:     sessions.SessionInput{Event: "end_session"},
	})
	require.NoError(t, err)
	assert.True(t, out.SessionEnded)
}

func TestRunSession_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":{"code":400,"message":"bad request"}}`))
	}))
	defer srv.Close()

	client, _ := sessions.NewClient(context.Background(), newTestAuth(), sessions.WithBaseURL(srv.URL))
	_, err := client.Run(context.Background(), sessions.RunSessionRequest{
		AppName:   "projects/p/locations/l/apps/a",
		SessionID: "sess-1",
		Input:     sessions.SessionInput{Text: "Hi"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}
