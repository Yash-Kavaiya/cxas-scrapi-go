// Package ir defines the intermediate representation (IR) data models used
// during DFCX → CXAS migration.
package ir

// MigrationStatus represents the lifecycle state of a migration artifact.
type MigrationStatus string

const (
	StatusCompiled  MigrationStatus = "COMPILED"
	StatusGenerated MigrationStatus = "GENERATED"
	StatusDeployed  MigrationStatus = "DEPLOYED"
	StatusFailed    MigrationStatus = "FAILED"
	StatusError     MigrationStatus = "ERROR"
	StatusPending   MigrationStatus = "PENDING"
)

// DFCXPageModel represents a single Dialogflow CX page.
type DFCXPageModel struct {
	PageID   string                 `json:"page_id"`
	PageData map[string]interface{} `json:"page_data"`
}

// DFCXFlowModel represents a Dialogflow CX flow with its pages.
type DFCXFlowModel struct {
	FlowID   string                 `json:"flow_id"`
	FlowData map[string]interface{} `json:"flow_data"`
	Pages    []DFCXPageModel        `json:"pages"`
}

// DFCXAgentIR is the intermediate representation of a Dialogflow CX agent,
// capturing all resources needed for migration.
type DFCXAgentIR struct {
	Name                       string                            `json:"name"`
	DisplayName                string                            `json:"display_name"`
	DefaultLanguageCode        string                            `json:"default_language_code"`
	SupportedLanguageCodes     []string                          `json:"supported_language_codes"`
	TimeZone                   string                            `json:"time_zone,omitempty"`
	Description                string                            `json:"description,omitempty"`
	StartFlow                  string                            `json:"start_flow,omitempty"`
	StartPlaybook              string                            `json:"start_playbook,omitempty"`
	Intents                    []map[string]interface{}          `json:"intents"`
	Tools                      []map[string]interface{}          `json:"tools"`
	EntityTypes                []map[string]interface{}          `json:"entity_types"`
	Webhooks                   []map[string]interface{}          `json:"webhooks"`
	Flows                      []DFCXFlowModel                   `json:"flows"`
	Playbooks                  []map[string]interface{}          `json:"playbooks"`
	TestCases                  []map[string]interface{}          `json:"test_cases"`
	GenerativeSettings         map[string]map[string]interface{} `json:"generative_settings"`
	PlaybookGenerativeSettings map[string]interface{}            `json:"playbook_generative_settings,omitempty"`
	Generators                 []map[string]interface{}          `json:"generators"`
	AgentTransitionRouteGroups []map[string]interface{}          `json:"agent_transition_route_groups"`
}

// IRMetadata holds target-side metadata for the migration run.
type IRMetadata struct {
	AppName         string `json:"app_name"`
	AppID           string `json:"app_id,omitempty"`
	AppResourceName string `json:"app_resource_name,omitempty"`
	DefaultModel    string `json:"default_model"`
}

// IRTool represents a tool artifact in the migration IR.
// Type is one of "TOOLSET", "TOOL", or "PYTHON".
type IRTool struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Payload      map[string]interface{} `json:"payload"`
	OperationIDs []string               `json:"operation_ids"`
	Status       MigrationStatus        `json:"status"`
}

// IRAgent represents a generative agent artifact in the migration IR.
// Type is "FLOW" or "PLAYBOOK".
type IRAgent struct {
	Type          string                   `json:"type"`
	DisplayName   string                   `json:"display_name"`
	Description   string                   `json:"description,omitempty"`
	Instruction   string                   `json:"instruction"`
	Tools         []string                 `json:"tools"`
	Toolsets      []map[string]interface{} `json:"toolsets"`
	ModelSettings map[string]interface{}   `json:"model_settings"`
	RawData       map[string]interface{}   `json:"raw_data,omitempty"`
	Blueprint     map[string]interface{}   `json:"blueprint,omitempty"`
	Callbacks     map[string]interface{}   `json:"callbacks,omitempty"`
	Status        MigrationStatus          `json:"status"`
	ResourceName  string                   `json:"resource_name,omitempty"`
}

// MigrationIR is the top-level workspace state for a migration run.
type MigrationIR struct {
	Metadata     IRMetadata             `json:"metadata"`
	Parameters   map[string]interface{} `json:"parameters"`
	Tools        map[string]IRTool      `json:"tools"`
	Agents       map[string]IRAgent     `json:"agents"`
	RoutingEdges []map[string]interface{} `json:"routing_edges"`
	TestCases    map[string]interface{} `json:"test_cases"`
	TestRuns     map[string]interface{} `json:"test_runs"`
}
