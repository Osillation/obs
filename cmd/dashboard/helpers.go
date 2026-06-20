package dashboard

import (
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/osillation/obs/internal/platform"
	"github.com/spf13/cobra"
)

func resolvePlatformForDashboard(cmd *cobra.Command) *platform.Platform {
	dir, _ := cmd.Root().PersistentFlags().GetString("platform-dir")
	if dir == "" {
		dir = platform.DefaultDir()
	}
	return platform.New(dir)
}

func runShell(script string) error {
	return runShellWithOutput(script, os.Stdout)
}

func runShellWithOutput(script string, out io.Writer) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("wsl", "bash", "-c", script)
	} else {
		cmd = exec.Command("bash", "-c", script)
	}
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}
