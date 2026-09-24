package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type theme struct {
	text    color.Color
	subtle  color.Color
	faint   color.Color
	track   color.Color
	accent  color.Color
	good    color.Color
	warn    color.Color
	bad     color.Color
	claude  color.Color
	codex   color.Color
	onBadge color.Color
	levels  []color.Color
}

func newTheme(dark bool) theme {
	pick := lipgloss.LightDark(dark)
	t := theme{
		text:    pick(lipgloss.Color("#1E1E2E"), lipgloss.Color("#E4E4EF")),
		subtle:  pick(lipgloss.Color("#6C6F85"), lipgloss.Color("#9399B2")),
		faint:   pick(lipgloss.Color("#BCC0CC"), lipgloss.Color("#45475A")),
		track:   pick(lipgloss.Color("#DCE0E8"), lipgloss.Color("#313244")),
		accent:  pick(lipgloss.Color("#7C3AED"), lipgloss.Color("#A78BFA")),
		good:    pick(lipgloss.Color("#16A34A"), lipgloss.Color("#4ADE80")),
		warn:    pick(lipgloss.Color("#CA8A04"), lipgloss.Color("#FACC15")),
		bad:     pick(lipgloss.Color("#DC2626"), lipgloss.Color("#F87171")),
		claude:  pick(lipgloss.Color("#C15F3C"), lipgloss.Color("#E08A68")),
		codex:   pick(lipgloss.Color("#0E8A6A"), lipgloss.Color("#34D399")),
		onBadge: lipgloss.Color("#FFFFFF"),
	}
	t.levels = t.gradient(101)
	return t
}

func (t theme) provider(name string) color.Color {
	if name == "codex" {
		return t.codex
	}
	return t.claude
}

func (t theme) fg(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

func (t theme) badge(label string, bg color.Color) string {
	return lipgloss.NewStyle().Background(bg).Foreground(t.onBadge).Bold(true).Padding(0, 1).Render(label)
}
