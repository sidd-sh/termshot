package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/homeport/termshot/internal/img"
)

func TestLoadGlobalConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	configDir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "termshot")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	if err := os.WriteFile(configPath, []byte(`{"font":{"name":"hack","size":15,"dpi":144,"regular_path":"Hack-Regular.ttf"}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := loadGlobalConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Font.Name != "hack" {
		t.Fatalf("expected font name hack, got %q", config.Font.Name)
	}
	if config.Font.Size != 15 {
		t.Fatalf("expected font size 15, got %v", config.Font.Size)
	}
	if config.Font.DPI != 144 {
		t.Fatalf("expected font dpi 144, got %v", config.Font.DPI)
	}
	if config.Font.RegularPath != "Hack-Regular.ttf" {
		t.Fatalf("expected font path Hack-Regular.ttf, got %q", config.Font.RegularPath)
	}
	_ = img.FontConfig{}
}

func TestLoadGlobalConfigMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	config, err := loadGlobalConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Font != (img.FontConfig{}) {
		t.Fatalf("expected empty config, got %#v", config)
	}
}
