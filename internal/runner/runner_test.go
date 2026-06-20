package runner_test

import (
	"strings"
	"testing"

	"github.com/osillation/obs/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCapturesOutput(t *testing.T) {
	r := runner.New("/tmp")
	out, err := r.Capture("echo hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", strings.TrimSpace(out))
}

func TestRunWithEnv(t *testing.T) {
	r := runner.New("/tmp")
	r.SetEnv("TEST_VAR", "world")
	out, err := r.Capture("echo $TEST_VAR")
	require.NoError(t, err)
	assert.Equal(t, "world", strings.TrimSpace(out))
}

func TestRunNonZeroExitReturnsError(t *testing.T) {
	r := runner.New("/tmp")
	_, err := r.Capture("exit 1")
	assert.Error(t, err)
}

func TestScriptRunCapturesJSON(t *testing.T) {
	r := runner.New("/tmp")
	out, err := r.Capture(`echo '{"frameworks":["nextjs"]}'`)
	require.NoError(t, err)
	assert.Contains(t, out, "nextjs")
}
