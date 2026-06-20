package config

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func ReadEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, scanner.Err()
}

func WriteEnv(path string, cfg map[string]string) error {
	keys := make([]string, 0, len(cfg))
	for k := range cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(cfg[k])
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0600)
}

func ClientDirExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}
