package cloudwatch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect",
		Short: "Remove CloudWatch integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := resolveClientForCW(cmd)
			if client == "" {
				return fmt.Errorf("--client is required")
			}

			confirmed, err := ui.PromptConfirm("Remove cloudwatch.yml and stop polling?")
			if err != nil || !confirmed {
				return nil
			}

			p := resolvePlatformForCW(cmd)
			cfgPath := filepath.Join(p.ClientDir(client), "cloudwatch.yml")
			receiverPath := filepath.Join(p.PlatformDir(), "config", "otel", "receivers", "cloudwatch.yaml")

			os.Remove(cfgPath)
			os.Remove(receiverPath)

			ui.Success("CloudWatch disconnected")
			ui.Dim("Restart the collector to apply: obs dashboard stop && obs dashboard start")
			return nil
		},
	}
}
