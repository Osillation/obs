package main

import (
	"embed"

	"github.com/osillation/obs/cmd"
	"github.com/osillation/obs/internal/platform"
)

//go:embed all:platform
var embeddedPlatform embed.FS

func main() {
	platform.SetEmbeddedFS(embeddedPlatform, "platform")
	cmd.Execute()
}
