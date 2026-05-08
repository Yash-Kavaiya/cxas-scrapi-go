// Package gemini provides a Vertex AI Gemini client with a semaphore-based
// concurrency limiter and exponential back-off for rate-limit retries.
// It is used internally by the SimulationEvals engine to generate LLM-driven
// multi-turn conversations against a CXAS app.
package gemini
