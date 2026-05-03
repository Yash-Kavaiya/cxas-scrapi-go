package apps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestAuth() auth.Config {
	return auth.Config{TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-tok"})}
}

func TestListApps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/apps")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"apps": []map[string]interface{}{
				{"name": "projects/p/locations/l/apps/a1", "displayName": "App One"},
			},
		})
	}))
	defer srv.Close()

	client, err := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))
	require.NoError(t, err)

	list, err := client.ListApps(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "App One", list[0].DisplayName)
	assert.Equal(t, "projects/p/locations/l/apps/a1", list[0].Name)
}

func TestGetApp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		json.NewEncoder(w).Encode(apps.App{Name: "projects/p/locations/l/apps/a1", DisplayName: "My App"})
	}))
	defer srv.Close()

	client, err := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))
	require.NoError(t, err)

	app, err := client.GetApp(context.Background(), "projects/p/locations/l/apps/a1")
	require.NoError(t, err)
	assert.Equal(t, "My App", app.DisplayName)
}

func TestGetAppByDisplayName_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"apps": []map[string]interface{}{
				{"name": "projects/p/locations/l/apps/a1", "displayName": "Target App"},
				{"name": "projects/p/locations/l/apps/a2", "displayName": "Other App"},
			},
		})
	}))
	defer srv.Close()

	client, err := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))
	require.NoError(t, err)

	app, err := client.GetAppByDisplayName(context.Background(), "Target App")
	require.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, "projects/p/locations/l/apps/a1", app.Name)
}

func TestGetAppByDisplayName_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"apps": []interface{}{}})
	}))
	defer srv.Close()

	client, _ := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))
	app, err := client.GetAppByDisplayName(context.Background(), "Nonexistent")
	require.NoError(t, err)
	assert.Nil(t, app)
}

func TestDeleteApp(t *testing.T) {
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			deleted = true
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	client, _ := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))
	err := client.DeleteApp(context.Background(), "projects/p/locations/l/apps/a1")
	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestGetAppsMap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"apps": []map[string]interface{}{
				{"name": "projects/p/locations/l/apps/a1", "displayName": "App One"},
			},
		})
	}))
	defer srv.Close()

	client, _ := apps.NewClient(context.Background(), "p", "l", newTestAuth(), apps.WithBaseURL(srv.URL))

	m, err := client.GetAppsMap(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, "App One", m["projects/p/locations/l/apps/a1"])

	m2, err := client.GetAppsMap(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, "projects/p/locations/l/apps/a1", m2["App One"])
}
