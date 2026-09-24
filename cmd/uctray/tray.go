//go:build windows

package main

import (
	"context"
	"errors"
	"time"

	"fyne.io/systray"

	"github.com/elliot40404/uc/internal/app"
	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/tray"
	"github.com/elliot40404/uc/internal/usage"
)

const minEvery = time.Minute

type trayApp struct {
	refresh chan struct{}
	menu    *menu
}

func newTray() *trayApp {
	t := &trayApp{refresh: make(chan struct{}, 1)}
	t.menu = &menu{open: t.open, refresh: t.requestRefresh}
	return t
}

func (t *trayApp) start() {
	systray.SetIcon(tray.Icon(nil))
	systray.SetTooltip("uc\nloading")
	t.menu.set("Loading", nil)
	go t.loop()
}

func (t *trayApp) loop() {
	for {
		every := t.update()
		timer := time.NewTimer(every)
		select {
		case <-timer.C:
		case <-t.refresh:
			timer.Stop()
		}
	}
}

func (t *trayApp) requestRefresh() {
	select {
	case t.refresh <- struct{}{}:
	default:
	}
}

func (t *trayApp) update() time.Duration {
	env, err := app.New()
	if err != nil {
		t.fail("Config error: " + err.Error())
		return config.DefaultRefresh
	}
	t.menu.setStatus("Refreshing")
	reports, now, err := env.Reports(context.Background())
	if err != nil && !errors.Is(err, app.ErrCacheSave) {
		t.fail("Error: " + err.Error())
		return max(env.Cfg.Defaults.Refresh, minEvery)
	}
	t.show(reports, now)
	return max(env.Cfg.Defaults.Refresh, minEvery)
}

func (t *trayApp) show(reports []usage.Report, now time.Time) {
	picks := tray.Picks(reports, now)
	lines := make([]string, len(reports))
	for i, r := range reports {
		lines[i] = tray.Line(r, tray.IsPick(r, picks))
	}
	systray.SetIcon(tray.Icon(tray.Bars(reports, picks)))
	systray.SetTooltip(tray.Tooltip(reports, picks))
	status := "Updated " + now.Format("15:04")
	if len(reports) == 0 {
		status = "No accounts found"
	}
	t.menu.set(status, lines)
}

func (t *trayApp) fail(msg string) {
	systray.SetIcon(tray.Icon([]tray.Bar{{}}))
	systray.SetTooltip(tray.Tip(msg))
	t.menu.set(tray.MenuText(msg), nil)
}
