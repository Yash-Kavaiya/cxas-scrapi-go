package resource_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_FullAppName(t *testing.T) {
	rn, err := resource.Parse("projects/my-proj/locations/us/apps/app-123")
	require.NoError(t, err)
	assert.Equal(t, "my-proj", rn.ProjectID)
	assert.Equal(t, "us", rn.Location)
	assert.Equal(t, "app-123", rn.AppID)
	assert.Empty(t, rn.SubType)
	assert.Empty(t, rn.SubID)
}

func TestParse_AgentSubresource(t *testing.T) {
	rn, err := resource.Parse("projects/p/locations/l/apps/a/agents/agent-1")
	require.NoError(t, err)
	assert.Equal(t, "agents", rn.SubType)
	assert.Equal(t, "agent-1", rn.SubID)
}

func TestParse_ToolsSubresource(t *testing.T) {
	rn, err := resource.Parse("projects/p/locations/l/apps/a/tools/tool-xyz")
	require.NoError(t, err)
	assert.Equal(t, "p", rn.ProjectID)
	assert.Equal(t, "tools", rn.SubType)
	assert.Equal(t, "tool-xyz", rn.SubID)
}

func TestParse_EmptyString(t *testing.T) {
	_, err := resource.Parse("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestParse_TooShort(t *testing.T) {
	_, err := resource.Parse("my-app")
	assert.Error(t, err)
}

func TestParse_WrongStructure(t *testing.T) {
	_, err := resource.Parse("projects/p/locations/l/WRONG/a")
	assert.Error(t, err)
}

func TestAppParent(t *testing.T) {
	rn, err := resource.Parse("projects/p/locations/l/apps/a/agents/x")
	require.NoError(t, err)
	assert.Equal(t, "projects/p/locations/l/apps/a", rn.AppParent())
}

func TestLocationParent(t *testing.T) {
	rn, err := resource.Parse("projects/p/locations/l/apps/a")
	require.NoError(t, err)
	assert.Equal(t, "projects/p/locations/l", rn.LocationParent())
}

func TestSessionName(t *testing.T) {
	got := resource.SessionName("projects/p/locations/l/apps/a", "sess-abc")
	assert.Equal(t, "projects/p/locations/l/apps/a/sessions/sess-abc", got)
}

func TestAPIEndpoint(t *testing.T) {
	assert.Equal(t, "ces.googleapis.com", resource.APIEndpoint("us"))
	assert.Equal(t, "ces.googleapis.com", resource.APIEndpoint("global"))
}
