package certs

import (
	"fmt"
	"os"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Issue a signed TLS certificate for an employee",
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
			r.SetLiveOutput(os.Stdout)
			_, err := r.Capture(p.ScriptPath("gen-certs.sh") + " add-employee " + name)
			if err != nil {
				return err
			}
			ui.Success("Certificate: " + clientDir + "/tls/" + name + ".crt")
			ui.Success("Key (keep private): " + clientDir + "/tls/" + name + ".key")
			return nil
		},
	}
}
