package runner_test

import (
	"testing"

	"github.com/osillation/obs/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerEnvInjection(t *testing.T) {
	r := runner.New("/platform")
	r.SetEnv("OBS_CLIENT_DIR", "/clients/mari")
	out, err := r.Capture("echo $OBS_CLIENT_DIR")
	require.NoError(t, err)
	assert.Equal(t, "/clients/mari\n", out)
}

func TestRunnerPlatformDirInEnv(t *testing.T) {
	r := runner.New("/platform-test")
	out, err := r.Capture("echo $OBS_PLATFORM_DIR")
	require.NoError(t, err)
	assert.Equal(t, "/platform-test\n", out)
}

func TestRunnerExitCode(t *testing.T) {
	r := runner.New("/tmp")
	_, err := r.Capture("exit 42")
	assert.Error(t, err)
}
