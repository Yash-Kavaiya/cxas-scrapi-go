package linter

import (
	"encoding/json"
	"strings"
)

// missingDisplayNameRule checks that every resource has a displayName.
type missingDisplayNameRule struct{}

func (r *missingDisplayNameRule) ID() string              { return "CXL001" }
func (r *missingDisplayNameRule) Name() string            { return "missing-display-name" }
func (r *missingDisplayNameRule) Category() string        { return "structure" }
func (r *missingDisplayNameRule) DefaultSeverity() Severity { return SeverityError }

func (r *missingDisplayNameRule) Check(filePath, content string, ctx *LintContext) []LintResult {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		return nil
	}
	if dn, ok := m["displayName"].(string); !ok || strings.TrimSpace(dn) == "" {
		return []LintResult{{
			File:       filePath,
			RuleID:     r.ID(),
			Severity:   severityFor(ctx, r.ID(), r.DefaultSeverity()),
			Message:    "resource is missing a displayName",
			FixSuggest: `add "displayName": "<name>" to the resource JSON`,
		}}
	}
	return nil
}

// emptyInstructionRule checks that LLM agents have a non-empty instruction.
type emptyInstructionRule struct{}

func (r *emptyInstructionRule) ID() string              { return "CXL002" }
func (r *emptyInstructionRule) Name() string            { return "empty-instruction" }
func (r *emptyInstructionRule) Category() string        { return "instructions" }
func (r *emptyInstructionRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *emptyInstructionRule) Check(filePath, content string, ctx *LintContext) []LintResult {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		return nil
	}
	if _, isLLM := m["llmAgent"]; !isLLM {
		return nil
	}
	inst, _ := m["instruction"].(string)
	if strings.TrimSpace(inst) == "" {
		return []LintResult{{
			File:     filePath,
			RuleID:   r.ID(),
			Severity: severityFor(ctx, r.ID(), r.DefaultSeverity()),
			Message:  "LLM agent has an empty instruction",
		}}
	}
	return nil
}

// longInstructionRule warns when an instruction exceeds 8000 characters.
type longInstructionRule struct{}

func (r *longInstructionRule) ID() string              { return "CXL003" }
func (r *longInstructionRule) Name() string            { return "long-instruction" }
func (r *longInstructionRule) Category() string        { return "instructions" }
func (r *longInstructionRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *longInstructionRule) Check(filePath, content string, ctx *LintContext) []LintResult {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		return nil
	}
	inst, _ := m["instruction"].(string)
	if len(inst) > 8000 {
		return []LintResult{{
			File:       filePath,
			RuleID:     r.ID(),
			Severity:   severityFor(ctx, r.ID(), r.DefaultSeverity()),
			Message:    "agent instruction exceeds 8000 characters — consider splitting into multiple agents",
			FixSuggest: "split complex instructions across agents using routing",
		}}
	}
	return nil
}

// missingDescriptionRule warns when a tool or toolset has no description.
type missingDescriptionRule struct{}

func (r *missingDescriptionRule) ID() string              { return "CXL004" }
func (r *missingDescriptionRule) Name() string            { return "missing-description" }
func (r *missingDescriptionRule) Category() string        { return "structure" }
func (r *missingDescriptionRule) DefaultSeverity() Severity { return SeverityInfo }

func (r *missingDescriptionRule) Check(filePath, content string, ctx *LintContext) []LintResult {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		return nil
	}
	// Only check tools/toolsets (files with openApiSpec or openApiYaml).
	if _, hasTool := m["openApiSpec"]; !hasTool {
		if _, hasToolset := m["openApiYaml"]; !hasToolset {
			return nil
		}
	}
	desc, _ := m["description"].(string)
	if strings.TrimSpace(desc) == "" {
		return []LintResult{{
			File:     filePath,
			RuleID:   r.ID(),
			Severity: severityFor(ctx, r.ID(), r.DefaultSeverity()),
			Message:  "tool/toolset is missing a description",
		}}
	}
	return nil
}

// evalFileStructureRule checks that evaluation YAML files have at least one test.
type evalFileStructureRule struct{}

func (r *evalFileStructureRule) ID() string              { return "CXL005" }
func (r *evalFileStructureRule) Name() string            { return "eval-file-structure" }
func (r *evalFileStructureRule) Category() string        { return "evals" }
func (r *evalFileStructureRule) DefaultSeverity() Severity { return SeverityWarning }

func (r *evalFileStructureRule) Check(filePath, content string, ctx *LintContext) []LintResult {
	if !strings.HasSuffix(filePath, ".yaml") && !strings.HasSuffix(filePath, ".yml") {
		return nil
	}
	// Very lightweight check: file must contain "turns:" or "conversations:" or "tests:"
	if !strings.Contains(content, "turns:") && !strings.Contains(content, "conversations:") && !strings.Contains(content, "tests:") {
		return nil
	}
	if !strings.Contains(content, "input:") {
		return []LintResult{{
			File:     filePath,
			RuleID:   r.ID(),
			Severity: severityFor(ctx, r.ID(), r.DefaultSeverity()),
			Message:  "eval file appears to have no turns with 'input' fields",
		}}
	}
	return nil
}

// severityFor returns the configured severity for a rule, falling back to defaultSev.
func severityFor(ctx *LintContext, ruleID string, defaultSev Severity) Severity {
	if ctx == nil || ctx.Severities == nil {
		return defaultSev
	}
	if s, ok := ctx.Severities[ruleID]; ok {
		return s
	}
	return defaultSev
}
