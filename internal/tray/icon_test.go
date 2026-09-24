package tray

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"testing"

	"github.com/elliot40404/uc/internal/usage"
)

func TestBarsFollowPicks(t *testing.T) {
	reports := []usage.Report{report("claude", "a", 10, 70), report("codex", "b", 0, 100)}
	got := Bars(reports, Picks(reports, now))
	if len(got) != 2 || got[0] != (Bar{Used: 70, OK: true}) || got[1].OK {
		t.Errorf("got %+v", got)
	}
}

func TestIconIsValidICO(t *testing.T) {
	data := Icon([]Bar{{Used: 95, OK: true}, {}})
	var head [3]uint16
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &head); err != nil {
		t.Fatal(err)
	}
	if head != [3]uint16{0, 1, uint16(len(sizes))} {
		t.Fatalf("header %v", head)
	}
	for i, s := range sizes {
		e := data[6+16*i:]
		size := binary.LittleEndian.Uint32(e[8:])
		off := binary.LittleEndian.Uint32(e[12:])
		img, err := png.Decode(bytes.NewReader(data[off : off+size]))
		if err != nil || img.Bounds().Dx() != s || int(e[0]) != s {
			t.Fatalf("size %d: %v %v", s, err, img.Bounds())
		}
	}
}

func TestDrawColorsByLevel(t *testing.T) {
	img := draw([]Bar{{Used: 95, OK: true}}, 32)
	if c := img.NRGBAAt(8, 16); c != red {
		t.Errorf("want red fill, got %v", c)
	}
	img = draw([]Bar{{Used: 10, OK: true}}, 32)
	if c := img.NRGBAAt(24, 16); c.A == 0 || c == green {
		t.Errorf("want track past the fill, got %v", c)
	}
	if c := img.NRGBAAt(0, 0); c.A != 0 {
		t.Errorf("corner should be clear, got %v", c)
	}
}
