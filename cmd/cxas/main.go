package main

import (
	"os"

	"github.com/GoogleCloudPlatform/cxas-go/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
