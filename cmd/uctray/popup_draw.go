//go:build windows

package main

import (
	"image"
	"unsafe"

	"golang.org/x/sys/windows/registry"

	"github.com/elliot40404/uc/internal/flyout"
)

const personalizeKey = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`

func (p *popup) render() {
	p.mu.Lock()
	v := p.view
	p.mu.Unlock()
	theme, dark := flyout.Dark(), int32(1)
	if lightTheme() {
		theme, dark = flyout.Light(), 0
	}
	call("DwmSetWindowAttribute", uintptr(p.hwnd), dwmDarkMode, ptr(&dark), 4)
	p.img, p.hits = flyout.Render(v, theme, p.fonts, p.scale, p.hover)
	p.bgra = toBGRA(p.img, p.bgra)
}

func (p *popup) redraw() {
	old := p.img.Bounds()

	p.render()
	if p.img.Bounds() != old {
		p.shown = false
		p.show()
		return
	}
	call("InvalidateRect", uintptr(p.hwnd), 0, 0)
}

func (p *popup) paint() {
	var ps paintStruct
	hdc := call("BeginPaint", uintptr(p.hwnd), ptr(&ps))
	if p.img != nil {
		w, h := p.img.Bounds().Dx(), p.img.Bounds().Dy()
		bi := bitmapHeader{Width: int32(w), Height: -int32(h), Planes: 1, BitCount: 32}
		bi.Size = uint32(unsafe.Sizeof(bi))
		call("SetDIBitsToDevice", hdc, 0, 0, uintptr(w), uintptr(h), 0, 0, 0, uintptr(h), ptr(&p.bgra[0]), ptr(&bi), 0)
	}
	call("EndPaint", uintptr(p.hwnd), ptr(&ps))
}

func (p *popup) onMove(x, y int) {
	if !p.tracking {
		tm := trackMouse{Flags: tmeLeave, Hwnd: p.hwnd}
		tm.Size = uint32(unsafe.Sizeof(tm))
		call("TrackMouseEvent", ptr(&tm))
		p.tracking = true
	}
	p.setHover(p.hitAt(x, y))
}

func (p *popup) setHover(action string) {
	if action == p.hover {
		return
	}
	p.hover = action
	p.render()
	call("InvalidateRect", uintptr(p.hwnd), 0, 0)
}

func (p *popup) onClick(x, y int) {
	if p.hitAt(x, y) == flyout.RefrHit {
		go p.onRefresh()
	}
}

func (p *popup) hitAt(x, y int) string {
	for _, h := range p.hits {
		if image.Pt(x, y).In(h.Rect) {
			return h.Action
		}
	}
	return ""
}

func toBGRA(img *image.RGBA, buf []byte) []byte {
	n := len(img.Pix)
	if cap(buf) < n {
		buf = make([]byte, n)
	}
	buf = buf[:n]
	for i := 0; i < n; i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = img.Pix[i+2], img.Pix[i+1], img.Pix[i], img.Pix[i+3]
	}
	return buf
}

func lightTheme() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, personalizeKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("SystemUsesLightTheme")
	return err == nil && v == 1
}
