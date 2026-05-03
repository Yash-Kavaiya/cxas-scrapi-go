package github_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/github"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCIWorkflow_ContainsRequiredFields(t *testing.T) {
	yaml := github.GenerateCIWorkflow(github.CIWorkflowOptions{
		AppDir:                   "./my-app",
		AppName:                  "projects/p/locations/l/apps/a",
		WorkloadIdentityProvider: "projects/123/locations/global/workloadIdentityPools/pool/providers/prov",
		ServiceAccount:           "sa@project.iam.gserviceaccount.com",
		ProjectID:                "my-project",
		Location:                 "us",
	})
	assert.Contains(t, yaml, "CXAS CI")
	assert.Contains(t, yaml, "workload_identity_provider")
	assert.Contains(t, yaml, "sa@project.iam.gserviceaccount.com")
	assert.Contains(t, yaml, "my-app")
	assert.Contains(t, yaml, "my-project")
}
