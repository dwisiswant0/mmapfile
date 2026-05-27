// Package mmapfile provides an [os.File]-like type backed by memory-mapped I/O.
//
// [MmapFile] implements [io.ReadWriteSeeker], [io.ReaderAt], [io.WriterAt],
// [io.Closer], [io.ReaderFrom], [io.WriterTo], and [io.StringWriter].
//
// Limitations:
//   - File size is fixed at open time; the file cannot grow.
//   - Runtime truncation after opening is not supported.
//   - Directory operations are not supported.
package mmapfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

var (
	ErrClosed           = errors.New("mmapfile: file is closed")
	ErrReadOnly         = errors.New("mmapfile: file is read-only")
	ErrInvalidWhence    = errors.New("mmapfile: invalid whence")
	ErrNegativeOffset   = errors.New("mmapfile: negative offset")
	ErrOffsetTooLarge   = errors.New("mmapfile: offset too large")
	ErrWriteOutOfBounds = errors.New("mmapfile: write would exceed file size")
)

const (
	maxInt64 = int64(1<<63 - 1)
	minInt64 = -maxInt64 - 1

	supportedOpenFileFlags = os.O_RDONLY | os.O_WRONLY | os.O_RDWR | os.O_APPEND | os.O_CREATE | os.O_EXCL | os.O_SYNC | os.O_TRUNC
)

// MmapFile represents a fixed-size memory-mapped file.
//
// Methods are safe for concurrent use. [Read], [Write], and [Seek] share a
// cursor, so concurrent cursor-based operations may interleave in the same way
// concurrent operations on an [os.File] can.
type MmapFile struct {
	mu       sync.RWMutex
	file     *os.File
	data     []byte
	offset   int64
	name     string
	writable bool
	closed   bool
}

type openConfig struct {
	writable  bool
	create    bool
	exclusive bool
	truncate  bool
	sync      bool
}

// Open opens name as a read-only memory-mapped file.
func Open(name string) (*MmapFile, error) {
	return OpenFile(name, os.O_RDONLY, 0, 0)
}

// OpenFile opens name as a fixed-size memory-mapped file.
//
// Supported flags are [os.O_RDONLY], [os.O_RDWR], [os.O_CREATE],
// [os.O_EXCL], [os.O_SYNC], and [os.O_TRUNC]. [os.O_APPEND] and
// [os.O_WRONLY] are not supported because mapped files have fixed size and
// mmap does not have write-only mappings.
//
// When creating a new file or truncating an existing file, size is the target
// mapped size. For existing files opened without [os.O_TRUNC], size is ignored.
// [os.O_TRUNC] and creating a non-empty file require [os.O_RDWR].
func OpenFile(name string, flag int, perm os.FileMode, size int64) (*MmapFile, error) {
	cfg, err := parseOpenFlags(flag, size)
	if err != nil {
		return nil, err
	}

	file, fileSize, err := openFixedSizeFile(name, flagForOS(cfg), perm, size, cfg)
	if err != nil {
		return nil, err
	}

	mf := &MmapFile{
		file:     file,
		name:     name,
		writable: cfg.writable,
	}

	closeOnError := true
	defer func() {
		if closeOnError {
			_ = file.Close()
		}
	}()

	if fileSize > 0 {
		data, err := mapFile(file, int(fileSize), cfg.writable)
		if err != nil {
			return nil, err
		}
		mf.data = data
	}

	closeOnError = false
	runtime.SetFinalizer(mf, (*MmapFile).Close)

	return mf, nil
}

func parseOpenFlags(flag int, size int64) (openConfig, error) {
	if flag&os.O_APPEND != 0 {
		return openConfig{}, fmt.Errorf("mmapfile: O_APPEND is not supported")
	}
	if flag&os.O_WRONLY != 0 {
		return openConfig{}, fmt.Errorf("mmapfile: O_WRONLY is not supported")
	}
	if unsupported := flag &^ supportedOpenFileFlags; unsupported != 0 {
		return openConfig{}, fmt.Errorf("mmapfile: unsupported open flag bits %#x", unsupported)
	}
	if size < 0 {
		return openConfig{}, fmt.Errorf("mmapfile: negative size")
	}

	cfg := openConfig{
		writable:  flag&os.O_RDWR != 0,
		create:    flag&os.O_CREATE != 0,
		exclusive: flag&os.O_EXCL != 0,
		truncate:  flag&os.O_TRUNC != 0,
		sync:      flag&os.O_SYNC != 0,
	}

	if cfg.exclusive && !cfg.create {
		return openConfig{}, fmt.Errorf("mmapfile: O_EXCL requires O_CREATE")
	}
	if cfg.truncate && !cfg.writable {
		return openConfig{}, fmt.Errorf("mmapfile: O_TRUNC requires O_RDWR")
	}
	if cfg.create && size > 0 && !cfg.writable {
		return openConfig{}, fmt.Errorf("mmapfile: O_CREATE with size > 0 requires O_RDWR")
	}

	return cfg, nil
}

func flagForOS(cfg openConfig) int {
	flag := os.O_RDONLY
	if cfg.writable {
		flag = os.O_RDWR
	}
	if cfg.truncate {
		flag |= os.O_TRUNC
	}
	if cfg.sync {
		flag |= os.O_SYNC
	}

	return flag
}

func openFixedSizeFile(name string, osFlag int, perm os.FileMode, size int64, cfg openConfig) (*os.File, int64, error) {
	var (
		file    *os.File
		err     error
		created bool
	)

	switch {
	case cfg.create && cfg.exclusive:
		file, err = os.OpenFile(name, osFlag|os.O_CREATE|os.O_EXCL, perm)
		created = err == nil
	case cfg.create:
		file, err = os.OpenFile(name, osFlag, perm)
		if errors.Is(err, os.ErrNotExist) {
			file, created, err = createFixedSizeFile(name, osFlag, perm)
		}
	default:
		file, err = os.OpenFile(name, osFlag, perm)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("mmapfile: open %q: %w", name, err)
	}

	closeOnError := true
	defer func() {
		if closeOnError {
			_ = file.Close()
		}
	}()

	if (created || cfg.truncate) && size > 0 {
		if err := file.Truncate(size); err != nil {
			return nil, 0, fmt.Errorf("mmapfile: resize %q: %w", name, err)
		}
	}

	info, err := file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("mmapfile: stat %q: %w", name, err)
	}
	if info.IsDir() {
		return nil, 0, fmt.Errorf("mmapfile: %q is a directory", name)
	}

	fileSize := info.Size()
	if fileSize < 0 {
		return nil, 0, fmt.Errorf("mmapfile: file %q has negative size", name)
	}
	if fileSize != int64(int(fileSize)) {
		return nil, 0, fmt.Errorf("mmapfile: file %q is too large", name)
	}

	closeOnError = false

	return file, fileSize, nil
}

func createFixedSizeFile(name string, osFlag int, perm os.FileMode) (*os.File, bool, error) {
	file, err := os.OpenFile(name, osFlag|os.O_CREATE|os.O_EXCL, perm)
	switch {
	case err == nil:
		return file, true, nil
	case errors.Is(err, os.ErrExist):
		file, err = os.OpenFile(name, osFlag, perm)
		if err != nil {
			return nil, false, err
		}
		return file, false, nil
	default:
		return nil, false, err
	}
}

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

// Name returns the name passed to [Open] or [OpenFile].
func (f *MmapFile) Name() string {
	return f.name
}

// Len returns the length of the mapped region.
func (f *MmapFile) Len() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.data)
}

// Bytes returns direct access to the mapped bytes.
//
// The returned slice is valid only until [Close]. The caller is responsible for
// synchronization when using it concurrently with other operations.
func (f *MmapFile) Bytes() []byte {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.data
}

// Read reads from the shared file cursor.
func (f *MmapFile) Read(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if len(b) == 0 {
		return 0, nil
	}
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n := copy(b, f.data[f.offset:])
	f.offset += int64(n)
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// ReadAt reads from the mapped bytes without changing the shared cursor.
func (f *MmapFile) ReadAt(b []byte, off int64) (int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return 0, ErrClosed
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if len(b) == 0 {
		return 0, nil
	}
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}

	n := copy(b, f.data[off:])
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// Write writes to the mapped bytes at the shared file cursor.
func (f *MmapFile) Write(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if !f.writable {
		return 0, ErrReadOnly
	}
	if len(b) == 0 {
		return 0, nil
	}

	available := int64(len(f.data)) - f.offset
	if available <= 0 {
		return 0, ErrWriteOutOfBounds
	}
	if int64(len(b)) > available {
		n := copy(f.data[f.offset:], b[:available])
		f.offset += int64(n)
		return n, ErrWriteOutOfBounds
	}

	n := copy(f.data[f.offset:], b)
	f.offset += int64(n)

	return n, nil
}

// WriteAt writes to the mapped bytes without changing the shared cursor.
func (f *MmapFile) WriteAt(b []byte, off int64) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if off < 0 {
		return 0, ErrNegativeOffset
	}
	if !f.writable {
		return 0, ErrReadOnly
	}
	if len(b) == 0 {
		return 0, nil
	}
	if off >= int64(len(f.data)) {
		return 0, ErrWriteOutOfBounds
	}

	available := int64(len(f.data)) - off
	if int64(len(b)) > available {
		n := copy(f.data[off:], b[:available])
		return n, ErrWriteOutOfBounds
	}

	return copy(f.data[off:], b), nil
}

// WriteString is like [Write], but writes string s.
func (f *MmapFile) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

// Seek sets the shared cursor for the next [Read] or [Write].
func (f *MmapFile) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}

	var (
		next int64
		ok   bool
	)

	switch whence {
	case io.SeekStart:
		next = offset
		ok = true
	case io.SeekCurrent:
		next, ok = addOffset(f.offset, offset)
	case io.SeekEnd:
		next, ok = addOffset(int64(len(f.data)), offset)
	default:
		return 0, ErrInvalidWhence
	}
	if !ok {
		return 0, ErrOffsetTooLarge
	}
	if next < 0 {
		return 0, ErrNegativeOffset
	}

	f.offset = next

	return next, nil
}

func addOffset(base, offset int64) (int64, bool) {
	if offset > 0 && base > maxInt64-offset {
		return 0, false
	}
	if offset < 0 && base < minInt64-offset {
		return 0, false
	}

	return base + offset, true
}

// ReadFrom reads from r into the mapped bytes at the shared cursor until EOF.
func (f *MmapFile) ReadFrom(r io.Reader) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return 0, ErrClosed
	}
	if !f.writable {
		return 0, ErrReadOnly
	}

	var total int64
	for f.offset < int64(len(f.data)) {
		dst := f.data[f.offset:]
		n, readErr := r.Read(dst)
		if n < 0 || n > len(dst) {
			return total, fmt.Errorf("mmapfile: invalid read count %d", n)
		}
		if n > 0 {
			f.offset += int64(n)
			total += int64(n)
		}
		if readErr == io.EOF {
			return total, nil
		}
		if readErr != nil {
			return total, readErr
		}
	}

	var probe [1]byte
	n, readErr := r.Read(probe[:])
	if n < 0 || n > 1 {
		return total, fmt.Errorf("mmapfile: invalid read count %d", n)
	}
	if n > 0 {
		return total, ErrWriteOutOfBounds
	}
	if readErr == io.EOF {
		return total, nil
	}
	if readErr != nil {
		return total, readErr
	}

	return total, ErrWriteOutOfBounds
}

// WriteTo writes the mapped bytes to w.
func (f *MmapFile) WriteTo(w io.Writer) (int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return 0, ErrClosed
	}

	n, err := w.Write(f.data)
	if n < 0 || n > len(f.data) {
		return int64(n), fmt.Errorf("mmapfile: invalid write count %d", n)
	}
	if err != nil {
		return int64(n), err
	}
	if n != len(f.data) {
		return int64(n), io.ErrShortWrite
	}

	return int64(n), nil
}

// Stat returns the underlying file metadata.
func (f *MmapFile) Stat() (os.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return nil, ErrClosed
	}

	return f.file.Stat()
}

// Sync flushes changes to the underlying file.
func (f *MmapFile) Sync() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.closed {
		return ErrClosed
	}
	if !f.writable || len(f.data) == 0 {
		return nil
	}

	return syncFile(f.file, f.data)
}

// Close unmaps and closes the file. It is safe to call Close more than once.
func (f *MmapFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	runtime.SetFinalizer(f, nil)

	file := f.file
	data := f.data
	writable := f.writable
	f.file = nil
	f.data = nil

	if file == nil {
		return nil
	}

	var err error
	if len(data) > 0 {
		if unmapErr := unmapFile(file, data, writable); unmapErr != nil {
			err = unmapErr
		}
	}
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}
