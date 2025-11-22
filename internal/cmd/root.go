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

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"

	"github.com/homeport/termshot/internal/ansi"
	"github.com/homeport/termshot/internal/img"
	"github.com/homeport/termshot/internal/ptexec"
	"github.com/homeport/termshot/internal/theme"

	"github.com/spf13/cobra"
)

// version string will be injected by automation
var version string

// saveToClipboard function will be implemented by OS specific code
var saveToClipboard func(img.Scaffold) error

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [%s flags] [--] command [command flags] [command arguments] [...]", executableName(), executableName()),
	Short: "Creates a screenshot of terminal command output",
	Long: `Executes the provided command as-is with all flags and arguments in a pseudo
terminal and captures the generated output. The result is printed as it was
produced. Additionally, an image will be rendered in a lookalike terminal
window including all terminal colors and text decorations.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion, err := cmd.Flags().GetBool("version"); showVersion && err == nil {
			if len(version) == 0 {
				version = "(development)"
			}

			// #nosec G104
			// nolint:all
			bunt.Printf("Lime{*%s*} version DimGray{%s}\n",
				executableName(),
				version,
			)

			return nil
		}

		rawRead, _ := cmd.Flags().GetString("raw-read")
		rawWrite, _ := cmd.Flags().GetString("raw-write")

		if len(args) == 0 && rawRead == "" {
			return cmd.Usage()
		}

		var scaffold = img.NewImageCreator()
		var buf bytes.Buffer
		var pt = ptexec.New()

		// Load theme
		themeName, _ := cmd.Flags().GetString("theme")
		themeFile, _ := cmd.Flags().GetString("theme-file")
		
		var selectedTheme theme.Theme
		if themeFile != "" {
			loadedTheme, err := theme.LoadThemeFromFile(themeFile)
			if err != nil {
				return fmt.Errorf("failed to load theme file: %w", err)
			}
			selectedTheme = loadedTheme
		} else {
			selectedTheme = theme.GetTheme(themeName)
		}
		
		// Apply theme to scaffold
		scaffold.SetTheme(selectedTheme)

		// Check for custom prompt
		customPrompt, _ := cmd.Flags().GetString("prompt")
		if customPrompt != "" {
			scaffold.SetPrompt(customPrompt)
		}

		// Check for syntax highlighting
		syntaxHighlight, _ := cmd.Flags().GetBool("syntax-highlight")
		scaffold.EnableSyntaxHighlighting(syntaxHighlight)

		// Check for prompt detection
		noPromptDetect, _ := cmd.Flags().GetBool("no-prompt-detect")
		scaffold.DisablePromptDetection(noPromptDetect)

		// Configure shell if specified
		shellPath, _ := cmd.Flags().GetString("shell")
		shellConfig, _ := cmd.Flags().GetString("shell-config")
		shellOpts, _ := cmd.Flags().GetStringSlice("shell-opts")
		
		if shellPath != "" {
			pt.SetShell(shellPath)
		}
		if shellConfig != "" {
			pt.SetShellConfig(shellConfig)
		}
		if len(shellOpts) > 0 {
			pt.SetShellOpts(shellOpts)
		}

		// Initialise scaffold with a column sizing so that the
		// content can be wrapped accordingly
		//
		if columns, err := cmd.Flags().GetInt("columns"); err == nil && columns > 0 {
			scaffold.SetColumns(columns)
			pt.Cols(uint16(columns))
		}

		// Disable window shadow if requested
		//
		if val, err := cmd.Flags().GetBool("no-shadow"); err == nil {
			scaffold.DrawShadow(!val)
		}

		// Disable window decorations (buttons) if requested
		//
		if val, err := cmd.Flags().GetBool("no-decoration"); err == nil {
			scaffold.DrawDecorations(!val)
		}

		// Configure that canvas is clipped at the end
		//
		if val, err := cmd.Flags().GetBool("clip-canvas"); err == nil {
			scaffold.ClipCanvas(val)
		}

		// Optional: Prepend command line arguments to output content
		//
		if includeCommand, err := cmd.Flags().GetBool("show-cmd"); err == nil && includeCommand && rawRead == "" {
			if err := scaffold.AddCommand(args...); err != nil {
				return err
			}
		}

		// Get the actual content for the screenshot
		//
		if rawRead == "" {
			// Run the provided command in a pseudo terminal and capture
			// the output to be later rendered into the screenshot
			bytes, err := pt.Command(args[0], args[1:]...).Run()
			if err != nil {
				return fmt.Errorf("failed to run command in pseudo terminal: %w", err)
			}
			buf.Write(bytes)

		} else {
			// Read the content from an existing file instead of
			// executing a command to read its output
			bytes, err := readFile(rawRead)
			if err != nil {
				return fmt.Errorf("failed to read contents: %w", err)
			}
			buf.Write(bytes)
		}

		// Allow manual override of command output content
		//
		if edit, err := cmd.Flags().GetBool("edit"); err == nil && edit && rawRead == "" {
			tmpFile, tmpErr := os.CreateTemp("", executableName())
			if tmpErr != nil {
				return tmpErr
			}

			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if err := os.WriteFile(tmpFile.Name(), buf.Bytes(), os.FileMode(0644)); err != nil {
				return err
			}

			editor := os.Getenv("EDITOR")
			if len(editor) == 0 {
				editor = "vi"
			}

			if _, err := ptexec.New().Command(editor, tmpFile.Name()).Run(); err != nil {
				return err
			}

			bytes, tmpErr := os.ReadFile(tmpFile.Name())
			if tmpErr != nil {
				return tmpErr
			}

			buf.Reset()
			buf.Write(bytes)
		}

		// Use improved ANSI parser if enabled
		improvedANSI, _ := cmd.Flags().GetBool("improved-ansi")
		if improvedANSI && rawRead != "" {
			// Only use improved parser for raw-read files, not for live commands
			// Use a large default to avoid unwanted wrapping, unless --columns is explicitly set
			columns := 500 // Large default to preserve line lengths
			if explicitCols, err := cmd.Flags().GetInt("columns"); err == nil && explicitCols > 0 {
				columns = explicitCols
			}
			vt := ansi.NewVirtualTerminal(columns)
			parsed, err := vt.Parse(&buf)
			if err != nil {
				// If parsing fails, continue with original content
				fmt.Fprintf(os.Stderr, "Warning: ANSI parsing failed, using original content: %v\n", err)
			} else {
				// Replace buffer content with parsed output
				buf.Reset()
				buf.WriteString(parsed.String())
			}
		}

		//
		if rawWrite != "" {
			// For raw-write, temporarily disable column wrapping
			originalColumns := scaffold.GetColumns()
			scaffold.SetColumns(0)
			if err := scaffold.AddContent(&buf); err != nil {
				return err
			}
			scaffold.SetColumns(originalColumns)
			
			var output *os.File
			var err error
			switch rawWrite {
			case "-":
				output = os.Stdout

			default:
				output, err = os.Create(filepath.Clean(rawWrite))
				if err != nil {
					return fmt.Errorf("failed to create file: %w", err)
				}

				defer func() { _ = output.Close() }()
			}

			return scaffold.WriteRaw(output)
		}

		// Add the captured output to the scaffold
		//
		if err := scaffold.AddContent(&buf); err != nil {
			return err
		}

		// Optional: Save image to clipboard
		//
		if toClipboard, err := cmd.Flags().GetBool("clipboard"); err == nil && toClipboard {
			return saveToClipboard(scaffold)
		}

		// Save image to file
		//
		filename, err := cmd.Flags().GetString("filename")
		if filename == "" || err != nil {
			fmt.Fprintf(os.Stderr, "failed to read filename from command-line, defaulting to out.png")
			filename = "out.png"
		}

		if extension := filepath.Ext(filename); extension != ".png" {
			return fmt.Errorf("file extension %q of filename %q is not supported, only png is supported", extension, filename)
		}

		file, err := os.Create(filepath.Clean(filename))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		defer func() { _ = file.Close() }()
		return scaffold.WritePNG(file)
	},
}

// Execute is the main entry point into the CLI code
func Execute() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, e error) error {
		return fmt.Errorf("unknown %s flag %w",
			executableName(),
			fmt.Errorf("issue with %v\n\nIn order to differentiate between program flags and command flags,\nuse '--' before the command so that all flags before the separator\nbelong to %s, while all others are used for the command.\n\n%s", e, executableName(), c.UsageString()),
		)
	})

	if err := rootCmd.Execute(); err != nil {
		var headline, content string

		type wrappedError interface {
			Error() string
			Unwrap() error
		}

		switch err := err.(type) {
		case wrappedError:
			headline = strings.SplitN(err.Error(), ":", 2)[0]
			content = err.Unwrap().Error()

		default:
			headline = "Error occurred"
			content = err.Error()
		}

		fmt.Fprint(os.Stderr, neat.ContentBox(
			headline,
			content,
			neat.HeadlineColor(bunt.OrangeRed),
			neat.ContentColor(bunt.LightCoral),
			neat.NoLineWrap(),
		))

		os.Exit(1)
	}
}

func executableName() string {
	if executable, err := os.Executable(); err == nil {
		return filepath.Clean(filepath.Base(executable))
	}

	return "termshot"
}

func readFile(name string) ([]byte, error) {
	switch name {
	case "-":
		return io.ReadAll(os.Stdin)

	default:
		return os.ReadFile(filepath.Clean(name))
	}
}

func init() {
	rootCmd.Flags().SortFlags = false

	// flags to control content
	rootCmd.Flags().BoolP("edit", "e", false, "edit content before creating screenshot")

	// flags to control look
	rootCmd.Flags().BoolP("show-cmd", "c", false, "include command in screenshot")
	rootCmd.Flags().IntP("columns", "C", 0, "force fixed number of columns in screenshot")
	rootCmd.Flags().Bool("no-decoration", false, "do not draw window decorations")
	rootCmd.Flags().Bool("no-shadow", false, "do not draw window shadow")
	rootCmd.Flags().BoolP("clip-canvas", "s", false, "clip canvas to visible image area (no margin)")

	// flags for shell configuration
	rootCmd.Flags().String("shell", "", "shell to use for command execution (e.g., /bin/zsh, /bin/bash)")
	rootCmd.Flags().String("shell-config", "", "shell configuration file to source (e.g., ~/.zshrc)")
	rootCmd.Flags().StringSlice("shell-opts", []string{}, "additional shell options")

	// flags for theming
	rootCmd.Flags().String("theme", "default", "color theme to use (default, catppuccin-mocha, nord, dracula, tokyo-night, gruvbox-dark, solarized-dark)")
	rootCmd.Flags().String("theme-file", "", "path to custom theme JSON file")

	// flags for prompt customization
	rootCmd.Flags().String("prompt", "", "custom prompt string (overrides default)")
	rootCmd.Flags().Bool("syntax-highlight", false, "enable syntax highlighting for command")

	// flags for output related settings
	rootCmd.Flags().StringP("filename", "f", "out.png", "filename of the screenshot")

	// flags for raw output processing
	rootCmd.Flags().String("raw-write", "", "write raw output to file instead of creating a screenshot")
	rootCmd.Flags().String("raw-read", "", "read raw input from file instead of executing a command")

	// flags for cursor handling
	rootCmd.Flags().Bool("improved-ansi", false, "use improved ANSI parser with cursor handling (only for --raw-read)")
	rootCmd.Flags().Bool("no-prompt-detect", false, "disable automatic prompt detection and command highlighting")

	// internals
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}
