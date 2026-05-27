//go:build windows

package mmapfile

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func mapFile(file *os.File, size int, writable bool) ([]byte, error) {
	protect := uint32(syscall.PAGE_READONLY)
	access := uint32(syscall.FILE_MAP_READ)
	if writable {
		protect = syscall.PAGE_READWRITE
		access = syscall.FILE_MAP_WRITE
	}

	fileSize := uint64(size)
	mapping, err := syscall.CreateFileMapping(
		syscall.Handle(file.Fd()),
		nil,
		protect,
		uint32(fileSize>>32),
		uint32(fileSize),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("mmapfile: CreateFileMapping %q: %w", file.Name(), err)
	}
	defer syscall.CloseHandle(mapping)

	ptr, err := syscall.MapViewOfFile(mapping, access, 0, 0, uintptr(size))
	if err != nil {
		return nil, fmt.Errorf("mmapfile: MapViewOfFile %q: %w", file.Name(), err)
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size), nil //nolint:gosec
}

func unmapFile(file *os.File, data []byte, _ bool) error {
	if len(data) == 0 {
		return nil
	}

	addr := uintptr(unsafe.Pointer(&data[0]))
	if err := syscall.UnmapViewOfFile(addr); err != nil {
		return fmt.Errorf("mmapfile: UnmapViewOfFile %q: %w", file.Name(), err)
	}

	return nil
}

func syncFile(file *os.File, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	err := flushViewOfFile(uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	if syncErr := file.Sync(); syncErr != nil && err == nil {
		err = syncErr
	}
	if err != nil {
		return fmt.Errorf("mmapfile: sync %q: %w", file.Name(), err)
	}

	return nil
}

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procFlushViewOfFile = modkernel32.NewProc("FlushViewOfFile")
)

func flushViewOfFile(addr, length uintptr) error {
	r1, _, err := procFlushViewOfFile.Call(addr, length)
	if r1 == 0 {
		return err
	}

	return nil
}
