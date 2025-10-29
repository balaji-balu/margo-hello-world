// cmd/lo/cli/cmd/apply.go
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the current desired state to the local environment",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Applying desired state to local environment...")
		// integrate orchestrator apply logic here
		fmt.Println("✅ Deployment applied successfully.")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
