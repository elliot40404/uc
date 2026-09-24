package tui

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

const barRune = "━"

func (t theme) gradient(width int) []color.Color {
	return lipgloss.Blend1D(max(width, 2), t.good, t.warn, t.bad)
}

func (t theme) level(pct float64) color.Color {
	return t.levels[int(math.Round(min(max(pct, 0), 100)))]
}

func (t theme) bar(pct float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := filledCells(pct, width)
	colors := t.gradient(width)
	var b strings.Builder
	for i := range filled {
		b.WriteString(t.fg(colors[i]).Render(barRune))
	}
	b.WriteString(t.fg(t.track).Render(strings.Repeat(barRune, width-filled)))
	return b.String()
}

func filledCells(pct float64, width int) int {
	n := int(math.Round(pct / 100 * float64(width)))
	if pct > 0 && n == 0 {
		n = 1
	}
	return min(max(n, 0), width)
}
