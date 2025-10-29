package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show deployment status of edge sites",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("✅ site1: applied successfully")
		fmt.Println("⚙ site2: applying...")
		fmt.Println("❌ site3: failed - container pull error")
	},
}
