package dashboard

import (
	"os/exec"
	"runtime"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open",
		Short: "Open SigNoz and PostHog in the default browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			urls := []string{"http://localhost:3301", "http://localhost:8000"}
			for _, url := range urls {
				if err := openBrowser(url); err != nil {
					ui.Warn("Could not open " + url + " — open it manually")
				}
			}
			ui.Success("Opened SigNoz (localhost:3301) and PostHog (localhost:8000)")
			return nil
		},
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
