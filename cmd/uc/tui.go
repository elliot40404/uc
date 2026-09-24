package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/term"
	"github.com/elliot40404/uc/internal/tui"
)

const (
	minEvery     = time.Minute
	defaultWidth = 120
)

func (e env) tui(opts options) error {
	model := tui.New(e.Reports, opts.tui())
	if opts.mini {
		model = tui.NewMini(e.Reports, opts.tui())
	}
	return runProgram(model, opts)
}

func runProgram(model tui.Model, opts options) error {
	if opts.every < minEvery {
		return fmt.Errorf("--every must be at least %s", minEvery)
	}
	_, err := tea.NewProgram(model).Run()
	return err
}

func (e env) mini(opts options) error {
	reports, now, err := e.Reports(context.Background())
	if err := showCacheWarning(err, now); err != nil {
		return err
	}
	dark := true
	if term.IsTerminal(os.Stdout) && term.IsTerminal(os.Stdin) {
		dark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	}
	return tui.Print(os.Stdout, reports, now, term.Width(os.Stdout, defaultWidth), dark, opts.tui())
}

func interactive(opts options) bool {
	return !opts.print && !opts.json && term.IsTerminal(os.Stdin) && term.IsTerminal(os.Stdout)
}
