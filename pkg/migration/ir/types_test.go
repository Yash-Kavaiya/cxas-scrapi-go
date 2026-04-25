package ir_test

import (
	"encoding/json"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationStatus_Constants(t *testing.T) {
	assert.Equal(t, ir.MigrationStatus("COMPILED"), ir.StatusCompiled)
	assert.Equal(t, ir.MigrationStatus("DEPLOYED"), ir.StatusDeployed)
	assert.Equal(t, ir.MigrationStatus("FAILED"), ir.StatusFailed)
	assert.Equal(t, ir.MigrationStatus("PENDING"), ir.StatusPending)
}

func TestIRMetadata_JSONRoundTrip(t *testing.T) {
	meta := ir.IRMetadata{
		AppName:         "My App",
		AppID:           "app-123",
		AppResourceName: "projects/p/locations/l/apps/app-123",
		DefaultModel:    "gemini-2.0-flash",
	}
	b, err := json.Marshal(meta)
	require.NoError(t, err)

	var got ir.IRMetadata
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, meta, got)
}

func TestIRTool_JSONRoundTrip(t *testing.T) {
	tool := ir.IRTool{
		ID:           "tool_billing",
		Name:         "projects/p/locations/l/apps/a/tools/t1",
		Type:         "TOOLSET",
		Payload:      map[string]interface{}{"schema": "openapi"},
		OperationIDs: []string{"getBill", "payBill"},
		Status:       ir.StatusCompiled,
	}
	b, err := json.Marshal(tool)
	require.NoError(t, err)

	var got ir.IRTool
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, tool.ID, got.ID)
	assert.Equal(t, tool.Status, got.Status)
	assert.Equal(t, tool.OperationIDs, got.OperationIDs)
}

func TestIRAgent_JSONRoundTrip(t *testing.T) {
	agent := ir.IRAgent{
		Type:          "FLOW",
		DisplayName:   "Support Agent",
		Instruction:   "<pif>...</pif>",
		Tools:         []string{"projects/p/.../tools/t1"},
		ModelSettings: map[string]interface{}{"model": "gemini-2.0-flash"},
		Status:        ir.StatusDeployed,
	}
	b, err := json.Marshal(agent)
	require.NoError(t, err)

	var got ir.IRAgent
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, agent.Type, got.Type)
	assert.Equal(t, agent.DisplayName, got.DisplayName)
	assert.Equal(t, agent.Status, got.Status)
}

func TestMigrationIR_JSONRoundTrip(t *testing.T) {
	ir2 := ir.MigrationIR{
		Metadata:   ir.IRMetadata{AppName: "test", DefaultModel: "gemini-2.0-flash"},
		Parameters: map[string]interface{}{"lang": "en"},
		Tools:      map[string]ir.IRTool{},
		Agents:     map[string]ir.IRAgent{},
	}
	b, err := json.Marshal(ir2)
	require.NoError(t, err)

	var got ir.MigrationIR
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, "test", got.Metadata.AppName)
}

func TestDFCXFlowModel_HasPages(t *testing.T) {
	flow := ir.DFCXFlowModel{
		FlowID:   "flow-1",
		FlowData: map[string]interface{}{"name": "Default Flow"},
		Pages: []ir.DFCXPageModel{
			{PageID: "page-1", PageData: map[string]interface{}{"displayName": "Start"}},
		},
	}
	assert.Len(t, flow.Pages, 1)
	assert.Equal(t, "page-1", flow.Pages[0].PageID)
}
