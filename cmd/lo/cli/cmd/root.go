// cmd/lo/cli/cmd/root.go
package cmd

import (
	"fmt"
	//"os"

	"github.com/spf13/cobra"
)

var (
	version = "v0.1.0"
)

// rootCmd is the base command for the Local Orchestrator CLI
var rootCmd = &cobra.Command{
	Use:   "lo",
	Short: "Local Orchestrator CLI",
	Long: `The Local Orchestrator CLI manages deployments, synchronization,
and polling operations for the edge orchestrator.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'lo --help' to see available commands")
	},
}

// Execute runs the CLI
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "config.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
