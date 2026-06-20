package dashboard

import (
	"fmt"
	"path/filepath"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the local observability stack (SigNoz + PostHog + OTel)",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := resolvePlatformForDashboard(cmd)
			if err := p.EnsureExtracted(); err != nil {
				return err
			}

			ui.Info("Starting SigNoz + PostHog + OTel Collector ...")

			stackDir := filepath.Join(p.PlatformDir(), "stacks", "compose", "base")
			composeCmd := fmt.Sprintf(
				"docker compose -f %s/signoz.yml -f %s/posthog.yml -f %s/otel-collector.yml up -d --pull missing",
				stackDir, stackDir, stackDir)

			if err := runShell(composeCmd); err != nil {
				return fmt.Errorf("starting stack: %w", err)
			}

			fmt.Println()
			ui.Success("Stack running")
			fmt.Println()
			fmt.Println("  SigNoz   → http://localhost:3301")
			fmt.Println("  PostHog  → http://localhost:8000")
			fmt.Println("  OTel     → http://localhost:4317 (gRPC)")
			fmt.Println("             http://localhost:4318 (HTTP)")
			fmt.Println()
			ui.Bold("Set in your app:")
			ui.Dim("  OBS_LOCAL=true")
			ui.Dim("  OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318")
			fmt.Println()
			ui.Dim("  obs dashboard open     open both UIs in browser")
			ui.Dim("  obs dashboard logs     tail collector logs")
			ui.Dim("  obs dashboard stop     stop all containers")
			return nil
		},
	}
}
