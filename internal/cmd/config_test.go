package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/homeport/termshot/internal/img"
	"github.com/spf13/pflag"
)

func useTempConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "termshot", configFileName)
	t.Setenv(configEnvVar, path)
	return path
}

func testFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("theme", "default", "")
	flags.Int("columns", defaultColumns, "")
	flags.Bool("no-shadow", false, "")
	flags.Float64("font-size", 12, "")
	flags.StringSlice("shell-opts", []string{}, "")
	flags.String("filename", "", "")
	flags.Bool("no-improved-ansi", false, "")
	_ = flags.MarkDeprecated("no-improved-ansi", "deprecated")
	return flags
}

func TestLoadGlobalConfig(t *testing.T) {
	configPath := useTempConfig(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

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
}

func TestLoadGlobalConfigMissing(t *testing.T) {
	useTempConfig(t)

	config, err := loadGlobalConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Font != (img.FontConfig{}) || len(config.Flags) != 0 {
		t.Fatalf("expected empty config, got %#v", config)
	}
}

func TestEnsureGlobalConfigWritesDefaults(t *testing.T) {
	configPath := useTempConfig(t)

	path, created, err := ensureGlobalConfig(testFlags())
	if err != nil || !created || path != configPath {
		t.Fatalf("expected config to be created at %q, got %q, %v, %v", configPath, path, created, err)
	}

	var content struct {
		Flags map[string]any `json:"flags"`
	}
	data, _ := os.ReadFile(configPath)
	if err := json.Unmarshal(data, &content); err != nil {
		t.Fatalf("parse generated config: %v", err)
	}

	if content.Flags["theme"] != "default" || content.Flags["columns"] != float64(defaultColumns) || content.Flags["no-shadow"] != false {
		t.Fatalf("unexpected generated defaults: %v", content.Flags)
	}

	for _, name := range []string{"filename", "no-improved-ansi"} {
		if _, ok := content.Flags[name]; ok {
			t.Fatalf("flag %q must not be part of the generated config", name)
		}
	}

	if _, created, err := ensureGlobalConfig(testFlags()); err != nil || created {
		t.Fatalf("expected existing config to be kept, got %v, %v", created, err)
	}
}

func TestApplyConfigFlags(t *testing.T) {
	useTempConfig(t)

	flags := testFlags()
	if err := flags.Parse([]string{"--theme", "dracula"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	config := GlobalConfig{Flags: map[string]any{
		"theme":      "nord",
		"columns":    json.Number("80"),
		"no-shadow":  true,
		"shell-opts": []any{"-i", "-l"},
	}}

	if err := applyConfigFlags(flags, config); err != nil {
		t.Fatalf("apply config flags: %v", err)
	}

	theme, _ := flags.GetString("theme")
	columns, _ := flags.GetInt("columns")
	noShadow, _ := flags.GetBool("no-shadow")
	shellOpts, _ := flags.GetStringSlice("shell-opts")

	if theme != "dracula" {
		t.Fatalf("command line flag must take precedence, got theme %q", theme)
	}
	if columns != 80 || !noShadow || len(shellOpts) != 2 {
		t.Fatalf("config defaults not applied: columns=%d no-shadow=%v shell-opts=%v", columns, noShadow, shellOpts)
	}

	for _, invalid := range []map[string]any{{"unknown": "x"}, {"filename": "x.png"}, {"columns": "many"}} {
		if err := applyConfigFlags(testFlags(), GlobalConfig{Flags: invalid}); err == nil {
			t.Fatalf("expected error for config flags %v", invalid)
		}
	}
}

func TestUpdateConfigFlags(t *testing.T) {
	configPath := useTempConfig(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"font":{"name":"hack"},"flags":{"theme":"nord"}}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := updateConfigFlags(testFlags(), []string{"columns=120", "no-shadow=true"}, []string{"theme"}); err != nil {
		t.Fatalf("update config: %v", err)
	}

	config, err := loadGlobalConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Font.Name != "hack" {
		t.Fatalf("expected font section to be kept, got %#v", config.Font)
	}
	if _, ok := config.Flags["theme"]; ok {
		t.Fatalf("expected theme to be removed, got %v", config.Flags)
	}
	if config.Flags["columns"] != json.Number("120") || config.Flags["no-shadow"] != true {
		t.Fatalf("unexpected flags %v", config.Flags)
	}

	for _, invalid := range []string{"columns", "unknown=1", "filename=x.png", "columns=many"} {
		if _, err := updateConfigFlags(testFlags(), []string{invalid}, nil); err == nil {
			t.Fatalf("expected error for %q", invalid)
		}
	}
}

func TestRawWriteFilename(t *testing.T) {
	name, err := rawWriteFilename()
	if err != nil {
		t.Fatalf("raw-write filename: %v", err)
	}

	if !regexp.MustCompile(`^\d{8}-\d{6}-[0-9a-f]{8}\.log$`).MatchString(name) {
		t.Fatalf("unexpected raw-write filename %q", name)
	}
}
