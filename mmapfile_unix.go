//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package mmapfile

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// Open memory-maps the named file for reading.
// The returned MmapFile implements io.ReadSeeker and io.ReaderAt.
func Open(name string) (*MmapFile, error) {
	return OpenFile(name, os.O_RDONLY, 0, 0)
}

// OpenFile opens a memory-mapped file with the specified flags and permissions.
//
// Supported flags:
//   - [os.O_RDONLY]: Open for reading only
//   - [os.O_RDWR]: Open for reading and writing
//   - [os.O_CREATE]: Create the file if it doesn't exist (requires size > 0)
//   - [os.O_TRUNC]: Truncate the file to the specified size
//
// The size parameter is used when creating a new file or when [os.O_TRUNC] is
// specified. For existing files opened without [os.O_TRUNC], size is ignored
// and the file's current size is used.
//
// Note: [os.O_APPEND] is not supported as mmap does not support growing files.
func OpenFile(name string, flag int, perm os.FileMode, size int64) (*MmapFile, error) {
	writable := flag&os.O_RDWR != 0 || flag&os.O_WRONLY != 0
	create := flag&os.O_CREATE != 0
	trunc := flag&os.O_TRUNC != 0

	if flag&os.O_APPEND != 0 {
		return nil, fmt.Errorf("mmapfile: O_APPEND is not supported")
	}

	osFlag := os.O_RDONLY
	if writable {
		osFlag = os.O_RDWR
	}
	if create {
		osFlag |= os.O_CREATE
	}

	f, err := os.OpenFile(name, osFlag, perm)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := fi.Size()

	if create && fileSize == 0 && size > 0 {
		if err := f.Truncate(size); err != nil {
			return nil, fmt.Errorf("mmapfile: failed to set file size: %w", err)
		}
		fileSize = size
	} else if trunc && size > 0 {
		if err := f.Truncate(size); err != nil {
			return nil, fmt.Errorf("mmapfile: failed to truncate file: %w", err)
		}
		fileSize = size
	}

	if fileSize == 0 {
		return &MmapFile{
			data:     nil,
			name:     name,
			writable: writable,
		}, nil
	}

	if fileSize < 0 {
		return nil, fmt.Errorf("mmapfile: file %q has negative size", name)
	}
	if fileSize != int64(int(fileSize)) {
		return nil, fmt.Errorf("mmapfile: file %q is too large", name)
	}

	prot := syscall.PROT_READ
	if writable {
		prot |= syscall.PROT_WRITE
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(fileSize), prot, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmapfile: mmap failed: %w", err)
	}

	mf := &MmapFile{
		data:     data,
		name:     name,
		writable: writable,
	}

	runtime.SetFinalizer(mf, (*MmapFile).Close)

	return mf, nil
}

// Close closes the memory-mapped file.
//
// After Close, the [MmapFile] should not be used.
func (f *MmapFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}
	f.closed = true

	runtime.SetFinalizer(f, nil)

	if len(f.data) == 0 {
		f.data = nil
		return nil
	}

	data := f.data
	f.data = nil

	return syscall.Munmap(data)
}

// Sync flushes changes to the underlying file.
//
// This is a no-op for read-only files.
func (f *MmapFile) Sync() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return ErrClosed
	}
	if !f.writable || len(f.data) == 0 {
		return nil
	}

	// MS_SYNC: synchronous write
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&f.data[0])),
		uintptr(len(f.data)),
		uintptr(syscall.MS_SYNC))
	if errno != 0 {
		return errno
	}

	return nil
}
