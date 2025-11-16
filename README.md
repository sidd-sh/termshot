# termshot

[![License](https://img.shields.io/github/license/homeport/termshot.svg)](https://github.com/homeport/termshot/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/homeport/termshot)](https://goreportcard.com/report/github.com/homeport/termshot)
[![Tests](https://github.com/homeport/termshot/workflows/Tests/badge.svg)](https://github.com/homeport/termshot/actions?query=workflow%3A%22Tests%22)
[![Codecov](https://img.shields.io/codecov/c/github/homeport/termshot/main.svg)](https://codecov.io/gh/homeport/termshot)
[![Go Reference](https://pkg.go.dev/badge/github.com/homeport/termshot.svg)](https://pkg.go.dev/github.com/homeport/termshot)
[![Release](https://img.shields.io/github/release/homeport/termshot.svg)](https://github.com/homeport/termshot/releases/latest)

## Fork Improvements

This fork adds several enhancements:
- **8 built-in themes** (catppuccin-mocha, nord, dracula, etc.) + custom theme support
- **Shell configuration** support (--shell, --shell-config, --shell-opts) for custom prompts
- **Improved ANSI parsing** with virtual terminal for better cursor handling (powerlevel10k compatible)
- **Syntax highlighting** for shell commands with customizable prompts
- **--no-prompt-detect** flag for interactive shells with autosuggestions
- **True RGB color** preservation from modern terminals

Generate beautiful screenshots of your terminal, from your terminal.

```sh
termshot lolcat -f <(figlet -f big termshot)
```

This command generates this screenshot:

![example](.doc/example-cmd-figlet.png)

## Installation

To install with Homebrew on macOS or Linux:

```sh
brew install homeport/tap/termshot
```

See [Releases](https://github.com/homeport/termshot/releases/) for pre-compiled binaries for Darwin and Linux.

## Usage

This tool reads the console output and renders an output image that resembles a user interface window. It's inspired by some other web-based tools like [carbon.now.sh](https://carbon.now.sh/), and [codekeep.io/screenshot](https://codekeep.io/screenshot). Unlike those tools, `termshot` does not blindly apply syntax highlighting to some provided text; instead it reads the ANSI escape codes ("rich text") logged by most command-line tools and uses it to generate a high-fidelity "screenshot" of your terminal output.

Like `time`, `watch`, or `perf`, just prefix the command you want to screenshot with `termshot`.

```sh
termshot ls -a
```

This will generate an image file called `out.png` in the current directory.

![basic termshot](.doc/example-cmd-ls-a.png)

In some cases, if your target command contains _pipes_—there may still be ambiguity, even with `--`. In these cases, wrap your command in double quotes.

```sh
termshot -- "ls -1 | grep go"
```

![termshot with pipes](.doc/example-cmd-ls-pipe-grep.png)

### Flags to control the look

#### `--show-cmd`/`-c`

Include the target command in the screenshot.

```sh
termshot --show-cmd -- "ls -a"
```

![termshot that shows command](.doc/example-cmd-ls-a.png)

#### `--columns`/`-C`

Enforce that screenshot is wrapped after the provided number of columns. Use this flag to make sure that the screenshot does not exceed a certain horizontal length.

#### `--no-decoration`

Do not draw window decorations (minimize, maximize, and close button).

#### `--no-shadow`

Do not draw window shadow.

#### `--theme`

Specify a color theme for the terminal screenshot. Built-in themes include:
- `default` - Default theme
- `catppuccin-mocha` - Catppuccin Mocha theme
- `catppuccin-latte` - Catppuccin Latte theme
- `nord` - Nord theme
- `dracula` - Dracula theme
- `tokyo-night` - Tokyo Night theme
- `gruvbox-dark` - Gruvbox Dark theme
- `solarized-dark` - Solarized Dark theme

```sh
termshot --theme catppuccin-mocha -- "ls -a"
```

#### `--theme-file`

Load a custom theme from a JSON file. The JSON file should define colors for background, foreground, window decorations, and ANSI colors.

```sh
termshot --theme-file my-theme.json -- "ls -a"
```

#### `--prompt`

Customize the command prompt indicator (default is "➜").

```sh
termshot --show-cmd --prompt "❯" -- "ls -a"
```

#### `--syntax-highlight`

Enable syntax highlighting for the command line. When enabled, commands, keywords, flags, strings, and other tokens are colorized.

```sh
termshot --show-cmd --syntax-highlight -- "docker ps -a | grep running"
```

### Flags for output related settings

#### `--clipboard`/`-b` (only on selected platforms)

Do not create an output file with the screenshot, but save the screenshot image into the operating system clipboard.

_Note:_ Only available on some platforms. Check `termshot` help to see if flag is available.

#### `--filename`/`-f`

Specify a path where the screenshot should be generated. This can be an absolute path or a relative path; relative paths will be resolved relative to the current working directory. Defaults to `out.png`.

```sh
termshot -- "ls -a" # defaults to <cwd>/out.png
termshot --filename my-image.png -- "ls -a"
termshot --filename screenshots/my-image.png -- "ls -a"
termshot --filename /Desktop/my-image.png -- "ls -a"
```

### Flags for shell configuration

#### `--shell`

Specify a custom shell to use for command execution (e.g., `/bin/zsh`, `/bin/bash`).

```sh
termshot --shell /bin/zsh -- "ls -a"
```

#### `--shell-config`

Specify a shell configuration file to source before running the command (e.g., `~/.zshrc`, `~/.bashrc`). This is useful for loading custom prompts like powerlevel10k.

```sh
termshot --shell /bin/zsh --shell-config ~/.zshrc -- "ls -a"
```

#### `--shell-opts`

Specify additional shell options as a comma-separated list.

```sh
termshot --shell /bin/zsh --shell-opts "-i,-l" -- "ls -a"
```

### Flags to control content

#### `--edit`/`-e`

Edit the output before generating the screenshot. This will open the rich text output in the editor configured in `$EDITOR`, using `vi` as a fallback. Use this flag to remove unwanted or sensitive output.

```sh
termshot --edit -- "ls -a"
```

### Miscellaneous flags

#### `--raw-write <file>`

Write command output as-is into the file that is specified as the flag argument. No screenshot is being created. The command-line flag `--filename` has no effect, when `--raw-write` is used.

#### `--raw-read <file>`

Read input from provided file instead of running a command. If this flag is being used, no pseudo terminal is being created to execute a command. The command-line flags `--show-cmd`, and `--edit` have no effect, when `--raw-read` is used.

#### `--improved-ansi`

Enable improved ANSI parser with better cursor handling (enabled by default). This helps with prompts that use cursor positioning like powerlevel10k.

```sh
termshot --improved-ansi -- "ls -a"
```

#### `--version`/`-v`

Print the version of `termshot` installed.

```sh
$ termshot --version
termshot version 0.2.5
```

### Multiple commands

In order to work, `termshot` uses a pseudo terminal for the command to be executed. For advanced use cases, you can invoke a fully interactive shell, run several commands, and capture the entire output. The screenshot will be created once you terminate the shell.

```sh
termshot /bin/zsh
```

> _Please note:_ This project is work in progress. The improved ANSI parser now handles most cursor positioning sequences, including those used by powerlevel10k and similar prompts. You can customize the appearance with themes, custom prompts, and syntax highlighting.

## Advanced Examples

### Using with ZSH and powerlevel10k

```sh
termshot --shell /bin/zsh --shell-config ~/.zshrc --theme catppuccin-mocha --show-cmd -- "git status"
```

### Custom prompt and syntax highlighting

```sh
termshot --show-cmd --prompt "❯" --syntax-highlight -- "docker ps -a | grep running"
```

### Creating a custom theme

Create a JSON file (e.g., `my-theme.json`):

```json
{
  "name": "My Theme",
  "background": "#1e1e2e",
  "foreground": "#cdd6f4",
  "window_red": "#f38ba8",
  "window_yellow": "#f9e2af",
  "window_green": "#a6e3a1",
  "window_border": "#45475a",
  "shadow": "#11111b66",
  "black": "#45475a",
  "red": "#f38ba8",
  "green": "#a6e3a1",
  "yellow": "#f9e2af",
  "blue": "#89b4fa",
  "magenta": "#f5c2e7",
  "cyan": "#94e2d5",
  "white": "#bac2de",
  "bright_black": "#585b70",
  "bright_red": "#f38ba8",
  "bright_green": "#a6e3a1",
  "bright_yellow": "#f9e2af",
  "bright_blue": "#89b4fa",
  "bright_magenta": "#f5c2e7",
  "bright_cyan": "#94e2d5",
  "bright_white": "#a6adc8"
}
```

Then use it:

```sh
termshot --theme-file my-theme.json -- "ls -la"
```
