// Package linter provides a linting engine for CXAS app repositories.
package linter

import "fmt"

// Severity represents the importance level of a lint finding.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
	SeverityOff
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityOff:
		return "off"
	}
	return "unknown"
}

// LintResult holds a single finding from a rule check.
type LintResult struct {
	File       string
	RuleID     string
	Severity   Severity
	Message    string
	Line       int
	FixSuggest string
}

func (r LintResult) String() string {
	return fmt.Sprintf("[%s] %s:%d %s: %s", r.Severity, r.File, r.Line, r.RuleID, r.Message)
}

// LintContext carries shared context available to all rules during a run.
type LintContext struct {
	// AppRoot is the path to the app directory being linted.
	AppRoot string
	// Severities overrides default severity per rule ID.
	Severities map[string]Severity
	// ParsedContent is the pre-parsed JSON content, populated by LintFile.
	ParsedContent map[string]interface{}
}

// Rule is the interface every lint rule must implement.
type Rule interface {
	ID() string
	Name() string
	Category() string
	DefaultSeverity() Severity
	Check(filePath, content string, ctx *LintContext) []LintResult
}

// RuleRegistry holds all registered rules.
type RuleRegistry struct {
	rules []Rule
}

// Register adds a rule to the registry.
func (r *RuleRegistry) Register(rule Rule) { r.rules = append(r.rules, rule) }

// All returns all registered rules.
func (r *RuleRegistry) All() []Rule { return r.rules }

// defaultRules are the built-in lint rules.
var defaultRules = []Rule{
	&missingDisplayNameRule{},
	&emptyInstructionRule{},
	&longInstructionRule{},
	&missingDescriptionRule{},
	&evalFileStructureRule{},
}

// NewDefaultRegistry creates a RuleRegistry pre-loaded with all built-in rules.
func NewDefaultRegistry() *RuleRegistry {
	r := &RuleRegistry{}
	for _, rule := range defaultRules {
		r.Register(rule)
	}
	return r
}
