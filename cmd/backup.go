package cmd

import (
	"github.com/spf13/cobra"

	"github.com/jr0d/cert-manager-upgrade/pkg/app"
	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup cert-manager resources for conversion and restoration on upgrade failure",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Backup(kubeconfig)
	},
}

func init() {
	backupCmd.Flags().BoolVar(
		&config.AppConfig.PreserveCRDs, "keep-crds", false, "do not delete crds")

	backupCmd.Flags().BoolVar(
		&config.AppConfig.SkipFixSecrets, "skip-fix-secrets", false, "do not modify secrets")
}
