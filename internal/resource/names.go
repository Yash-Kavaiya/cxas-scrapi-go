// Package resource provides helpers for parsing and building GCP resource names.
package resource

import (
	"errors"
	"fmt"
	"strings"
)

// ResourceName holds the parsed components of a CXAS resource name.
// Expected format: projects/{P}/locations/{L}/apps/{A}[/{subtype}/{id}]
type ResourceName struct {
	ProjectID string
	Location  string
	AppID     string
	SubType   string // e.g. "agents", "tools", "sessions"
	SubID     string
}

// Parse parses a CXAS resource name string into a ResourceName.
// Returns an error if the name doesn't have at least the apps/{A} segment.
func Parse(name string) (ResourceName, error) {
	if name == "" {
		return ResourceName{}, errors.New("resource name is empty")
	}
	parts := strings.Split(name, "/")
	if len(parts) < 6 || parts[0] != "projects" || parts[2] != "locations" || parts[4] != "apps" {
		return ResourceName{}, fmt.Errorf("invalid resource name %q: expected projects/{P}/locations/{L}/apps/{A}", name)
	}
	rn := ResourceName{
		ProjectID: parts[1],
		Location:  parts[3],
		AppID:     parts[5],
	}
	if len(parts) >= 8 {
		rn.SubType = parts[6]
		rn.SubID = parts[7]
	}
	return rn, nil
}

// AppParent returns the app-level resource name: "projects/{P}/locations/{L}/apps/{A}".
func (r ResourceName) AppParent() string {
	return fmt.Sprintf("projects/%s/locations/%s/apps/%s", r.ProjectID, r.Location, r.AppID)
}

// LocationParent returns the location-level parent: "projects/{P}/locations/{L}".
func (r ResourceName) LocationParent() string {
	return fmt.Sprintf("projects/%s/locations/%s", r.ProjectID, r.Location)
}

// SessionName builds a session resource name from an app name and session ID.
func SessionName(appName, sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s", appName, sessionID)
}

// APIEndpoint returns the CES API endpoint hostname. Currently always ces.googleapis.com.
func APIEndpoint(_ string) string {
	return "ces.googleapis.com"
}
