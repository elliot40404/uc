package flyout

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

func sample() []usage.Report {
	win := func(name string, used float64, in time.Duration) usage.Window {
		return usage.Window{Name: name, UsedPct: used, ResetsAt: now.Add(in)}
	}
	return []usage.Report{
		{Provider: "claude", Account: "personal", Dir: "a", Plan: "pro", Status: usage.StatusOK,
			Windows: []usage.Window{win("5h", 53, 3*time.Hour+55*time.Minute), win("7d", 7, 76*time.Hour)}},
		{Provider: "claude", Account: "work", Dir: "b", Plan: "team", Status: usage.StatusOK, Stale: true, UpdatedAt: now,
			Windows: []usage.Window{win("5h", 96, 40*time.Minute), win("7d", 71, 30*time.Hour)}},
		{Provider: "codex", Account: "default", Dir: "c", Plan: "plus", Status: usage.StatusOK,
			Windows: []usage.Window{win("5h", 0, 5*time.Hour), win("7d", 9, 60*time.Hour)}},
		{Provider: "codex", Account: "old", Dir: "d", Status: usage.StatusExpired},
	}
}

func TestBuildGroupsByProvider(t *testing.T) {
	v := Build(sample(), now)
	if len(v.Sections) != 2 || len(v.Sections[0].Accounts) != 2 || len(v.Sections[1].Accounts) != 2 {
		t.Fatalf("got %+v", v)
	}
	a := v.Sections[0].Accounts[0]
	if !a.Pick || a.Bars[0].Resets != "in 3h 55m" || a.Note != "" {
		t.Errorf("got %+v", a)
	}
	if old := v.Sections[1].Accounts[1]; !old.Problem || len(old.Bars) != 0 {
		t.Errorf("got %+v", old)
	}
}

func TestRenderSizeAndHits(t *testing.T) {
	v := Build(sample(), now)
	img, hits := Render(v, Dark(), LoadFonts(), 1.5, "")
	if img.Bounds().Dx() != Width*3/2 || float64(img.Bounds().Dy()) < Height(v)*1.5-1 {
		t.Errorf("size %v", img.Bounds())
	}
	if len(hits) != 1 || hits[0].Action != RefrHit || !hits[0].Rect.In(img.Bounds()) {
		t.Errorf("hits %+v", hits)
	}
}

func TestPreview(t *testing.T) {
	dir := os.Getenv("FLYOUT_PREVIEW")
	if dir == "" {
		t.Skip("set FLYOUT_PREVIEW to a folder to write preview images")
	}
	fonts := LoadFonts()
	for name, theme := range map[string]Theme{"dark": Dark(), "light": Light()} {
		v := Build(sample(), now)
		v.Status = "Updated 04:45"
		img, _ := Render(v, theme, fonts, 1.5, "")
		f, err := os.Create(filepath.Join(dir, "flyout_"+name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		_ = png.Encode(f, img)
		f.Close()
	}
}
