package cmd

import (
	"os"

	"github.com/osillation/obs/cmd/certs"
	"github.com/osillation/obs/cmd/cloudwatch"
	"github.com/osillation/obs/cmd/dashboard"
	"github.com/spf13/cobra"
)

var (
	platformDir string
	clientName  string
	noColor     bool
	quiet       bool
)

var rootCmd = &cobra.Command{
	Use:   "obs",
	Short: "obs — open-source observability stack (SigNoz + PostHog)",
	Long: `obs deploys and manages a self-hosted observability stack that replaces Datadog.
Run 'obs dashboard start' to get a local Datadog-equivalent in 2 minutes.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&platformDir, "platform-dir", "", "override ~/.obs/ working directory")
	rootCmd.PersistentFlags().StringVar(&clientName, "client", "", "client name (reads/writes ~/.obs/clients/<name>/)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress all output except errors")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newDetectCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newDeployCmd())
	rootCmd.AddCommand(newInstrumentCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(certs.NewCertsCmd())
	rootCmd.AddCommand(dashboard.NewDashboardCmd())
	rootCmd.AddCommand(cloudwatch.NewCloudwatchCmd())
}
