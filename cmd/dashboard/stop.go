package dashboard

import (
	"fmt"
	"path/filepath"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the local observability stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := resolvePlatformForDashboard(cmd)
			stackDir := filepath.Join(p.PlatformDir(), "stacks", "compose", "base")
			composeCmd := fmt.Sprintf(
				"docker compose -f %s/signoz.yml -f %s/posthog.yml -f %s/otel-collector.yml down",
				stackDir, stackDir, stackDir)
			if err := runShell(composeCmd); err != nil {
				return err
			}
			ui.Success("Stack stopped")
			return nil
		},
	}
}
