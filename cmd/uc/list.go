package main

import (
	"context"
	"os"

	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/term"
)

func (e env) list(opts options) error {
	reports, now, err := e.Reports(context.Background())
	if err := showCacheWarning(err, now); err != nil {
		return err
	}
	reports = opts.visible(reports)
	if opts.json {
		return render.JSON(os.Stdout, reports, now)
	}
	return opts.view()(os.Stdout, reports, now, term.ColorOK(os.Stdout))
}
