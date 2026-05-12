// Package gemini provides a rate-limited client for Google Vertex AI Gemini
// model inference. It supports synchronous generation with structured output,
// exponential backoff retry on quota errors, and configurable concurrency limits.
package gemini

// GenerateRequest holds parameters for a single generation call.
type GenerateRequest struct {
	Prompt         string
	SystemPrompt   string
	ModelName      string      // override default model
	ResponseSchema interface{} // if set, request JSON output conforming to this schema
}

// GenerateResponse holds the result of a generation call.
type GenerateResponse struct {
	Text string
	Raw  map[string]interface{}
}
