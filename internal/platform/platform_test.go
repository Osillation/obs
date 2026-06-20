package platform_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/osillation/obs/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed all:testdata/platform
var testFS embed.FS

func TestMain(m *testing.M) {
	platform.SetEmbeddedFS(testFS, "testdata/platform")
	os.Exit(m.Run())
}

func TestDefaultDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".obs"), platform.DefaultDir())
}

func TestClientDir(t *testing.T) {
	p := platform.New("/tmp/obs-test")
	assert.Equal(t, "/tmp/obs-test/clients/mari", p.ClientDir("mari"))
}

func TestScriptPath(t *testing.T) {
	p := platform.New("/tmp/obs-test")
	assert.Equal(t, "/tmp/obs-test/platform/scripts/detect.sh", p.ScriptPath("detect.sh"))
}

func TestEnsureExtracted(t *testing.T) {
	dir := t.TempDir()
	p := platform.New(dir)
	err := p.EnsureExtracted()
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "platform", "scripts", "detect.sh"))
	assert.FileExists(t, filepath.Join(dir, "platform", ".version"))
}
