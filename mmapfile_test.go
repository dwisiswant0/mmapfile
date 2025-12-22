package mmapfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type failingWriter struct {
	limit   int
	written int
}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	if w.written+len(p) > w.limit {
		n = w.limit - w.written
		w.written += n
		return n, errors.New("write failed")
	}
	w.written += len(p)
	return len(p), nil
}

func TestOpen(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		f, err := Open("testdata/hello.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		if f.Name() != "testdata/hello.txt" {
			t.Errorf("Name() = %q, want %q", f.Name(), "testdata/hello.txt")
		}

		if f.Len() == 0 {
			t.Error("Len() = 0, want > 0")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := Open("testdata/nonexistent.txt")
		if err == nil {
			t.Error("Open should fail for non-existent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		f, err := Open("testdata/empty.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		if f.Len() != 0 {
			t.Errorf("Len() = %d, want 0", f.Len())
		}
	})

	t.Run("unix permission denied", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Unix permission test")
		}

		path := filepath.Join(t.TempDir(), "no_perm.txt")

		// Create file
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		// Remove read permission
		if err := os.Chmod(path, 0000); err != nil {
			t.Skipf("Cannot change permissions: %v", err)
		}
		defer os.Chmod(path, 0644) // Restore for cleanup

		_, err := Open(path)
		if err == nil {
			t.Error("Open should fail for permission denied")
		}
	})

}

func TestOpenFile(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "new.txt")

		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		if f.Len() != 100 {
			t.Errorf("Len() = %d, want 100", f.Len())
		}
	})

	t.Run("O_APPEND not supported", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "append.txt")
		_, err := OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644, 100)
		if err == nil {
			t.Error("OpenFile should fail with O_APPEND")
		}
	})

	t.Run("truncate existing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "trunc.txt")

		// Create a file with some content
		if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		f, err := OpenFile(path, os.O_RDWR|os.O_TRUNC, 0644, 50)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		if f.Len() != 50 {
			t.Errorf("Len() = %d, want 50", f.Len())
		}
	})
}

func TestRead(t *testing.T) {
	f, err := Open("testdata/hello.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	t.Run("read all", func(t *testing.T) {
		buf := make([]byte, f.Len())
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			t.Errorf("Read failed: %v", err)
		}
		if n != f.Len() {
			t.Errorf("Read %d bytes, want %d", n, f.Len())
		}
		if !strings.HasPrefix(string(buf), "Hello, World!") {
			t.Errorf("unexpected content: %q", buf[:min(20, len(buf))])
		}
	})

	t.Run("read at EOF", func(t *testing.T) {
		// Seek to end
		f.Seek(0, io.SeekEnd)
		buf := make([]byte, 10)
		n, err := f.Read(buf)
		if n != 0 || err != io.EOF {
			t.Errorf("Read at EOF: got n=%d, err=%v, want n=0, err=EOF", n, err)
		}
	})

	t.Run("partial read", func(t *testing.T) {
		f.Seek(0, io.SeekStart)
		buf := make([]byte, 5)
		n, err := f.Read(buf)
		if err != nil {
			t.Errorf("Read failed: %v", err)
		}
		if n != 5 {
			t.Errorf("Read %d bytes, want 5", n)
		}
		if string(buf) != "Hello" {
			t.Errorf("Read %q, want %q", buf, "Hello")
		}
	})
}

func TestReadAt(t *testing.T) {
	f, err := Open("testdata/binary.dat")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	t.Run("read at offset", func(t *testing.T) {
		buf := make([]byte, 10)
		n, err := f.ReadAt(buf, 0)
		if err != nil {
			t.Errorf("ReadAt failed: %v", err)
		}
		if n != 10 {
			t.Errorf("ReadAt %d bytes, want 10", n)
		}
		if string(buf) != "ABCDEFGHIJ" {
			t.Errorf("ReadAt got %q, want %q", buf, "ABCDEFGHIJ")
		}
	})

	t.Run("read at middle", func(t *testing.T) {
		buf := make([]byte, 5)
		n, err := f.ReadAt(buf, 10)
		if err != nil {
			t.Errorf("ReadAt failed: %v", err)
		}
		if n != 5 {
			t.Errorf("ReadAt %d bytes, want 5", n)
		}
		if string(buf) != "KLMNO" {
			t.Errorf("ReadAt got %q, want %q", buf, "KLMNO")
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		buf := make([]byte, 5)
		_, err := f.ReadAt(buf, -1)
		if !errors.Is(err, ErrNegativeOffset) {
			t.Errorf("ReadAt with negative offset: got %v, want ErrNegativeOffset", err)
		}
	})

	t.Run("offset past EOF", func(t *testing.T) {
		buf := make([]byte, 5)
		_, err := f.ReadAt(buf, int64(f.Len()+10))
		if err != io.EOF {
			t.Errorf("ReadAt past EOF: got %v, want io.EOF", err)
		}
	})

	t.Run("concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(offset int) {
				defer wg.Done()
				buf := make([]byte, 3)
				_, err := f.ReadAt(buf, int64(offset))
				if err != nil && err != io.EOF {
					t.Errorf("concurrent ReadAt failed: %v", err)
				}
			}(i * 3)
		}
		wg.Wait()
	})
}

func TestWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "write.txt")

	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	t.Run("write and read back", func(t *testing.T) {
		data := []byte("Hello, mmap!")
		n, err := f.Write(data)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Write %d bytes, want %d", n, len(data))
		}

		// Read back
		f.Seek(0, io.SeekStart)
		buf := make([]byte, len(data))
		_, err = f.Read(buf)
		if err != nil {
			t.Errorf("Read failed: %v", err)
		}
		if string(buf) != "Hello, mmap!" {
			t.Errorf("Read %q, want %q", buf, "Hello, mmap!")
		}
	})

	t.Run("write past end", func(t *testing.T) {
		f.Seek(0, io.SeekEnd)
		_, err := f.Write([]byte("x"))
		if !errors.Is(err, ErrWriteOutOfBounds) {
			t.Errorf("Write past end: got %v, want ErrWriteOutOfBounds", err)
		}
	})
}

func TestWriteAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "writeat.txt")

	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	t.Run("write at offset", func(t *testing.T) {
		data := []byte("HELLO")
		n, err := f.WriteAt(data, 10)
		if err != nil {
			t.Errorf("WriteAt failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("WriteAt %d bytes, want %d", n, len(data))
		}

		// Read back
		buf := make([]byte, 5)
		_, err = f.ReadAt(buf, 10)
		if err != nil {
			t.Errorf("ReadAt failed: %v", err)
		}
		if string(buf) != "HELLO" {
			t.Errorf("ReadAt got %q, want %q", buf, "HELLO")
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		_, err := f.WriteAt([]byte("x"), -1)
		if !errors.Is(err, ErrNegativeOffset) {
			t.Errorf("WriteAt with negative offset: got %v, want ErrNegativeOffset", err)
		}
	})

	t.Run("offset past EOF", func(t *testing.T) {
		_, err := f.WriteAt([]byte("x"), int64(f.Len()+10))
		if !errors.Is(err, ErrWriteOutOfBounds) {
			t.Errorf("WriteAt past EOF: got %v, want ErrWriteOutOfBounds", err)
		}
	})

	t.Run("concurrent writes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "concurrent.txt")

		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		var wg sync.WaitGroup
		numGoroutines := 10
		dataSize := 5

		for i := range numGoroutines {
			id := i
			wg.Go(func() {
				data := fmt.Appendf(nil, "data%d", id)
				offset := int64(id * dataSize)
				_, err := f.WriteAt(data, offset)
				if err != nil {
					t.Errorf("WriteAt failed: %v", err)
				}
			})
		}
		wg.Wait()

		// Verify data
		for i := range numGoroutines {
			expected := fmt.Sprintf("data%d", i)
			buf := make([]byte, len(expected))
			_, err := f.ReadAt(buf, int64(i*dataSize))
			if err != nil {
				t.Errorf("ReadAt failed: %v", err)
			}
			if string(buf) != expected {
				t.Errorf("ReadAt got %q, want %q", buf, expected)
			}
		}
	})
}

func TestWriteString(t *testing.T) {
	path := filepath.Join(t.TempDir(), "writestring.txt")

	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 50)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	n, err := f.WriteString("Hello, World!")
	if err != nil {
		t.Errorf("WriteString failed: %v", err)
	}
	if n != 13 {
		t.Errorf("WriteString wrote %d bytes, want 13", n)
	}
}

func TestReadOnlyWrite(t *testing.T) {
	f, err := Open("testdata/hello.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	_, err = f.Write([]byte("x"))
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("Write on read-only file: got %v, want ErrReadOnly", err)
	}

	_, err = f.WriteAt([]byte("x"), 0)
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("WriteAt on read-only file: got %v, want ErrReadOnly", err)
	}
}

func TestSeek(t *testing.T) {
	f, err := Open("testdata/binary.dat")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	t.Run("SeekStart", func(t *testing.T) {
		pos, err := f.Seek(5, io.SeekStart)
		if err != nil {
			t.Errorf("Seek failed: %v", err)
		}
		if pos != 5 {
			t.Errorf("Seek returned %d, want 5", pos)
		}
	})

	t.Run("SeekCurrent", func(t *testing.T) {
		f.Seek(10, io.SeekStart)
		pos, err := f.Seek(5, io.SeekCurrent)
		if err != nil {
			t.Errorf("Seek failed: %v", err)
		}
		if pos != 15 {
			t.Errorf("Seek returned %d, want 15", pos)
		}
	})

	t.Run("SeekEnd", func(t *testing.T) {
		pos, err := f.Seek(-5, io.SeekEnd)
		if err != nil {
			t.Errorf("Seek failed: %v", err)
		}
		expected := int64(f.Len() - 5)
		if pos != expected {
			t.Errorf("Seek returned %d, want %d", pos, expected)
		}
	})

	t.Run("negative result", func(t *testing.T) {
		_, err := f.Seek(-100, io.SeekStart)
		if !errors.Is(err, ErrNegativeOffset) {
			t.Errorf("Seek to negative: got %v, want ErrNegativeOffset", err)
		}
	})

	t.Run("invalid whence", func(t *testing.T) {
		_, err := f.Seek(0, 999)
		if !errors.Is(err, ErrInvalidWhence) {
			t.Errorf("Seek with invalid whence: got %v, want ErrInvalidWhence", err)
		}
	})

	t.Run("seek past end (allowed)", func(t *testing.T) {
		pos, err := f.Seek(int64(f.Len()+100), io.SeekStart)
		if err != nil {
			t.Errorf("Seek past end failed: %v", err)
		}
		if pos != int64(f.Len()+100) {
			t.Errorf("Seek returned %d, want %d", pos, f.Len()+100)
		}
	})
}

func TestClose(t *testing.T) {
	t.Run("double close is safe", func(t *testing.T) {
		f, err := Open("testdata/hello.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}

		if err := f.Close(); err != nil {
			t.Errorf("first Close failed: %v", err)
		}

		if err := f.Close(); err != nil {
			t.Errorf("second Close failed: %v", err)
		}
	})

	t.Run("operations after close", func(t *testing.T) {
		f, err := Open("testdata/hello.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		f.Close()

		buf := make([]byte, 10)
		_, err = f.Read(buf)
		if !errors.Is(err, ErrClosed) {
			t.Errorf("Read after close: got %v, want ErrClosed", err)
		}

		_, err = f.Seek(0, io.SeekStart)
		if !errors.Is(err, ErrClosed) {
			t.Errorf("Seek after close: got %v, want ErrClosed", err)
		}
	})
}

func TestBytes(t *testing.T) {
	f, err := Open("testdata/binary.dat")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	data := f.Bytes()
	if len(data) != f.Len() {
		t.Errorf("Bytes() len = %d, want %d", len(data), f.Len())
	}
	if !strings.HasPrefix(string(data), "ABCDEFGHIJ") {
		t.Errorf("unexpected Bytes() content: %q", data[:min(10, len(data))])
	}
}

func TestStat(t *testing.T) {
	f, err := Open("testdata/hello.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if fi.Name() != "hello.txt" {
		t.Errorf("Stat().Name() = %q, want %q", fi.Name(), "hello.txt")
	}

	if fi.Size() != int64(f.Len()) {
		t.Errorf("Stat().Size() = %d, want %d", fi.Size(), f.Len())
	}
}

func TestSync(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.txt")

	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	// Write some data
	f.WriteString("Hello, Sync!")

	// Sync should not error
	if err := f.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	// Verify data is persisted by reopening
	f.Close()

	f2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer f2.Close()

	buf := make([]byte, 12)
	f2.Read(buf)
	if string(buf) != "Hello, Sync!" {
		t.Errorf("after Sync, got %q, want %q", buf, "Hello, Sync!")
	}
}

func TestReadFrom(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readfrom.txt")

	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	reader := strings.NewReader("Data from reader")
	n, err := f.ReadFrom(reader)
	if err != nil {
		t.Errorf("ReadFrom failed: %v", err)
	}
	if n != 16 {
		t.Errorf("ReadFrom read %d bytes, want 16", n)
	}

	// Verify
	f.Seek(0, io.SeekStart)
	buf := make([]byte, 16)
	f.Read(buf)
	if string(buf) != "Data from reader" {
		t.Errorf("ReadFrom content: got %q, want %q", buf, "Data from reader")
	}

	t.Run("with excess data", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "readfrom_excess.txt")

		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 10) // Small file
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		// Reader with more data than file can hold
		reader := strings.NewReader("This is more than 10 bytes of data")
		n, err := f.ReadFrom(reader)
		if err != ErrWriteOutOfBounds {
			t.Errorf("ReadFrom got err %v, want ErrWriteOutOfBounds", err)
		}
		if n != 10 {
			t.Errorf("ReadFrom read %d bytes, want 10", n)
		}
	})
}

func TestWriteTo(t *testing.T) {
	f, err := Open("testdata/binary.dat")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	n, err := f.WriteTo(&buf)
	if err != nil {
		t.Errorf("WriteTo failed: %v", err)
	}
	if n != int64(f.Len()) {
		t.Errorf("WriteTo wrote %d bytes, want %d", n, f.Len())
	}
	if buf.Len() != f.Len() {
		t.Errorf("buffer len = %d, want %d", buf.Len(), f.Len())
	}

	t.Run("with error", func(t *testing.T) {
		f, err := Open("testdata/hello.txt")
		if err != nil {
			t.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		// Writer that fails after some bytes
		failingWriter := &failingWriter{limit: 5}
		n, err := f.WriteTo(failingWriter)
		if err == nil {
			t.Error("WriteTo should fail with failing writer")
		}
		if n != 5 {
			t.Errorf("WriteTo wrote %d bytes, want 5", n)
		}
	})
}

func TestEmptyFile(t *testing.T) {
	f, err := Open("testdata/empty.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	if f.Len() != 0 {
		t.Errorf("Len() = %d, want 0", f.Len())
	}

	buf := make([]byte, 10)
	n, err := f.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("Read empty file: got n=%d, err=%v, want n=0, err=EOF", n, err)
	}

	n, err = f.ReadAt(buf, 0)
	if n != 0 || err != io.EOF {
		t.Errorf("ReadAt empty file: got n=%d, err=%v, want n=0, err=EOF", n, err)
	}
}

func TestInterfaceCompliance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "interface.txt")
	f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 100)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	// Test that MmapFile can be used where these interfaces are expected
	var _ io.Reader = f
	var _ io.Writer = f
	var _ io.Seeker = f
	var _ io.ReaderAt = f
	var _ io.WriterAt = f
	var _ io.Closer = f
	var _ io.ReaderFrom = f
	var _ io.WriterTo = f
	var _ io.StringWriter = f
	var _ io.ReadWriteSeeker = f
}
