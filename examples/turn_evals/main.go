// Package main demonstrates running TurnEvals from a YAML file.
//
// Usage:
//
//	go run ./examples/turn_evals \
//	  --project my-project \
//	  --app "My App" \
//	  --file evals.yaml
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/evals/turn"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
)

func main() {
	project := flag.String("project", "", "GCP project ID (required)")
	location := flag.String("location", "us", "CXAS location")
	appName := flag.String("app", "", "App display name (required)")
	file := flag.String("file", "evals.yaml", "Path to TurnEvals YAML file")
	flag.Parse()

	if *project == "" || *appName == "" {
		log.Fatal("--project and --app are required")
	}

	ctx := context.Background()
	authCfg := auth.Config{}

	appsClient, err := apps.NewClient(ctx, *project, *location, authCfg)
	if err != nil {
		log.Fatalf("apps client: %v", err)
	}

	app, err := appsClient.GetAppByDisplayName(ctx, *appName)
	if err != nil {
		log.Fatalf("GetAppByDisplayName: %v", err)
	}

	sessClient, err := sessions.NewClient(ctx, authCfg)
	if err != nil {
		log.Fatalf("sessions client: %v", err)
	}

	f, err := turn.LoadFile(*file)
	if err != nil {
		log.Fatalf("LoadFile: %v", err)
	}
	turn.ApplyGlobalConfig(f)

	result, err := turn.RunTests(ctx, f.AllTests(), app.Name, func(ctx context.Context, appN, sessID, text string) (*sessions.SessionOutput, error) {
		return sessClient.Run(ctx, sessions.RunSessionRequest{
			AppName:   appN,
			SessionID: sessID,
			Input:     sessions.SessionInput{Text: text},
		})
	})
	if err != nil {
		log.Fatalf("RunTests: %v", err)
	}

	fmt.Printf("\nResults: %d passed / %d total\n", result.Passed, result.Total)
	for _, r := range result.Results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		fmt.Printf("  [%s] %s\n", status, r.TestID)
	}
}
