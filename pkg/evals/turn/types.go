// Package turn provides TurnEvals: expectation types and YAML config models.
package turn

import "gopkg.in/yaml.v3"

// TurnOperator defines which signal a turn expectation checks.
type TurnOperator string

const (
	OperatorContains      TurnOperator = "contains"
	OperatorEquals        TurnOperator = "equals"
	OperatorToolCalled    TurnOperator = "tool_called"
	OperatorToolInput     TurnOperator = "tool_input"
	OperatorToolOutput    TurnOperator = "tool_output"
	OperatorNoToolsCalled TurnOperator = "no_tools_called"
	OperatorAgentTransfer TurnOperator = "agent_transfer"
)

// TurnExpectation is a structured expectation for a single turn.
type TurnExpectation struct {
	Operator TurnOperator `yaml:"operator"`
	// Value is the expected value (string, map, etc.).
	Value     interface{} `yaml:"value,omitempty"`
	ToolName  string      `yaml:"tool_name,omitempty"`
	FieldPath string      `yaml:"field_path,omitempty"`
}

// TurnExpectationOrStr holds either a plain LLM-evaluated string expectation
// or a structured TurnExpectation — matches Python's Union[str, TurnExpectation].
type TurnExpectationOrStr struct {
	Str        string
	Structured *TurnExpectation
}

// UnmarshalYAML handles the union: scalar → LLM expectation string, map → TurnExpectation.
func (e *TurnExpectationOrStr) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		e.Str = value.Value
		return nil
	}
	var s TurnExpectation
	if err := value.Decode(&s); err != nil {
		return err
	}
	e.Structured = &s
	return nil
}

// IsLLM returns true when the expectation should be evaluated by an LLM judge.
func (e *TurnExpectationOrStr) IsLLM() bool { return e.Str != "" && e.Structured == nil }

// Turn represents a single conversation turn in a test case.
type Turn struct {
	// Input is the user utterance.
	Input string `yaml:"input"`
	// ExpectedAgent is the agent that should handle this turn (optional).
	ExpectedAgent string `yaml:"expected_agent,omitempty"`
	// Expectations is a list of expectations for the turn response.
	Expectations []TurnExpectationOrStr `yaml:"expectations,omitempty"`
	// DependsOn lists IDs of test cases that must run first.
	DependsOn []string `yaml:"depends_on,omitempty"`
}

// TestCase represents a full multi-turn conversation test.
type TestCase struct {
	// ID uniquely identifies this test case for dependency resolution.
	ID          string `yaml:"id"`
	Description string `yaml:"description,omitempty"`
	Turns       []Turn `yaml:"turns"`
	// Tags allows filtering test cases.
	Tags []string `yaml:"tags,omitempty"`
}

// GlobalConfig holds default settings applied to all test cases in a file.
type GlobalConfig struct {
	DefaultAgent string   `yaml:"default_agent,omitempty"`
	Tags         []string `yaml:"tags,omitempty"`
}

// TurnEvalFile is the top-level YAML structure.
// Supports both "conversations" and "tests" as the list key.
type TurnEvalFile struct {
	GlobalConfig GlobalConfig `yaml:"global_config,omitempty"`
	Tests        []TestCase   `yaml:"tests,omitempty"`
	Conversations []TestCase  `yaml:"conversations,omitempty"`
}

// AllTests returns the combined test list regardless of which key was used.
func (f *TurnEvalFile) AllTests() []TestCase {
	if len(f.Tests) > 0 {
		return f.Tests
	}
	return f.Conversations
}
