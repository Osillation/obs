package dashboard

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var service string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Tail logs from the local stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := resolvePlatformForDashboard(cmd)
			stackDir := filepath.Join(p.PlatformDir(), "stacks", "compose", "base")
			composeCmd := fmt.Sprintf(
				"docker compose -f %s/signoz.yml -f %s/posthog.yml -f %s/otel-collector.yml logs -f --tail=100",
				stackDir, stackDir, stackDir)
			if service != "" {
				composeCmd += " " + service
			}
			return runShellWithOutput(composeCmd, os.Stdout)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "filter to a specific service name")
	return cmd
}
