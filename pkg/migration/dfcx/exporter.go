// Package dfcx provides tools for exporting and converting Dialogflow CX agents
// to the CXAS intermediate representation format.
package dfcx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/internal/httpclient"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/migration/ir"
)

const dfcxBaseURL = "https://dialogflow.googleapis.com/v3"

// Exporter fetches all data from a Dialogflow CX agent.
type Exporter struct {
	httpClient *http.Client
}

// NewExporter creates a new DFCX exporter.
func NewExporter(ctx context.Context, cfg auth.Config) (*Exporter, error) {
	ts, err := auth.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Exporter{
		httpClient: httpclient.New(httpclient.Options{TokenSource: ts}),
	}, nil
}

// ExportAgent fetches the full agent structure into a DFCXAgentIR.
func (e *Exporter) ExportAgent(ctx context.Context, agentName string) (*ir.DFCXAgentIR, error) {
	// GET the agent metadata
	agentURL := fmt.Sprintf("%s/%s", dfcxBaseURL, agentName)
	var agentData map[string]interface{}
	if err := httpclient.DoJSON(ctx, e.httpClient, "GET", agentURL, nil, &agentData); err != nil {
		return nil, fmt.Errorf("get agent %s: %w", agentName, err)
	}

	displayName, _ := agentData["displayName"].(string)
	defaultLang, _ := agentData["defaultLanguageCode"].(string)

	agentIR := &ir.DFCXAgentIR{
		Name:                agentName,
		DisplayName:         displayName,
		DefaultLanguageCode: defaultLang,
	}

	// Fetch flows
	flows, err := e.listFlows(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("list flows: %w", err)
	}
	agentIR.Flows = flows

	// Fetch playbooks
	playbooks, err := e.listPlaybooks(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("list playbooks: %w", err)
	}
	agentIR.Playbooks = playbooks

	// Fetch webhooks
	webhooks, err := e.listWebhooks(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	agentIR.Webhooks = webhooks

	return agentIR, nil
}

func (e *Exporter) listFlows(ctx context.Context, agentName string) ([]ir.DFCXFlowModel, error) {
	url := fmt.Sprintf("%s/%s/flows", dfcxBaseURL, agentName)
	var resp struct {
		Flows []map[string]interface{} `json:"flows"`
	}
	if err := httpclient.DoJSON(ctx, e.httpClient, "GET", url, nil, &resp); err != nil {
		return nil, err
	}
	flows := make([]ir.DFCXFlowModel, len(resp.Flows))
	for i, f := range resp.Flows {
		name, _ := f["name"].(string)
		flows[i] = ir.DFCXFlowModel{
			FlowID:   name,
			FlowData: f,
		}
	}
	return flows, nil
}

func (e *Exporter) listPlaybooks(ctx context.Context, agentName string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/playbooks", dfcxBaseURL, agentName)
	var resp struct {
		Playbooks []map[string]interface{} `json:"playbooks"`
	}
	if err := httpclient.DoJSON(ctx, e.httpClient, "GET", url, nil, &resp); err != nil {
		// Not all agents have playbooks; tolerate 404
		if httpclient.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp.Playbooks, nil
}

func (e *Exporter) listWebhooks(ctx context.Context, agentName string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/webhooks", dfcxBaseURL, agentName)
	var resp struct {
		Webhooks []map[string]interface{} `json:"webhooks"`
	}
	if err := httpclient.DoJSON(ctx, e.httpClient, "GET", url, nil, &resp); err != nil {
		if httpclient.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp.Webhooks, nil
}
