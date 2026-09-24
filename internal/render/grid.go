package render

import (
	"strings"
	"unicode/utf8"
)

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	red    = "\x1b[31m"
	dim    = "\x1b[2m"
)

type seg struct {
	color string
	text  string
}

type cell []seg

func text(color, s string) cell {
	return cell{{color, s}}
}

func (c cell) width() int {
	n := 0
	for _, s := range c {
		n += utf8.RuneCountInString(s.text)
	}
	return n
}

func (c cell) render(color bool) string {
	var b strings.Builder
	for _, s := range c {
		if color && s.color != "" {
			b.WriteString(s.color + s.text + reset)
		} else {
			b.WriteString(s.text)
		}
	}
	return b.String()
}

func grid(rows [][]cell, color bool) []string {
	widths := map[int]int{}
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], c.width())
		}
	}
	lines := make([]string, len(rows))
	for i, r := range rows {
		var b strings.Builder
		for j, c := range r {
			b.WriteString(c.render(color))
			if j < len(r)-1 {
				b.WriteString(strings.Repeat(" ", widths[j]-c.width()+2))
			}
		}
		lines[i] = b.String()
	}
	return lines
}
