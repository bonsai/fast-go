package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const digitRows = 5

// Block-style numeral font, 5 rows tall. Each glyph's rows share one width.
var digitFont = map[byte][]string{
	'0': {
		" █████ ",
		"██   ██",
		"██   ██",
		"██   ██",
		" █████ ",
	},
	'1': {
		"  ██ ",
		" ███ ",
		"  ██ ",
		"  ██ ",
		" ████",
	},
	'2': {
		" █████ ",
		"     ██",
		"  ████ ",
		"██     ",
		"███████",
	},
	'3': {
		"██████ ",
		"     ██",
		" █████ ",
		"     ██",
		"██████ ",
	},
	'4': {
		"██   ██",
		"██   ██",
		"███████",
		"     ██",
		"     ██",
	},
	'5': {
		"███████",
		"██     ",
		"██████ ",
		"     ██",
		"██████ ",
	},
	'6': {
		" █████ ",
		"██     ",
		"██████ ",
		"██   ██",
		" █████ ",
	},
	'7': {
		"███████",
		"     ██",
		"    ██ ",
		"   ██  ",
		"  ██   ",
	},
	'8': {
		" █████ ",
		"██   ██",
		" █████ ",
		"██   ██",
		" █████ ",
	},
	'9': {
		" █████ ",
		"██   ██",
		" ██████",
		"     ██",
		" █████ ",
	},
	'.': {
		"  ",
		"  ",
		"  ",
		"  ",
		"██",
	},
}

// RenderBigDigits renders a numeric string (digits and '.') as block text
// with a vertical gradient from top to bottom.
func RenderBigDigits(s string, top, bottom string) string {
	var out strings.Builder
	for row := 0; row < digitRows; row++ {
		var line strings.Builder
		for i := 0; i < len(s); i++ {
			glyph, ok := digitFont[s[i]]
			if !ok {
				continue
			}
			if line.Len() > 0 {
				line.WriteByte(' ')
			}
			line.WriteString(glyph[row])
		}
		c := lerpHex(top, bottom, float64(row)/float64(digitRows-1))
		out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(line.String()))
		if row < digitRows-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}
