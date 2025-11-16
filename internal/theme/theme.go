package theme

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
)

// Theme defines the color scheme for the terminal screenshot
type Theme struct {
	Name       string `json:"name"`
	Background string `json:"background"`
	Foreground string `json:"foreground"`
	
	// Window decorations
	WindowRed    string `json:"window_red"`
	WindowYellow string `json:"window_yellow"`
	WindowGreen  string `json:"window_green"`
	WindowBorder string `json:"window_border"`
	
	// Shadow
	Shadow string `json:"shadow"`
	
	// ANSI colors (0-15)
	Black        string `json:"black"`
	Red          string `json:"red"`
	Green        string `json:"green"`
	Yellow       string `json:"yellow"`
	Blue         string `json:"blue"`
	Magenta      string `json:"magenta"`
	Cyan         string `json:"cyan"`
	White        string `json:"white"`
	BrightBlack  string `json:"bright_black"`
	BrightRed    string `json:"bright_red"`
	BrightGreen  string `json:"bright_green"`
	BrightYellow string `json:"bright_yellow"`
	BrightBlue   string `json:"bright_blue"`
	BrightMagenta string `json:"bright_magenta"`
	BrightCyan   string `json:"bright_cyan"`
	BrightWhite  string `json:"bright_white"`
}

// Preset themes
var themes = map[string]Theme{
	"default": {
		Name:          "Default",
		Background:    "#151515",
		Foreground:    "#D3D3D3",
		WindowRed:     "#ED655A",
		WindowYellow:  "#E1C04C",
		WindowGreen:   "#71BD47",
		WindowBorder:  "#404040",
		Shadow:        "#10101066",
		Black:         "#000000",
		Red:           "#E06C75",
		Green:         "#98C379",
		Yellow:        "#E5C07B",
		Blue:          "#61AFEF",
		Magenta:       "#C678DD",
		Cyan:          "#56B6C2",
		White:         "#ABB2BF",
		BrightBlack:   "#5C6370",
		BrightRed:     "#E06C75",
		BrightGreen:   "#98C379",
		BrightYellow:  "#E5C07B",
		BrightBlue:    "#61AFEF",
		BrightMagenta: "#C678DD",
		BrightCyan:    "#56B6C2",
		BrightWhite:   "#FFFFFF",
	},
	"catppuccin-mocha": {
		Name:          "Catppuccin Mocha",
		Background:    "#1e1e2e",
		Foreground:    "#cdd6f4",
		WindowRed:     "#f38ba8",
		WindowYellow:  "#f9e2af",
		WindowGreen:   "#a6e3a1",
		WindowBorder:  "#45475a",
		Shadow:        "#11111b66",
		Black:         "#45475a",
		Red:           "#f38ba8",
		Green:         "#a6e3a1",
		Yellow:        "#f9e2af",
		Blue:          "#89b4fa",
		Magenta:       "#f5c2e7",
		Cyan:          "#94e2d5",
		White:         "#bac2de",
		BrightBlack:   "#585b70",
		BrightRed:     "#f38ba8",
		BrightGreen:   "#a6e3a1",
		BrightYellow:  "#f9e2af",
		BrightBlue:    "#89b4fa",
		BrightMagenta: "#f5c2e7",
		BrightCyan:    "#94e2d5",
		BrightWhite:   "#a6adc8",
	},
	"catppuccin-latte": {
		Name:          "Catppuccin Latte",
		Background:    "#eff1f5",
		Foreground:    "#4c4f69",
		WindowRed:     "#d20f39",
		WindowYellow:  "#df8e1d",
		WindowGreen:   "#40a02b",
		WindowBorder:  "#acb0be",
		Shadow:        "#e6e9ef66",
		Black:         "#5c5f77",
		Red:           "#d20f39",
		Green:         "#40a02b",
		Yellow:        "#df8e1d",
		Blue:          "#1e66f5",
		Magenta:       "#ea76cb",
		Cyan:          "#179299",
		White:         "#acb0be",
		BrightBlack:   "#6c6f85",
		BrightRed:     "#d20f39",
		BrightGreen:   "#40a02b",
		BrightYellow:  "#df8e1d",
		BrightBlue:    "#1e66f5",
		BrightMagenta: "#ea76cb",
		BrightCyan:    "#179299",
		BrightWhite:   "#bcc0cc",
	},
	"nord": {
		Name:          "Nord",
		Background:    "#2e3440",
		Foreground:    "#d8dee9",
		WindowRed:     "#bf616a",
		WindowYellow:  "#ebcb8b",
		WindowGreen:   "#a3be8c",
		WindowBorder:  "#4c566a",
		Shadow:        "#2e344066",
		Black:         "#3b4252",
		Red:           "#bf616a",
		Green:         "#a3be8c",
		Yellow:        "#ebcb8b",
		Blue:          "#81a1c1",
		Magenta:       "#b48ead",
		Cyan:          "#88c0d0",
		White:         "#e5e9f0",
		BrightBlack:   "#4c566a",
		BrightRed:     "#bf616a",
		BrightGreen:   "#a3be8c",
		BrightYellow:  "#ebcb8b",
		BrightBlue:    "#81a1c1",
		BrightMagenta: "#b48ead",
		BrightCyan:    "#8fbcbb",
		BrightWhite:   "#eceff4",
	},
	"dracula": {
		Name:          "Dracula",
		Background:    "#282a36",
		Foreground:    "#f8f8f2",
		WindowRed:     "#ff5555",
		WindowYellow:  "#f1fa8c",
		WindowGreen:   "#50fa7b",
		WindowBorder:  "#44475a",
		Shadow:        "#21222c66",
		Black:         "#21222c",
		Red:           "#ff5555",
		Green:         "#50fa7b",
		Yellow:        "#f1fa8c",
		Blue:          "#bd93f9",
		Magenta:       "#ff79c6",
		Cyan:          "#8be9fd",
		White:         "#f8f8f2",
		BrightBlack:   "#6272a4",
		BrightRed:     "#ff6e6e",
		BrightGreen:   "#69ff94",
		BrightYellow:  "#ffffa5",
		BrightBlue:    "#d6acff",
		BrightMagenta: "#ff92df",
		BrightCyan:    "#a4ffff",
		BrightWhite:   "#ffffff",
	},
	"tokyo-night": {
		Name:          "Tokyo Night",
		Background:    "#1a1b26",
		Foreground:    "#c0caf5",
		WindowRed:     "#f7768e",
		WindowYellow:  "#e0af68",
		WindowGreen:   "#9ece6a",
		WindowBorder:  "#414868",
		Shadow:        "#16161e66",
		Black:         "#15161e",
		Red:           "#f7768e",
		Green:         "#9ece6a",
		Yellow:        "#e0af68",
		Blue:          "#7aa2f7",
		Magenta:       "#bb9af7",
		Cyan:          "#7dcfff",
		White:         "#a9b1d6",
		BrightBlack:   "#414868",
		BrightRed:     "#f7768e",
		BrightGreen:   "#9ece6a",
		BrightYellow:  "#e0af68",
		BrightBlue:    "#7aa2f7",
		BrightMagenta: "#bb9af7",
		BrightCyan:    "#7dcfff",
		BrightWhite:   "#c0caf5",
	},
	"gruvbox-dark": {
		Name:          "Gruvbox Dark",
		Background:    "#282828",
		Foreground:    "#ebdbb2",
		WindowRed:     "#cc241d",
		WindowYellow:  "#d79921",
		WindowGreen:   "#98971a",
		WindowBorder:  "#504945",
		Shadow:        "#1d202166",
		Black:         "#282828",
		Red:           "#cc241d",
		Green:         "#98971a",
		Yellow:        "#d79921",
		Blue:          "#458588",
		Magenta:       "#b16286",
		Cyan:          "#689d6a",
		White:         "#a89984",
		BrightBlack:   "#928374",
		BrightRed:     "#fb4934",
		BrightGreen:   "#b8bb26",
		BrightYellow:  "#fabd2f",
		BrightBlue:    "#83a598",
		BrightMagenta: "#d3869b",
		BrightCyan:    "#8ec07c",
		BrightWhite:   "#ebdbb2",
	},
	"solarized-dark": {
		Name:          "Solarized Dark",
		Background:    "#002b36",
		Foreground:    "#839496",
		WindowRed:     "#dc322f",
		WindowYellow:  "#b58900",
		WindowGreen:   "#859900",
		WindowBorder:  "#073642",
		Shadow:        "#002b3666",
		Black:         "#073642",
		Red:           "#dc322f",
		Green:         "#859900",
		Yellow:        "#b58900",
		Blue:          "#268bd2",
		Magenta:       "#d33682",
		Cyan:          "#2aa198",
		White:         "#eee8d5",
		BrightBlack:   "#002b36",
		BrightRed:     "#cb4b16",
		BrightGreen:   "#586e75",
		BrightYellow:  "#657b83",
		BrightBlue:    "#839496",
		BrightMagenta: "#6c71c4",
		BrightCyan:    "#93a1a1",
		BrightWhite:   "#fdf6e3",
	},
}

// GetTheme returns a theme by name, or the default theme if not found
func GetTheme(name string) Theme {
	if theme, ok := themes[name]; ok {
		return theme
	}
	return themes["default"]
}

// LoadThemeFromFile loads a theme from a JSON file
func LoadThemeFromFile(path string) (Theme, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Theme{}, fmt.Errorf("failed to read theme file: %w", err)
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return Theme{}, fmt.Errorf("failed to parse theme file: %w", err)
	}

	return theme, nil
}

// ListThemes returns a list of all available preset themes
func ListThemes() []string {
	var names []string
	for name := range themes {
		names = append(names, name)
	}
	return names
}

// ParseColor converts a hex color string to a color.Color
func ParseColor(hex string) (color.Color, error) {
	if len(hex) == 0 {
		return nil, fmt.Errorf("empty color string")
	}

	// Remove # prefix if present
	if hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse RGB or RGBA
	var r, g, b, a uint8 = 0, 0, 0, 255
	
	switch len(hex) {
	case 6: // RGB
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RGB color: %w", err)
		}
	case 8: // RGBA
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RGBA color: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid color format: %s", hex)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
