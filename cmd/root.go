package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

var (
	kubeconfig string

	rootCmd = &cobra.Command{
		Use:   "cert-manager-upgrade",
		Short: "Throwaway CLI which can safely upgrade cert-manager",
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		println("something terrible happened...")
		panic(err)
	}
}

func stringVarWithEnv(p *string, name, env, value, help string) {
	var v string

	if v = os.Getenv(env); v == "" {
		v = value
	}
	rootCmd.PersistentFlags().StringVar(p, name, v, help)
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&config.AppConfig.TargetVersion, "target-version", config.TargetVersion, "target version to upgrade cert-manager to")

	stringVarWithEnv(
		&kubeconfig, "kubeconfig", "KUBECONFIG", "", "a kubernetes configuration to upgrade")

	stringVarWithEnv(
		&config.AppConfig.CertManagerNamespace, "cert-manager-namespace", "CERT_MANAGER_NAMESPACE", config.DefaultCertManagerNamespace, "")

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}
