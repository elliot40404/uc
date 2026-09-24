package term

import (
	"os"
	"syscall"
	"unsafe"
)

const enableVirtualTerminal = 0x0004

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode = kernel32.NewProc("GetConsoleMode")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func enableVT(f *os.File) bool {
	var mode uint32
	h := f.Fd()
	if ok, _, _ := getConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); ok == 0 {
		return false
	}
	ok, _, _ := setConsoleMode.Call(h, uintptr(mode|enableVirtualTerminal))
	return ok != 0
}
