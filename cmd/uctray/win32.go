//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	gdi32  = windows.NewLazySystemDLL("gdi32.dll")
	dwmapi = windows.NewLazySystemDLL("dwmapi.dll")
	shcore = windows.NewLazySystemDLL("shcore.dll")
	kernel = windows.NewLazySystemDLL("kernel32.dll")
	procs  = map[string]*windows.LazyProc{}
	libs   = map[*windows.LazyDLL][]string{
		user32: {"RegisterClassExW", "CreateWindowExW", "DefWindowProcW", "GetMessageW", "TranslateMessage",
			"DispatchMessageW", "PostMessageW", "ShowWindow", "SetWindowPos", "SetForegroundWindow", "GetCursorPos",
			"MonitorFromPoint", "GetMonitorInfoW", "BeginPaint", "EndPaint", "InvalidateRect", "TrackMouseEvent",
			"LoadCursorW", "SetCursor", "SetProcessDpiAwarenessContext"},
		gdi32:  {"SetDIBitsToDevice"},
		dwmapi: {"DwmSetWindowAttribute", "DwmExtendFrameIntoClientArea"},
		shcore: {"GetDpiForMonitor"},
		kernel: {"GetModuleHandleW"},
	}
)

func init() {
	for dll, names := range libs {
		for _, n := range names {
			procs[n] = dll.NewProc(n)
		}
	}
}

func call(name string, args ...uintptr) uintptr {
	r, _, _ := procs[name].Call(args...)
	return r
}

const (
	wmActivate     = 0x0006
	wmPaint        = 0x000F
	wmEraseBkgnd   = 0x0014
	wmSetCursor    = 0x0020
	wmKeyDown      = 0x0100
	wmMouseMove    = 0x0200
	wmLButtonUp    = 0x0202
	wmMouseLeave   = 0x02A3
	wmApp          = 0x8000
	wmToggle       = wmApp + 1
	wmData         = wmApp + 2
	wsPopup        = 0x80000000
	wsExTopmost    = 0x00000008
	wsExToolWindow = 0x00000080
	swHide         = 0
	swShow         = 5
	swpNoZOrder    = 0x0004
	swpShowWindow  = 0x0040
	vkEscape       = 0x1B
	tmeLeave       = 0x00000002
	idcArrow       = 32512
	idcHand        = 32649
	dpiAwareV2     = ^uintptr(3)
	monitorNearest = 2
	mdtEffective   = 0
	dwmDarkMode    = 20
	dwmCorners     = 33
	dwmRound       = 2
)

type wndClass struct {
	Size, Style                uint32
	WndProc                    uintptr
	ClsExtra, WndExtra         int32
	Instance, Icon, Cursor, Bg windows.Handle
	MenuName, ClassName        *uint16
	IconSm                     windows.Handle
}

type msg struct {
	Hwnd    windows.HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type point struct{ X, Y int32 }

type rect struct{ Left, Top, Right, Bottom int32 }

type monitorInfo struct {
	Size          uint32
	Monitor, Work rect
	Flags         uint32
}

type paintStruct struct {
	Hdc       windows.Handle
	Erase     int32
	Paint     rect
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}

type trackMouse struct {
	Size, Flags uint32
	Hwnd        windows.HWND
	HoverTime   uint32
}

type bitmapHeader struct {
	Size                   uint32
	Width, Height          int32
	Planes, BitCount       uint16
	Compression, SizeImage uint32
	XPels, YPels           int32
	ClrUsed, ClrImportant  uint32
}

type margins struct{ Left, Right, Top, Bottom int32 }

func ptr[T any](v *T) uintptr { return uintptr(unsafe.Pointer(v)) }

func loword(v uintptr) int { return int(int16(v & 0xFFFF)) }

func hiword(v uintptr) int { return int(int16((v >> 16) & 0xFFFF)) }
