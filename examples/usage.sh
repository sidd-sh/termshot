#!/bin/bash
# Example usage of termshot with new features

echo "=== Termshot Enhanced Features Demo ==="
echo ""

# Example 1: Basic usage with Catppuccin Mocha theme
echo "1. Using Catppuccin Mocha theme:"
echo "   termshot --theme catppuccin-mocha --show-cmd -- \"ls -la\""
echo ""

# Example 2: Custom shell configuration (ZSH with config)
echo "2. Using ZSH with custom configuration:"
echo "   termshot --shell /bin/zsh --shell-config ~/.zshrc --show-cmd -- \"git status\""
echo ""

# Example 3: Custom prompt with syntax highlighting
echo "3. Custom prompt with syntax highlighting:"
echo "   termshot --show-cmd --prompt \"❯\" --syntax-highlight -- \"docker ps -a | grep running\""
echo ""

# Example 4: Nord theme with custom prompt
echo "4. Nord theme with custom prompt:"
echo "   termshot --theme nord --show-cmd --prompt \"λ\" -- \"cat package.json\""
echo ""

# Example 5: Using custom theme file
echo "5. Using custom theme file:"
echo "   termshot --theme-file examples/catppuccin-mocha.json --show-cmd -- \"npm run build\""
echo ""

# Example 6: Dracula theme without window decorations
echo "6. Dracula theme without decorations:"
echo "   termshot --theme dracula --no-decoration --show-cmd -- \"python --version\""
echo ""

# Example 7: Tokyo Night theme with fixed columns
echo "7. Tokyo Night theme with fixed width:"
echo "   termshot --theme tokyo-night --columns 100 --show-cmd -- \"kubectl get pods\""
echo ""

# Example 8: Gruvbox Dark theme with no shadow
echo "8. Gruvbox Dark theme without shadow:"
echo "   termshot --theme gruvbox-dark --no-shadow --show-cmd -- \"cargo build --release\""
echo ""

# Example 9: Improved ANSI parsing for complex prompts
echo "9. With improved ANSI parsing (for powerlevel10k):"
echo "   termshot --shell /bin/zsh --shell-config ~/.zshrc --improved-ansi --theme catppuccin-mocha -- \"echo 'Hello World'\""
echo ""

# Example 10: All features combined
echo "10. All features combined:"
echo "    termshot --shell /bin/zsh \\"
echo "             --shell-config ~/.zshrc \\"
echo "             --theme catppuccin-mocha \\"
echo "             --show-cmd \\"
echo "             --prompt \"➜\" \\"
echo "             --syntax-highlight \\"
echo "             --improved-ansi \\"
echo "             --filename my-screenshot.png \\"
echo "             -- \"git log --oneline -5\""
echo ""

echo "=== Theme Options ==="
echo "Available built-in themes:"
echo "  - default"
echo "  - catppuccin-mocha"
echo "  - catppuccin-latte"
echo "  - nord"
echo "  - dracula"
echo "  - tokyo-night"
echo "  - gruvbox-dark"
echo "  - solarized-dark"
echo ""

echo "=== Creating Custom Themes ==="
echo "Create a JSON file with the following structure:"
echo ""
cat << 'EOF'
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
EOF
