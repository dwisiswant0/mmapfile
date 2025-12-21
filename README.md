# mmapfile

[![tests](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml/badge.svg?branch=master)](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.dw1.io/mmapfile.svg)](https://pkg.go.dev/go.dw1.io/mmapfile)

An [`*os.File`](https://pkg.go.dev/os#File)-like type backed by memory-mapped I/O for Go.

**mmapfile** provides a drop-in replacement for `*os.File` in many contexts, offering significantly faster I/O operations by avoiding syscall overhead on every read/write.

## Features

- **[`*os.File`](https://pkg.go.dev/os#File)-compatible interface**: implements [`io.Reader`](https://pkg.go.dev/io#Reader), [`io.Writer`](https://pkg.go.dev/io#Writer), [`io.Seeker`](https://pkg.go.dev/io#Seeker), [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt), [`io.WriterAt`](https://pkg.go.dev/io#WriterAt), [`io.Closer`](https://pkg.go.dev/io#Closer), [`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom), [`io.WriterTo`](https://pkg.go.dev/io#WriterTo), and [`io.StringWriter`](https://pkg.go.dev/io#StringWriter)
- **Zero-copy reads**: direct access to file contents via `Bytes()` method
- **Cross-platform**: native mmap on Linux, Darwin, FreeBSD, OpenBSD, NetBSD, DragonFly, and Windows; fallback for other platforms
- **Thread-safe**: concurrent `ReadAt`/`WriteAt` operations are safe.
- **Zero allocations**: all I/O operations are allocation-free.

## Install

```bash
go get go.dw1.io/mmapfile
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "go.dw1.io/mmapfile"
)

func main() {
    // open a file for reading (like os.Open)
    f, err := mmapfile.Open("data.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // use it like *os.File.
    buf := make([]byte, 100)
    n, err := f.Read(buf)
    fmt.Printf("Read %d bytes: %s\n", n, buf[:n])

    // or get zero-copy access to the entire file.
    data := f.Bytes()
    fmt.Printf("File contents: %s\n", data)
}
```

## API

### Opening Files

```go
// open a file for reading (read-only)
f, err := mmapfile.Open("file.txt")

// open with flags (like os.OpenFile)
//
// size parameter is required for os.O_CREATE.
f, err := mmapfile.OpenFile("file.txt", os.O_RDWR|os.O_CREATE, 0644, 1024*1024)
```

### Supported Flags

| Flag | Description |
|------|-------------|
| [`os.O_RDONLY`](https://pkg.go.dev/os#O_RDONLY) | Open for reading only |
| [`os.O_RDWR`](https://pkg.go.dev/os#O_RDWR) | Open for reading and writing |
| [`os.O_CREATE`](https://pkg.go.dev/os#O_CREATE) | Create if doesn't exist (requires `size > 0`) |
| [`os.O_TRUNC`](https://pkg.go.dev/os#O_TRUNC) | Truncate to specified size |

> [!NOTE]
> [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND) is not supported — mmap files have fixed size.

### Methods

| Method | Description |
|--------|-------------|
| `Read([]byte)` | Read bytes, advancing cursor |
| `ReadAt([]byte, int64)` | Read at offset (cursor unchanged) |
| `Write([]byte)` | Write bytes, advancing cursor |
| `WriteAt([]byte, int64)` | Write at offset (cursor unchanged) |
| `WriteString(string)` | Write string |
| `Seek(int64, int)` | Set cursor position |
| `ReadFrom(io.Reader)` | Read from reader into file |
| `WriteTo(io.Writer)` | Write file contents to writer |
| `Close()` | Close and unmap the file |
| `Sync()` | Flush changes to disk |
| `Stat()` | Get file info |
| `Name()` | Get file name |
| `Len()` | Get file size |
| `Bytes()` | Get direct access to mapped memory ⚠️ |

### Zero-Copy Access

```go
// get direct access to the memory-mapped region
data := f.Bytes()

// WARNING: This slice is only valid until Close() is called.
// Modifying a read-only file's bytes will cause a segfault.
```

## Benchmarks

<details open>
  <summary><code>benchstat</code></summary>

  ```
  goos: linux
  goarch: amd64
  pkg: go.dw1.io/mmapfile
  cpu: AMD EPYC 7763 64-Core Processor                
                  │       os       │                 mmap                  │
                  │     sec/op     │    sec/op     vs base                 │
  Read-2              1138.00n ± 0%    24.06n ± 0%   -97.89% (p=0.000 n=10)
  ReadAt-2            639.800n ± 0%    9.367n ± 1%   -98.54% (p=0.000 n=10)
  ReadAtParallel-2     354.70n ± 1%    35.02n ± 4%   -90.13% (p=0.000 n=10)
  Write-2             1508.50n ± 0%    24.33n ± 0%   -98.39% (p=0.000 n=10)
  WriteAt-2          1024.500n ± 1%    9.664n ± 0%   -99.06% (p=0.000 n=10)
  Seek-2               357.10n ± 0%    11.84n ± 0%   -96.68% (p=0.000 n=10)
  ReadFrom-2          1683.00n ± 1%    61.16n ± 9%   -96.37% (p=0.000 n=10)
  WriteTo-2          1940.500n ± 0%    7.578n ± 0%   -99.61% (p=0.000 n=10)
  Stat-2                679.8n ± 0%   1606.0n ± 1%  +136.26% (p=0.000 n=10)
  Sync-2                849.8n ± 0%    910.4n ± 1%    +7.13% (p=0.000 n=10)
  Close-2               5.915µ ± 1%   12.665µ ± 0%  +114.13% (p=0.000 n=10)
  geomean               1.043µ         69.70n        -93.32%

                  │      os      │                  mmap                   │
                  │     B/op     │    B/op     vs base                     │
  Read-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Seek-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom-2         56.00 ± 0%     48.00 ± 0%   -14.29% (p=0.000 n=10)
  WriteTo-2          40.00 ± 0%      0.00 ± 0%  -100.00% (p=0.000 n=10)
  Stat-2             208.0 ± 0%     232.0 ± 0%   +11.54% (p=0.000 n=10)
  Sync-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2            216.0 ± 7%     536.0 ± 0%  +148.15% (p=0.000 n=10)
  geomean                       ²               ?                       ² ³
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ³ ratios must be >0 to compute geomean

                  │      os      │                  mmap                   │
                  │  allocs/op   │ allocs/op   vs base                     │
  Read-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Seek-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom-2         2.000 ± 0%     1.000 ± 0%   -50.00% (p=0.000 n=10)
  WriteTo-2          3.000 ± 0%     0.000 ± 0%  -100.00% (p=0.000 n=10)
  Stat-2             1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  Sync-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2            4.000 ± 0%     6.000 ± 0%   +50.00% (p=0.000 n=10)
  geomean                       ²               ?                       ² ³
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ³ ratios must be >0 to compute geomean
  ```
</details>

### Summary

| Operation | os.File (ns/op) | mmap (ns/op) | Improvement | Allocations |
|-----------|-----------------|--------------|-------------|-------------|
| `Read` | 1138 | 24 | **47x faster** | 0 → 0 |
| `ReadAt` | 640 | 9 | **68x faster** | 0 → 0 |
| `ReadAt` (parallel) | 355 | 35 | **10x faster** | 0 → 0 |
| `Write` | 1509 | 24 | **62x faster** | 0 → 0 |
| `WriteAt` | 1025 | 10 | **106x faster** | 0 → 0 |
| `Seek` | 357 | 12 | **30x faster** | 0 → 0 |
| `ReadFrom` | 1683 | 61 | **28x faster** | 2 → 1 |
| `WriteTo` | 1941 | 8 | **256x faster** | 3 → 0 |
| `Stat` | 680 | 1606 | 2.4x slower | 1 → 2 |
| `Sync` | 850 | 910 | 1.1x slower | 0 → 0 |
| `Close` | 5.9 µs | 12.7 µs | 2.1x slower | 4 → 6 |
| **Geomean** | **1043 ns** | **70 ns** | **15x faster** | — |

**Key takeaway:** mmap eliminates syscall overhead, delivering **10-200x++ speedups** for I/O operations. Once mapped, reads and writes are simple memory copies with zero allocations.

Run benchmarks yourself:

```bash
make bench
make -C benchdata/
```

## When to Use `mmapfile`

### Good Use Cases

- **Large file random access**: databases, indexes, binary file parsing.
- **Read-heavy workloads**: config files, static data, lookup tables.
- **Memory-mapped databases**: fixed-size arenas, append-only logs.
- **Shared memory IPC**: multiple processes reading the same file.
- **High-frequency I/O**: avoiding syscall overhead.

### When to Stick with `os.File`

- **Growing files**: mmap requires fixed size upfront.
- **Small files with single read**: mmap setup overhead not worth it.
- **Streaming data**: network, pipes, stdin.
- **Infrequent access**: syscall overhead is negligible.

## Limitations

1. **Fixed size**: Files cannot grow after opening. Use `size` parameter with [`os.O_CREATE`](https://pkg.go.dev/os#O_CREATE).
2. **No Truncate**: Changing file size requires closing and reopening.
3. **No [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND)**: Appending is not supported.
4. **Cursor operations are slower than positional**: Use `ReadAt`/`WriteAt` for best performance.

## Platform Support

| Platform | Implementation |
|----------|----------------|
| Linux | `mmap`/`munmap`/`msync` |
| Darwin (macOS) | `mmap`/`munmap`/`msync` |
| FreeBSD, OpenBSD, NetBSD, DragonFly | `mmap`/`munmap`/`msync` |
| Windows | `CreateFileMapping`/`MapViewOfFile`/`FlushViewOfFile` |
| Other | Fallback (reads file into memory) |

## Thread Safety

- `ReadAt` and `WriteAt` are safe for concurrent use.
- `Read`, `Write`, and `Seek` share a cursor, concurrent use will interleave unpredictably.
- `Close` should not be called concurrently with other operations.

## License

**mmapfile** is released with ♡ by [**@dwisiswant0**](https://github.com/dwisiswant0) under the Apache 2.0 license. See [LICENSE](/LICENSE).

## Acknowledgments

Inspired by [`golang.org/x/exp/mmap`](https://pkg.go.dev/golang.org/x/exp/mmap).
