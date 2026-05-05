package turn_test

import (
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/evals/turn"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const yamlTests = `
global_config:
  default_agent: main-agent
  tags: [smoke]

tests:
  - id: greet-test
    description: basic greeting
    turns:
      - input: "Hello"
        expectations:
          - operator: contains
            value: "Hello"
          - operator: no_tools_called
  - id: tool-test
    description: tool is called
    turns:
      - input: "Search for weather"
        expectations:
          - operator: tool_called
            value: search_tool
`

func TestParseBytes_TestsKey(t *testing.T) {
	f, err := turn.ParseBytes([]byte(yamlTests))
	require.NoError(t, err)
	tests := f.AllTests()
	require.Len(t, tests, 2)
	assert.Equal(t, "greet-test", tests[0].ID)
	assert.Equal(t, "tool-test", tests[1].ID)
}

const yamlConversations = `
conversations:
  - id: conv-1
    turns:
      - input: "Hi"
        expectations:
          - "The agent greets the user"
`

func TestParseBytes_ConversationsKey(t *testing.T) {
	f, err := turn.ParseBytes([]byte(yamlConversations))
	require.NoError(t, err)
	tests := f.AllTests()
	require.Len(t, tests, 1)
	assert.Equal(t, "conv-1", tests[0].ID)
	// LLM expectation (string form)
	require.Len(t, tests[0].Turns[0].Expectations, 1)
	assert.True(t, tests[0].Turns[0].Expectations[0].IsLLM())
}

func TestApplyGlobalConfig_MergesTags(t *testing.T) {
	f, _ := turn.ParseBytes([]byte(yamlTests))
	turn.ApplyGlobalConfig(f)
	for _, tc := range f.AllTests() {
		found := false
		for _, tag := range tc.Tags {
			if tag == "smoke" {
				found = true
				break
			}
		}
		assert.True(t, found, "global tag 'smoke' should be merged into test %s", tc.ID)
	}
}

func TestRunTests_ContainsPass(t *testing.T) {
	runner := func(ctx context.Context, appName, sessID, text string) (*sessions.SessionOutput, error) {
		return &sessions.SessionOutput{Text: "Hello there!"}, nil
	}

	tests := []turn.TestCase{
		{
			ID: "t1",
			Turns: []turn.Turn{
				{
					Input: "Hello",
					Expectations: []turn.TurnExpectationOrStr{
						{Structured: &turn.TurnExpectation{Operator: turn.OperatorContains, Value: "Hello"}},
					},
				},
			},
		},
	}

	result, err := turn.RunTests(context.Background(), tests, "projects/p/locations/l/apps/a", runner)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 0, result.Failed)
}

func TestRunTests_ContainsFail(t *testing.T) {
	runner := func(ctx context.Context, appName, sessID, text string) (*sessions.SessionOutput, error) {
		return &sessions.SessionOutput{Text: "Goodbye!"}, nil
	}

	tests := []turn.TestCase{
		{
			ID: "t1",
			Turns: []turn.Turn{
				{
					Input: "Hello",
					Expectations: []turn.TurnExpectationOrStr{
						{Structured: &turn.TurnExpectation{Operator: turn.OperatorContains, Value: "Hello"}},
					},
				},
			},
		},
	}

	result, err := turn.RunTests(context.Background(), tests, "projects/p/locations/l/apps/a", runner)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Passed)
	assert.Equal(t, 1, result.Failed)
}

func TestRunTests_ToolCalled(t *testing.T) {
	runner := func(ctx context.Context, appName, sessID, text string) (*sessions.SessionOutput, error) {
		return &sessions.SessionOutput{
			ToolCalls: []sessions.ToolCall{{ToolName: "search_tool"}},
		}, nil
	}

	tests := []turn.TestCase{
		{
			ID: "t1",
			Turns: []turn.Turn{
				{
					Input: "Search",
					Expectations: []turn.TurnExpectationOrStr{
						{Structured: &turn.TurnExpectation{Operator: turn.OperatorToolCalled, Value: "search_tool"}},
					},
				},
			},
		},
	}

	result, err := turn.RunTests(context.Background(), tests, "projects/p/locations/l/apps/a", runner)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Passed)
}

func TestTopoSort_DetectsCycle(t *testing.T) {
	tests := []turn.TestCase{
		{ID: "a", Turns: []turn.Turn{{Input: "x", DependsOn: []string{"b"}}}},
		{ID: "b", Turns: []turn.Turn{{Input: "y", DependsOn: []string{"a"}}}},
	}
	_, err := turn.RunTests(context.Background(), tests, "app", func(ctx context.Context, a, s, t string) (*sessions.SessionOutput, error) {
		return &sessions.SessionOutput{}, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}
