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
	"image/color"
	"io"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/gonvenience/bunt"
)

const (
	// emulatorDefaultColumns is the emulator width when no fixed column count
	// is configured, wide enough that captured lines keep their length
	emulatorDefaultColumns = 500

	// emulatorRows is the height of the visible emulator screen, everything
	// that scrolls off the top ends up in the scrollback buffer
	emulatorRows = 100

	// emulatorScrollbackLines is the scrollback limit, large enough to never
	// drop lines of a realistic capture
	emulatorScrollbackLines = 1 << 20
)

// Bits of the bunt.ColoredRune settings, see bunt's model
const (
	fgMask        = 0x1
	bgMask        = 0x2
	boldMask      = 0x4
	italicMask    = 0x8
	underlineMask = 0x10
)

// emulate feeds the captured terminal output into a virtual terminal and
// returns what the screen (including scrollback) looks like in the end, so
// that cursor movements, erase sequences and line rewrites are applied
// instead of being rendered as-is
func (s *Scaffold) emulate(input []byte) bunt.String {
	columns := s.columns
	if columns <= 0 {
		columns = emulatorDefaultColumns
	}

	emu := vt.NewEmulator(columns, emulatorRows)
	emu.SetScrollbackSize(emulatorScrollbackLines)

	// Replies to terminal queries (e.g. cursor position reports) go into an
	// unbuffered pipe, discard them so that writing to the emulator never blocks
	replies := emu.InputPipe()
	go func() { _, _ = io.Copy(io.Discard, emu) }()
	defer func() {
		if closer, ok := replies.(io.Closer); ok {
			_ = closer.Close()
			return
		}

		_ = emu.Close()
	}()

	_, _ = emu.WriteString(s.emulatorSetup(columns))
	_, _ = emu.Write(input)

	var lines []uv.Line
	if sb := emu.Scrollback(); sb != nil {
		lines = append(lines, sb.Lines()...)
	}

	for y := 0; y < emu.Height(); y++ {
		line := make(uv.Line, emu.Width())
		for x := range line {
			if cell := emu.CellAt(x, y); cell != nil {
				line[x] = *cell
			}
		}

		lines = append(lines, line)
	}

	rows := make([]bunt.String, len(lines))
	lastRow := -1
	for i, line := range lines {
		rows[i] = s.convertLine(emu, line)
		if len(rows[i]) > 0 {
			lastRow = i
		}
	}

	// Keep a final newline if the output ended with one, but drop all
	// trailing blank rows of the screen
	cursorRow := len(lines) - emu.Height() + emu.CursorPosition().Y

	var result bunt.String
	for i := 0; i <= lastRow; i++ {
		result = append(result, rows[i]...)
		if i < lastRow || cursorRow > lastRow {
			result = append(result, bunt.ColoredRune{Symbol: '\n'})
		}
	}

	return result
}

// emulatorSetup returns the sequences to prepare the emulator: a line feed
// also returns the carriage (input often only uses \n), and tab stops match
// the configured tab width
func (s *Scaffold) emulatorSetup(columns int) string {
	var setup strings.Builder
	setup.WriteString(ansi.SetMode(ansi.ModeLineFeedNewLine))

	if s.tabSpaces > 0 {
		setup.WriteString(ansi.TabClear(3))
		for col := s.tabSpaces; col < columns; col += s.tabSpaces {
			setup.WriteString(ansi.CursorPosition(col+1, 1) + ansi.HorizontalTabSet)
		}
	}

	setup.WriteString(ansi.CursorHomePosition)
	return setup.String()
}

// convertLine converts a line of emulator cells into colored runes, with
// trailing blank cells removed
func (s *Scaffold) convertLine(emu *vt.Emulator, line uv.Line) bunt.String {
	var result bunt.String
	for _, cell := range line {
		// Zero width cells are the placeholders after a wide character
		if cell.Width == 0 && cell.Content == "" {
			continue
		}

		content := cell.Content
		if content == "" {
			content = " "
		}

		settings := cellSettings(emu, cell.Style)
		for _, r := range content {
			result = append(result, bunt.ColoredRune{Symbol: r, Settings: settings})
		}
	}

	for len(result) > 0 {
		last := result[len(result)-1]
		if last.Symbol != ' ' || last.Settings&(bgMask|underlineMask) != 0 {
			break
		}

		result = result[:len(result)-1]
	}

	return result
}

// cellSettings translates the style of an emulator cell into bunt settings
func cellSettings(emu *vt.Emulator, style uv.Style) uint64 {
	var settings uint64

	if style.Fg != nil {
		r, g, b := rgb(emu, style.Fg)
		settings |= fgMask | r<<8 | g<<16 | b<<24
	}

	if style.Bg != nil {
		r, g, b := rgb(emu, style.Bg)
		settings |= bgMask | r<<32 | g<<40 | b<<48
	}

	if style.Attrs&uv.AttrBold != 0 {
		settings |= boldMask
	}

	if style.Attrs&uv.AttrItalic != 0 {
		settings |= italicMask
	}

	if style.Underline != uv.UnderlineNone {
		settings |= underlineMask
	}

	return settings
}

// rgb returns the 8-bit color channels, resolving palette colors through the
// emulator so that the standard ANSI colors can be remapped to theme colors
func rgb(emu *vt.Emulator, c color.Color) (r, g, b uint64) {
	switch idx := c.(type) {
	case ansi.BasicColor:
		c = emu.IndexedColor(int(idx))
	case ansi.IndexedColor:
		c = emu.IndexedColor(int(idx))
	}

	if c == nil {
		return 0, 0, 0
	}

	cr, cg, cb, _ := c.RGBA()
	return uint64(cr >> 8), uint64(cg >> 8), uint64(cb >> 8)
}
