package mmapfile

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type byteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB byteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b byteSize) Human() string {
	if b == 0 {
		return "0B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	i := 0
	val := float64(b)
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	return fmt.Sprintf("%.0f%s", val, units[i])
}

var sizes = []byteSize{
	1 * KB,
	10 * KB,
	100 * KB,
	1 * MB,
	10 * MB,
	100 * MB,
	500 * MB,
	1 * GB,
}

func BenchmarkRead(b *testing.B) {
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			tempDir := b.TempDir()
			path := filepath.Join(tempDir, fmt.Sprintf("bench_read_%s.dat", sizeStr))

			f, err := os.Create(path)
			if err != nil {
				b.Fatalf("Create failed: %v", err)
			}

			// Fill with data
			buf := make([]byte, 1024)
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			for written := int64(0); written < sizeInt; {
				n, err := f.Write(buf[:min(1024, int(sizeInt-written))])
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				written += int64(n)
			}
			f.Close()

			b.Run("mmap", func(b *testing.B) {
				f, err := Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				buf := make([]byte, 4096)
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					for {
						n, err := f.Read(buf)
						if err == io.EOF {
							break
						}
						if err != nil {
							b.Fatalf("Read failed: %v", err)
						}
						if n == 0 {
							break
						}
					}
				}
			})

			b.Run("os", func(b *testing.B) {
				f, err := os.Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				buf := make([]byte, 4096)
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					for {
						n, err := f.Read(buf)
						if err == io.EOF {
							break
						}
						if err != nil {
							b.Fatalf("Read failed: %v", err)
						}
						if n == 0 {
							break
						}
					}
				}
			})
		})
	}
}

func BenchmarkReadAt(b *testing.B) {
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			tempDir := b.TempDir()
			path := filepath.Join(tempDir, fmt.Sprintf("bench_readat_%s.dat", sizeStr))

			f, err := os.Create(path)
			if err != nil {
				b.Fatalf("Create failed: %v", err)
			}

			// Fill with data
			buf := make([]byte, 1024)
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			for written := int64(0); written < sizeInt; {
				n, err := f.Write(buf[:min(1024, int(sizeInt-written))])
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				written += int64(n)
			}
			f.Close()

			b.Run("mmap", func(b *testing.B) {
				f, err := Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				buf := make([]byte, 4096)
				b.ResetTimer()

				for b.Loop() {
					for offset := int64(0); offset < sizeInt; offset += 4096 {
						n, err := f.ReadAt(buf, offset)
						if err != nil && err != io.EOF {
							b.Fatalf("ReadAt failed: %v", err)
						}
						if n == 0 && err != io.EOF {
							break
						}
					}
				}
			})

			b.Run("os", func(b *testing.B) {
				f, err := os.Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				buf := make([]byte, 4096)
				b.ResetTimer()

				for b.Loop() {
					for offset := int64(0); offset < sizeInt; offset += 4096 {
						n, err := f.ReadAt(buf, offset)
						if err != nil && err != io.EOF {
							b.Fatalf("ReadAt failed: %v", err)
						}
						if n == 0 && err != io.EOF {
							break
						}
					}
				}
			})
		})
	}
}

func BenchmarkReadAtParallel(b *testing.B) {
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			tempDir := b.TempDir()
			path := filepath.Join(tempDir, fmt.Sprintf("bench_readat_parallel_%s.dat", sizeStr))

			f, err := os.Create(path)
			if err != nil {
				b.Fatalf("Create failed: %v", err)
			}

			// Fill with data
			buf := make([]byte, 1024)
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			for written := int64(0); written < sizeInt; {
				n, err := f.Write(buf[:min(1024, int(sizeInt-written))])
				if err != nil {
					b.Fatalf("Write failed: %v", err)
				}
				written += int64(n)
			}
			f.Close()

			b.Run("mmap", func(b *testing.B) {
				f, err := Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					buf := make([]byte, 4096)
					for pb.Next() {
						f.ReadAt(buf, 0)
					}
				})
			})

			b.Run("os", func(b *testing.B) {
				f, err := os.Open(path)
				if err != nil {
					b.Fatalf("Open failed: %v", err)
				}
				defer f.Close()

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					buf := make([]byte, 4096)
					for pb.Next() {
						f.ReadAt(buf, 0)
					}
				})
			})
		})
	}
}

func BenchmarkWrite(b *testing.B) {
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			b.Run("mmap", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("bench_mmap_%s.txt", sizeStr))
				f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, sizeInt)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					f.Write(data)
				}
			})

			b.Run("os", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("bench_os_%s.txt", sizeStr))
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				// Pre-allocate to match mmap behavior
				f.Truncate(sizeInt)

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					f.Write(data)
				}
			})
		})
	}
}

func BenchmarkWriteAt(b *testing.B) {
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			b.Run("mmap", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("bench_mmap_%s.txt", sizeStr))
				f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, sizeInt)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.WriteAt(data, 0)
				}
			})

			b.Run("os", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("bench_os_%s.txt", sizeStr))
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				f.Truncate(sizeInt)

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.WriteAt(data, 0)
				}
			})
		})
	}
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
	for _, size := range sizes {
		sizeStr := byteSize(size).Human()
		sizeInt := int64(size)
		b.Run(sizeStr, func(b *testing.B) {
			b.Run("mmap", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("readfrom_mmap_%s.txt", sizeStr))
				f, err := OpenFile(path, os.O_RDWR|os.O_CREATE, 0644, sizeInt)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					f.ReadFrom(bytes.NewReader(data))
				}
			})

			b.Run("os", func(b *testing.B) {
				path := filepath.Join(b.TempDir(), fmt.Sprintf("readfrom_os_%s.txt", sizeStr))
				f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					b.Fatalf("OpenFile failed: %v", err)
				}
				defer f.Close()

				f.Truncate(sizeInt)

				data := make([]byte, sizeInt)
				for i := range data {
					data[i] = byte(i % 256)
				}
				b.ResetTimer()

				for b.Loop() {
					f.Seek(0, io.SeekStart)
					f.ReadFrom(bytes.NewReader(data))
				}
			})
		})
	}
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
