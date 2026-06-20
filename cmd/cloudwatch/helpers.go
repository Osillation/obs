package cloudwatch

import (
	"github.com/osillation/obs/internal/platform"
	"github.com/spf13/cobra"
)

func resolvePlatformForCW(cmd *cobra.Command) *platform.Platform {
	dir, _ := cmd.Root().PersistentFlags().GetString("platform-dir")
	if dir == "" {
		dir = platform.DefaultDir()
	}
	return platform.New(dir)
}

func resolveClientForCW(cmd *cobra.Command) string {
	name, _ := cmd.Root().PersistentFlags().GetString("client")
	return name
}
