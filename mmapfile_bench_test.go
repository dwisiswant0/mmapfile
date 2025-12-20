package mmapfile

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkRead(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		buf := make([]byte, 10)
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.Read(buf)
		}
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		buf := make([]byte, 10)
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.Read(buf)
		}
	})
}

func BenchmarkReadAt(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		buf := make([]byte, 10)
		b.ResetTimer()

		for b.Loop() {
			f.ReadAt(buf, 0)
		}
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		buf := make([]byte, 10)
		b.ResetTimer()

		for b.Loop() {
			f.ReadAt(buf, 0)
		}
	})
}

func BenchmarkReadAtParallel(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			buf := make([]byte, 10)
			for pb.Next() {
				f.ReadAt(buf, 0)
			}
		})
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			buf := make([]byte, 10)
			for pb.Next() {
				f.ReadAt(buf, 0)
			}
		})
	})
}

func BenchmarkWrite(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "bench_mmap.txt")
		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 1024*1024)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		data := []byte("Hello, mmap!")
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.Write(data)
		}
	})

	b.Run("os", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "bench_os.txt")
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		// Pre-allocate to match mmap behavior
		f.Truncate(1024 * 1024)

		data := []byte("Hello, mmap!")
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.Write(data)
		}
	})
}

func BenchmarkWriteAt(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "bench_mmap.txt")
		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 1024*1024)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		data := []byte("Hello, mmap!")
		b.ResetTimer()

		for b.Loop() {
			f.WriteAt(data, 0)
		}
	})

	b.Run("os", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "bench_os.txt")
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		f.Truncate(1024 * 1024)

		data := []byte("Hello, mmap!")
		b.ResetTimer()

		for b.Loop() {
			f.WriteAt(data, 0)
		}
	})
}

func BenchmarkSeek(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			f.Seek(int64(i%10), io.SeekStart)
		}
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			f.Seek(int64(i%10), io.SeekStart)
		}
	})
}

func BenchmarkBytes(b *testing.B) {
	f, err := Open("testdata/binary.dat")
	if err != nil {
		b.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	b.ResetTimer()
	for b.Loop() {
		_ = f.Bytes()
	}
}

func BenchmarkReadFrom(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "readfrom_mmap.txt")
		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 1024)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		data := []byte("Hello, ReadFrom benchmark!")
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.ReadFrom(bytes.NewReader(data))
		}
	})

	b.Run("os", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "readfrom_os.txt")
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		f.Truncate(1024)

		data := []byte("Hello, ReadFrom benchmark!")
		b.ResetTimer()

		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.ReadFrom(bytes.NewReader(data))
		}
	})
}

func BenchmarkWriteTo(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for b.Loop() {
			f.WriteTo(io.Discard)
		}
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for b.Loop() {
			f.Seek(0, io.SeekStart)
			f.WriteTo(io.Discard)
		}
	})
}

func BenchmarkStat(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		f, err := Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for b.Loop() {
			f.Stat()
		}
	})

	b.Run("os", func(b *testing.B) {
		f, err := os.Open("testdata/binary.dat")
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		defer f.Close()

		b.ResetTimer()
		for b.Loop() {
			f.Stat()
		}
	})
}

func BenchmarkSync(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "sync_mmap.txt")
		f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 1024)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		f.WriteString("data")
		b.ResetTimer()

		for b.Loop() {
			f.Sync()
		}
	})

	b.Run("os", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "sync_os.txt")
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			b.Fatalf("OpenFile failed: %v", err)
		}
		defer f.Close()

		f.WriteString("data")
		b.ResetTimer()

		for b.Loop() {
			f.Sync()
		}
	})
}

func BenchmarkClose(b *testing.B) {
	b.Run("mmap", func(b *testing.B) {
		dir := b.TempDir()
		b.ResetTimer()

		for i := 0; b.Loop(); i++ {
			path := filepath.Join(dir, "close_mmap.txt")
			f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, 1024)
			if err != nil {
				b.Fatalf("OpenFile failed: %v", err)
			}
			f.Close()
		}
	})

	b.Run("os", func(b *testing.B) {
		dir := b.TempDir()
		b.ResetTimer()

		for i := 0; b.Loop(); i++ {
			path := filepath.Join(dir, "close_os.txt")
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				b.Fatalf("OpenFile failed: %v", err)
			}
			f.Close()
		}
	})
}
