package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/elliot40404/uc/internal/app"
	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

type env struct {
	app.Env
}

func newEnv() (env, error) {
	e, err := app.New()
	return env{e}, err
}

func (e env) saveConfig(c config.Config) error {
	return config.Save(e.ConfigPath, c, e.Home)
}

func (e env) reportsWith(ctx context.Context, c config.Config) ([]usage.Report, time.Time, error) {
	e.Cfg = c
	return e.Reports(ctx)
}

func showCacheWarning(err error, at time.Time) error {
	if errors.Is(err, app.ErrCacheSave) && !at.IsZero() {
		fmt.Fprintln(os.Stderr, "uc:", app.ErrCacheSave)
		return nil
	}
	return err
}
