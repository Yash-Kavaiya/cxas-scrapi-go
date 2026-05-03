package workflows_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/workflows"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowAgent_ToMap(t *testing.T) {
	agent := workflows.WorkflowAgent{RootStepID: "step-1"}
	agent.AddStep(workflows.WorkflowStep{
		ID:          "step-1",
		DisplayName: "First Step",
		Action: workflows.WorkflowAction{
			Reasoning: &workflows.Reasoning{Prompt: "Think carefully"},
		},
	})
	m := agent.ToMap()
	assert.Equal(t, "step-1", m["root_step_id"])
	steps := m["workflow_steps"].([]map[string]interface{})
	assert.Len(t, steps, 1)
	assert.Equal(t, "First Step", steps[0]["display_name"])
}

func TestPythonCode_ToMap(t *testing.T) {
	pc := workflows.PythonCode{Code: "def foo(): pass"}
	m := pc.ToMap()
	assert.Equal(t, "def foo(): pass", m["code"])
}

func TestTransition_ToMap_WithCondition(t *testing.T) {
	tr := workflows.Transition{WorkflowStepID: "step-2", ConditionType: "SUCCESS"}
	m := tr.ToMap()
	assert.Equal(t, "step-2", m["workflow_step_id"])
	assert.Equal(t, "SUCCESS", m["condition_type"])
}
