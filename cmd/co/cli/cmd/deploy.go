package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy desired state to edge sites",
	Long:  `Pushes the desired deployment spec to one or more registered edge sites.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		site, _ := cmd.Flags().GetString("site")
		file, _ := cmd.Flags().GetString("file")

		if site == "" || file == "" {
			return fmt.Errorf("both --site and --file are required")
		}

		logger.Info("Starting deployment",
			zap.String("site", site),
			zap.String("file", file),
		)

		// simulate deploy
		fmt.Printf("🚀 Deploying %s using %s\n", site, file)

		logger.Info("Deployment completed successfully",
			zap.String("site", site),
		)
		return nil
	},
}

func init() {
	deployCmd.Flags().String("site", "", "Target edge site ID")
	deployCmd.Flags().String("file", "", "Path to desired state YAML")
}
