// cmd/lo/cli/main.go
package main

import (
	"log"

	"github.com/balaji-balu/margo-hello-world/cmd/lo/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
