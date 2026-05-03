// Package sessions provides clients for CXAS text and audio session execution.
package sessions

// SessionInput holds the input for a RunSession call.
// Exactly one of Text, Event, Blob, DTMF, or ToolResponses should be set.
type SessionInput struct {
	Text          string                 `json:"text,omitempty"`
	Event         string                 `json:"event,omitempty"`
	EventVars     map[string]interface{} `json:"eventVars,omitempty"`
	Blob          []byte                 `json:"blob,omitempty"`
	BlobMimeType  string                 `json:"blobMimeType,omitempty"`
	DTMF          string                 `json:"dtmf,omitempty"`
	ToolResponses []ToolResponse         `json:"toolResponses,omitempty"`
}

// ToolResponse holds the result of a tool invocation.
type ToolResponse struct {
	ToolName string      `json:"toolName"`
	Output   interface{} `json:"output"`
}

// SessionOutput holds a structured response parsed from a session turn.
type SessionOutput struct {
	// Text is the concatenated plain-text response.
	Text string
	// ToolCalls contains any tool invocations requested by the agent.
	ToolCalls []ToolCall
	// AgentTransfer is set when the agent hands off to another agent.
	AgentTransfer string
	// SessionEnded is true when the session has been terminated.
	SessionEnded bool
	// Raw is the full API response for callers that need raw data.
	Raw map[string]interface{}
}

// ToolCall represents a single tool invocation requested by the agent.
type ToolCall struct {
	ToolName  string                 `json:"toolName"`
	ToolInput map[string]interface{} `json:"toolInput"`
}

// RunSessionRequest holds all parameters for a text-mode session call.
type RunSessionRequest struct {
	AppName   string
	SessionID string
	Input     SessionInput
}

// BidiConfig holds configuration for a WebSocket BidiSession.
type BidiConfig struct {
	AppName   string
	Location  string
	SessionID string
	// AudioSampleRate is the sample rate in Hz (default: 16000).
	AudioSampleRate int
}
