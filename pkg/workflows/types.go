// Package workflows provides data models for workflow-type CXAS agents.
package workflows

// VariableMetadata describes a variable used by a workflow step.
type VariableMetadata struct {
	Variable    string `json:"variable"`
	Description string `json:"description,omitempty"`
	IsRequired  bool   `json:"is_required,omitempty"`
}

// Reasoning defines an LLM reasoning step in a workflow.
type Reasoning struct {
	Prompt          string             `json:"prompt"`
	OutputVariables []VariableMetadata `json:"output_variables,omitempty"`
}

// ToMap converts Reasoning to a map for JSON API payloads.
func (r Reasoning) ToMap() map[string]interface{} {
	m := map[string]interface{}{"prompt": r.Prompt}
	if len(r.OutputVariables) > 0 {
		vars := make([]map[string]interface{}, len(r.OutputVariables))
		for i, v := range r.OutputVariables {
			vars[i] = map[string]interface{}{"variable": v.Variable, "description": v.Description, "is_required": v.IsRequired}
		}
		m["output_variables"] = vars
	}
	return m
}

// PythonCode defines a Python function step.
type PythonCode struct {
	Code string `json:"code"`
}

// ToMap converts PythonCode to a map for JSON API payloads.
func (p PythonCode) ToMap() map[string]interface{} {
	return map[string]interface{}{"code": p.Code}
}

// WorkflowAction is a step action — either Reasoning or PythonCode (mutually exclusive).
type WorkflowAction struct {
	Reasoning  *Reasoning  `json:"reasoning,omitempty"`
	PythonCode *PythonCode `json:"python_code,omitempty"`
}

// ToMap converts WorkflowAction to a map for JSON API payloads.
func (a WorkflowAction) ToMap() map[string]interface{} {
	m := map[string]interface{}{}
	if a.Reasoning != nil {
		m["reasoning"] = a.Reasoning.ToMap()
	}
	if a.PythonCode != nil {
		m["python_code"] = a.PythonCode.ToMap()
	}
	return m
}

// Transition defines an outgoing edge from a workflow step.
type Transition struct {
	WorkflowStepID string `json:"workflow_step_id"`
	ConditionType  string `json:"condition_type,omitempty"`
}

// ToMap converts Transition to a map.
func (t Transition) ToMap() map[string]interface{} {
	m := map[string]interface{}{"workflow_step_id": t.WorkflowStepID}
	if t.ConditionType != "" {
		m["condition_type"] = t.ConditionType
	}
	return m
}

// WorkflowStep is a single step in a workflow.
type WorkflowStep struct {
	ID          string         `json:"id"`
	DisplayName string         `json:"display_name"`
	Action      WorkflowAction `json:"action"`
	Transitions []Transition   `json:"transitions,omitempty"`
}

// ToMap converts WorkflowStep to a map.
func (s WorkflowStep) ToMap() map[string]interface{} {
	transitions := make([]map[string]interface{}, len(s.Transitions))
	for i, t := range s.Transitions {
		transitions[i] = t.ToMap()
	}
	return map[string]interface{}{
		"id":           s.ID,
		"display_name": s.DisplayName,
		"action":       s.Action.ToMap(),
		"transitions":  transitions,
	}
}

// WorkflowAgent is the top-level workflow agent configuration.
type WorkflowAgent struct {
	WorkflowSteps []WorkflowStep `json:"workflow_steps,omitempty"`
	RootStepID    string         `json:"root_step_id,omitempty"`
}

// AddStep appends a step to the workflow.
func (w *WorkflowAgent) AddStep(step WorkflowStep) {
	w.WorkflowSteps = append(w.WorkflowSteps, step)
}

// ToMap converts WorkflowAgent to a map for the CES API payload.
func (w WorkflowAgent) ToMap() map[string]interface{} {
	steps := make([]map[string]interface{}, len(w.WorkflowSteps))
	for i, s := range w.WorkflowSteps {
		steps[i] = s.ToMap()
	}
	return map[string]interface{}{
		"workflow_steps": steps,
		"root_step_id":   w.RootStepID,
	}
}
