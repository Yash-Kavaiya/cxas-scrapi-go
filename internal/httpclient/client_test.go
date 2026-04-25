package httpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNew_InjectsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{})
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Contains(t, gotUA, "cxas-go/")
}

func TestNew_InjectsBearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "tok-abc"})
	client := httpclient.New(httpclient.Options{TokenSource: ts})
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "Bearer tok-abc", gotAuth)
}

func TestNew_InjectsQuotaProject(t *testing.T) {
	var gotQP string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQP = r.Header.Get("x-goog-user-project")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{QuotaProject: "my-project"})
	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "my-project", gotQP)
}

func TestDoJSON_DecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"name": "my-app", "displayName": "My App"})
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{})
	var out map[string]string
	err := httpclient.DoJSON(context.Background(), client, "GET", srv.URL, nil, &out)
	require.NoError(t, err)
	assert.Equal(t, "My App", out["displayName"])
}

func TestDoJSON_Returns404AsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":{"code":404,"message":"not found"}}`))
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{})
	err := httpclient.DoJSON(context.Background(), client, "GET", srv.URL+"/missing", nil, nil)
	require.Error(t, err)
	assert.True(t, httpclient.IsNotFound(err))
}

func TestDoJSON_Returns403AsPermissionDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{})
	err := httpclient.DoJSON(context.Background(), client, "GET", srv.URL, nil, nil)
	require.Error(t, err)
	assert.True(t, httpclient.IsPermissionDenied(err))
}

func TestDoJSON_SendsJSONBody(t *testing.T) {
	var gotContentType string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := httpclient.New(httpclient.Options{})
	body := map[string]string{"displayName": "Test"}
	err := httpclient.DoJSON(context.Background(), client, "POST", srv.URL, body, nil)
	require.NoError(t, err)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "Test", gotBody["displayName"])
}

func TestMockRoundTripper_ReturnsConfiguredResponse(t *testing.T) {
	mock := &httpclient.MockRoundTripper{
		Responses: map[string]*http.Response{
			"GET https://example.com/apps": httpclient.NewMockResponse(200, `{"apps":[]}`),
		},
	}
	client := &http.Client{Transport: mock}
	var out map[string]interface{}
	err := httpclient.DoJSON(context.Background(), client, "GET", "https://example.com/apps", nil, &out)
	require.NoError(t, err)
	assert.Len(t, mock.Calls, 1)
	assert.Equal(t, "GET https://example.com/apps", mock.Calls[0])
}
