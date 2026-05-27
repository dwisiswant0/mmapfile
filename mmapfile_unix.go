//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package mmapfile

import (
	"fmt"
	"os"
	"syscall"
)

func mapFile(file *os.File, size int, writable bool) ([]byte, error) {
	prot := syscall.PROT_READ
	if writable {
		prot |= syscall.PROT_WRITE
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, size, prot, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmapfile: mmap %q: %w", file.Name(), err)
	}

	return data, nil
}

func unmapFile(file *os.File, data []byte, _ bool) error {
	if len(data) == 0 {
		return nil
	}
	if err := syscall.Munmap(data); err != nil {
		return fmt.Errorf("mmapfile: munmap %q: %w", file.Name(), err)
	}

	return nil
}

func syncFile(file *os.File, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	err := msync(data)
	if syncErr := file.Sync(); syncErr != nil && err == nil {
		err = syncErr
	}
	if err != nil {
		return fmt.Errorf("mmapfile: sync %q: %w", file.Name(), err)
	}

	return nil
}
