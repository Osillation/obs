package cmd

import (
	"fmt"
	"strings"

	"github.com/osillation/obs/internal/runner"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var receivers string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Pre-deploy config validation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(receivers)
		},
	}
	cmd.Flags().StringVar(&receivers, "receivers", "hostmetrics docker_stats", "space-separated receiver names")
	return cmd
}

func runValidate(receivers string) error {
	p := resolvePlatform()
	if err := p.EnsureExtracted(); err != nil {
		return err
	}
	if clientName == "" {
		return fmt.Errorf("--client is required")
	}

	clientDir := p.ClientDir(clientName)
	ui.Info("Validating " + clientDir + " ...")

	r := runner.New(p.PlatformDir())
	r.SetEnv("OBS_CLIENT_DIR", clientDir)
	r.SetEnv("OBS_PLATFORM_DIR", p.PlatformDir())

	script := fmt.Sprintf("%s --client-dir %s --receivers %q",
		p.ScriptPath("validate.sh"), clientDir, receivers)

	out, err := r.Capture(script)

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "Validation passed") {
			ui.Success(line)
		} else if strings.HasPrefix(line, "  x ") || strings.HasPrefix(line, "  ✗ ") {
			ui.Error(strings.TrimPrefix(strings.TrimPrefix(line, "  x "), "  ✗ "))
		} else {
			ui.Dim(line)
		}
	}

	return err
}
