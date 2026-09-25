package flyout

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

const kappa = 0.5523

type canvas struct {
	img   *image.RGBA
	scale float64
	fonts *Fonts
	ras   vector.Rasterizer
}

type box struct{ x, y, w, h float64 }

func (b box) px() image.Rectangle {
	return image.Rect(int(math.Floor(b.x)), int(math.Floor(b.y)), int(math.Ceil(b.x+b.w)), int(math.Ceil(b.y+b.h)))
}

func (b box) has(x, y float64) bool {
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

func (c *canvas) s(v float64) float64 { return v * c.scale }

func (c *canvas) round(b box, radius float64, col color.RGBA) {
	b = box{c.s(b.x), c.s(b.y), c.s(b.w), c.s(b.h)}
	r := min(c.s(radius), b.w/2, b.h/2)
	area := b.px()
	ox, oy := float64(area.Min.X), float64(area.Min.Y)
	c.ras.Reset(area.Dx(), area.Dy())
	x0, y0 := float32(b.x-ox), float32(b.y-oy)
	x1, y1 := float32(b.x+b.w-ox), float32(b.y+b.h-oy)
	rr, k := float32(r), float32(r*kappa)
	c.ras.MoveTo(x0+rr, y0)
	c.ras.LineTo(x1-rr, y0)
	c.ras.CubeTo(x1-rr+k, y0, x1, y0+rr-k, x1, y0+rr)
	c.ras.LineTo(x1, y1-rr)
	c.ras.CubeTo(x1, y1-rr+k, x1-rr+k, y1, x1-rr, y1)
	c.ras.LineTo(x0+rr, y1)
	c.ras.CubeTo(x0+rr-k, y1, x0, y1-rr+k, x0, y1-rr)
	c.ras.LineTo(x0, y0+rr)
	c.ras.CubeTo(x0, y0+rr-k, x0+rr-k, y0, x0+rr, y0)
	c.ras.ClosePath()
	c.ras.Draw(c.img, area, image.NewUniform(col), image.Point{})
}

func (c *canvas) text(s string, x, baseline, size float64, bold bool, col color.RGBA) {
	d := font.Drawer{Dst: c.img, Src: image.NewUniform(col), Face: c.fonts.face(bold, c.s(size))}
	d.Dot = fixed.P(int(math.Round(c.s(x))), int(math.Round(c.s(baseline))))
	d.DrawString(s)
}

func (c *canvas) right(s string, x, baseline, size float64, bold bool, col color.RGBA) {
	c.text(s, x-c.width(s, size, bold), baseline, size, bold, col)
}

func (c *canvas) center(s string, b box, size float64, bold bool, col color.RGBA) {
	c.text(s, b.x+(b.w-c.width(s, size, bold))/2, b.y+b.h/2+size*0.36, size, bold, col)
}

func (c *canvas) width(s string, size float64, bold bool) float64 {
	return float64(font.MeasureString(c.fonts.face(bold, c.s(size)), s)) / 64 / c.scale
}

func (c *canvas) fit(s string, maxW, size float64, bold bool) string {
	if c.width(s, size, bold) <= maxW {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && c.width(string(r)+"…", size, bold) > maxW {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
