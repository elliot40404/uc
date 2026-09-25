package tui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/usage"
)

type Fetch func(ctx context.Context, cfg config.Config) ([]usage.Report, time.Time, error)

type Options struct {
	Every     time.Duration
	Compact   bool
	Emails    bool
	Live      bool
	AltScreen bool
	Config    config.Config
	Save      func(config.Config) error
	Found     []discover.Account
}

type reportsMsg struct {
	reports []usage.Report
	at      time.Time
	err     error
}

type tickMsg time.Time

type Model struct {
	fetch     Fetch
	every     time.Duration
	theme     theme
	keys      keyMap
	help      help.Model
	spin      spinner.Model
	reports   []usage.Report
	fetchedAt time.Time
	now       time.Time
	err       error
	warning   bool
	loading   bool
	selected  int
	compact   bool
	emails    bool
	inline    bool
	live      bool
	altScreen bool
	quitting  bool
	settings  bool
	setRow    int
	cfg       config.Config
	save      func(config.Config) error
	saveErr   error
	found     []discover.Account
	width     int
	height    int
}

func New(fetch Fetch, o Options) Model {
	m := Model{fetch: fetch, every: o.Every, compact: o.Compact, emails: o.Emails, live: o.Live, altScreen: o.AltScreen, cfg: o.Config, save: o.Save, found: o.Found, keys: newKeys(), help: help.New(), now: time.Now(), loading: true, width: 80, height: 24}
	m.spin = spinner.New(spinner.WithSpinner(spinner.MiniDot))
	return m.withTheme(true)
}

func (m Model) withTheme(dark bool) Model {
	m.theme = newTheme(dark)
	m.spin.Style = m.theme.fg(m.theme.accent)
	m.help.Styles = help.DefaultStyles(dark)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.load(), m.spin.Tick, tick(), tea.RequestBackgroundColor)
}

func (m Model) load() tea.Cmd {
	fetch, cfg := m.fetch, m.cfg
	return func() tea.Msg {
		reports, at, err := fetch(context.Background(), cfg)
		return reportsMsg{reports, at, err}
	}
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(msg.Width - 2*marginX)
	case tea.BackgroundColorMsg:
		m = m.withTheme(msg.IsDark())
	case tea.KeyPressMsg:
		return m.onKey(msg)
	case reportsMsg:
		return m.onReports(msg), nil
	case tickMsg:
		return m.onTick(time.Time(msg))
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) onKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.settings {
		return m.onSettingsKey(msg)
	}
	switch {
	case key.Matches(msg, m.keys.quit):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.up):
		m.selected = wrap(m.selected-1, len(m.reports))
	case key.Matches(msg, m.keys.down):
		m.selected = wrap(m.selected+1, len(m.reports))
	case key.Matches(msg, m.keys.compact):
		m.compact = !m.compact
	case key.Matches(msg, m.keys.emails):
		m.emails = !m.emails
	case key.Matches(msg, m.keys.settings):
		m.settings = true
	case key.Matches(msg, m.keys.help):
		m.help.ShowAll = !m.help.ShowAll
	case key.Matches(msg, m.keys.refresh):
		return m.startLoad()
	}
	return m, nil
}

func (m Model) onReports(msg reportsMsg) Model {
	m.loading = false
	m.warning = msg.err != nil && !msg.at.IsZero()
	if m.warning {
		m.err = nil
	} else {
		m.err = msg.err
	}
	if msg.err == nil || m.warning {
		m.reports, m.fetchedAt = usage.TerminalReports(msg.reports), msg.at
		m.selected = min(m.selected, max(len(m.reports)-1, 0))
	}
	return m
}

func (m Model) onTick(t time.Time) (tea.Model, tea.Cmd) {
	m.now = t
	if m.live && !m.loading && !m.fetchedAt.IsZero() && t.Sub(m.fetchedAt) >= m.every {
		m, cmd := m.startLoad()
		return m, tea.Batch(cmd, tick())
	}
	return m, tick()
}

func (m Model) startLoad() (Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.load(), m.spin.Tick)
}

func wrap(i, n int) int {
	if n == 0 {
		return 0
	}
	return (i%n + n) % n
}
