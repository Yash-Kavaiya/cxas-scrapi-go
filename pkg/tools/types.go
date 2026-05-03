// Package tools provides a client for CXAS tool and toolset management.
package tools

// Tool represents a CXAS tool resource (OpenAPI or function-based).
type Tool struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	OpenAPISpec map[string]interface{} `json:"openApiSpec,omitempty"`
}

// Toolset represents a CXAS toolset resource (collection of tools sharing an OpenAPI spec).
type Toolset struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	OpenAPIYAML string `json:"openApiYaml,omitempty"`
}

// ExecuteToolRequest holds parameters for invoking a tool directly.
type ExecuteToolRequest struct {
	// AppName is the parent app resource name.
	AppName string
	// Tool is the resource name of an atomic tool (mutually exclusive with ToolsetTool).
	Tool string
	// ToolsetTool specifies a toolset + tool ID (mutually exclusive with Tool).
	ToolsetTool *ToolsetToolRef
	// Input is the JSON-serializable input for the tool.
	Input map[string]interface{}
}

// ToolsetToolRef identifies a tool within a toolset.
type ToolsetToolRef struct {
	Toolset string `json:"toolset"`
	ToolID  string `json:"toolId"`
}

// ExecuteToolResponse holds the raw response from executing a tool.
type ExecuteToolResponse struct {
	Result interface{} `json:"result"`
}
