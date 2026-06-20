package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osillation/obs/internal/platform"
	"github.com/spf13/cobra"
)

const cliVersion = "0.1.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version and embedded platform version",
		Run: func(cmd *cobra.Command, args []string) {
			p := resolvePlatform()
			platformVersion := "unknown"
			versionFile := filepath.Join(p.PlatformDir(), ".version")
			if data, err := os.ReadFile(versionFile); err == nil {
				platformVersion = strings.TrimSpace(string(data))
			}
			fmt.Printf("obs v%s\n", cliVersion)
			fmt.Printf("platform v%s (embedded)\n", platformVersion)
		},
	}
}

// resolvePlatform returns a Platform using the --platform-dir flag or default.
func resolvePlatform() *platform.Platform {
	dir := platformDir
	if dir == "" {
		dir = platform.DefaultDir()
	}
	return platform.New(dir)
}
