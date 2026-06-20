package cloudwatch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show CloudWatch integration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := resolveClientForCW(cmd)
			if client == "" {
				return fmt.Errorf("--client is required")
			}
			p := resolvePlatformForCW(cmd)
			cfgPath := filepath.Join(p.ClientDir(client), "cloudwatch.yml")

			data, err := os.ReadFile(cfgPath)
			if err != nil {
				ui.Warn("CloudWatch not connected. Run: obs cloudwatch connect --client " + client)
				return nil
			}

			var cfg cloudwatchConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return err
			}

			fmt.Println()
			ui.Bold(fmt.Sprintf("CloudWatch → SigNoz (%s)", cfg.Region))
			fmt.Println()

			if len(cfg.LogGroups) > 0 {
				rows := make([][2]string, len(cfg.LogGroups))
				for i, lg := range cfg.LogGroups {
					rows[i] = [2]string{lg, "active"}
				}
				ui.Table([2]string{"Log group", "Status"}, rows)
				fmt.Println()
			}

			if len(cfg.Namespaces) > 0 {
				rows := make([][2]string, len(cfg.Namespaces))
				for i, ns := range cfg.Namespaces {
					rows[i] = [2]string{ns, "active"}
				}
				ui.Table([2]string{"Metric namespace", "Status"}, rows)
				fmt.Println()
			}

			fmt.Printf("  Auth: %s", cfg.AuthMethod)
			if cfg.Profile != "" {
				fmt.Printf("=%s", cfg.Profile)
			}
			fmt.Printf("  Region: %s  Poll: %s\n", cfg.Region, cfg.PollInterval)
			return nil
		},
	}
}
