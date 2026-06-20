package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/osillation/obs/internal/config"
	"github.com/osillation/obs/internal/ui"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		domain string
		token  string
	)

	cmd := &cobra.Command{
		Use:   "init <client>",
		Short: "Scaffold a new client directory with interactive config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(args[0], domain, token)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "client domain (skips prompt)")
	cmd.Flags().StringVar(&token, "token", "", "ingest token (skips prompt)")
	return cmd
}

func runInit(client, domain, token string) error {
	p := resolvePlatform()
	if err := p.EnsureExtracted(); err != nil {
		return err
	}

	clientDir := p.ClientDir(client)
	if config.ClientDirExists(clientDir) {
		return fmt.Errorf("client '%s' already exists at %s", client, clientDir)
	}

	ui.Info("Setting up client: " + client)

	if domain == "" {
		var err error
		domain, err = ui.Prompt("Client domain (e.g. mari.fleet.io)", "")
		if err != nil {
			return err
		}
	}

	if token == "" {
		genNew, err := ui.PromptConfirm("Generate a new ingest token?")
		if err != nil {
			return err
		}
		if genNew {
			out, err := exec.Command("openssl", "rand", "-hex", "32").Output()
			if err != nil {
				return fmt.Errorf("generating token: %w", err)
			}
			token = string(out[:64])
			ui.Dim("Generated: " + token)
		} else {
			token, err = ui.Prompt("Ingest token (32+ chars)", "")
			if err != nil {
				return err
			}
		}
	}

	cfg := map[string]string{
		"CLIENT_DOMAIN":    domain,
		"OBS_INGEST_TOKEN": token,
	}

	pgHost, _ := ui.Prompt("PostgreSQL host (enter to skip)", "")
	if pgHost != "" {
		cfg["POSTGRES_HOST"] = pgHost
		cfg["POSTGRES_PORT"], _ = ui.Prompt("PostgreSQL port", "5432")
		cfg["POSTGRES_DB"], _ = ui.Prompt("PostgreSQL database", "")
		cfg["POSTGRES_OTEL_USER"], _ = ui.Prompt("PostgreSQL OTel username", "otel_reader")
		cfg["POSTGRES_OTEL_PASSWORD"], _ = ui.PromptPassword("PostgreSQL OTel password")
	}

	redisHost, _ := ui.Prompt("Redis host (enter to skip)", "")
	if redisHost != "" {
		cfg["REDIS_HOST"] = redisHost
		cfg["REDIS_PORT"], _ = ui.Prompt("Redis port", "6379")
		cfg["REDIS_PASSWORD"], _ = ui.PromptPassword("Redis password (enter if none)")
	}

	mongoHost, _ := ui.Prompt("MongoDB host (enter to skip)", "")
	if mongoHost != "" {
		cfg["MONGO_HOST"] = mongoHost
		cfg["MONGO_PORT"], _ = ui.Prompt("MongoDB port", "27017")
		cfg["MONGO_OTEL_USER"], _ = ui.Prompt("MongoDB OTel username", "otel_reader")
		cfg["MONGO_OTEL_PASSWORD"], _ = ui.PromptPassword("MongoDB OTel password")
	}

	templateDir := filepath.Join(p.PlatformDir(), "clients", "_template")
	if err := copyDir(templateDir, clientDir); err != nil {
		return fmt.Errorf("copying template: %w", err)
	}

	if err := config.WriteEnv(filepath.Join(clientDir, "config.env"), cfg); err != nil {
		return err
	}

	ui.Success("Created " + clientDir + "/config.env")
	ui.Success("Created " + clientDir + "/acl.conf")
	ui.Success("Created " + clientDir + "/deployment.yml")
	fmt.Println()
	ui.Bold("Next steps:")
	ui.Dim("  obs certs init-ca --client " + client)
	ui.Dim("  obs certs add <engineer-name> --client " + client)

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
