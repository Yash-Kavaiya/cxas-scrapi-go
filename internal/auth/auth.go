// Package auth resolves Google Cloud credentials for the CXAS API.
package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// CloudPlatformScope is the OAuth2 scope required for all CXAS API calls.
const CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// Config holds credential resolution inputs. Fields are tried in priority order:
// TokenSource > CredsPath > CredsJSON > CXAS_OAUTH_TOKEN env var > ADC.
type Config struct {
	// CredsPath is a path to a service account JSON key file.
	CredsPath string
	// CredsJSON is a service account JSON key as raw bytes.
	CredsJSON []byte
	// TokenSource is a pre-built token source (takes highest priority).
	TokenSource oauth2.TokenSource
	// ExtraScopes are additional OAuth2 scopes to request.
	ExtraScopes []string
}

// NewTokenSource resolves credentials using the same priority chain as the Python SDK's
// Common.__init__: explicit TokenSource > file path > JSON bytes > CXAS_OAUTH_TOKEN env > ADC.
func NewTokenSource(ctx context.Context, cfg Config) (oauth2.TokenSource, error) {
	scopes := append([]string{CloudPlatformScope}, cfg.ExtraScopes...)

	if cfg.TokenSource != nil {
		return cfg.TokenSource, nil
	}
	if cfg.CredsPath != "" {
		data, err := os.ReadFile(cfg.CredsPath)
		if err != nil {
			return nil, fmt.Errorf("read creds file %s: %w", cfg.CredsPath, err)
		}
		return tokenSourceFromJSON(ctx, data, scopes)
	}
	if len(cfg.CredsJSON) > 0 {
		return tokenSourceFromJSON(ctx, cfg.CredsJSON, scopes)
	}
	if tok := os.Getenv("CXAS_OAUTH_TOKEN"); tok != "" {
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok}), nil
	}
	return google.DefaultTokenSource(ctx, scopes...)
}

func tokenSourceFromJSON(ctx context.Context, data []byte, scopes []string) (oauth2.TokenSource, error) {
	creds, err := google.CredentialsFromJSON(ctx, data, scopes...)
	if err != nil {
		return nil, fmt.Errorf("credentials from JSON: %w", err)
	}
	return creds.TokenSource, nil
}
