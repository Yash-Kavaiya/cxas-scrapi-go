package agents_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestAuth() auth.Config {
	return auth.Config{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-tok"})}
}

const appName = "projects/p/locations/l/apps/a"

func TestListAgents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agents": []map[string]interface{}{
				{"name": appName + "/agents/ag1", "displayName": "Agent One"},
			},
		})
	}))
	defer srv.Close()

	client, err := agents.NewClient(context.Background(), appName, newTestAuth(), agents.WithBaseURL(srv.URL))
	require.NoError(t, err)

	list, err := client.ListAgents(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "Agent One", list[0].DisplayName)
}

func TestCreateAgent_WorkflowUsesRESTPOST(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(agents.Agent{Name: appName + "/agents/wf-1"})
	}))
	defer srv.Close()

	client, err := agents.NewClient(context.Background(), appName, newTestAuth(), agents.WithBaseURL(srv.URL))
	require.NoError(t, err)

	ag, err := client.CreateAgent(context.Background(), agents.CreateAgentRequest{
		DisplayName: "Workflow Agent",
		AgentType:   agents.AgentTypeWorkflow,
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Contains(t, gotPath, "/agents")
	assert.NotNil(t, gotBody["workflowAgent"])
	assert.Equal(t, appName+"/agents/wf-1", ag.Name)
}

func TestCreateAgent_LLM(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(agents.Agent{Name: appName + "/agents/llm-1"})
	}))
	defer srv.Close()

	client, _ := agents.NewClient(context.Background(), appName, newTestAuth(), agents.WithBaseURL(srv.URL))
	_, err := client.CreateAgent(context.Background(), agents.CreateAgentRequest{
		DisplayName: "LLM Agent",
		AgentType:   agents.AgentTypeLLM,
		Model:       "gemini-2.0-flash",
		Instruction: "Be helpful.",
	})
	require.NoError(t, err)
	assert.NotNil(t, gotBody["llmAgent"])
	assert.Equal(t, "Be helpful.", gotBody["instruction"])
}

func TestCreateAgent_DFCX(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(agents.Agent{Name: appName + "/agents/dfcx-1"})
	}))
	defer srv.Close()

	client, _ := agents.NewClient(context.Background(), appName, newTestAuth(), agents.WithBaseURL(srv.URL))
	_, err := client.CreateAgent(context.Background(), agents.CreateAgentRequest{
		DisplayName:       "DFCX Agent",
		AgentType:         agents.AgentTypeDFCX,
		DFCXAgentResource: "projects/p/locations/l/agents/my-dfcx",
	})
	require.NoError(t, err)
	rda, _ := gotBody["remoteDialogflowAgent"].(map[string]interface{})
	assert.Equal(t, "projects/p/locations/l/agents/my-dfcx", rda["agent"])
}

func TestDeleteAgent(t *testing.T) {
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			deleted = true
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	client, _ := agents.NewClient(context.Background(), appName, newTestAuth(), agents.WithBaseURL(srv.URL))
	err := client.DeleteAgent(context.Background(), appName+"/agents/ag1")
	require.NoError(t, err)
	assert.True(t, deleted)
}
