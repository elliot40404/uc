package tui

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

func TestProgramRunsAndQuits(t *testing.T) {
	loaded := make(chan struct{})
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) {
		defer close(loaded)
		return sample(), now, nil
	}
	in, keys := io.Pipe()
	var out bytes.Buffer
	p := tea.NewProgram(New(fetch, Options{Every: 5 * time.Minute, Emails: true}),
		tea.WithInput(in), tea.WithOutput(&out), tea.WithWindowSize(120, 40))
	go func() {
		<-loaded
		time.Sleep(100 * time.Millisecond)
		keys.Write([]byte("jcq"))
	}()
	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		p.Kill()
		t.Fatal("program did not quit on q")
	}
	screen := plain(out.String())
	for _, want := range []string{"usage limits", "personal", "codex"} {
		if !strings.Contains(screen, want) {
			t.Errorf("screen missing %q", want)
		}
	}
}
