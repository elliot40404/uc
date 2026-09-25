//go:build windows

package main

import (
	"context"
	"errors"
	"time"

	"fyne.io/systray"

	"github.com/elliot40404/uc/internal/app"
	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/flyout"
	"github.com/elliot40404/uc/internal/tray"
	"github.com/elliot40404/uc/internal/usage"
)

const minEvery = time.Minute

type trayApp struct {
	refresh chan struct{}
	popup   *popup
	last    flyout.View
}

func newTray() *trayApp {
	t := &trayApp{refresh: make(chan struct{}, 1)}
	t.popup = &popup{onRefresh: t.requestRefresh}
	return t
}

func (t *trayApp) start() {
	ready := make(chan struct{})
	go t.popup.run(ready)
	<-ready
	systray.SetIcon(tray.Loading())
	systray.SetTooltip("uc\nloading")
	systray.SetOnTapped(t.popup.toggle)
	buildMenu(t.requestRefresh)
	t.popup.setView(flyout.View{Status: "Loading"})
	go t.loop()
}

func (t *trayApp) loop() {
	for {
		timer := time.NewTimer(t.update())
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
	every := max(env.Cfg.Defaults.Refresh, minEvery)
	t.last.Status = "Refreshing…"
	t.popup.setView(t.last)
	reports, now, err := env.Reports(context.Background())
	if err != nil && !errors.Is(err, app.ErrCacheSave) {
		t.fail("Error: " + err.Error())
		return every
	}
	t.show(reports, now)
	return every
}

func (t *trayApp) show(reports []usage.Report, now time.Time) {
	systray.SetIcon(tray.Icon(tray.GaugeOf(reports, now)))
	systray.SetTooltip(tray.Tooltip(reports, tray.Picks(reports, now)))
	t.last = flyout.Build(reports, now)
	t.last.Status = "Updated " + now.Format("15:04")
	t.popup.setView(t.last)
}

func (t *trayApp) fail(msg string) {
	systray.SetIcon(tray.Icon(tray.Gauge{}))
	systray.SetTooltip(tray.Tip(msg))
	t.last = flyout.View{Status: "Failed " + time.Now().Format("15:04"), Error: usage.TerminalText(msg)}
	t.popup.setView(t.last)
}
