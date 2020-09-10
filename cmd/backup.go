package cmd

import (
	"fmt"
	"github.com/jr0d/cert-manager-upgrade/pkg/app"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use: "backup",
	Short: "backup cert-manager resources for conversion and restoration on upgrade failure",
	Long: `\
This command will attempt to use the default storage class to backup v1alpha2
cert-manager resources so that they can be converted to v1 resource types
after the upgrade occurs. If a default storage class is not available, the
upgrade operation will fail.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("I am a turtle!\n")
		return app.Backup(kubeconfig)
	},
}
