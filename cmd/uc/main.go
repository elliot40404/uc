package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/tui"
	"github.com/elliot40404/uc/internal/usage"
)

const usageText = `uc shows usage limits for your Claude and Codex accounts.

Usage:
  uc [--compact] [--every 5m]     open the live dashboard
  uc --mini                       open the compact live dashboard in a terminal
  uc --mini --live=false          print the compact dashboard once
  uc --full                       open the full dashboard even if mini is the default
  --live=false                    disable automatic refresh (r still refreshes)
  --alt-screen=false              draw in the normal terminal screen
  uc --print [--compact]          print usage once as plain text
  --emails                        show emails, hidden by default
  uc --json                       print usage once as JSON
  uc best [claude|codex] [--json] print the account with the most room left
  uc init                         write a starter config with every found account
  uc --version                    print the build revision
  uc help                         show this help

Config: %s
`

type options struct {
	json      bool
	compact   bool
	print     bool
	mini      bool
	live      bool
	altScreen bool
	full      bool
	emails    bool
	every     time.Duration
}

type view func(io.Writer, []usage.Report, time.Time, bool) error

func (o options) view() view {
	if o.compact {
		return render.Compact
	}
	return render.Grouped
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "uc:", usage.TerminalText(err.Error()))
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-version") {
		fmt.Println(version())
		return nil
	}
	env, err := newEnv()
	if err != nil {
		return err
	}
	opts, pos, err := parse(args, env.configPath, env.cfg.Defaults)
	if err != nil {
		return err
	}
	cmd, rest := "", []string(nil)
	if len(pos) > 0 {
		cmd, rest = pos[0], pos[1:]
	}
	switch cmd {
	case "":
		if interactive(opts) && (!opts.mini || opts.live) {
			return env.tui(opts)
		}
		if opts.mini && !opts.json && !opts.print {
			return env.mini(opts)
		}
		return env.list(opts)
	case "best":
		return env.best(opts, rest)
	case "init":
		return env.init()
	case "help":
		fmt.Printf(usageText, env.configPath)
		return nil
	default:
		return fmt.Errorf("unknown command %q, see uc help", cmd)
	}
}

func parse(args []string, configPath string, d config.Defaults) (options, []string, error) {
	var opts options
	fs := flag.NewFlagSet("uc", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprintf(fs.Output(), usageText, configPath) }
	fs.BoolVar(&opts.json, "json", false, "")
	fs.BoolVar(&opts.compact, "compact", d.Compact, "")
	fs.BoolVar(&opts.print, "print", false, "")
	fs.BoolVar(&opts.mini, "mini", d.Mini, "")
	fs.BoolVar(&opts.live, "live", d.Live, "")
	fs.BoolVar(&opts.altScreen, "alt-screen", d.AltScreen, "")
	fs.BoolVar(&opts.full, "full", false, "")
	fs.BoolVar(&opts.emails, "emails", d.ShowEmails, "")
	fs.DurationVar(&opts.every, "every", d.Refresh, "")
	var pos []string
	for {
		if err := fs.Parse(args); err != nil {
			return opts, nil, err
		}
		if fs.NArg() == 0 {
			return opts.withFull(), pos, nil
		}
		pos = append(pos, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

func (o options) tui() tui.Options {
	return tui.Options{Every: o.every, Compact: o.compact, Emails: o.emails, Live: o.live, AltScreen: o.altScreen}
}

func (o options) visible(reports []usage.Report) []usage.Report {
	if o.emails {
		return reports
	}
	return usage.WithoutEmails(reports)
}

func (o options) withFull() options {
	if o.full {
		o.mini = false
	}
	return o
}
