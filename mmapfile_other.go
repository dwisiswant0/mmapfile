//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly && !windows

package mmapfile

import (
	"fmt"
	"io"
	"os"
)

// Open memory-maps the named file for reading.
//
// On unsupported platforms, this falls back to regular file I/O.
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

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	fileSize := fi.Size()

	if create && fileSize == 0 && size > 0 {
		if err := f.Truncate(size); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("mmapfile: failed to set file size: %w", err)
		}
		fileSize = size
	} else if trunc && size > 0 {
		if err := f.Truncate(size); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("mmapfile: failed to truncate file: %w", err)
		}
		fileSize = size
	}

	if fileSize == 0 {
		return &MmapFile{
			data:     nil,
			name:     name,
			writable: writable,
			platform: &fileHolder{file: f},
		}, nil
	}

	if fileSize < 0 {
		_ = f.Close()
		return nil, fmt.Errorf("mmapfile: file %q has negative size", name)
	}
	if fileSize != int64(int(fileSize)) {
		_ = f.Close()
		return nil, fmt.Errorf("mmapfile: file %q is too large", name)
	}

	// Fallback: read entire file into memory
	data := make([]byte, fileSize)
	if _, err := io.ReadFull(f, data); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("mmapfile: failed to read file: %w", err)
	}

	mf := &MmapFile{
		data:     data,
		name:     name,
		writable: writable,
		platform: &fileHolder{file: f},
	}

	return mf, nil
}

// Close closes the memory-mapped file.
func (f *MmapFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}
	f.closed = true

	var err error
	if fh, ok := f.platform.(*fileHolder); ok && fh != nil && fh.file != nil {
		if f.writable && len(f.data) > 0 {
			if _, seekErr := fh.file.Seek(0, io.SeekStart); seekErr != nil {
				err = seekErr
			} else if _, writeErr := fh.file.Write(f.data); writeErr != nil {
				err = writeErr
			}
		}
		if closeErr := fh.file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		f.platform = nil
	}

	f.data = nil

	return err
}

// Sync flushes changes to the underlying file.
func (f *MmapFile) Sync() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return ErrClosed
	}

	fh, ok := f.platform.(*fileHolder)
	if !f.writable || !ok || fh == nil || fh.file == nil || len(f.data) == 0 {
		return nil
	}

	if _, err := fh.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := fh.file.Write(f.data); err != nil {
		return err
	}

	return fh.file.Sync()
}
