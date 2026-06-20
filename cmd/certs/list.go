package certs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/osillation/obs/internal/runner"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active employee certificates and their expiry dates",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := resolveClient(cmd)
			if client == "" {
				return fmt.Errorf("--client is required")
			}
			p := resolvePlatformForCerts(cmd)
			if err := p.EnsureExtracted(); err != nil {
				return err
			}
			clientDir := p.ClientDir(client)
			tlsDir := filepath.Join(clientDir, "tls")

			entries, err := os.ReadDir(tlsDir)
			if err != nil {
				return fmt.Errorf("reading tls dir: %w", err)
			}

			rows := [][2]string{}
			for _, entry := range entries {
				name := entry.Name()
				if !strings.HasSuffix(name, ".crt") {
					continue
				}
				if name == "ca.crt" || name == "server.crt" {
					continue
				}

				certPath := filepath.Join(tlsDir, name)
				r := runner.New(p.PlatformDir())
				expiry, err := r.Capture(fmt.Sprintf(
					`openssl x509 -noout -enddate -in %s | cut -d= -f2`, certPath))
				if err != nil {
					expiry = "unknown"
				}
				expiry = strings.TrimSpace(expiry)

				t, err := time.Parse("Jan  2 15:04:05 2006 MST", expiry)
				if err != nil {
					t, _ = time.Parse("Jan 2 15:04:05 2006 MST", expiry)
				}
				if !t.IsZero() {
					d := int(time.Until(t).Hours() / 24)
					expiry = t.Format("2006-01-02") + fmt.Sprintf(" (%d days)", d)
				}

				rows = append(rows, [2]string{strings.TrimSuffix(name, ".crt"), expiry})
			}

			if len(rows) == 0 {
				ui.Dim("No employee certificates found in " + tlsDir)
				return nil
			}

			fmt.Println()
			ui.Table([2]string{"Employee", "Expires"}, rows)
			return nil
		},
	}
}
