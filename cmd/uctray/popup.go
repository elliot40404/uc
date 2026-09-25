//go:build windows

package main

import (
	"image"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/elliot40404/uc/internal/flyout"
)

const reopenGuard = 300 * time.Millisecond

type popup struct {
	mu        sync.Mutex
	view      flyout.View
	hwnd      windows.HWND
	fonts     *flyout.Fonts
	img       *image.RGBA
	bgra      []byte
	hits      []flyout.Hit
	hover     string
	scale     float64
	shown     bool
	hiddenAt  time.Time
	tracking  bool
	onRefresh func()
}

func (p *popup) run(ready chan<- struct{}) {
	runtime.LockOSThread()
	p.hwnd = p.create()
	close(ready)
	var m msg
	for int32(call("GetMessageW", ptr(&m), 0, 0, 0)) > 0 {
		call("TranslateMessage", ptr(&m))
		call("DispatchMessageW", ptr(&m))
	}
}

func (p *popup) create() windows.HWND {
	name := windows.StringToUTF16Ptr("ucTrayPopup")
	inst := windows.Handle(call("GetModuleHandleW", 0))
	wc := wndClass{
		WndProc:   windows.NewCallback(p.proc),
		Instance:  inst,
		Cursor:    windows.Handle(call("LoadCursorW", 0, idcArrow)),
		ClassName: name,
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	call("RegisterClassExW", ptr(&wc))
	h := windows.HWND(call("CreateWindowExW", wsExTopmost|wsExToolWindow, uintptr(unsafe.Pointer(name)), 0, wsPopup,
		0, 0, 1, 1, 0, 0, uintptr(inst), 0))
	corner := int32(dwmRound)
	call("DwmSetWindowAttribute", uintptr(h), dwmCorners, ptr(&corner), 4)
	call("DwmExtendFrameIntoClientArea", uintptr(h), ptr(&margins{0, 0, 0, 1}))
	return h
}

func (p *popup) setView(v flyout.View) {
	p.mu.Lock()
	p.view = v
	p.mu.Unlock()
	if p.hwnd != 0 {
		call("PostMessageW", uintptr(p.hwnd), wmData, 0, 0)
	}
}

func (p *popup) toggle() {
	call("PostMessageW", uintptr(p.hwnd), wmToggle, 0, 0)
}

func (p *popup) proc(hwnd windows.HWND, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmToggle:
		p.onToggle()
	case wmData:
		if p.shown {
			p.redraw()
		}
	case wmActivate:
		if loword(wParam) == 0 {
			p.hide()
		}
	case wmPaint:
		p.paint()
		return 0
	case wmEraseBkgnd:
		return 1
	case wmMouseMove:
		p.onMove(loword(lParam), hiword(lParam))
	case wmMouseLeave:
		p.tracking = false
		p.setHover("")
	case wmSetCursor:
		if p.hover != "" {
			call("SetCursor", call("LoadCursorW", 0, idcHand))
			return 1
		}
	case wmLButtonUp:
		p.onClick(loword(lParam), hiword(lParam))
	case wmKeyDown:
		if wParam == vkEscape {
			p.hide()
		}
	}
	return call("DefWindowProcW", uintptr(hwnd), uintptr(message), wParam, lParam)
}

func (p *popup) onToggle() {
	if p.shown {
		p.hide()
		return
	}
	if time.Since(p.hiddenAt) < reopenGuard {
		return
	}
	p.show()
}

func (p *popup) show() {
	var cursor point
	call("GetCursorPos", ptr(&cursor))
	mon := call("MonitorFromPoint", uintptr(uint32(cursor.X))|uintptr(uint32(cursor.Y))<<32, monitorNearest)
	var dpiX, dpiY uint32
	call("GetDpiForMonitor", mon, mdtEffective, ptr(&dpiX), ptr(&dpiY))
	p.scale = max(float64(dpiX), 96) / 96
	p.hover = ""
	if p.fonts == nil {
		p.fonts = flyout.LoadFonts()
	}
	p.render()
	info := monitorInfo{}
	info.Size = uint32(unsafe.Sizeof(info))
	call("GetMonitorInfoW", mon, ptr(&info))
	w, h := int32(p.img.Bounds().Dx()), int32(p.img.Bounds().Dy())
	x, y := place(cursor, info.Work, w, h, int32(12*p.scale))
	p.shown = true
	call("SetWindowPos", uintptr(p.hwnd), 0, uintptr(x), uintptr(y), uintptr(w), uintptr(h), swpNoZOrder|swpShowWindow)
	call("SetForegroundWindow", uintptr(p.hwnd))
}

func (p *popup) hide() {
	if !p.shown {
		return
	}
	call("ShowWindow", uintptr(p.hwnd), swHide)
	p.shown, p.hiddenAt = false, time.Now()
	p.img, p.bgra, p.hits, p.fonts = nil, nil, nil, nil
	go debug.FreeOSMemory()
}

func place(cursor point, work rect, w, h, gap int32) (int32, int32) {
	x := min(max(cursor.X-w/2, work.Left+gap), work.Right-w-gap)
	y := work.Bottom - h - gap
	switch {
	case cursor.Y < work.Top:
		y = work.Top + gap
	case cursor.X < work.Left:
		x, y = work.Left+gap, min(max(cursor.Y-h/2, work.Top+gap), work.Bottom-h-gap)
	case cursor.X >= work.Right:
		x, y = work.Right-w-gap, min(max(cursor.Y-h/2, work.Top+gap), work.Bottom-h-gap)
	}
	return x, max(y, work.Top)
}
