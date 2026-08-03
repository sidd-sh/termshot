package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/homeport/termshot/internal/img"
)

const configFileName = "config.json"

// GlobalConfig stores optional defaults loaded from ~/.config/termshot/config.json.
type GlobalConfig struct {
	Font img.FontConfig `json:"font,omitempty"`
}

func loadGlobalConfig() (GlobalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return GlobalConfig{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return GlobalConfig{}, nil
		}

		return GlobalConfig{}, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var config GlobalConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return GlobalConfig{}, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	return config, nil
}

func globalConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine config directory: %w", err)
	}

	return filepath.Join(dir, "termshot", configFileName), nil
}
