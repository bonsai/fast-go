package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const trackColor = "#3A3A3A"

func parseHex(s string) (r, g, b int) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 255, 255, 255
	}
	pr, _ := strconv.ParseInt(s[0:2], 16, 32)
	pg, _ := strconv.ParseInt(s[2:4], 16, 32)
	pb, _ := strconv.ParseInt(s[4:6], 16, 32)
	return int(pr), int(pg), int(pb)
}

// lerpHex blends two #RRGGBB colors. t is clamped to [0, 1].
func lerpHex(a, b string, t float64) string {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	ar, ag, ab := parseHex(a)
	br, bg, bb := parseHex(b)
	mix := func(x, y int) int { return x + int(float64(y-x)*t) }
	return fmt.Sprintf("#%02X%02X%02X", mix(ar, br), mix(ag, bg), mix(ab, bb))
}

// gradientBar draws a progress bar whose filled cells fade from one color
// to another along its length.
func gradientBar(width int, ratio float64, from, to string) string {
	if width < 1 {
		width = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))

	var b strings.Builder
	for i := 0; i < filled; i++ {
		c := lerpHex(from, to, float64(i)/float64(width-1))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render("━"))
	}
	if filled < width {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(trackColor)).
			Render(strings.Repeat("╌", width-filled)))
	}
	return b.String()
}

var sparkRunes = []rune("▁▂▃▄▅▆▇█")

// sparkline renders the last `width` values as bars scaled to the peak,
// each bar colored by its own height.
func sparkline(values []float64, width int, from, to string) string {
	if len(values) == 0 || width < 1 {
		return ""
	}
	if len(values) > width {
		values = values[len(values)-width:]
	}
	max := 0.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		max = 1
	}
	var b strings.Builder
	for _, v := range values {
		t := v / max
		idx := int(t * float64(len(sparkRunes)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkRunes) {
			idx = len(sparkRunes) - 1
		}
		c := lerpHex(from, to, t)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(string(sparkRunes[idx])))
	}
	return b.String()
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func spinner(frame int) string {
	return spinnerFrames[(frame/3)%len(spinnerFrames)]
}
