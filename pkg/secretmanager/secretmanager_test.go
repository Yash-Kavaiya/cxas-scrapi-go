package secretmanager_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/secretmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestAuth() auth.Config {
	return auth.Config{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-tok"})}
}

func TestGetOrCreateSecret_ExistingSecret(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, "GET", r.Method)
		json.NewEncoder(w).Encode(map[string]interface{}{"name": "projects/p/secrets/my-secret"})
	}))
	defer srv.Close()

	client, err := secretmanager.NewClientWithBaseURL(context.Background(), "p", newTestAuth(), srv.URL)
	require.NoError(t, err)

	name, err := client.GetOrCreateSecret(context.Background(), "my-secret")
	require.NoError(t, err)
	assert.Equal(t, "projects/p/secrets/my-secret", name)
	assert.True(t, called)
}

func TestGetOrCreateSecret_CreatesWhenNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(404)
			w.Write([]byte(`{"error":{"code":404,"message":"not found"}}`))
			return
		}
		// POST to create
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"name": "projects/p/secrets/new-secret"})
	}))
	defer srv.Close()

	client, err := secretmanager.NewClientWithBaseURL(context.Background(), "p", newTestAuth(), srv.URL)
	require.NoError(t, err)

	name, err := client.GetOrCreateSecret(context.Background(), "new-secret")
	require.NoError(t, err)
	assert.Equal(t, "projects/p/secrets/new-secret", name)
}
