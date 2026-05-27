//go:build linux || darwin || freebsd || openbsd || dragonfly

package mmapfile

import (
	"syscall"
	"unsafe"
)

func msync(data []byte) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(syscall.MS_SYNC),
	)
	if errno != 0 {
		return errno
	}

	return nil
}
