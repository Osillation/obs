package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Runner struct {
	platformDir string
	env         map[string]string
	stdout      io.Writer
}

func New(platformDir string) *Runner {
	return &Runner{
		platformDir: platformDir,
		env:         make(map[string]string),
	}
}

func (r *Runner) SetEnv(key, value string) {
	r.env[key] = value
}

func (r *Runner) SetLiveOutput(w io.Writer) {
	r.stdout = w
}

func (r *Runner) Capture(script string) (string, error) {
	var buf bytes.Buffer
	err := r.run(script, &buf)
	return buf.String(), err
}

func (r *Runner) Stream(scriptPath string, args ...string) error {
	snippet := scriptPath + " " + strings.Join(args, " ")
	w := r.stdout
	if w == nil {
		w = os.Stdout
	}
	return r.run(snippet, w)
}

func (r *Runner) run(script string, out io.Writer) error {
	cmd := r.buildCmd(script)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func (r *Runner) buildCmd(script string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if !wslAvailable() {
			os.Stderr.WriteString("obs requires WSL on Windows.\nInstall: https://learn.microsoft.com/en-us/windows/wsl/install\n")
			os.Exit(1)
		}
		cmd = exec.Command("wsl", "bash", "-c", script)
	} else {
		cmd = exec.Command("bash", "-c", script)
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "OBS_PLATFORM_DIR="+r.platformDir)
	for k, v := range r.env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	return cmd
}

func wslAvailable() bool {
	return exec.Command("wsl", "--status").Run() == nil
}
