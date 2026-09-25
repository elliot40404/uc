# uc

`uc` shows usage limits for multiple Claude and Codex logins in one terminal. It finds local accounts, reads their existing logins, and fetches usage from each provider. Windows is the main target; the code also builds on Linux and macOS.

```text
CLAUDE
  default   team  ok
    5h  █░░░░░░░░░░░░░░░░░░░    5%  resets 06:30 (in 4h 9m)
    7d  ███████░░░░░░░░░░░░░   33%  resets Mon 06:30 (in 3d 4h)

CODEX
  default   plus  ok
    5h  ░░░░░░░░░░░░░░░░░░░░    0%  resets 07:21 (in 5h)
    7d  ██░░░░░░░░░░░░░░░░░░    9%  resets Sun 15:02 (in 2d 12h)
```

This is an example of `uc --print`. Percentages show quota **used**, not quota left. Plain `uc` opens a live dashboard with account cards when run in a terminal; when piped, it prints text once.

## Install

Requires Go 1.27.1 or newer:

```sh
go install github.com/elliot40404/uc/cmd/uc@latest
```

Or clone the repository and run `just install` to install from the local source. Make sure your Go bin directory is on `PATH`. Run `just` to list build, test, lint, and other local tasks.

On macOS, Claude Code may store its login in Keychain. `uc` currently reads Claude credentials from files only, so those accounts can appear as `not logged in`. Codex reads its local `auth.json`.

## Commands and views

```text
uc                           live full dashboard in a terminal
uc --full --compact          full dashboard with one row per account
uc --mini                    live compact dashboard in a terminal
uc --mini --live=false       print the compact dashboard once
uc --mini --alt-screen=false keep the live mini view in the normal screen
uc --print                   print grouped text once
uc --print --compact         print a one-row-per-account table once
uc --json                    print JSON once
uc best                      suggest an account across both providers
uc best claude               suggest a Claude account
uc best codex --json         suggest a Codex account as JSON
uc init                      write a starter config if none exists
uc help                      show all options and the config path
uc --version                 print the build revision, or uc dev without Git metadata
```

Use `--every 2m` to change the live refresh interval (minimum 1 minute). Use `--live=false` to stop automatic refresh in a full dashboard; `r` still refreshes. Live dashboards use the alternate screen by default; `--alt-screen=false` leaves them in the normal screen. `--full` forces the full dashboard if mini is configured as the default. Flags can go before or after `best` and its provider.

The dashboard supports `↑`/`↓` or `j`/`k` to select, `r` to refresh, `c` to switch between cards and compact rows, `e` to show or hide emails, `?` for help, and `q` or `Esc` to quit. When output is piped, `--mini` prints once even with live mode enabled. `--print` and `--json` always exit after one fetch.

`--json` includes `generated_at` and an `accounts` array. Each account contains its provider, name, local directory, status and available usage windows. `left_pct` is the percentage left in the most-used window; it is `null` when usage is unavailable. Emails are omitted unless `--emails` is set or `show_emails` is enabled in config. JSON and `uc best` can reveal local account paths, so check output before sharing it.

## Tray icon (Windows, experimental)

`uctray` puts a ring in the Windows notification area. The ring fills with the most-used window of the suggested account across all providers: green under 60%, yellow under 90%, red above. A gray ring with a red dot means no account is usable.

```sh
go install -ldflags="-H=windowsgui" github.com/elliot40404/uc/cmd/uctray@latest
```

Or run `just install-tray`. Without `-H=windowsgui` a console window stays open next to the icon.

Hover shows the suggested account per provider. Left click opens a popup with a card per account, its usage bars and reset times,, with a refresh icon in the header. The popup follows the Windows light or dark taskbar theme and closes on Esc or a click elsewhere. Right click has `Refresh now`, `Start with Windows` and `Quit`. The tray reads the same config and cache as `uc`, refreshes on `defaults.every` (minimum 1 minute), rereads config on every refresh and never shows emails. Only one copy runs at a time.

## Which account to use first?

The dashboard suggests one account per provider. `uc best` chooses the highest-ranked account across providers, or only the provider you name. The ranking is **unused 7-day quota divided by days until its reset**; a reset less than 5 hours away is treated as 5 hours away. This favors quota that would otherwise expire soon.

| Account | 7d used | Resets in | Score |
| --- | ---: | ---: | ---: |
| a | 30% | 4 days | 17.5 |
| c | 10% | 1.5 days | 60, picked |

Accounts with no usable 7-day data, a full 7-day window, expired logins, or errors are skipped. A full 5-hour window does **not** remove an account from consideration: the suggestion shows when that window becomes usable again. The score is only for ranking; the bars and percentages still show usage.

## Config and accounts

The config file is `~/.config/uc/config.json`, or `$XDG_CONFIG_HOME/uc/config.json` if set. It is optional; run `uc init` to create one with the accounts found on your machine. `uc init` does not overwrite an existing file. Command-line flags override the matching defaults for that run.

Without a config file, `uc` uses the full, live dashboard, the alternate screen, hidden emails, and a 5-minute refresh interval. A config example:

```json
{
  "defaults": {
    "mini": true,
    "live": true,
    "alt_screen": true,
    "compact": false,
    "show_emails": false,
    "every": "5m"
  },
  "accounts": [
    {"provider": "claude", "dir": "~/.claude-backup", "hide": true},
    {"provider": "codex", "dir": "~/.codex-work", "name": "work"},
    {"provider": "codex", "dir": "D:/other/codex", "name": "side"}
  ]
}
```

In this example, `mini: true` makes plain `uc` open the compact view instead of the full one. `live: false` disables automatic refresh; for mini mode it prints once, while the full dashboard stays open for manual refresh. `compact` selects rows instead of cards in the full dashboard. `show_emails: true` shows emails when available; `--emails` enables them for one run. `every` controls live refresh. You can omit any setting to keep its default.

Accounts are found automatically in directories matching `~/.claude*` and `~/.codex*`, plus the paths in `CLAUDE_CONFIG_DIR` and `CODEX_HOME`. The `accounts` list can rename a discovered directory, hide it, or add another directory. Hidden accounts are not fetched. You must already be logged in through the relevant CLI; `uc` does not log in or renew tokens.

## Privacy and refresh behavior

- Login files are read but never written or refreshed. Access tokens are sent only to the corresponding provider's usage API.
- Successful usage results are cached in your user cache directory (typically `%LOCALAPPDATA%/uc/usage.json` on Windows). The cache stores usage, plan, status and refresh times, but not tokens, emails, account names or account paths. Emails shown with cached data are read again from local login files.
- Results less than a minute old are reused. After an HTTP 429, requests to that provider pause for 2 minutes, doubling on further rate limits up to 30 minutes. Recent cached usage (less than 6 hours old) is shown as cached and rate limited; older usage is not shown.
- `expired` means the saved login has expired. Run `claude` or `codex` once with that account to renew it.
