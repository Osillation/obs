package cmd

import (
	"fmt"
	"os"

	"github.com/osillation/obs/internal/runner"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newDeployCmd() *cobra.Command {
	var projectPath string
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the observability stack for a client project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(projectPath)
		},
	}
	cmd.Flags().StringVar(&projectPath, "project", "", "path to the client's application project")
	cmd.MarkFlagRequired("project")
	return cmd
}

func runDeploy(projectPath string) error {
	if clientName == "" {
		return fmt.Errorf("--client is required")
	}

	p := resolvePlatform()
	if err := p.EnsureExtracted(); err != nil {
		return err
	}

	clientDir := p.ClientDir(clientName)

	r := runner.New(p.PlatformDir())
	r.SetEnv("OBS_CLIENT_DIR", clientDir)
	r.SetEnv("OBS_PLATFORM_DIR", p.PlatformDir())
	r.SetLiveOutput(os.Stdout)

	script := fmt.Sprintf("%s --client %s --project %s",
		p.ScriptPath("deploy.sh"), clientName, projectPath)

	ui.Info("Deploying " + clientName + " ...")
	return r.Stream(script)
}
