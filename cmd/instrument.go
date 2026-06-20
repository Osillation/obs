package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osillation/obs/internal/config"
	"github.com/osillation/obs/internal/runner"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newInstrumentCmd() *cobra.Command {
	var (
		projectPath string
		localMode   bool
	)
	cmd := &cobra.Command{
		Use:   "instrument",
		Short: "Copy instrumentation templates into a client project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstrument(projectPath, localMode)
		},
	}
	cmd.Flags().StringVar(&projectPath, "project", "", "path to the client's application project")
	cmd.MarkFlagRequired("project")
	cmd.Flags().BoolVar(&localMode, "local", false, "configure for local dashboard (localhost:4318)")
	return cmd
}

func runInstrument(projectPath string, localMode bool) error {
	p := resolvePlatform()
	if err := p.EnsureExtracted(); err != nil {
		return err
	}

	r := runner.New(p.PlatformDir())
	raw, err := r.Capture(p.ScriptPath("detect.sh") + " " + projectPath)
	if err != nil {
		return fmt.Errorf("detecting project stack: %w", err)
	}

	var result detectionResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return fmt.Errorf("parsing detection: %w", err)
	}

	fmt.Printf("\nDetected: %s / %s / %s\n\n",
		join(result.Frameworks, ", "),
		join(result.Databases, ", "),
		result.DeploymentMode)

	ui.Info("Copying templates:")

	instrDir := filepath.Join(p.PlatformDir(), "instrumentation")

	for _, fw := range result.Frameworks {
		srcDir := filepath.Join(instrDir, fw)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			ui.Warn("No template for " + fw + " — skipping")
			continue
		}
		if err := copyDir(srcDir, filepath.Join(projectPath, "obs-instrumentation", fw)); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("instrumentation/%s/ → %s/obs-instrumentation/%s/", fw, projectPath, fw))
	}

	fmt.Println()
	ui.Bold("Env vars to set:")

	if localMode {
		ui.Dim("  OBS_LOCAL=true")
		ui.Dim("  OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318")
	} else {
		if clientName != "" {
			clientDir := p.ClientDir(clientName)
			cfg, _ := config.ReadEnv(filepath.Join(clientDir, "config.env"))
			domain := cfg["CLIENT_DOMAIN"]
			if domain != "" {
				ui.Dim("  OTEL_EXPORTER_OTLP_ENDPOINT=https://ingest." + domain)
				ui.Dim("  OBS_INGEST_TOKEN=" + cfg["OBS_INGEST_TOKEN"])
			}
		}
		ui.Dim("  OTEL_SERVICE_NAME=<your-service-name>")
		ui.Dim("  OTEL_SERVICE_NAMESPACE=" + clientName)
	}

	if contains(result.Frameworks, "nextjs") {
		fmt.Println()
		ui.Bold("Frontend (public env vars):")
		ui.Dim("  NEXT_PUBLIC_POSTHOG_KEY=phc_<get from PostHog UI after first deploy>")
		ui.Dim("  NEXT_PUBLIC_POSTHOG_HOST=/ingest")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}
