package tray

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"testing"

	"github.com/elliot40404/uc/internal/usage"
)

func TestGaugeFollowsBestAccount(t *testing.T) {
	reports := []usage.Report{report("claude", "a", 10, 71), report("codex", "b", 0, 100)}
	if got := GaugeOf(reports, now); got != (Gauge{Used: 70, OK: true}) {
		t.Errorf("got %+v", got)
	}
	if got := GaugeOf(reports[1:], now); got.OK {
		t.Errorf("full account should not be usable: %+v", got)
	}
}

func TestIconIsValidICO(t *testing.T) {
	data := Icon(Gauge{Used: 95, OK: true})
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

func TestRingFillsClockwiseFromTop(t *testing.T) {
	img := drawIcon(Gauge{Used: 30, OK: true}, 64)
	if c := img.NRGBAAt(58, 32); c != green {
		t.Errorf("right side should be filled, got %v", c)
	}
	if c := img.NRGBAAt(5, 32); c == green || c.A == 0 {
		t.Errorf("left side should be track, got %v", c)
	}
	if c := img.NRGBAAt(32, 32); c.A != 0 {
		t.Errorf("center should be clear, got %v", c)
	}
}
