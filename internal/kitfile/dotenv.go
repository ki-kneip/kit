package kitfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DotEnv reads root/.env if present and returns its KEY=VALUE pairs.
// A missing file is not an error. Supports comments, blank lines, an
// optional "export " prefix and single/double quoted values.
func DotEnv(root string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, ".env"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var pairs []string
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf(".env:%d: expected KEY=VALUE", i+1)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		pairs = append(pairs, key+"="+value)
	}
	return pairs, nil
}
