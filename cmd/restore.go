package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jr0d/cert-manager-upgrade/pkg/app"
	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "convert and restore cert-manager resources from backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Restore(kubeconfig)
	},
}

func init() {
	restoreCmd.Flags().BoolVar(
		&config.AppConfig.PreserveBackups, "keep-backups", false, "do not delete backups")
}
