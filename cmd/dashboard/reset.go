package dashboard

import (
	"fmt"
	"path/filepath"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Stop and remove all local stack containers and volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := resolvePlatformForDashboard(cmd)
			if err := p.EnsureExtracted(); err != nil {
				return err
			}
			stackDir := filepath.Join(p.PlatformDir(), "stacks", "compose", "base")
			composeCmd := fmt.Sprintf(
				"docker compose -f %s/signoz.yml -f %s/posthog.yml -f %s/otel-collector.yml down -v --remove-orphans",
				stackDir, stackDir, stackDir)
			if err := runShell(composeCmd); err != nil {
				return fmt.Errorf("resetting stack: %w", err)
			}
			ui.Success("Stack reset — all containers and volumes removed")
			return nil
		},
	}
}
