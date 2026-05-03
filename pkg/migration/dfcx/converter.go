package dfcx

import (
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/ir"
)

// PlaybookConverter converts DFCX playbooks to CXAS IRAgent entries.
type PlaybookConverter struct {
	DefaultModel string
}

// NewPlaybookConverter creates a converter with the given default model.
func NewPlaybookConverter(defaultModel string) *PlaybookConverter {
	if defaultModel == "" {
		defaultModel = "gemini-2.0-flash"
	}
	return &PlaybookConverter{DefaultModel: defaultModel}
}

// ConvertPlaybook converts a single DFCX playbook map into an IRAgent.
func (c *PlaybookConverter) ConvertPlaybook(playbook map[string]interface{}) ir.IRAgent {
	displayName, _ := playbook["displayName"].(string)
	goal, _ := playbook["goal"].(string)
	name, _ := playbook["name"].(string)

	instruction := buildPlaybookInstruction(displayName, goal, playbook)

	return ir.IRAgent{
		Type:        "PLAYBOOK",
		DisplayName: displayName,
		Instruction: instruction,
		Tools:       []string{},
		Toolsets:    []map[string]interface{}{},
		ModelSettings: map[string]interface{}{
			"model": c.DefaultModel,
		},
		RawData:      playbook,
		Status:       ir.StatusCompiled,
		ResourceName: name,
	}
}

func buildPlaybookInstruction(name, goal string, playbook map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<agent name=\"%s\">\n", name))
	if goal != "" {
		sb.WriteString(fmt.Sprintf("  <goal>%s</goal>\n", goal))
	}
	if steps, ok := playbook["steps"].([]interface{}); ok {
		sb.WriteString("  <steps>\n")
		for _, s := range steps {
			if step, ok := s.(map[string]interface{}); ok {
				text, _ := step["text"].(string)
				sb.WriteString(fmt.Sprintf("    <step>%s</step>\n", text))
			}
		}
		sb.WriteString("  </steps>\n")
	}
	sb.WriteString("</agent>")
	return sb.String()
}

// ToolConverter converts DFCX webhooks to CXAS IRTool (OpenAPI toolset) entries.
type ToolConverter struct{}

// NewToolConverter creates a ToolConverter.
func NewToolConverter() *ToolConverter {
	return &ToolConverter{}
}

// ConvertWebhook converts a DFCX webhook into an IRTool.
func (c *ToolConverter) ConvertWebhook(webhook map[string]interface{}) ir.IRTool {
	displayName, _ := webhook["displayName"].(string)
	name, _ := webhook["name"].(string)

	// Build a minimal OpenAPI schema for the webhook
	toolID := sanitizeID(displayName)
	operationID := sanitizeID(displayName) + "_call"

	payload := map[string]interface{}{
		"openapi": "3.0.0",
		"info":    map[string]interface{}{"title": displayName, "version": "1.0.0"},
		"paths": map[string]interface{}{
			"/invoke": map[string]interface{}{
				"post": map[string]interface{}{
					"operationId": operationID,
					"summary":     fmt.Sprintf("Invoke %s webhook", displayName),
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{"type": "object"},
							},
						},
					},
				},
			},
		},
	}

	return ir.IRTool{
		ID:           toolID,
		Name:         name,
		Type:         "TOOLSET",
		Payload:      payload,
		OperationIDs: []string{operationID},
		Status:       ir.StatusCompiled,
	}
}

// ParameterExtractor extracts DFCX parameters for CXAS variable declarations.
type ParameterExtractor struct{}

// NewParameterExtractor creates a ParameterExtractor.
func NewParameterExtractor() *ParameterExtractor { return &ParameterExtractor{} }

// ExtractParameters extracts variable declarations from DFCX flows and playbooks.
func (e *ParameterExtractor) ExtractParameters(agentIR *ir.DFCXAgentIR) map[string]interface{} {
	params := map[string]interface{}{}
	for _, flow := range agentIR.Flows {
		for _, page := range flow.Pages {
			if form, ok := page.PageData["form"].(map[string]interface{}); ok {
				if parameters, ok := form["parameters"].([]interface{}); ok {
					for _, p := range parameters {
						if param, ok := p.(map[string]interface{}); ok {
							name, _ := param["displayName"].(string)
							if name != "" {
								params[name] = map[string]interface{}{
									"type":    "STRING",
									"default": "",
								}
							}
						}
					}
				}
			}
		}
	}
	return params
}

func sanitizeID(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	return sb.String()
}
