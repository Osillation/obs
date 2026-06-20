package platform

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	embeddedFS     embed.FS
	embeddedPrefix string // root path within the FS (e.g. "platform" or "testdata/platform")
)

// SetEmbeddedFS must be called from main() (or TestMain) before any platform
// operations. prefix is the root directory inside fsys that contains the
// platform assets (e.g. "platform").
func SetEmbeddedFS(fsys embed.FS, prefix string) {
	embeddedFS = fsys
	embeddedPrefix = prefix
}

// Platform manages the obs working directory and embedded asset extraction.
type Platform struct {
	baseDir string // e.g. ~/.obs
}

// New returns a Platform rooted at baseDir.
func New(baseDir string) *Platform {
	return &Platform{baseDir: baseDir}
}

// DefaultDir returns ~/.obs.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".obs"
	}
	return filepath.Join(home, ".obs")
}

// ClientDir returns the directory for a named client.
func (p *Platform) ClientDir(client string) string {
	return filepath.Join(p.baseDir, "clients", client)
}

// ScriptPath returns the full path to a named platform script.
func (p *Platform) ScriptPath(script string) string {
	return filepath.Join(p.baseDir, "platform", "scripts", script)
}

// PlatformDir returns the platform assets directory under baseDir.
func (p *Platform) PlatformDir() string {
	return filepath.Join(p.baseDir, "platform")
}

// EnsureExtracted extracts embedded platform assets into baseDir/platform/.
// If the assets are already present and unchanged (SHA-256 hash match) it is
// a no-op.
func (p *Platform) EnsureExtracted() error {
	if embeddedPrefix == "" {
		return fmt.Errorf("embedded FS not initialised — call platform.SetEmbeddedFS() from main()")
	}

	destDir := filepath.Join(p.baseDir, "platform")
	versionHashFile := filepath.Join(destDir, ".version.hash")

	embeddedHash, err := p.embeddedHash()
	if err != nil {
		return fmt.Errorf("computing embedded hash: %w", err)
	}

	if existingHash, err := os.ReadFile(versionHashFile); err == nil {
		if strings.TrimSpace(string(existingHash)) == embeddedHash {
			return nil // already extracted, same version
		}
	}

	if err := p.extract(destDir); err != nil {
		return fmt.Errorf("extracting platform assets: %w", err)
	}

	return os.WriteFile(versionHashFile, []byte(embeddedHash), 0644)
}

func (p *Platform) extract(destDir string) error {
	return fs.WalkDir(embeddedFS, embeddedPrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Strip the embedded prefix to get a path relative to destDir.
		rel := strings.TrimPrefix(path, embeddedPrefix)
		if rel == "" {
			return nil // root entry itself
		}
		rel = strings.TrimPrefix(rel, "/")
		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		data, err := embeddedFS.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}

		perm := fs.FileMode(0644)
		if strings.HasSuffix(path, ".sh") {
			perm = 0755
		}
		return os.WriteFile(dest, data, perm)
	})
}

func (p *Platform) embeddedHash() (string, error) {
	h := sha256.New()
	err := fs.WalkDir(embeddedFS, embeddedPrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := embeddedFS.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		return err
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
