// Package main demonstrates running a single text session turn against a CXAS app.
//
// Usage:
//
//	go run ./examples/run_session \
//	  --project my-project \
//	  --app "My App" \
//	  --text "Hello, can you help me?"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/sessions"
)

func main() {
	project := flag.String("project", "", "GCP project ID (required)")
	location := flag.String("location", "us", "CXAS location")
	appName := flag.String("app", "", "App display name (required)")
	text := flag.String("text", "Hello", "Text input to send")
	flag.Parse()

	if *project == "" || *appName == "" {
		log.Fatal("--project and --app are required")
	}

	ctx := context.Background()
	authCfg := auth.Config{}

	appsClient, err := apps.NewClient(ctx, *project, *location, authCfg)
	if err != nil {
		log.Fatalf("failed to create apps client: %v", err)
	}

	app, err := appsClient.GetAppByDisplayName(ctx, *appName)
	if err != nil {
		log.Fatalf("GetAppByDisplayName: %v", err)
	}

	sessClient, err := sessions.NewClient(ctx, authCfg)
	if err != nil {
		log.Fatalf("failed to create sessions client: %v", err)
	}

	out, err := sessClient.Run(ctx, sessions.RunSessionRequest{
		AppName:   app.Name,
		SessionID: "example-session-001",
		Input:     sessions.SessionInput{Text: *text},
	})
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	fmt.Printf("Response: %s\n", out.Text)
	if len(out.ToolCalls) > 0 {
		for _, tc := range out.ToolCalls {
			fmt.Printf("Tool called: %s\n", tc.ToolName)
		}
	}
}
