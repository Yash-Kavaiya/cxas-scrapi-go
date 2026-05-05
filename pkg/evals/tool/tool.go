// Package tool provides ToolEvals: structured assertions on tool execution results.
package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/tools"
)

// Assertion defines a check on tool execution output.
type Assertion struct {
	// FieldPath is a dot-separated path into the output (e.g. "result.status").
	FieldPath string `yaml:"field_path"`
	// Operator is one of: equals, contains, exists, not_exists.
	Operator string `yaml:"operator"`
	// Value is the expected value.
	Value interface{} `yaml:"value,omitempty"`
}

// ToolTestCase defines a single tool invocation + assertions.
type ToolTestCase struct {
	ID         string                 `yaml:"id"`
	AppName    string                 `yaml:"app_name"`
	Tool       string                 `yaml:"tool,omitempty"`
	ToolsetRef *tools.ToolsetToolRef  `yaml:"toolset_ref,omitempty"`
	Input      map[string]interface{} `yaml:"input"`
	Assertions []Assertion            `yaml:"assertions"`
}

// ToolTestResult holds the outcome of a single tool test case.
type ToolTestResult struct {
	CaseID  string
	Passed  bool
	Reason  string
}

// ToolRunResult aggregates all tool test case results.
type ToolRunResult struct {
	Total   int
	Passed  int
	Failed  int
	Results []ToolTestResult
}

// ToolExecutor runs a tool and returns the raw output.
type ToolExecutor func(ctx context.Context, req tools.ExecuteToolRequest) (*tools.ExecuteToolResponse, error)

// RunToolTests executes all tool test cases and checks assertions.
func RunToolTests(ctx context.Context, cases []ToolTestCase, executor ToolExecutor) (*ToolRunResult, error) {
	rr := &ToolRunResult{}
	for _, tc := range cases {
		req := tools.ExecuteToolRequest{
			AppName:     tc.AppName,
			Tool:        tc.Tool,
			ToolsetTool: tc.ToolsetRef,
			Input:       tc.Input,
		}
		resp, err := executor(ctx, req)
		tr := ToolTestResult{CaseID: tc.ID}
		if err != nil {
			tr.Passed = false
			tr.Reason = fmt.Sprintf("execution error: %v", err)
		} else {
			// Convert response to map for assertion checking.
			raw := map[string]interface{}{"result": resp.Result}
			allPassed := true
			for _, a := range tc.Assertions {
				if !checkAssertion(raw, a) {
					allPassed = false
					tr.Reason = fmt.Sprintf("assertion failed: %s %s %v", a.FieldPath, a.Operator, a.Value)
					break
				}
			}
			tr.Passed = allPassed
		}
		rr.Results = append(rr.Results, tr)
		rr.Total++
		if tr.Passed {
			rr.Passed++
		} else {
			rr.Failed++
		}
	}
	return rr, nil
}

// checkAssertion evaluates a single assertion against the output map.
func checkAssertion(m map[string]interface{}, a Assertion) bool {
	v := getPath(m, a.FieldPath)
	switch a.Operator {
	case "exists":
		return v != nil
	case "not_exists":
		return v == nil
	case "equals":
		return fmt.Sprintf("%v", v) == fmt.Sprintf("%v", a.Value)
	case "contains":
		str, ok := v.(string)
		if !ok {
			return false
		}
		expected, _ := a.Value.(string)
		return strings.Contains(str, expected)
	}
	return false
}

// getPath walks a dot-separated path in a nested map.
func getPath(m map[string]interface{}, path string) interface{} {
	if path == "" {
		return nil
	}
	parts := strings.SplitN(path, ".", 2)
	v, ok := m[parts[0]]
	if !ok {
		return nil
	}
	if len(parts) == 1 {
		return v
	}
	sub, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return getPath(sub, parts[1])
}
