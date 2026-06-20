package certs

import (
	"fmt"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <name>",
		Short: "Remove an employee's certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := resolveClient(cmd)
			if client == "" {
				return fmt.Errorf("--client is required")
			}
			name := args[0]
			p := resolvePlatformForCerts(cmd)
			if err := p.EnsureExtracted(); err != nil {
				return err
			}
			clientDir := p.ClientDir(client)
			r := newRunner(p)
			r.SetEnv("OBS_CLIENT_DIR", clientDir)
			out, err := r.Capture(p.ScriptPath("gen-certs.sh") + " revoke " + name)
			if err != nil {
				ui.Error(out)
				return err
			}
			ui.Success("Revoked: " + name)
			ui.Dim("Reload nginx to apply: docker exec obs-nginx nginx -s reload")
			return nil
		},
	}
}
