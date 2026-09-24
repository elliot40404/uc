package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/elliot40404/uc/internal/usage"
)

type Bar struct {
	Used float64
	OK   bool
}

var (
	sizes  = []int{16, 24, 32, 48, 64}
	green  = color.NRGBA{0x3f, 0xb9, 0x50, 0xff}
	yellow = color.NRGBA{0xd2, 0x99, 0x22, 0xff}
	red    = color.NRGBA{0xf8, 0x51, 0x49, 0xff}
	track  = color.NRGBA{0x80, 0x80, 0x80, 0x90}
	failed = color.NRGBA{0xf8, 0x51, 0x49, 0x70}
)

const (
	samples = 4
	step    = 5
)

func Bars(reports []usage.Report, picks map[string]usage.Pick) []Bar {
	var out []Bar
	for _, p := range Providers(reports) {
		if pick, ok := picks[p]; ok {
			out = append(out, Bar{Used: math.Round((100-pick.Report.Left())/step) * step, OK: true})
		} else {
			out = append(out, Bar{})
		}
	}
	return out
}

func Icon(bars []Bar) []byte {
	if len(bars) == 0 {
		bars = []Bar{{OK: true}}
	}
	images := make([][]byte, len(sizes))
	for i, s := range sizes {
		var buf bytes.Buffer
		_ = png.Encode(&buf, draw(bars, s))
		images[i] = buf.Bytes()
	}
	return ico(images)
}

func draw(bars []Bar, size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	n := float64(len(bars))
	s := float64(size)
	margin, gap := s/16, s/8
	h := (s - 2*margin - (n-1)*gap) / n
	for i, b := range bars {
		y := margin + float64(i)*(h+gap)
		r := rect{margin, y, s - margin, y + h}
		bg := track
		if !b.OK {
			bg = failed
		}
		fill(img, r, r.x1, bg)
		if b.OK && b.Used > 0 {
			fill(img, r, r.x0+max(r.w()*min(b.Used, 100)/100, margin), level(b.Used))
		}
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

type rect struct{ x0, y0, x1, y1 float64 }

func (r rect) w() float64 { return r.x1 - r.x0 }
func (r rect) h() float64 { return r.y1 - r.y0 }

func (r rect) has(x, y float64) bool {
	if x < r.x0 || x >= r.x1 || y < r.y0 || y >= r.y1 {
		return false
	}
	rad := min(r.h(), r.w()) / 2
	cx := min(max(x, r.x0+rad), r.x1-rad)
	cy := min(max(y, r.y0+rad), r.y1-rad)
	return (x-cx)*(x-cx)+(y-cy)*(y-cy) <= rad*rad
}

func fill(img *image.NRGBA, shape rect, until float64, c color.NRGBA) {
	b := img.Bounds()
	for py := b.Min.Y; py < b.Max.Y; py++ {
		for px := b.Min.X; px < b.Max.X; px++ {
			hits := 0
			for sy := range samples {
				for sx := range samples {
					x := float64(px) + (float64(sx)+0.5)/samples
					y := float64(py) + (float64(sy)+0.5)/samples
					if x < until && shape.has(x, y) {
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
