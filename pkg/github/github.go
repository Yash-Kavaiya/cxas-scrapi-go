// Package github generates GitHub Actions workflow templates for CXAS CI/CD.
package github

import (
	"fmt"
	"runtime/debug"
)

// sdkVersion returns the current module version from build info.
func sdkVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "v1.0.0"
}

// CIWorkflowOptions configures GitHub Actions workflow generation.
type CIWorkflowOptions struct {
	AppDir                   string
	AppName                  string
	WorkloadIdentityProvider string
	ServiceAccount           string
	ProjectID                string
	Location                 string
}

// GenerateCIWorkflow returns a GitHub Actions YAML workflow for CXAS CI/CD.
func GenerateCIWorkflow(opts CIWorkflowOptions) string {
	return fmt.Sprintf(`name: CXAS CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  ci-test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write

    steps:
    - uses: actions/checkout@v4

    - name: Authenticate to Google Cloud
      uses: google-github-actions/auth@v2
      with:
        workload_identity_provider: %s
        service_account: %s

    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v2

    - name: Install cxas CLI
      run: |
        VERSION=%s
        go install github.com/GoogleCloudPlatform/cxas-go/cmd/cxas@${VERSION}

    - name: Run CXAS CI Tests
      run: |
        cxas ci-test \
          --app-dir %s \
          --app-name %s \
          --project-id %s \
          --location %s
`, opts.WorkloadIdentityProvider, opts.ServiceAccount,
		sdkVersion(), opts.AppDir, opts.AppName, opts.ProjectID, opts.Location)
}
