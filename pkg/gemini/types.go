// Package gemini provides a client for generating text using Google's Gemini models via Vertex AI.
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
