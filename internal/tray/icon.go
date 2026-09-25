package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

type Gauge struct {
	Used float64
	OK   bool
}

var (
	sizes  = []int{16, 20, 24, 32, 40, 48, 64}
	green  = color.NRGBA{0x3f, 0xb9, 0x50, 0xff}
	yellow = color.NRGBA{0xe3, 0xa0, 0x08, 0xff}
	red    = color.NRGBA{0xf8, 0x51, 0x49, 0xff}
	track  = color.NRGBA{0x9a, 0x9a, 0x9a, 0x70}
)

const (
	samples   = 4
	step      = 5
	thickness = 0.2
)

func GaugeOf(reports []usage.Report, now time.Time) Gauge {
	pick, ok := usage.Best(reports, "", now)
	if !ok {
		return Gauge{}
	}
	return Gauge{Used: math.Round((100-pick.Report.Left())/step) * step, OK: true}
}

func Icon(g Gauge) []byte {
	images := make([][]byte, len(sizes))
	for i, s := range sizes {
		var buf bytes.Buffer
		_ = png.Encode(&buf, drawIcon(g, s))
		images[i] = buf.Bytes()
	}
	return ico(images)
}

func Loading() []byte {
	return Icon(Gauge{OK: true})
}

func drawIcon(g Gauge, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	r := ring{c: float64(size) / 2, width: float64(size) * thickness}
	r.outer = r.c - float64(size)/32
	r.inner = r.outer - r.width
	sweep := min(max(g.Used, 0), 100) / 100 * 2 * math.Pi
	paint(img, func(x, y float64) bool { return r.onRing(x, y) }, track)
	if !g.OK {
		paint(img, func(x, y float64) bool { return math.Hypot(x-r.c, y-r.c) <= r.width }, red)
	}
	if g.OK && g.Used > 0 {
		paint(img, func(x, y float64) bool { return r.onArc(x, y, sweep) }, level(g.Used))
	}
	return img
}

func level(pct float64) color.NRGBA {
	switch {
	case pct >= 90:
		return red
	case pct >= 60:
		return yellow
	default:
		return green
	}
}

type ring struct{ c, outer, inner, width float64 }

func (r ring) onRing(x, y float64) bool {
	d := math.Hypot(x-r.c, y-r.c)
	return d >= r.inner && d <= r.outer
}

func (r ring) onArc(x, y, sweep float64) bool {
	if sweep >= 2*math.Pi {
		return r.onRing(x, y)
	}
	a := math.Atan2(x-r.c, r.c-y)
	if a < 0 {
		a += 2 * math.Pi
	}
	return r.onRing(x, y) && a <= sweep || r.onCap(x, y, 0) || r.onCap(x, y, sweep)
}

func (r ring) onCap(x, y, angle float64) bool {
	mid := (r.outer + r.inner) / 2
	cx := r.c + mid*math.Sin(angle)
	cy := r.c - mid*math.Cos(angle)
	return math.Hypot(x-cx, y-cy) <= r.width/2
}

func paint(img *image.NRGBA, inside func(x, y float64) bool, c color.NRGBA) {
	b := img.Bounds()
	for py := b.Min.Y; py < b.Max.Y; py++ {
		for px := b.Min.X; px < b.Max.X; px++ {
			hits := 0
			for sy := range samples {
				for sx := range samples {
					if inside(float64(px)+(float64(sx)+0.5)/samples, float64(py)+(float64(sy)+0.5)/samples) {
						hits++
					}
				}
			}
			if hits > 0 {
				blend(img, px, py, c, float64(hits)/(samples*samples))
			}
		}
	}
}

func blend(img *image.NRGBA, x, y int, c color.NRGBA, cover float64) {
	dst := img.NRGBAAt(x, y)
	sa := float64(c.A) / 255 * cover
	da := float64(dst.A) / 255
	oa := sa + da*(1-sa)
	mix := func(s, d uint8) uint8 {
		return uint8((float64(s)*sa + float64(d)*da*(1-sa)) / oa)
	}
	img.SetNRGBA(x, y, color.NRGBA{mix(c.R, dst.R), mix(c.G, dst.G), mix(c.B, dst.B), uint8(oa * 255)})
}

func ico(images [][]byte) []byte {
	var buf bytes.Buffer
	le := binary.LittleEndian
	_ = binary.Write(&buf, le, [3]uint16{0, 1, uint16(len(images))})
	offset := 6 + 16*len(images)
	for i, data := range images {
		s := uint8(sizes[i] % 256)
		_ = binary.Write(&buf, le, [4]uint8{s, s, 0, 0})
		_ = binary.Write(&buf, le, [2]uint16{1, 32})
		_ = binary.Write(&buf, le, [2]uint32{uint32(len(data)), uint32(offset)})
		offset += len(data)
	}
	for _, data := range images {
		buf.Write(data)
	}
	return buf.Bytes()
}
