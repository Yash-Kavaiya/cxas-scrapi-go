// Package agents provides a client for CXAS Agent resource CRUD operations.
package agents

// AgentType identifies which kind of agent to create.
type AgentType string

const (
	AgentTypeLLM      AgentType = "llm"
	AgentTypeDFCX     AgentType = "dfcx"
	AgentTypeWorkflow AgentType = "workflow"
)

// Agent represents a CX Agent Studio agent resource.
type Agent struct {
	Name                  string                 `json:"name"`
	DisplayName           string                 `json:"displayName"`
	Instruction           string                 `json:"instruction,omitempty"`
	ModelSettings         map[string]interface{} `json:"modelSettings,omitempty"`
	LlmAgent              map[string]interface{} `json:"llmAgent,omitempty"`
	WorkflowAgent         map[string]interface{} `json:"workflowAgent,omitempty"`
	RemoteDialogflowAgent map[string]interface{} `json:"remoteDialogflowAgent,omitempty"`
}

// CreateAgentRequest holds parameters for creating a new Agent.
type CreateAgentRequest struct {
	DisplayName       string
	AgentID           string
	AgentType         AgentType
	Model             string
	Instruction       string
	DFCXAgentResource string
	WorkflowConfig    map[string]interface{}
	Extra             map[string]interface{}
}
