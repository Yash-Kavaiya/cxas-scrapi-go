# cxas-go — CX Agent Studio SDK for Go

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25-blue)](go.mod)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)](#)

A production-grade **Go SDK and CLI** for [Google Cloud CX Agent Studio](https://cloud.google.com/cx-agent-studio), ported from the Python [cxas-scrapi](https://github.com/GoogleCloudPlatform/cxas-scrapi) library.

> **Module path:** `github.com/GoogleCloudPlatform/cxas-go`

---

## Features

- **Full REST SDK** for every CX Agent Studio resource — Apps, Agents, Tools, Toolsets, Sessions, Callbacks, Variables, Guardrails, Evaluations, Versions, Deployments, and Changelogs
- **WebSocket BidiSession** for streaming audio conversations
- **TurnEvals engine** — YAML-driven conversation tests with 7 assertion operators and topological dependency resolution
- **SimulationEvals** — parallel LLM-driven multi-turn simulation (up to 25 workers via `errgroup`)
- **Linter** — offline rule engine (`CXL001`–`CXL005`) for CXAS app directories
- **Migration service** — DFCX → CXAS orchestrator with parallel agent/tool deployment
- **Cobra CLI** — `apps`, `create`, `delete`, `pull`, `push`, `branch`, `run`, `lint`, `eval run`
- **5-path credential chain** — explicit token → credentials file → JSON → `CXAS_OAUTH_TOKEN` env → ADC
- **Retrying HTTP client** — `go-retryablehttp` + OAuth2 transport + User-Agent injection

---

## Installation

### Binary (CLI)

```bash
go install github.com/GoogleCloudPlatform/cxas-go/cmd/cxas@latest
```

### SDK

```bash
go get github.com/GoogleCloudPlatform/cxas-go
```

**Requirements:** Go 1.25+

---

## Authentication

The SDK resolves credentials in this priority order:

| Priority | Source |
|----------|--------|
| 1 | Explicit `auth.Config{TokenSource: ...}` |
| 2 | `auth.Config{CredsPath: "/path/to/sa.json"}` |
| 3 | `auth.Config{CredsJSON: []byte(...)}` |
| 4 | `CXAS_OAUTH_TOKEN` environment variable |
| 5 | Application Default Credentials (ADC) |

**Quickest setup** — run `gcloud auth application-default login` and the SDK picks up ADC automatically.

```bash
# Set a token directly
export CXAS_OAUTH_TOKEN=$(gcloud auth print-access-token)

# Or pass via CLI flag
cxas --oauth-token $(gcloud auth print-access-token) apps list --project my-project
```

---

## CLI Usage

```bash
cxas --help
```

```
Usage:
  cxas [command]

Available Commands:
  apps        Manage CXAS apps
  branch      Create a copy (branch) of an app
  create      Create a new CXAS app
  delete      Delete a CXAS app
  eval        Run evaluations against a CXAS app
  lint        Lint a CXAS app directory for common issues
  pull        Export (pull) an app to a local directory
  push        Import (push) an app from a local file
  run         Run a single session turn against a CXAS app

Global Flags:
  -l, --location string      CXAS location (default "us")
      --oauth-token string   OAuth2 token (overrides CXAS_OAUTH_TOKEN)
  -p, --project string       GCP project ID
  -o, --output string        Output format: table, json, yaml
```

### Common Commands

```bash
# List all apps
cxas apps list --project my-project

# Get a specific app
cxas apps get "My App Name" --project my-project

# Create a new app
cxas create --name "My New App" --project my-project

# Pull (export) an app to a local ZIP
cxas pull --app "My App" --out ./exports --project my-project

# Push (import) an app from a ZIP
cxas push --app "My App" --file ./exports/My_App.zip --project my-project

# Branch (copy) an app
cxas branch --app "Production App" --name "Staging App" --project my-project

# Run a single session turn
cxas run --app "My App" --text "Hello" --project my-project

# Lint an app directory
cxas lint --dir ./my-app-dir

# Run turn evaluations
cxas eval run --app "My App" --file evals.yaml --project my-project
```

---

## SDK Usage

### Apps

```go
import (
    "context"
    "github.com/GoogleCloudPlatform/cxas-go/internal/auth"
    "github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

client, err := apps.NewClient(ctx, "my-project", "us", auth.Config{})

// List all apps
list, err := client.ListApps(ctx)

// Get by display name
app, err := client.GetAppByDisplayName(ctx, "My App")

// Create
app, err := client.CreateApp(ctx, apps.CreateAppRequest{
    DisplayName: "New App",
    Description: "Created by cxas-go",
})

// Delete
err = client.DeleteApp(ctx, app.Name)
```

### Sessions (Text)

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"

client, err := sessions.NewClient(ctx, auth.Config{})

out, err := client.Run(ctx, sessions.RunSessionRequest{
    AppName:   "projects/my-project/locations/us/apps/my-app",
    SessionID: "session-001",
    Input:     sessions.SessionInput{Text: "Hello, how can you help?"},
})

fmt.Println(out.Text)
for _, tc := range out.ToolCalls {
    fmt.Printf("Tool called: %s\n", tc.ToolName)
}
```

### Agents

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/agents"

client, err := agents.NewClient(ctx, appName, auth.Config{})

// Create an LLM agent
ag, err := client.CreateAgent(ctx, agents.CreateAgentRequest{
    DisplayName: "Customer Support Agent",
    AgentType:   agents.AgentTypeLLM,
    Model:       "gemini-2.0-flash",
    Instruction: "You are a helpful customer support agent.",
})

// Create a workflow agent (uses REST POST, not gRPC)
wf, err := client.CreateAgent(ctx, agents.CreateAgentRequest{
    DisplayName: "Routing Agent",
    AgentType:   agents.AgentTypeWorkflow,
})
```

### Tools

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/tools"

client, err := tools.NewClient(ctx, appName, auth.Config{})

// Execute an atomic tool
resp, err := client.ExecuteTool(ctx, tools.ExecuteToolRequest{
    AppName: appName,
    Tool:    "projects/.../tools/search-tool",
    Input:   map[string]interface{}{"query": "latest offers"},
})

// Execute a toolset tool
resp, err := client.ExecuteTool(ctx, tools.ExecuteToolRequest{
    AppName: appName,
    ToolsetTool: &tools.ToolsetToolRef{
        Toolset: "projects/.../toolsets/crm-tools",
        ToolID:  "get_customer",
    },
    Input: map[string]interface{}{"customer_id": "12345"},
})
```

### TurnEvals

```yaml
# evals.yaml
tests:
  - id: greeting-test
    turns:
      - input: "Hello"
        expectations:
          - operator: contains
            value: "help"
          - operator: no_tools_called
  - id: search-test
    turns:
      - input: "Find me a flight to Paris"
        expectations:
          - operator: tool_called
            value: search_flights
```

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/evals/turn"

f, err := turn.LoadFile("evals.yaml")
turn.ApplyGlobalConfig(f)

result, err := turn.RunTests(ctx, f.AllTests(), appName, func(ctx context.Context, appName, sessID, text string) (*sessions.SessionOutput, error) {
    return sessClient.Run(ctx, sessions.RunSessionRequest{...})
})

fmt.Printf("Passed: %d/%d\n", result.Passed, result.Total)
```

**Supported operators:** `contains`, `equals`, `tool_called`, `no_tools_called`, `agent_transfer`, `tool_input`, `tool_output`

### Linter

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/linter"

l := linter.New("./my-app-dir", nil)
report, err := l.LintDirectory("./my-app-dir")

for _, r := range report.Results {
    fmt.Println(r) // [error] agent.json:0 CXL001: resource is missing a displayName
}
fmt.Printf("%d errors, %d warnings\n", report.Errors, report.Warnings)
```

| Rule | Severity | Description |
|------|----------|-------------|
| `CXL001` | error | Resource missing `displayName` |
| `CXL002` | warning | LLM agent has empty instruction |
| `CXL003` | warning | Instruction exceeds 8000 characters |
| `CXL004` | info | Tool/toolset missing `description` |
| `CXL005` | warning | Eval YAML file has no turn `input` fields |

### Migration (DFCX → CXAS)

```go
import "github.com/GoogleCloudPlatform/cxas-go/pkg/migration/service"

svc, err := service.New(service.MigrationConfig{
    DFCXAgent:      "projects/my-project/locations/global/agents/my-dfcx-agent",
    TargetProject:  "my-project",
    TargetLocation: "us",
    DisplayName:    "Migrated App",
    DefaultModel:   "gemini-2.0-flash",
    MaxParallel:    5,
}, auth.Config{})

result, err := svc.Run(ctx)
fmt.Printf("Created app: %s\n", result.AppName)
fmt.Printf("Deployed %d agents, %d tools\n", result.AgentCount, result.ToolCount)
```

---

## Package Reference

| Package | Description |
|---------|-------------|
| `internal/auth` | 5-path credential resolver |
| `internal/httpclient` | Retrying, authenticated HTTP client + `DoJSON` helper |
| `internal/resource` | Resource name parser (`projects/{P}/locations/{L}/apps/{A}`) |
| `internal/textproto` | Protocol Buffer text format parser |
| `pkg/apps` | App CRUD + Export/Import |
| `pkg/agents` | Agent CRUD (LLM, DFCX, Workflow) |
| `pkg/sessions` | Text RunSession + WebSocket BidiSession |
| `pkg/tools` | Tool/Toolset CRUD + ExecuteTool |
| `pkg/callbacks` | Callback CRUD |
| `pkg/variables` | Variable declaration CRUD |
| `pkg/guardrails` | Guardrail CRUD |
| `pkg/evaluations` | Evaluation CRUD + RunEvaluation LRO |
| `pkg/versions` | Version snapshot CRUD |
| `pkg/deployments` | Deployment CRUD |
| `pkg/changelogs` | Changelog read access |
| `pkg/workflows` | WorkflowAgent / WorkflowStep data models |
| `pkg/insights` | Contact Center Insights client |
| `pkg/scorecards` | Scorecard client |
| `pkg/gemini` | Vertex AI Gemini client (semaphore + exponential backoff) |
| `pkg/github` | GitHub Actions CI workflow template generation |
| `pkg/secretmanager` | Secret Manager get-or-create utility |
| `pkg/linter` | Offline lint rule engine |
| `pkg/evals/turn` | TurnEvals: YAML loader + 7-operator validator + dependency runner |
| `pkg/evals/simulation` | SimulationEvals: parallel LLM-driven multi-turn |
| `pkg/evals/tool` | ToolEvals: structured tool assertion engine |
| `pkg/evals/callback` | CallbackEvals: callback test runner |
| `pkg/migration/ir` | MigrationIR data models |
| `pkg/migration/dfcx` | DFCX exporter + playbook/tool converters |
| `pkg/migration/service` | DFCX → CXAS migration orchestrator |

---

## Building from Source

```bash
git clone https://github.com/Yash-Kavaiya/cxas-scrapi-go.git
cd cxas-scrapi-go

# Build the CLI binary
go build -o bin/cxas ./cmd/cxas

# Run all tests
go test ./...

# Run with the built binary
./bin/cxas --help
```

---

## Known Limitations

| Limitation | Notes |
|------------|-------|
| `callbacks.ExecuteCallback` | Python's `exec()`-based in-process callback execution cannot be ported to Go. Returns `ErrNotSupported`. Use external callback testing. |
| CES gRPC client | No public Go gRPC client exists for CES v1beta. All API calls use REST. If Google publishes a Go gRPC client, `internal/httpclient` calls can be swapped without changing `pkg/*` interfaces. |
| `pandas` / DataFrames | Replaced with structured slices and `tablewriter` for tabular CLI output. |
| Race detector (`-race`) | Requires CGO. On Windows without a C compiler, run `go test ./...` without `-race`. |

---

## License

Apache 2.0 — see [LICENSE](LICENSE)
