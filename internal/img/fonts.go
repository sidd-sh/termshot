// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package img

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/font"
	imgfont "golang.org/x/image/font"
)

const (
	DefaultFontFamily  = "firacode-mono"
	FallbackFontFamily = "hack"
)

// FontConfig controls how termshot resolves fonts.
type FontConfig struct {
	Name string `json:"name,omitempty"`

	Size float64 `json:"size,omitempty"`
	DPI  float64 `json:"dpi,omitempty"`

	RegularPath    string `json:"regular_path,omitempty"`
	BoldPath       string `json:"bold_path,omitempty"`
	ItalicPath     string `json:"italic_path,omitempty"`
	BoldItalicPath string `json:"bold_italic_path,omitempty"`
}

// ConfigureFonts applies a built-in font family or custom font files.
func (s *Scaffold) ConfigureFonts(config FontConfig) error {
	size := config.Size
	if size <= 0 {
		size = defaultFontSize
	}

	dpi := config.DPI
	if dpi <= 0 {
		dpi = defaultFontDPI
	}

	options := &truetype.Options{
		Size: s.factor * size,
		DPI:  dpi,
	}

	if config.RegularPath != "" || config.BoldPath != "" || config.ItalicPath != "" || config.BoldItalicPath != "" {
		return s.applyCustomFontFiles(config, options)
	}

	name := normalizeFontFamilyName(config.Name)
	if name == "" {
		name = DefaultFontFamily
	}

	switch name {
	case "hack":
		s.setFontFacesFromProvider(font.Hack, options)
		return nil
	case "go-mono", "go":
		s.setFontFacesFromProvider(font.Go, options)
		return nil
	case "firacode-mono", "fira-code", "firacode":
		return s.applyFiraCodeMono(options)
	default:
		return fmt.Errorf("unsupported font %q", config.Name)
	}
}

func (s *Scaffold) setFontFacesFromProvider(fontProvider font.Font, options *truetype.Options) {
	s.regular = fontProvider.Regular(options)
	s.bold = fontProvider.Bold(options)
	s.italic = fontProvider.Italic(options)
	s.boldItalic = fontProvider.BoldItalic(options)
}

func (s *Scaffold) applyCustomFontFiles(config FontConfig, options *truetype.Options) error {
	if config.RegularPath == "" {
		return fmt.Errorf("font-regular is required when using custom font files")
	}

	regular, err := loadFontFaceFromFile(config.RegularPath, options)
	if err != nil {
		return fmt.Errorf("failed to load regular font %q: %w", config.RegularPath, err)
	}
	s.regular = regular

	boldPath := config.BoldPath
	if boldPath == "" {
		boldPath = config.RegularPath
	}
	bold, err := loadFontFaceFromFile(boldPath, options)
	if err != nil {
		return fmt.Errorf("failed to load bold font %q: %w", boldPath, err)
	}
	s.bold = bold

	italicPath := config.ItalicPath
	if italicPath == "" {
		italicPath = config.RegularPath
	}
	italic, err := loadFontFaceFromFile(italicPath, options)
	if err != nil {
		return fmt.Errorf("failed to load italic font %q: %w", italicPath, err)
	}
	s.italic = italic

	boldItalicPath := config.BoldItalicPath
	if boldItalicPath == "" {
		boldItalicPath = boldPath
	}
	boldItalic, err := loadFontFaceFromFile(boldItalicPath, options)
	if err != nil {
		return fmt.Errorf("failed to load bold italic font %q: %w", boldItalicPath, err)
	}
	s.boldItalic = boldItalic

	return nil
}

func (s *Scaffold) applyFiraCodeMono(options *truetype.Options) error {
	fontPaths, err := detectFiraCodeMonoPaths()
	if err != nil {
		return err
	}

	return s.applyCustomFontFiles(FontConfig{
		RegularPath:    fontPaths.regular,
		BoldPath:       fontPaths.bold,
		ItalicPath:     fontPaths.italic,
		BoldItalicPath: fontPaths.boldItalic,
	}, options)
}

type firaCodeFontPaths struct {
	regular    string
	bold       string
	italic     string
	boldItalic string
}

func detectFiraCodeMonoPaths() (firaCodeFontPaths, error) {
	searchDirs := fontSearchDirectories()
	if len(searchDirs) == 0 {
		return firaCodeFontPaths{}, fmt.Errorf("could not determine font search directories")
	}

	regular, err := findFontFile(searchDirs, []string{
		"FiraCode-Regular.ttf",
		"FiraCodeNerdFontMono-Regular.ttf",
		"FiraCodeNerdFontMonoNerdFont-Regular.ttf",
	}, []string{
		"*Fira*Code*Mono*Regular*.ttf",
		"*Fira*Code*Regular*.ttf",
	})
	if err != nil {
		return firaCodeFontPaths{}, fmt.Errorf("failed to find FiraCode Mono regular font: %w", err)
	}

	bold, _ := findFontFile(searchDirs, []string{
		"FiraCode-Bold.ttf",
		"FiraCodeNerdFontMono-Bold.ttf",
		"FiraCodeNerdFontMonoNerdFont-Bold.ttf",
	}, []string{
		"*Fira*Code*Mono*Bold*.ttf",
		"*Fira*Code*Bold*.ttf",
	})
	if bold == "" {
		bold = regular
	}

	italic, _ := findFontFile(searchDirs, []string{
		"FiraCode-Italic.ttf",
		"FiraCodeNerdFontMono-Italic.ttf",
		"FiraCodeNerdFontMonoNerdFont-Italic.ttf",
	}, []string{
		"*Fira*Code*Mono*Italic*.ttf",
		"*Fira*Code*Italic*.ttf",
	})
	if italic == "" {
		italic = regular
	}

	boldItalic, _ := findFontFile(searchDirs, []string{
		"FiraCode-BoldItalic.ttf",
		"FiraCodeNerdFontMono-BoldItalic.ttf",
		"FiraCodeNerdFontMonoNerdFont-BoldItalic.ttf",
	}, []string{
		"*Fira*Code*Mono*Bold*Italic*.ttf",
		"*Fira*Code*Bold*Italic*.ttf",
	})
	if boldItalic == "" {
		boldItalic = bold
	}

	return firaCodeFontPaths{
		regular:    regular,
		bold:       bold,
		italic:     italic,
		boldItalic: boldItalic,
	}, nil
}

func loadFontFaceFromFile(path string, options *truetype.Options) (imgfont.Face, error) {
	resolvedPath, err := resolveFontFile(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath.Clean(resolvedPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	ttf, err := truetype.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font file: %w", err)
	}

	return truetype.NewFace(ttf, options), nil
}

func resolveFontFile(path string) (string, error) {
	cleaned := filepath.Clean(path)
	if info, err := os.Stat(cleaned); err == nil && !info.IsDir() {
		return cleaned, nil
	}

	searchDirs := fontSearchDirectories()
	for _, baseDir := range searchDirs {
		candidate := filepath.Join(baseDir, cleaned)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	baseName := filepath.Base(cleaned)
	found, err := findFontFile(searchDirs, []string{baseName}, []string{"*" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + "*" + filepath.Ext(baseName)})
	if err == nil && found != "" {
		return found, nil
	}

	return "", fmt.Errorf("font file %q not found in %v", path, searchDirs)
}

func findFontFile(searchDirs []string, preferredNames []string, globPatterns []string) (string, error) {
	for _, baseDir := range searchDirs {
		for _, name := range preferredNames {
			candidate := filepath.Join(baseDir, name)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}

	for _, baseDir := range searchDirs {
		for _, pattern := range globPatterns {
			matches, err := filepath.Glob(filepath.Join(baseDir, pattern))
			if err != nil {
				continue
			}
			for _, match := range matches {
				if info, err := os.Stat(match); err == nil && !info.IsDir() {
					return match, nil
				}
			}
		}
	}

	normalizedPreferred := map[string]struct{}{}
	for _, preferred := range preferredNames {
		normalizedPreferred[strings.ToLower(preferred)] = struct{}{}
	}

	for _, baseDir := range searchDirs {
		var found string
		walkErr := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			fileName := d.Name()
			lowerFileName := strings.ToLower(fileName)
			if _, ok := normalizedPreferred[lowerFileName]; ok {
				found = path
				return fs.SkipAll
			}

			for _, pattern := range globPatterns {
				matched, matchErr := filepath.Match(strings.ToLower(pattern), lowerFileName)
				if matchErr == nil && matched {
					found = path
					return fs.SkipAll
				}
			}

			return nil
		})

		if found != "" {
			return found, nil
		}

		if walkErr != nil && walkErr != fs.SkipAll {
			continue
		}
	}

	return "", fmt.Errorf("not found in %v", searchDirs)
}

func fontSearchDirectories() []string {
	var dirs []string

	if configDir, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(configDir, "termshot"))
		dirs = append(dirs, filepath.Join(configDir, "termshot", "fonts"))
	}

	switch runtime.GOOS {
	case "windows":
		windowsDir := os.Getenv("WINDIR")
		if windowsDir != "" {
			dirs = append(dirs, filepath.Join(windowsDir, "Fonts"))
		}
	case "darwin":
		dirs = append(dirs, "/Library/Fonts", "/System/Library/Fonts")
	default:
		dirs = append(dirs, "/usr/share/fonts", "/usr/local/share/fonts")
	}

	if home, err := os.UserHomeDir(); err == nil {
		switch runtime.GOOS {
		case "windows":
			dirs = append(dirs, filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"))
		case "darwin":
			dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
		default:
			dirs = append(dirs, filepath.Join(home, ".local", "share", "fonts"))
		}
	}

	unique := make([]string, 0, len(dirs))
	seen := map[string]struct{}{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		cleaned := filepath.Clean(dir)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		unique = append(unique, cleaned)
	}

	return unique
}

func normalizeFontFamilyName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return normalized
}
