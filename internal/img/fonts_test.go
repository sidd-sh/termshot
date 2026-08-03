package img

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func TestLoadFontFaceFromFileResolvesUserFontsDirectory(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config")
	fontsDir := filepath.Join(homeDir, ".local", "share", "fonts")

	if err := os.MkdirAll(fontsDir, 0o755); err != nil {
		t.Fatalf("mkdir fonts dir: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	fontPath := filepath.Join(fontsDir, "Goregular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o644); err != nil {
		t.Fatalf("write font file: %v", err)
	}

	face, err := loadFontFaceFromFile("Goregular.ttf", &truetype.Options{Size: 24, DPI: 144})
	if err != nil {
		t.Fatalf("resolve and load font: %v", err)
	}
	if face == nil {
		t.Fatal("expected font face")
	}
}

func TestResolveFontFileFindsFontsInConfigDirectory(t *testing.T) {
	configHome := t.TempDir()
	termshotDir := filepath.Join(configHome, "termshot")
	if err := os.MkdirAll(termshotDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", t.TempDir())

	fontPath := filepath.Join(termshotDir, "MyFont-Regular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o644); err != nil {
		t.Fatalf("write font file: %v", err)
	}

	resolved, err := resolveFontFile("MyFont-Regular.ttf")
	if err != nil {
		t.Fatalf("resolve font: %v", err)
	}
	if resolved != fontPath {
		t.Fatalf("expected %q, got %q", fontPath, resolved)
	}
}
