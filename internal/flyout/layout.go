package flyout

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
)

const refreshGlyph rune = 0xE72C

const (
	Width    = 360
	pad      = 12
	inner    = 12
	headerH  = 48
	labelH   = 28
	titleH   = 24
	rowH     = 22
	noteH    = 20
	cardGap  = 6
	emptyH   = 72
	bottomH  = 6
	iconBox  = 28
	RefrHit  = "refresh"
	barLeft  = 54
	barRight = 218
)

type Hit struct {
	Rect   image.Rectangle
	Action string
}

func Height(v View) float64 {
	h := float64(headerH + bottomH)
	if len(v.Sections) == 0 {
		return h + emptyH
	}
	for _, s := range v.Sections {
		h += labelH
		for _, a := range s.Accounts {
			h += cardHeight(a) + cardGap
		}
	}
	return h
}

func cardHeight(a Account) float64 {
	h := float64(inner + titleH + len(a.Bars)*rowH + inner - 4)
	if a.Note != "" {
		h += noteH
	}
	return h
}

func Render(v View, t Theme, f *Fonts, scale float64, hover string) (*image.RGBA, []Hit) {
	h := Height(v)
	img := image.NewRGBA(image.Rect(0, 0, int(math.Ceil(Width*scale)), int(math.Ceil(h*scale))))
	draw.Draw(img, img.Bounds(), image.NewUniform(t.Bg), image.Point{}, draw.Src)
	c := &canvas{img: img, scale: scale, fonts: f}
	c.text("Usage limits", 16, 30, 14, true, t.Text)
	hits := c.refreshButton(t, hover)
	c.right(c.fit(v.Status, 200, 12, false), Width-pad-iconBox-6, 30, 12, false, t.Dim)
	y := float64(headerH)
	if len(v.Sections) == 0 {
		c.empty(v, t, y)
		y += emptyH
	}
	for _, s := range v.Sections {
		y = c.section(s, t, y)
	}
	return img, hits
}

func (c *canvas) empty(v View, t Theme, y float64) {
	msg, col := "No accounts found", t.Dim
	if v.Error != "" {
		msg, col = v.Error, t.Yellow
	}
	c.center(c.fit(msg, Width-2*pad-2*inner, 13, false), box{0, y, Width, emptyH}, 13, false, col)
}

func (c *canvas) section(s Section, t Theme, y float64) float64 {
	c.round(box{17, y + 12, 7, 7}, 4, t.provider(s.Provider))
	c.text(strings.ToUpper(s.Provider), 30, y+19, 11, true, t.Dim)
	y += labelH
	for _, a := range s.Accounts {
		c.card(a, t, y)
		y += cardHeight(a) + cardGap
	}
	return y
}

func (c *canvas) card(a Account, t Theme, y float64) {
	right := float64(Width - pad - inner)
	c.round(box{pad, y, Width - 2*pad, cardHeight(a)}, 8, t.Card)
	top := y + inner
	planW := c.width(a.Plan, 12, false)
	pillW := 0.0
	if a.Pick {
		pillW = c.width("Use next", 11, true) + 22
	}
	name := c.fit(a.Name, right-pad-inner-planW-pillW-12, 14, true)
	c.text(name, pad+inner, top+15, 14, true, t.Text)
	if a.Pick {
		x := pad + inner + c.width(name, 14, true) + 8
		c.round(box{x, top + 1, pillW - 8, 19}, 10, mix(t.Accent, t.Card, 0.2))
		c.text("Use next", x+7, top+15, 11, true, t.Accent)
	}
	c.right(a.Plan, right, top+15, 12, false, t.Dim)
	row := top + titleH
	for _, b := range a.Bars {
		c.bar(b, t, row)
		row += rowH
	}
	if a.Note != "" {
		col := t.Faint
		if a.Problem {
			col = t.Yellow
		}
		c.text(c.fit(a.Note, right-pad-inner, 12, false), pad+inner, row+14, 12, false, col)
	}
}

func (c *canvas) bar(b Bar, t Theme, y float64) {
	c.text(b.Name, pad+inner, y+15, 12, false, t.Dim)
	width := float64(barRight - barLeft)
	c.round(box{barLeft, y + 8, width, 6}, 3, t.Track)
	if b.Used > 0 {
		c.round(box{barLeft, y + 8, max(width*min(b.Used, 100)/100, 6), 6}, 3, t.level(b.Used))
	}
	c.right(fmt.Sprintf("%.0f%%", b.Used), barRight+44, y+15, 12, true, t.Text)
	c.right(b.Resets, Width-pad-inner, y+15, 12, false, t.Faint)
}

func (c *canvas) refreshButton(t Theme, hover string) []Hit {
	b := box{Width - pad - iconBox, (headerH - iconBox) / 2, iconBox, iconBox}
	if hover == RefrHit {
		c.round(b, 6, t.Hover)
	}
	if c.fonts.icons != nil {
		c.glyph(string(refreshGlyph), b, 14, t.Dim)
	} else {
		c.center("↻", b, 16, false, t.Dim)
	}
	return []Hit{{c.pixels(b), RefrHit}}
}

func (c *canvas) pixels(b box) image.Rectangle {
	return box{c.s(b.x), c.s(b.y), c.s(b.w), c.s(b.h)}.px()
}

func mix(a, b color.RGBA, amount float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*amount + float64(y)*(1-amount)) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 0xff}
}
