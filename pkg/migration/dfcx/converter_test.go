package dfcx_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/dfcx"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/ir"
	"github.com/stretchr/testify/assert"
)

func TestPlaybookConverter_ConvertPlaybook(t *testing.T) {
	conv := dfcx.NewPlaybookConverter("gemini-2.0-flash")
	playbook := map[string]interface{}{
		"name":        "projects/p/agents/a/playbooks/pb1",
		"displayName": "Support Playbook",
		"goal":        "Help the user resolve issues",
	}
	agent := conv.ConvertPlaybook(playbook)
	assert.Equal(t, "PLAYBOOK", agent.Type)
	assert.Equal(t, "Support Playbook", agent.DisplayName)
	assert.Equal(t, ir.StatusCompiled, agent.Status)
	assert.Contains(t, agent.Instruction, "Support Playbook")
	assert.Contains(t, agent.Instruction, "Help the user resolve issues")
}

func TestToolConverter_ConvertWebhook(t *testing.T) {
	conv := dfcx.NewToolConverter()
	webhook := map[string]interface{}{
		"name":        "projects/p/agents/a/webhooks/wh1",
		"displayName": "Billing Webhook",
	}
	tool := conv.ConvertWebhook(webhook)
	assert.Equal(t, "TOOLSET", tool.Type)
	assert.Equal(t, ir.StatusCompiled, tool.Status)
	assert.NotEmpty(t, tool.OperationIDs)
	_, hasOpenAPI := tool.Payload["openapi"]
	assert.True(t, hasOpenAPI)
}

func TestParameterExtractor_ExtractParameters(t *testing.T) {
	ext := dfcx.NewParameterExtractor()
	agentIR := &ir.DFCXAgentIR{
		Flows: []ir.DFCXFlowModel{
			{
				FlowID: "flow-1",
				Pages: []ir.DFCXPageModel{
					{
						PageID: "page-1",
						PageData: map[string]interface{}{
							"form": map[string]interface{}{
								"parameters": []interface{}{
									map[string]interface{}{"displayName": "account_number"},
								},
							},
						},
					},
				},
			},
		},
	}
	params := ext.ExtractParameters(agentIR)
	assert.Contains(t, params, "account_number")
}
