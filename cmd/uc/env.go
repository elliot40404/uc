package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/elliot40404/uc/internal/app"
)

type env struct {
	app.Env
}

func newEnv() (env, error) {
	e, err := app.New()
	return env{e}, err
}

func showCacheWarning(err error, at time.Time) error {
	if errors.Is(err, app.ErrCacheSave) && !at.IsZero() {
		fmt.Fprintln(os.Stderr, "uc:", app.ErrCacheSave)
		return nil
	}
	return err
}
