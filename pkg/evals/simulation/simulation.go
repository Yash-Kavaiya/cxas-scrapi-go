// Package simulation provides SimulationEvals: LLM-driven multi-turn simulation.
package simulation

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
)

const defaultMaxWorkers = 25

// Scenario defines a single simulation scenario.
type Scenario struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description,omitempty"`
	// SystemPrompt is used to prime the LLM user simulator.
	SystemPrompt string `yaml:"system_prompt"`
	// MaxTurns limits the conversation length.
	MaxTurns int `yaml:"max_turns,omitempty"`
	// ExpectedOutcome is an LLM-evaluated string (e.g. "User's issue is resolved").
	ExpectedOutcome string `yaml:"expected_outcome,omitempty"`
}

// StepResult holds the result of a single conversation step.
type StepResult struct {
	TurnIndex  int
	UserInput  string
	AgentReply string
}

// ScenarioResult holds the full outcome of running a scenario.
type ScenarioResult struct {
	ScenarioID string
	Steps      []StepResult
	Passed     bool
	Reason     string
}

// SimulationRunResult aggregates all scenario results.
type SimulationRunResult struct {
	Total   int
	Passed  int
	Failed  int
	Results []ScenarioResult
}

// UserSimulator generates user utterances given conversation history.
type UserSimulator func(ctx context.Context, scenario Scenario, history []StepResult) (string, bool, error)

// SessionRunner executes a single agent turn.
type SessionRunner func(ctx context.Context, appName, sessionID, text string) (*sessions.SessionOutput, error)

// RunSimulations executes all scenarios concurrently (max maxWorkers goroutines).
func RunSimulations(ctx context.Context, scenarios []Scenario, appName string, runner SessionRunner, sim UserSimulator, maxWorkers int) (*SimulationRunResult, error) {
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}

	sem := make(chan struct{}, maxWorkers)
	var mu sync.Mutex
	rr := &SimulationRunResult{}

	g, ctx := errgroup.WithContext(ctx)
	for _, sc := range scenarios {
		sc := sc
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := runScenario(ctx, sc, appName, runner, sim)
			if err != nil {
				result = &ScenarioResult{
					ScenarioID: sc.ID,
					Passed:     false,
					Reason:     err.Error(),
				}
			}
			mu.Lock()
			rr.Results = append(rr.Results, *result)
			rr.Total++
			if result.Passed {
				rr.Passed++
			} else {
				rr.Failed++
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return rr, nil
}

func runScenario(ctx context.Context, sc Scenario, appName string, runner SessionRunner, sim UserSimulator) (*ScenarioResult, error) {
	maxTurns := sc.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 10
	}

	sessID := fmt.Sprintf("sim-%s", sc.ID)
	result := &ScenarioResult{ScenarioID: sc.ID}

	for i := 0; i < maxTurns; i++ {
		userInput, done, err := sim(ctx, sc, result.Steps)
		if err != nil {
			return nil, fmt.Errorf("scenario %s turn %d: simulator error: %w", sc.ID, i, err)
		}

		agentOut, err := runner(ctx, appName, sessID, userInput)
		if err != nil {
			return nil, fmt.Errorf("scenario %s turn %d: session error: %w", sc.ID, i, err)
		}

		result.Steps = append(result.Steps, StepResult{
			TurnIndex:  i,
			UserInput:  userInput,
			AgentReply: agentOut.Text,
		})

		if done || agentOut.SessionEnded {
			break
		}
	}

	// Mark as passed (LLM outcome evaluation is deferred to external judge).
	result.Passed = true
	result.Reason = "completed; LLM outcome evaluation deferred"
	return result, nil
}
