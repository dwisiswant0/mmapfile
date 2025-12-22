// Package mmapfile provides an [os.File]-like type backed by memory-mapped I/O.
//
// [MmapFile] implements [io.ReadWriteSeeker], [io.ReaderAt], [io.WriterAt], and
// [io.Closer], allowing it to be used as a drop-in replacement for [os.File] in
// many contexts.
//
// Limitations:
//   - File size is fixed at open time; the file cannot grow.
//   - Truncate is not supported.
//   - Directory operations are not supported.
package mmapfile

import (
	"errors"
	"io"
	"os"
	"sync"
)

// Common errors.
var (
	ErrClosed           = errors.New("mmapfile: file is closed")
	ErrReadOnly         = errors.New("mmapfile: file is read-only")
	ErrInvalidWhence    = errors.New("mmapfile: invalid whence")
	ErrNegativeOffset   = errors.New("mmapfile: negative offset")
	ErrOffsetTooLarge   = errors.New("mmapfile: offset too large")
	ErrWriteOutOfBounds = errors.New("mmapfile: write would exceed file size")
)

// MmapFile represents a memory-mapped file that implements an [os.File]-like
// interface.
//
// The methods of [MmapFile] are safe for concurrent use, with the exception
// that concurrent [Read]/[Write]/[Seek] ops may interleave unpredictably since
// they share a cursor.
//
// Use [ReadAt]/[WriteAt] for concurrent positional I/O.
type MmapFile struct {
	mu       sync.RWMutex
	data     []byte
	offset   int64
	name     string
	writable bool
	closed   bool
	platform any //nolint:unused // platform-specific data (e.g., file handle for fallback impl)
}

// fileHolder holds the underlying file.
type fileHolder struct {
	file *os.File
}

// Compile-time interface checks.
var (
	_ io.Reader       = (*MmapFile)(nil)
	_ io.Writer       = (*MmapFile)(nil)
	_ io.Seeker       = (*MmapFile)(nil)
	_ io.ReaderAt     = (*MmapFile)(nil)
	_ io.WriterAt     = (*MmapFile)(nil)
	_ io.Closer       = (*MmapFile)(nil)
	_ io.ReaderFrom   = (*MmapFile)(nil)
	_ io.WriterTo     = (*MmapFile)(nil)
	_ io.StringWriter = (*MmapFile)(nil)
)

// Name returns the name of the file as presented to [Open] or [OpenFile].
func (f *MmapFile) Name() string {
	return f.name
}

// Len returns the length of the memory-mapped region.
func (f *MmapFile) Len() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.data)
}

// Bytes returns direct access to the underlying memory-mapped byte slice.
//
// WARNING: The returned slice is only valid until [Close] is called.
// Modifying the slice on a read-only file will cause a panic/segfault.
// The caller is responsible for synchronization when using this method.
func (f *MmapFile) Bytes() []byte {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.data
}

// Read reads up to len(b) bytes from the file, advancing the file offset.
//
// It returns the number of bytes read and any error encountered.
// At end of file, Read returns 0, io.EOF.
func (f *MmapFile) Read(b []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n = copy(b, f.data[f.offset:])
	f.offset += int64(n)

	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

// ReadAt reads len(b) bytes from the file starting at byte offset off.
//
// It returns the number of bytes read and any error encountered.
// ReadAt does not affect the file offset used by [Read]/[Write]/[Seek].
//
// It is safe for concurrent use.
func (f *MmapFile) ReadAt(b []byte, off int64) (n int, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return 0, ErrClosed
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n = copy(b, f.data[off:])
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// Write writes len(b) bytes to the file, advancing the file offset.
//
// It returns the number of bytes written and any error encountered.
// Write returns an error if the file was opened read-only or if the
// write would exceed the file's size.
func (f *MmapFile) Write(b []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if !f.writable {
		return 0, ErrReadOnly
	}

	available := int64(len(f.data)) - f.offset
	if available <= 0 {
		return 0, ErrWriteOutOfBounds
	}

	if int64(len(b)) > available {
		n = copy(f.data[f.offset:], b[:available])
		f.offset += int64(n)
		return n, ErrWriteOutOfBounds
	}

	n = copy(f.data[f.offset:], b)
	f.offset += int64(n)

	return n, nil
}

// WriteAt writes len(b) bytes to the file starting at byte offset off.
//
// It returns the number of bytes written and any error encountered.
// WriteAt does not affect the file offset used by [Read]/[Write]/[Seek].
//
// It is safe for concurrent use (though overlapping writes MAY interleave).
func (f *MmapFile) WriteAt(b []byte, off int64) (n int, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return 0, ErrClosed
	}
	if !f.writable {
		return 0, ErrReadOnly
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if off >= int64(len(f.data)) {
		return 0, ErrWriteOutOfBounds
	}

	available := int64(len(f.data)) - off
	if int64(len(b)) > available {
		n = copy(f.data[off:], b[:available])
		return n, ErrWriteOutOfBounds
	}

	n = copy(f.data[off:], b)

	return n, nil
}

// WriteString is like Write, but writes the contents of string s.
func (f *MmapFile) WriteString(s string) (n int, err error) {
	return f.Write([]byte(s))
}

// Seek sets the offset for the next Read or Write on the file,
// interpreted according to whence:
//   - [io.SeekStart] (0): relative to the start of the file
//   - [io.SeekCurrent] (1): relative to the current offset
//   - [io.SeekEnd] (2): relative to the end of the file
//
// It returns the new offset and any error encountered.
func (f *MmapFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = f.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(f.data)) + offset
	default:
		return 0, ErrInvalidWhence
	}

	if newOffset < 0 {
		return 0, ErrNegativeOffset
	}

	f.offset = newOffset

	return newOffset, nil
}

// ReadFrom reads data from r until EOF and writes it to the file.
//
// It returns the number of bytes read and any error encountered.
func (f *MmapFile) ReadFrom(r io.Reader) (n int64, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if !f.writable {
		return 0, ErrReadOnly
	}

	for f.offset < int64(len(f.data)) {
		m, readErr := r.Read(f.data[f.offset:])
		n += int64(m)
		f.offset += int64(m)
		if readErr == io.EOF {
			return n, nil
		}
		if readErr != nil {
			return n, readErr
		}
	}

	// Check if there's more data in the reader
	var buf [1]byte
	_, readErr := r.Read(buf[:])
	if readErr == nil || readErr != io.EOF {
		return n, ErrWriteOutOfBounds
	}

	return n, nil
}

// WriteTo writes the entire file contents to w.
//
// It returns the number of bytes written and any error encountered.
func (f *MmapFile) WriteTo(w io.Writer) (n int64, err error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return 0, ErrClosed
	}

	written, err := w.Write(f.data)
	return int64(written), err
}

// Stat returns the FileInfo structure describing the file.
func (f *MmapFile) Stat() (os.FileInfo, error) {
	f.mu.RLock()
	closed := f.closed
	name := f.name
	f.mu.RUnlock()

	if closed {
		return nil, ErrClosed
	}

	if fh, ok := f.platform.(*fileHolder); ok && fh.file != nil {
		return fh.file.Stat()
	}

	return os.Stat(name)
}
