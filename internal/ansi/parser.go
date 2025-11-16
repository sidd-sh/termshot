package ansi

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gonvenience/bunt"
)

// VirtualTerminal represents a virtual terminal buffer that properly handles
// cursor positioning, clearing operations, and ANSI escape sequences
type VirtualTerminal struct {
	lines      [][]bunt.ColoredRune
	cursorX    int
	cursorY    int
	settings   uint64
	maxColumns int
}

// NewVirtualTerminal creates a new virtual terminal with the specified column limit
func NewVirtualTerminal(maxColumns int) *VirtualTerminal {
	if maxColumns <= 0 {
		maxColumns = 80
	}
	return &VirtualTerminal{
		lines:      [][]bunt.ColoredRune{{}},
		cursorX:    0,
		cursorY:    0,
		maxColumns: maxColumns,
	}
}

// Parse processes ANSI input and returns a bunt.String with proper handling
// of cursor movements and escape sequences
func (vt *VirtualTerminal) Parse(input io.Reader) (*bunt.String, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// First, parse with bunt to get the color information
	parsed, err := bunt.ParseStream(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ANSI stream: %w", err)
	}

	// Process each rune with cursor awareness
	i := 0
	for i < len(*parsed) {
		cr := (*parsed)[i]
		
		// Check for escape sequences
		if cr.Symbol == '\033' {
			// Handle CSI sequences (ESC [)
			if i+1 < len(*parsed) && (*parsed)[i+1].Symbol == '[' {
				consumed, handled := vt.handleCSI(parsed, i)
				if handled {
					i += consumed
					continue
				}
			}
			i++
			continue
		}

		// Handle carriage return
		if cr.Symbol == '\r' {
			vt.cursorX = 0
			i++
			continue
		}

		// Handle newline
		if cr.Symbol == '\n' {
			vt.cursorY++
			vt.cursorX = 0
			vt.ensureLineExists(vt.cursorY)
			i++
			continue
		}

		// Handle backspace
		if cr.Symbol == '\b' {
			if vt.cursorX > 0 {
				vt.cursorX--
			}
			i++
			continue
		}

		// Regular character - place it at cursor position
		vt.ensureLineExists(vt.cursorY)
		
		// Ensure the line has enough capacity
		for len(vt.lines[vt.cursorY]) <= vt.cursorX {
			vt.lines[vt.cursorY] = append(vt.lines[vt.cursorY], bunt.ColoredRune{Symbol: ' '})
		}
		
		// Check for line wrap
		if vt.cursorX >= vt.maxColumns {
			vt.cursorY++
			vt.cursorX = 0
			vt.ensureLineExists(vt.cursorY)
		}

		vt.lines[vt.cursorY][vt.cursorX] = cr
		vt.cursorX++
		i++
	}

	// Convert back to bunt.String
	return vt.toBuntString(), nil
}

// handleCSI processes CSI (Control Sequence Introducer) escape sequences
// Returns: (number of runes consumed, whether sequence was handled)
func (vt *VirtualTerminal) handleCSI(parsed *bunt.String, start int) (int, bool) {
	if start+2 >= len(*parsed) {
		return 0, false
	}

	// Start after ESC[
	i := start + 2
	var params []int
	var paramBuf strings.Builder

	// Parse parameters
	for i < len(*parsed) {
		ch := (*parsed)[i].Symbol
		
		if ch >= '0' && ch <= '9' {
			paramBuf.WriteRune(ch)
			i++
		} else if ch == ';' {
			if paramBuf.Len() > 0 {
				val, _ := strconv.Atoi(paramBuf.String())
				params = append(params, val)
				paramBuf.Reset()
			} else {
				params = append(params, 0)
			}
			i++
		} else {
			// Command letter found
			if paramBuf.Len() > 0 {
				val, _ := strconv.Atoi(paramBuf.String())
				params = append(params, val)
			}
			break
		}
	}

	if i >= len(*parsed) {
		return 0, false
	}

	command := (*parsed)[i].Symbol
	consumed := i - start + 1

	// Handle different CSI commands
	switch command {
	case 'A': // Cursor Up
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		vt.cursorY -= n
		if vt.cursorY < 0 {
			vt.cursorY = 0
		}
		return consumed, true

	case 'B': // Cursor Down
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		vt.cursorY += n
		vt.ensureLineExists(vt.cursorY)
		return consumed, true

	case 'C': // Cursor Forward
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		vt.cursorX += n
		return consumed, true

	case 'D': // Cursor Back
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		vt.cursorX -= n
		if vt.cursorX < 0 {
			vt.cursorX = 0
		}
		return consumed, true

	case 'G': // Cursor Horizontal Absolute
		n := 1
		if len(params) > 0 {
			n = params[0]
		}
		vt.cursorX = n - 1
		if vt.cursorX < 0 {
			vt.cursorX = 0
		}
		return consumed, true

	case 'H', 'f': // Cursor Position
		row, col := 1, 1
		if len(params) > 0 {
			row = params[0]
		}
		if len(params) > 1 {
			col = params[1]
		}
		vt.cursorY = row - 1
		vt.cursorX = col - 1
		if vt.cursorY < 0 {
			vt.cursorY = 0
		}
		if vt.cursorX < 0 {
			vt.cursorX = 0
		}
		vt.ensureLineExists(vt.cursorY)
		return consumed, true

	case 'J': // Erase in Display
		mode := 0
		if len(params) > 0 {
			mode = params[0]
		}
		switch mode {
		case 0: // Clear from cursor to end of screen
			vt.clearToEndOfScreen()
		case 1: // Clear from cursor to beginning of screen
			vt.clearToBeginningOfScreen()
		case 2, 3: // Clear entire screen
			vt.clearScreen()
		}
		return consumed, true

	case 'K': // Erase in Line
		mode := 0
		if len(params) > 0 {
			mode = params[0]
		}
		switch mode {
		case 0: // Clear from cursor to end of line
			vt.clearToEndOfLine()
		case 1: // Clear from cursor to beginning of line
			vt.clearToBeginningOfLine()
		case 2: // Clear entire line
			vt.clearLine()
		}
		return consumed, true

	case 's': // Save cursor position
		// Could implement if needed
		return consumed, true

	case 'u': // Restore cursor position
		// Could implement if needed
		return consumed, true
	}

	return 0, false
}

// ensureLineExists makes sure the line at the given index exists
func (vt *VirtualTerminal) ensureLineExists(line int) {
	for len(vt.lines) <= line {
		vt.lines = append(vt.lines, []bunt.ColoredRune{})
	}
}

// clearToEndOfLine clears from cursor to end of current line
func (vt *VirtualTerminal) clearToEndOfLine() {
	if vt.cursorY < len(vt.lines) {
		if vt.cursorX < len(vt.lines[vt.cursorY]) {
			vt.lines[vt.cursorY] = vt.lines[vt.cursorY][:vt.cursorX]
		}
	}
}

// clearToBeginningOfLine clears from beginning of line to cursor
func (vt *VirtualTerminal) clearToBeginningOfLine() {
	if vt.cursorY < len(vt.lines) {
		for i := 0; i <= vt.cursorX && i < len(vt.lines[vt.cursorY]); i++ {
			vt.lines[vt.cursorY][i] = bunt.ColoredRune{Symbol: ' '}
		}
	}
}

// clearLine clears the entire current line
func (vt *VirtualTerminal) clearLine() {
	if vt.cursorY < len(vt.lines) {
		vt.lines[vt.cursorY] = []bunt.ColoredRune{}
	}
}

// clearToEndOfScreen clears from cursor to end of screen
func (vt *VirtualTerminal) clearToEndOfScreen() {
	vt.clearToEndOfLine()
	if vt.cursorY+1 < len(vt.lines) {
		vt.lines = vt.lines[:vt.cursorY+1]
	}
}

// clearToBeginningOfScreen clears from beginning of screen to cursor
func (vt *VirtualTerminal) clearToBeginningOfScreen() {
	for i := 0; i < vt.cursorY && i < len(vt.lines); i++ {
		vt.lines[i] = []bunt.ColoredRune{}
	}
	vt.clearToBeginningOfLine()
}

// clearScreen clears the entire screen
func (vt *VirtualTerminal) clearScreen() {
	vt.lines = [][]bunt.ColoredRune{{}}
	vt.cursorX = 0
	vt.cursorY = 0
}

// toBuntString converts the virtual terminal buffer to a bunt.String
func (vt *VirtualTerminal) toBuntString() *bunt.String {
	var result bunt.String

	for lineIdx, line := range vt.lines {
		for _, cr := range line {
			result = append(result, cr)
		}
		
		// Add newline except for the last line
		if lineIdx < len(vt.lines)-1 {
			// Use the settings from the last character on the line, or default
			var settings uint64
			if len(line) > 0 {
				settings = line[len(line)-1].Settings
			}
			result = append(result, bunt.ColoredRune{Symbol: '\n', Settings: settings})
		}
	}

	return &result
}
