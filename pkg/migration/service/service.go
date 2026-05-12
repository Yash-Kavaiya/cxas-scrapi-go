// Package service orchestrates DFCX → CXAS migration.
package service

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/agents"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/dfcx"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/ir"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/tools"
)

// MigrationConfig holds parameters for a migration run.
type MigrationConfig struct {
	// DFCXAgent is the source Dialogflow CX agent resource name.
	DFCXAgent string
	// TargetProject is the GCP project for the new CXAS app.
	TargetProject string
	// TargetLocation is the CXAS location (e.g. "us").
	TargetLocation string
	// DisplayName for the new CXAS app.
	DisplayName string
	// DefaultModel is the LLM model for generated agents.
	DefaultModel string
	// MaxParallel controls concurrent flow processing (default: 5).
	MaxParallel int
}

// MigrationResult summarises the outcome of a migration run.
type MigrationResult struct {
	AppName    string
	IR         *ir.MigrationIR
	AgentCount int
	ToolCount  int
	Errors     []string
}

// MigrationService orchestrates a DFCX → CXAS migration.
type MigrationService struct {
	cfg      MigrationConfig
	authCfg  auth.Config
	exporter *dfcx.Exporter
	mu       sync.Mutex
}

// New creates a MigrationService.
func New(cfg MigrationConfig, authCfg auth.Config) (*MigrationService, error) {
	exp, err := dfcx.NewExporter(context.Background(), authCfg)
	if err != nil {
		return nil, fmt.Errorf("create dfcx exporter: %w", err)
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "gemini-2.0-flash"
	}
	if cfg.MaxParallel <= 0 {
		cfg.MaxParallel = 5
	}
	return &MigrationService{cfg: cfg, authCfg: authCfg, exporter: exp}, nil
}

// Run executes the full migration pipeline and returns the result.
func (s *MigrationService) Run(ctx context.Context) (*MigrationResult, error) {
	result := &MigrationResult{}

	// Phase 1: Export DFCX agent data.
	agentIR, err := s.exporter.ExportAgent(ctx, s.cfg.DFCXAgent)
	if err != nil {
		return nil, fmt.Errorf("phase 1 export: %w", err)
	}

	// Phase 2: Build MigrationIR.
	migIR := buildIR(agentIR, s.cfg)
	result.IR = migIR

	// Phase 3: Create the CXAS app.
	appsClient, err := apps.NewClient(ctx, s.cfg.TargetProject, s.cfg.TargetLocation, s.authCfg)
	if err != nil {
		return nil, fmt.Errorf("create apps client: %w", err)
	}

	app, err := appsClient.CreateApp(ctx, apps.CreateAppRequest{
		DisplayName: s.cfg.DisplayName,
	})
	if err != nil {
		return nil, fmt.Errorf("phase 3 create app: %w", err)
	}
	result.AppName = app.Name
	migIR.Metadata.AppResourceName = app.Name

	// Phase 4: Deploy tools in parallel.
	toolsClient, err := tools.NewClient(ctx, app.Name, s.authCfg)
	if err != nil {
		return nil, fmt.Errorf("create tools client: %w", err)
	}
	if err := s.deployTools(ctx, migIR, toolsClient, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("tools: %v", err))
	}

	// Phase 5: Deploy agents in parallel.
	agentsClient, err := agents.NewClient(ctx, app.Name, s.authCfg)
	if err != nil {
		return nil, fmt.Errorf("create agents client: %w", err)
	}
	if err := s.deployAgents(ctx, migIR, agentsClient, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("agents: %v", err))
	}

	migIR.Metadata.AppID = app.Name
	return result, nil
}

// deployTools creates all tools in the IR concurrently.
func (s *MigrationService) deployTools(ctx context.Context, migIR *ir.MigrationIR, client *tools.Client, result *MigrationResult) error {
	sem := make(chan struct{}, s.cfg.MaxParallel)
	g, ctx := errgroup.WithContext(ctx)

	for id, t := range migIR.Tools {
		id, t := id, t
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			created, err := client.CreateTool(ctx, tools.Tool{
				DisplayName: t.Name,
				Description: fmt.Sprintf("Migrated tool %s", id),
			}, id)
			if err != nil {
				return fmt.Errorf("create tool %s: %w", id, err)
			}
			s.mu.Lock()
			t.Status = ir.StatusDeployed
			migIR.Tools[id] = t
			s.mu.Unlock()
			_ = created
			result.ToolCount++
			return nil
		})
	}
	return g.Wait()
}

// deployAgents creates all agents in the IR concurrently.
func (s *MigrationService) deployAgents(ctx context.Context, migIR *ir.MigrationIR, client *agents.Client, result *MigrationResult) error {
	sem := make(chan struct{}, s.cfg.MaxParallel)
	g, ctx := errgroup.WithContext(ctx)

	for id, a := range migIR.Agents {
		id, a := id, a
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			created, err := client.CreateAgent(ctx, agents.CreateAgentRequest{
				DisplayName: a.DisplayName,
				AgentType:   agents.AgentTypeLLM,
				Instruction: a.Instruction,
				Model:       s.cfg.DefaultModel,
				AgentID:     id,
			})
			if err != nil {
				return fmt.Errorf("create agent %s: %w", id, err)
			}
			s.mu.Lock()
			a.Status = ir.StatusDeployed
			a.ResourceName = created.Name
			migIR.Agents[id] = a
			s.mu.Unlock()
			result.AgentCount++
			return nil
		})
	}
	return g.Wait()
}

// buildIR constructs a MigrationIR from exported DFCX data.
func buildIR(agentIR *ir.DFCXAgentIR, cfg MigrationConfig) *ir.MigrationIR {
	migIR := &ir.MigrationIR{
		Metadata: ir.IRMetadata{
			AppName:      cfg.DisplayName,
			DefaultModel: cfg.DefaultModel,
		},
		Parameters: map[string]interface{}{},
		Tools:      make(map[string]ir.IRTool),
		Agents:     make(map[string]ir.IRAgent),
	}

	// Convert tools.
	for _, wh := range agentIR.Webhooks {
		toolID := fmt.Sprintf("tool-%s", wh["name"])
		migIR.Tools[toolID] = ir.IRTool{
			ID:      toolID,
			Name:    fmt.Sprintf("%v", wh["displayName"]),
			Type:    "openapi",
			Payload: wh,
			Status:  ir.StatusCompiled,
		}
	}

	// Convert playbooks → agents.
	for _, pb := range agentIR.Playbooks {
		conv := dfcx.PlaybookConverter{}
		agent := conv.ConvertPlaybook(pb)
		migIR.Agents[agent.DisplayName] = agent
	}

	return migIR
}
