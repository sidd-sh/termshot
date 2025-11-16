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
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/font"
	"github.com/gonvenience/term"
	"github.com/homeport/termshot/internal/highlight"
	"github.com/homeport/termshot/internal/theme"
	imgfont "golang.org/x/image/font"
)

const (
	defaultFontSize = 12
	defaultFontDPI  = 144
)

// commandIndicator is the string to be used to indicate the command in the screenshot
var commandIndicator = func() string {
	if val, ok := os.LookupEnv("TS_COMMAND_INDICATOR"); ok {
		return val
	}

	return "❯"
}()

type Scaffold struct {
	content bunt.String

	factor float64

	columns int

	defaultForegroundColor color.Color

	clipCanvas bool

	drawDecorations bool
	drawShadow      bool

	shadowBaseColor string
	shadowRadius    uint8
	shadowOffsetX   float64
	shadowOffsetY   float64

	padding float64
	margin  float64

	regular     imgfont.Face
	bold        imgfont.Face
	italic      imgfont.Face
	boldItalic  imgfont.Face
	lineSpacing float64
	tabSpaces   int

	// Theme support
	currentTheme theme.Theme
	
	// Prompt customization
	customPrompt string
	
	// Syntax highlighting
	syntaxHighlight bool
	
	// Prompt detection
	noPromptDetect bool
}

func NewImageCreator() Scaffold {
	f := 2.0

	fontFaceOptions := &truetype.Options{
		Size: f * defaultFontSize,
		DPI:  defaultFontDPI,
	}

	defaultTheme := theme.GetTheme("default")

	return Scaffold{
		defaultForegroundColor: bunt.LightGray,

		factor: f,

		margin:  f * 48,
		padding: f * 24,

		drawDecorations: true,
		drawShadow:      true,

		shadowBaseColor: "#10101066",
		shadowRadius:    uint8(math.Min(f*16, 255)),
		shadowOffsetX:   f * 16,
		shadowOffsetY:   f * 16,

		regular:    font.Hack.Regular(fontFaceOptions),
		bold:       font.Hack.Bold(fontFaceOptions),
		italic:     font.Hack.Italic(fontFaceOptions),
		boldItalic: font.Hack.BoldItalic(fontFaceOptions),

		lineSpacing: 1.2,
		tabSpaces:   2,
		
		currentTheme:    defaultTheme,
		customPrompt:    "",
		syntaxHighlight: false, // Default to false for backward compatibility
	}
}

func (s *Scaffold) SetFontFaceRegular(face imgfont.Face) { s.regular = face }

func (s *Scaffold) SetFontFaceBold(face imgfont.Face) { s.bold = face }

func (s *Scaffold) SetFontFaceItalic(face imgfont.Face) { s.italic = face }

func (s *Scaffold) SetFontFaceBoldItalic(face imgfont.Face) { s.boldItalic = face }

func (s *Scaffold) SetColumns(columns int) { s.columns = columns }

func (s *Scaffold) DrawDecorations(value bool) { s.drawDecorations = value }

func (s *Scaffold) DrawShadow(value bool) { s.drawShadow = value }

func (s *Scaffold) ClipCanvas(value bool) { s.clipCanvas = value }

func (s *Scaffold) SetTheme(t theme.Theme) { 
	s.currentTheme = t 
	// Update foreground color based on theme
	if c, err := theme.ParseColor(t.Foreground); err == nil {
		s.defaultForegroundColor = c
	}
	// Update shadow color based on theme
	s.shadowBaseColor = t.Shadow
}

func (s *Scaffold) SetPrompt(prompt string) { s.customPrompt = prompt }

func (s *Scaffold) EnableSyntaxHighlighting(enable bool) { s.syntaxHighlight = enable }

func (s *Scaffold) DisablePromptDetection(disable bool) { s.noPromptDetect = disable }

func (s *Scaffold) GetFixedColumns() int {
	if s.columns != 0 {
		return s.columns
	}

	columns, _ := term.GetTerminalSize()
	return columns
}

func (s *Scaffold) AddCommand(args ...string) error {
	prompt := commandIndicator
	if s.customPrompt != "" {
		prompt = s.customPrompt
	}
	
	cmdString := strings.Join(args, " ")
	
	// Apply syntax highlighting if enabled
	if s.syntaxHighlight {
		return s.AddContent(strings.NewReader(
			s.syntaxHighlightCommand(prompt, cmdString) + "\n",
		))
	}
	
	// Default behavior without syntax highlighting
	return s.AddContent(strings.NewReader(
		bunt.Sprintf("Lime{%s} DimGray{%s}\n",
			prompt,
			cmdString,
		),
	))
}

func (s *Scaffold) syntaxHighlightCommand(prompt string, command string) string {
	lexer := highlight.NewLexer(command)
	tokens := lexer.Tokenize()
	
	var result strings.Builder
	result.WriteString(bunt.Sprintf("Lime{%s} ", prompt))
	
	for _, token := range tokens {
		coloredText := s.colorizeToken(token)
		result.WriteString(coloredText)
	}
	
	return result.String()
}

func (s *Scaffold) colorizeToken(token highlight.Token) string {
	switch token.Type {
	case highlight.TokenCommand:
		return bunt.Sprintf("Cyan{%s}", token.Value)
	case highlight.TokenKeyword:
		return bunt.Sprintf("Magenta{%s}", token.Value)
	case highlight.TokenFlag:
		return bunt.Sprintf("Yellow{%s}", token.Value)
	case highlight.TokenString:
		return bunt.Sprintf("Green{%s}", token.Value)
	case highlight.TokenVariable:
		return bunt.Sprintf("Blue{%s}", token.Value)
	case highlight.TokenOperator:
		return bunt.Sprintf("Red{%s}", token.Value)
	case highlight.TokenComment:
		return bunt.Sprintf("DimGray{%s}", token.Value)
	case highlight.TokenNumber:
		return bunt.Sprintf("Magenta{%s}", token.Value)
	case highlight.TokenPath:
		return bunt.Sprintf("Cyan{%s}", token.Value)
	default:
		return token.Value
	}
}

// enhanceRawContent detects prompts and commands in raw content, adds spacing and highlighting
func (s *Scaffold) enhanceRawContent(parsed *bunt.String) *bunt.String {
	if len(*parsed) == 0 {
		return parsed
	}

	// Skip enhancement if disabled
	if s.noPromptDetect {
		return parsed
	}

	// Check if first line looks like a prompt (starts with common prompt indicators)
	promptIndicators := []string{"➜", "❯", "$", "#", "λ", "→", "%"}
	firstLine := s.getFirstLine(parsed)
	hasPrompt := false

	for _, indicator := range promptIndicators {
		if strings.HasPrefix(firstLine, indicator) {
			hasPrompt = true
			break
		}
	}

	if !hasPrompt {
		return parsed
	}

	// Find the end of the first line (the command line)
	var result bunt.String
	var firstLineEnd int
	for i, cr := range *parsed {
		if cr.Symbol == '\n' {
			firstLineEnd = i
			break
		}
	}

	if firstLineEnd == 0 || firstLineEnd >= len(*parsed) {
		return parsed // No newline found or invalid position, return as-is
	}

	// Highlight the first line (command) with syntax highlighting
	commandLine := (*parsed)[:firstLineEnd]
	highlightedCommand := s.highlightCommandLine(commandLine)

	// Add the highlighted command
	result = append(result, highlightedCommand...)

	// Add a newline
	result = append(result, bunt.ColoredRune{Symbol: '\n'})

	// Check if there's any content after the command
	hasOutputAfter := firstLineEnd+1 < len(*parsed)
	if hasOutputAfter {
		// Check if the next line is also a prompt (multiple commands)
		nextLineStart := firstLineEnd + 1
		if nextLineStart >= len(*parsed) {
			return &result
		}
		nextLine := s.getLineAt(parsed, nextLineStart)
		isNextPrompt := false
		
		for _, indicator := range promptIndicators {
			if strings.HasPrefix(nextLine, indicator) {
				isNextPrompt = true
				break
			}
		}
		
		// Only add spacing if the next line is NOT another prompt
		if !isNextPrompt && strings.TrimSpace(nextLine) != "" {
			result = append(result, bunt.ColoredRune{Symbol: '\n'})
		}
	}

	// Add the rest of the content (skip the original newline)
	if hasOutputAfter {
		result = append(result, (*parsed)[firstLineEnd+1:]...)
	}

	return &result
}

// getFirstLine extracts the first line as a string
func (s *Scaffold) getFirstLine(parsed *bunt.String) string {
	var line strings.Builder
	for _, cr := range *parsed {
		if cr.Symbol == '\n' {
			break
		}
		line.WriteRune(cr.Symbol)
	}
	return line.String()
}

// getLineAt extracts a line starting from a specific position
func (s *Scaffold) getLineAt(parsed *bunt.String, start int) string {
	var line strings.Builder
	for i := start; i < len(*parsed); i++ {
		if (*parsed)[i].Symbol == '\n' {
			break
		}
		line.WriteRune((*parsed)[i].Symbol)
	}
	return line.String()
}

// highlightCommandLine applies syntax highlighting to a command line
func (s *Scaffold) highlightCommandLine(line bunt.String) bunt.String {
	// Extract plain text from the line
	var plainText strings.Builder
	for _, cr := range line {
		plainText.WriteRune(cr.Symbol)
	}
	text := plainText.String()

	// Find the prompt indicator and the command part
	var promptEnd int
	var promptCharCount int
	promptIndicators := []string{"➜", "❯", "$", "#", "λ", "→", "%"}
	for _, indicator := range promptIndicators {
		if strings.HasPrefix(text, indicator) {
			promptCharCount = len([]rune(indicator)) // Count actual runes, not bytes
			promptEnd = len(indicator)
			break
		}
	}

	if promptEnd == 0 {
		return line // No prompt found
	}

	// Skip whitespace after prompt
	whitespaceStart := promptEnd
	for promptEnd < len(text) && text[promptEnd] == ' ' {
		promptEnd++
	}

	// Find the first word (the command)
	commandStart := promptEnd
	commandEnd := commandStart
	for commandEnd < len(text) && text[commandEnd] != ' ' && text[commandEnd] != '\t' {
		commandEnd++
	}

	// Build the highlighted result
	var result bunt.String
	
	// Magenta color for prompt: rgb(203, 166, 247)
	var promptSettings uint64 = 1 // Foreground color enabled
	promptSettings |= uint64(203) << 8
	promptSettings |= uint64(166) << 16
	promptSettings |= uint64(247) << 24

	// Green color for command
	greenColor, _ := theme.ParseColor(s.currentTheme.Green)
	if greenColor == nil {
		greenColor = color.RGBA{R: 166, G: 227, B: 161, A: 255}
	}
	rgba := greenColor.(color.RGBA)
	var commandSettings uint64 = 1
	commandSettings |= uint64(rgba.R) << 8
	commandSettings |= uint64(rgba.G) << 16
	commandSettings |= uint64(rgba.B) << 24

	// Color the prompt in magenta
	for i := 0; i < promptCharCount && i < len(line); i++ {
		result = append(result, bunt.ColoredRune{
			Symbol:   line[i].Symbol,
			Settings: promptSettings,
		})
	}

	// Add whitespace after prompt (preserve original colors)
	runePos := promptCharCount
	bytePos := promptEnd
	for bytePos > whitespaceStart && runePos < len(line) {
		result = append(result, line[runePos])
		runePos++
		whitespaceStart++
	}

	// Find the actual rune position for command start
	commandStartRune := runePos
	commandEndRune := commandStartRune
	for i := commandStart; i < commandEnd && commandEndRune < len(line); {
		// Count the bytes in this rune
		runeLen := len(string(line[commandEndRune].Symbol))
		i += runeLen
		commandEndRune++
	}

	// Highlight the command in green
	for runePos < commandEndRune && runePos < len(line) {
		result = append(result, bunt.ColoredRune{
			Symbol:   line[runePos].Symbol,
			Settings: commandSettings,
		})
		runePos++
	}

	// Copy the rest as-is
	for runePos < len(line) {
		result = append(result, line[runePos])
		runePos++
	}

	return result
}

// remapAnsiColor maps standard ANSI colors to theme colors
// Only remaps the 16 standard ANSI colors, leaves true RGB colors unchanged
func (s *Scaffold) remapAnsiColor(r, g, b int) color.Color {
	// Standard ANSI 16 colors have very specific RGB values
	// We should only remap these exact values, not arbitrary RGB colors
	type ansiColor struct{ r, g, b int }
	
	// Map of exact standard ANSI color values to theme colors
	// These are the default values used by most terminal emulators
	standardColors := map[ansiColor]string{
		// Normal colors (30-37) - typical default values
		{0, 0, 0}:       s.currentTheme.Black,
		{128, 0, 0}:     s.currentTheme.Red,
		{0, 128, 0}:     s.currentTheme.Green,
		{128, 128, 0}:   s.currentTheme.Yellow,
		{0, 0, 128}:     s.currentTheme.Blue,
		{128, 0, 128}:   s.currentTheme.Magenta,
		{0, 128, 128}:   s.currentTheme.Cyan,
		{192, 192, 192}: s.currentTheme.White,
		
		// Bright colors (90-97) - typical default values
		{128, 128, 128}: s.currentTheme.BrightBlack,
		{255, 0, 0}:     s.currentTheme.BrightRed,
		{0, 255, 0}:     s.currentTheme.BrightGreen,
		{255, 255, 0}:   s.currentTheme.BrightYellow,
		{0, 0, 255}:     s.currentTheme.BrightBlue,
		{255, 0, 255}:   s.currentTheme.BrightMagenta,
		{0, 255, 255}:   s.currentTheme.BrightCyan,
		{255, 255, 255}: s.currentTheme.BrightWhite,
	}
	
	// Check for exact match with standard ANSI colors
	if hexColor, ok := standardColors[ansiColor{r, g, b}]; ok {
		if c, err := theme.ParseColor(hexColor); err == nil {
			return c
		}
	}
	
	// Check with small tolerance (5) for slight variations in ANSI colors
	tolerance := 5
	for ansi, hexColor := range standardColors {
		dr := r - ansi.r
		dg := g - ansi.g
		db := b - ansi.b
		if dr < 0 {
			dr = -dr
		}
		if dg < 0 {
			dg = -dg
		}
		if db < 0 {
			db = -db
		}
		
		if dr <= tolerance && dg <= tolerance && db <= tolerance {
			if c, err := theme.ParseColor(hexColor); err == nil {
				return c
			}
		}
	}
	
	// Not a standard ANSI color - this is likely a true RGB color from the terminal
	// Return it unchanged to preserve the terminal's actual color scheme
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}


func (s *Scaffold) AddContent(in io.Reader) error {
	parsed, err := bunt.ParseStream(in)
	if err != nil {
		return fmt.Errorf("failed to parse input stream: %w", err)
	}

	// Check if this is raw input that might have a prompt at the start
	enhanced := s.enhanceRawContent(parsed)

	var tmp bunt.String
	var counter int
	for _, cr := range *enhanced {
		counter++

		if cr.Symbol == '\n' {
			counter = 0
		}

		// Add an additional newline in case the column
		// count is reached and line wrapping is needed
		if counter > s.GetFixedColumns() {
			counter = 0
			tmp = append(tmp, bunt.ColoredRune{
				Settings: cr.Settings,
				Symbol:   '\n',
			})
		}

		tmp = append(tmp, cr)
	}

	s.content = append(s.content, tmp...)

	return nil
}

func (s *Scaffold) fontHeight() float64 {
	return float64(s.regular.Metrics().Height >> 6)
}

func (s *Scaffold) measureContent() (width float64, height float64) {
	var tmp = make([]rune, len(s.content))
	for i, cr := range s.content {
		tmp[i] = cr.Symbol
	}

	lines := strings.Split(
		strings.TrimSuffix(
			string(tmp),
			"\n",
		),
		"\n",
	)

	// temporary drawer for reference calucation
	tmpDrawer := &imgfont.Drawer{Face: s.regular}

	// width, either by using longest line, or by fixed column value
	switch s.columns {
	case 0: // unlimited: max width of all lines
		for _, line := range lines {
			advance := tmpDrawer.MeasureString(line)
			if lineWidth := float64(advance >> 6); lineWidth > width {
				width = lineWidth
			}
		}

	default: // fixed: max width based on column count
		width = float64(tmpDrawer.MeasureString(strings.Repeat("a", s.GetFixedColumns())) >> 6)
	}

	// height, lines times font height and line spacing
	height = float64(len(lines)) * s.fontHeight() * s.lineSpacing

	return width, height
}

func (s *Scaffold) image() (image.Image, error) {
	var f = func(value float64) float64 { return s.factor * value }

	var (
		corner   = f(6)
		radius   = f(9)
		distance = f(25)
	)

	contentWidth, contentHeight := s.measureContent()

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := s.margin, s.margin
	paddingX, paddingY := s.padding, s.padding

	xOffset := marginX
	yOffset := marginY

	var titleOffset float64
	if s.drawDecorations {
		titleOffset = f(40)
	}

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	// Optional: Apply blurred rounded rectangle to mimic the window shadow
	//
	if s.drawShadow {
		xOffset -= s.shadowOffsetX / 2
		yOffset -= s.shadowOffsetY / 2

		bc := gg.NewContext(int(width), int(height))
		bc.DrawRoundedRectangle(xOffset+s.shadowOffsetX, yOffset+s.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
		bc.SetHexColor(s.shadowBaseColor)
		bc.Fill()

		src := bc.Image()
		dst := image.NewNRGBA(src.Bounds())
		if err := stackblur.Process(dst, src, uint32(s.shadowRadius)); err != nil {
			return nil, err
		}

		dc.DrawImage(dst, 0, 0)
	}

	// Draw rounded rectangle with outline to produce impression of a window
	//
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(s.currentTheme.Background)
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(s.currentTheme.WindowBorder)
	dc.SetLineWidth(f(1))
	dc.Stroke()

	// Optional: Draw window decorations (i.e. three buttons) to produce the
	// impression of an actional window
	//
	if s.drawDecorations {
		colors := []string{s.currentTheme.WindowRed, s.currentTheme.WindowYellow, s.currentTheme.WindowGreen}
		for i, color := range colors {
			dc.DrawCircle(xOffset+paddingX+float64(i)*distance+f(4), yOffset+paddingY+f(4), radius)
			dc.SetHexColor(color)
			dc.Fill()
		}
	}

	// Apply the actual text into the prepared content area of the window
	//
	var x, y = xOffset + paddingX, yOffset + paddingY + titleOffset + s.fontHeight()
	for _, cr := range s.content {
		switch cr.Settings & 0x1C {
		case 4:
			dc.SetFontFace(s.bold)

		case 8:
			dc.SetFontFace(s.italic)

		case 12:
			dc.SetFontFace(s.boldItalic)

		default:
			dc.SetFontFace(s.regular)
		}

		str := string(cr.Symbol)
		w, h := dc.MeasureString(str)

		// background color
		switch cr.Settings & 0x02 { //nolint:gocritic
		case 2:
			bgR := int((cr.Settings >> 32) & 0xFF)
			bgG := int((cr.Settings >> 40) & 0xFF)
			bgB := int((cr.Settings >> 48) & 0xFF)
			
			// Remap ANSI colors to theme colors
			remappedBg := s.remapAnsiColor(bgR, bgG, bgB)
			dc.SetColor(remappedBg)

			dc.DrawRectangle(x, y-h+12, w, h)
			dc.Fill()
		}

		// foreground color
		switch cr.Settings & 0x01 {
		case 1:
			fgR := int((cr.Settings >> 8) & 0xFF)
			fgG := int((cr.Settings >> 16) & 0xFF)
			fgB := int((cr.Settings >> 24) & 0xFF)
			
			// Remap ANSI colors to theme colors
			remappedFg := s.remapAnsiColor(fgR, fgG, fgB)
			dc.SetColor(remappedFg)

		default:
			dc.SetColor(s.defaultForegroundColor)
		}

		switch str {
		case "\n":
			x = xOffset + paddingX
			y += h * s.lineSpacing
			continue

		case "\t":
			x += w * float64(s.tabSpaces)
			continue

		case "✗", "ˣ": // mitigate issue #1 by replacing it with a similar character
			str = "×"
		}

		dc.DrawString(str, x, y)

		// There seems to be no font face based way to do an underlined
		// string, therefore manually draw a line under each character
		if cr.Settings&0x1C == 16 {
			dc.DrawLine(x, y+f(4), x+w, y+f(4))
			dc.SetLineWidth(f(1))
			dc.Stroke()
		}

		x += w
	}

	return dc.Image(), nil
}

// Write writes the scaffold content as PNG into the provided writer
//
// Deprecated: Use [Scaffold.WritePNG] instead.
func (s *Scaffold) Write(w io.Writer) error {
	return s.WritePNG(w)
}

// WritePNG writes the scaffold content as PNG into the provided writer
func (s *Scaffold) WritePNG(w io.Writer) error {
	img, err := s.image()
	if err != nil {
		return err
	}

	// Optional: Clip image to minimum size by removing all surrounding transparent pixels
	//
	if s.clipCanvas {
		if imgRGBA, ok := img.(*image.RGBA); ok {
			var minX, minY = math.MaxInt, math.MaxInt
			var maxX, maxY = 0, 0

			var bounds = imgRGBA.Bounds()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					r, g, b, a := imgRGBA.At(x, y).RGBA()
					isTransparent := r == 0 && g == 0 && b == 0 && a == 0

					if !isTransparent {
						if x < minX {
							minX = x
						}

						if y < minY {
							minY = y
						}

						if x > maxX {
							maxX = x
						}

						if y > maxY {
							maxY = y
						}
					}
				}
			}

			img = imgRGBA.SubImage(image.Rect(minX, minY, maxX, maxY))
		}
	}

	return png.Encode(w, img)
}

// WriteRaw writes the scaffold content as-is into the provided writer
func (s *Scaffold) WriteRaw(w io.Writer) error {
	_, err := w.Write([]byte(s.content.String()))
	return err
}
