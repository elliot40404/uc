package tui

import (
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

const (
	rowLive = iota
	rowCompact
	rowEmails
	rowMini
	rowAltScreen
	rowEvery
	rowCount
)

var everyChoices = []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute, 10 * time.Minute, 15 * time.Minute, 30 * time.Minute}

func (m Model) onSettingsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.interrupt):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.back):
		m.settings = false
	case key.Matches(msg, m.keys.up):
		m.setRow = wrap(m.setRow-1, rowCount)
	case key.Matches(msg, m.keys.down):
		m.setRow = wrap(m.setRow+1, rowCount)
	case key.Matches(msg, m.keys.toggle):
		m = m.toggle()
	case key.Matches(msg, m.keys.left):
		m = m.setEvery(-1)
	case key.Matches(msg, m.keys.right):
		m = m.setEvery(1)
	}
	return m, nil
}

func (m Model) toggle() Model {
	d := &m.cfg.Defaults
	switch m.setRow {
	case rowLive:
		m.live = !m.live
		d.Live = m.live
	case rowCompact:
		m.compact = !m.compact
		d.Compact = m.compact
	case rowEmails:
		m.emails = !m.emails
		d.ShowEmails = m.emails
	case rowMini:
		d.Mini = !d.Mini
	case rowAltScreen:
		m.altScreen = !m.altScreen
		d.AltScreen = m.altScreen
	case rowEvery:
		return m.setEvery(1)
	}
	return m.persist()
}

func (m Model) setEvery(dir int) Model {
	if m.setRow != rowEvery {
		return m
	}
	next := stepEvery(m.every, dir)
	if next == m.every {
		return m
	}
	m.every = next
	m.cfg.Defaults.Refresh = next
	return m.persist()
}

func stepEvery(cur time.Duration, dir int) time.Duration {
	if dir > 0 {
		i := slices.IndexFunc(everyChoices, func(d time.Duration) bool { return d > cur })
		if i < 0 {
			return cur
		}
		return everyChoices[i]
	}
	for _, d := range slices.Backward(everyChoices) {
		if d < cur {
			return d
		}
	}
	return cur
}

func (m Model) persist() Model {
	if m.save != nil {
		m.saveErr = m.save(m.cfg)
	}
	return m
}
