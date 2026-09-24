package usage

import (
	"strings"
	"unicode"
)

func TerminalText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == 0x200e || r == 0x200f || r >= 0x202a && r <= 0x202e || r >= 0x2066 && r <= 0x2069 {
			return ' '
		}
		return r
	}, s)
}

func TerminalReport(r Report) Report {
	r.Provider = TerminalText(r.Provider)
	r.Account = TerminalText(r.Account)
	r.Dir = TerminalText(r.Dir)
	r.Plan = TerminalText(r.Plan)
	r.Email = TerminalText(r.Email)
	r.Error = TerminalText(r.Error)
	r.Windows = append([]Window(nil), r.Windows...)
	for i := range r.Windows {
		r.Windows[i].Name = TerminalText(r.Windows[i].Name)
	}
	return r
}

func TerminalReports(reports []Report) []Report {
	out := make([]Report, len(reports))
	for i, r := range reports {
		out[i] = TerminalReport(r)
	}
	return out
}
