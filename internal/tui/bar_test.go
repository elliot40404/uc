package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestFilledCells(t *testing.T) {
	cases := map[float64]int{0: 0, 0.5: 1, 50: 10, 100: 20, 140: 20, -3: 0}
	for pct, want := range cases {
		if got := filledCells(pct, 20); got != want {
			t.Errorf("filledCells(%v) = %d, want %d", pct, got, want)
		}
	}
}

func TestBarWidth(t *testing.T) {
	th := newTheme(true)
	for _, pct := range []float64{0, 33, 100} {
		if got := lipgloss.Width(th.bar(pct, 24)); got != 24 {
			t.Errorf("bar(%v) width = %d, want 24", pct, got)
		}
	}
	if th.bar(50, 0) != "" {
		t.Error("zero width bar must be empty")
	}
}

func TestLevelEnds(t *testing.T) {
	th := newTheme(true)
	if th.level(0) != th.levels[0] || th.level(250) != th.levels[100] {
		t.Error("level must clamp to gradient ends")
	}
}
