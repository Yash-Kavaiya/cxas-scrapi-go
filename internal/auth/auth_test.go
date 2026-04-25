package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewTokenSource_PreBuiltTokenSource(t *testing.T) {
	staticTS := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "pre-built-tok"})
	ts, err := auth.NewTokenSource(context.Background(), auth.Config{TokenSource: staticTS})
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "pre-built-tok", tok.AccessToken)
}

func TestNewTokenSource_EnvVarToken(t *testing.T) {
	os.Unsetenv("CXAS_OAUTH_TOKEN")
	os.Setenv("CXAS_OAUTH_TOKEN", "env-var-token-xyz")
	t.Cleanup(func() { os.Unsetenv("CXAS_OAUTH_TOKEN") })

	ts, err := auth.NewTokenSource(context.Background(), auth.Config{})
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "env-var-token-xyz", tok.AccessToken)
}

func TestNewTokenSource_PreBuiltTakesPriorityOverEnv(t *testing.T) {
	os.Setenv("CXAS_OAUTH_TOKEN", "env-tok")
	t.Cleanup(func() { os.Unsetenv("CXAS_OAUTH_TOKEN") })

	staticTS := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "explicit-tok"})
	ts, err := auth.NewTokenSource(context.Background(), auth.Config{TokenSource: staticTS})
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "explicit-tok", tok.AccessToken, "pre-built source should beat env var")
}

func TestNewTokenSource_InvalidCredsPath(t *testing.T) {
	_, err := auth.NewTokenSource(context.Background(), auth.Config{CredsPath: "/nonexistent/path/sa.json"})
	assert.Error(t, err)
}

func TestNewTokenSource_InvalidCredsJSON(t *testing.T) {
	_, err := auth.NewTokenSource(context.Background(), auth.Config{CredsJSON: []byte(`{"invalid": "json"}`)})
	assert.Error(t, err)
}

func TestNewTokenSource_NoCredsFallsToADC(t *testing.T) {
	// Skip if CXAS_OAUTH_TOKEN or GOOGLE_APPLICATION_CREDENTIALS is set
	// (would succeed for the wrong reason in CI)
	if os.Getenv("CXAS_OAUTH_TOKEN") != "" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		t.Skip("credentials env vars set; skipping ADC fallback test")
	}
	// ADC will fail on machines with no credentials — that's expected and correct
	_, err := auth.NewTokenSource(context.Background(), auth.Config{})
	// Either succeeds (dev machine with gcloud auth) or errors with ADC message
	if err != nil {
		assert.Contains(t, err.Error(), "credentials")
	}
}
