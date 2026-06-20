package certs

import (
	"github.com/osillation/obs/internal/platform"
	"github.com/osillation/obs/internal/runner"
	"github.com/spf13/cobra"
)

func resolvePlatformForCerts(cmd *cobra.Command) *platform.Platform {
	dir, _ := cmd.Root().PersistentFlags().GetString("platform-dir")
	if dir == "" {
		dir = platform.DefaultDir()
	}
	return platform.New(dir)
}

func resolveClient(cmd *cobra.Command) string {
	name, _ := cmd.Root().PersistentFlags().GetString("client")
	return name
}

func newRunner(p *platform.Platform) *runner.Runner {
	return runner.New(p.PlatformDir())
}
