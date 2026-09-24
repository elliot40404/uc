package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/usage"
)

func (e env) best(opts options, args []string) error {
	provider, err := bestProvider(args)
	if err != nil {
		return err
	}
	reports, now, err := e.reports(context.Background())
	if err := showCacheWarning(err, now); err != nil {
		return err
	}
	p, ok := usage.Best(reports, provider, now)
	if !ok {
		return errors.New("no account with usage available, all are full, expired or failing")
	}
	if opts.json {
		return render.JSON(os.Stdout, opts.visible([]usage.Report{p.Report}), now)
	}
	p.Report = usage.TerminalReport(p.Report)
	fmt.Printf("%s %s, %s\n%s\n", p.Report.Provider, p.Report.Account, render.PickSummary(p, now), p.Report.Dir)
	return nil
}

func bestProvider(args []string) (string, error) {
	switch {
	case len(args) == 0:
		return "", nil
	case len(args) == 1 && slices.Contains(config.Providers, args[0]):
		return args[0], nil
	default:
		return "", fmt.Errorf("best takes one of %v", config.Providers)
	}
}
