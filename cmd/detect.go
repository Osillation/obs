package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/osillation/obs/internal/runner"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newDetectCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "detect <project-path>",
		Short: "Scan a project directory and show detected stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDetect(args[0], jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output raw JSON")
	return cmd
}

type detectionResult struct {
	Frameworks     []string       `json:"frameworks"`
	Databases      []string       `json:"databases"`
	DeploymentMode string         `json:"deployment_mode"`
	ComposeFile    string         `json:"compose_file"`
	AppNetwork     string         `json:"app_network"`
	Credentials    map[string]any `json:"credentials"`
}

func runDetect(projectPath string, jsonOut bool) error {
	p := resolvePlatform()
	if err := p.EnsureExtracted(); err != nil {
		return err
	}

	r := runner.New(p.PlatformDir())
	stop := ui.Spinner("Scanning " + projectPath)
	raw, err := r.Capture(p.ScriptPath("detect.sh") + " " + projectPath)
	stop()
	if err != nil {
		return fmt.Errorf("detection failed: %w\n%s", err, raw)
	}

	if jsonOut {
		fmt.Println(raw)
		return nil
	}

	var result detectionResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return fmt.Errorf("parsing detection output: %w\n%s", err, raw)
	}

	fmt.Println()
	ui.Table([2]string{"Property", "Value"}, [][2]string{
		{"Frameworks", join(result.Frameworks, ", ")},
		{"Databases", join(result.Databases, ", ")},
		{"Mode", result.DeploymentMode},
		{"App network", result.AppNetwork},
	})

	if creds, ok := result.Credentials["postgresql"]; ok {
		if m, ok := creds.(map[string]any); ok {
			fmt.Printf("\n  PostgreSQL    host=%v  db=%v  port=%v\n", m["host"], m["db"], m["port"])
		}
	}
	if creds, ok := result.Credentials["redis"]; ok {
		if m, ok := creds.(map[string]any); ok {
			fmt.Printf("  Redis         host=%v  port=%v\n", m["host"], m["port"])
		}
	}

	return nil
}

func join(items []string, sep string) string {
	if len(items) == 0 {
		return "none"
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}
