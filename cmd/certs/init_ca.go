package certs

import (
	"fmt"
	"os"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCACmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-ca",
		Short: "Create the Certificate Authority for a client",
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
			r := newRunner(p)
			r.SetEnv("OBS_CLIENT_DIR", clientDir)
			r.SetLiveOutput(os.Stdout)
			out, err := r.Capture(p.ScriptPath("gen-certs.sh") + " init-ca")
			if err != nil {
				ui.Error(out)
				return err
			}
			ui.Success("CA created: " + clientDir + "/tls/ca.crt")
			ui.Dim("Next: obs certs add <name> --client " + client)
			return nil
		},
	}
}
