package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update obs to the latest release from GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate()
		},
	}
}

func runUpdate() error {
	ui.Info("Checking for latest release ...")

	latestURL := "https://github.com/osillation/obs/releases/latest/download/" + binaryName()
	stop := ui.Spinner("Downloading " + latestURL)

	resp, err := http.Get(latestURL)
	if err != nil {
		stop()
		return fmt.Errorf("downloading update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		stop()
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	self, err := os.Executable()
	if err != nil {
		stop()
		return fmt.Errorf("finding current binary: %w", err)
	}

	tmp := self + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		stop()
		return fmt.Errorf("writing update: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		stop()
		return fmt.Errorf("writing update: %w", err)
	}
	f.Close()
	stop()

	if err := os.Rename(tmp, self); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replacing binary: %w (try with sudo)", err)
	}

	ui.Success("Updated to latest release")
	return nil
}

func binaryName() string {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	name := fmt.Sprintf("obs_%s_%s", goos, arch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}
