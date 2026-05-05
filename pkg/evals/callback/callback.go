// Package callback provides CallbackEvals: structured test runners for callback handlers.
package callback

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/callbacks"
)

// CallbackTestCase defines a single callback invocation + expected outcome.
type CallbackTestCase struct {
	ID          string                 `yaml:"id"`
	AppName     string                 `yaml:"app_name"`
	CallbackID  string                 `yaml:"callback_id"`
	Input       map[string]interface{} `yaml:"input"`
	// ExpectedOutput is checked against the callback result if non-nil.
	ExpectedOutput map[string]interface{} `yaml:"expected_output,omitempty"`
}

// CallbackTestResult holds the outcome of a single callback test.
type CallbackTestResult struct {
	CaseID string
	Passed bool
	Reason string
}

// CallbackRunResult aggregates all callback test results.
type CallbackRunResult struct {
	Total   int
	Passed  int
	Failed  int
	Results []CallbackTestResult
}

// RunCallbackTests validates callback test cases.
// Since Go cannot exec() Python handlers in-process, this validates the callback
// resource exists and returns ErrNotSupported for actual execution.
func RunCallbackTests(ctx context.Context, cases []CallbackTestCase, cbClient *callbacks.Client) (*CallbackRunResult, error) {
	rr := &CallbackRunResult{}
	for _, tc := range cases {
		tr := CallbackTestResult{CaseID: tc.ID}

		name := fmt.Sprintf("%s/callbacks/%s", tc.AppName, tc.CallbackID)
		_, err := cbClient.GetCallback(ctx, name)
		if err != nil {
			tr.Passed = false
			tr.Reason = fmt.Sprintf("callback not found: %v", err)
		} else {
			// Actual execution is not supported in Go — see callbacks.ErrNotSupported.
			tr.Passed = false
			tr.Reason = callbacks.ErrNotSupported.Error()
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
