// Package main demonstrates a minimal quickstart for cxas-go.
// It lists all CX Agent Studio apps in a GCP project using
// Application Default Credentials (ADC).
//
// Usage:
//
//	gcloud auth application-default login
//	go run ./examples/quickstart --project my-project
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/cxas-go/internal/auth"
	"github.com/GoogleCloudPlatform/cxas-go/pkg/apps"
)

func main() {
	project := flag.String("project", "", "GCP project ID (required)")
	location := flag.String("location", "us", "CXAS location")
	flag.Parse()

	if *project == "" {
		log.Fatal("--project is required")
	}

	ctx := context.Background()

	client, err := apps.NewClient(ctx, *project, *location, auth.Config{})
	if err != nil {
		log.Fatalf("failed to create apps client: %v", err)
	}

	list, err := client.ListApps(ctx)
	if err != nil {
		log.Fatalf("ListApps: %v", err)
	}

	if len(list) == 0 {
		fmt.Println("No apps found.")
		return
	}

	fmt.Printf("Found %d app(s):\n", len(list))
	for _, a := range list {
		fmt.Printf("  - %s  (%s)\n", a.DisplayName, a.Name)
	}
}
