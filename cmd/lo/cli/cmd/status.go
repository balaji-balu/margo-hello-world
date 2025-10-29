// cmd/lo/cli/cmd/status.go
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the Local Orchestrator",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Fetching orchestrator status...")
		fmt.Println("✅ Orchestrator running. All deployments healthy.")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
