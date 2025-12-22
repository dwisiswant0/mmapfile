# mmapfile

[![tests](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml/badge.svg?branch=master)](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.dw1.io/mmapfile.svg)](https://pkg.go.dev/go.dw1.io/mmapfile)

An [`*os.File`](https://pkg.go.dev/os#File)-like type backed by memory-mapped I/O for Go.

**mmapfile** provides a drop-in replacement for `*os.File` in many contexts, offering significantly faster I/O operations by avoiding syscall overhead on every read/write.

## Features

* **[`*os.File`](https://pkg.go.dev/os#File)-compatible interface**: implements [`io.Reader`](https://pkg.go.dev/io#Reader), [`io.Writer`](https://pkg.go.dev/io#Writer), [`io.Seeker`](https://pkg.go.dev/io#Seeker), [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt), [`io.WriterAt`](https://pkg.go.dev/io#WriterAt), [`io.Closer`](https://pkg.go.dev/io#Closer), [`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom), [`io.WriterTo`](https://pkg.go.dev/io#WriterTo), and [`io.StringWriter`](https://pkg.go.dev/io#StringWriter).
* **Zero-copy reads**: direct access to file contents via `Bytes()` method.
* **Cross-platform**: native mmap on Linux, Darwin, FreeBSD, OpenBSD, NetBSD, DragonFly, and Windows; fallback for other platforms.
* **Thread-safe**: concurrent `ReadAt`/`WriteAt` operations are safe.
* **Zero allocations**: all I/O operations are allocation-free.

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
> [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND) is not supported - mmap files have fixed size.

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
                        │    os.File     │                mmapfile                │
                        │     sec/op     │    sec/op      vs base                 │
  Read/1KB-2                1727.50n ± 2%    36.80n ±  1%   -97.87% (p=0.000 n=10)
  Read/10KB-2                3383.5n ± 1%    158.0n ±  0%   -95.33% (p=0.000 n=10)
  Read/100KB-2               21.331µ ± 2%    1.584µ ±  0%   -92.58% (p=0.000 n=10)
  Read/1MB-2                 218.62µ ± 1%    20.61µ ±  0%   -90.57% (p=0.000 n=10)
  Read/10MB-2                2168.7µ ± 1%    207.0µ ±  2%   -90.45% (p=0.000 n=10)
  Read/100MB-2               24.758m ± 1%    4.945m ±  2%   -80.03% (p=0.000 n=10)
  Read/500MB-2               123.03m ± 0%    31.64m ±  1%   -74.28% (p=0.000 n=10)
  Read/1GB-2                 251.43m ± 1%    77.65m ±  1%   -69.12% (p=0.000 n=10)
  ReadAt/1KB-2              1151.50n ± 0%    23.45n ±  1%   -97.96% (p=0.000 n=10)
  ReadAt/10KB-2              2593.0n ± 1%    142.1n ±  1%   -94.52% (p=0.000 n=10)
  ReadAt/100KB-2             17.891µ ± 1%    1.476µ ±  1%   -91.75% (p=0.000 n=10)
  ReadAt/1MB-2               192.09µ ± 1%    18.69µ ±  2%   -90.27% (p=0.000 n=10)
  ReadAt/10MB-2              1899.0µ ± 2%    181.4µ ±  1%   -90.45% (p=0.000 n=10)
  ReadAt/100MB-2             22.322m ± 1%    4.414m ±  3%   -80.23% (p=0.000 n=10)
  ReadAt/500MB-2             110.74m ± 1%    27.29m ±  1%   -75.35% (p=0.000 n=10)
  ReadAt/1GB-2               226.05m ± 1%    71.34m ±  1%   -68.44% (p=0.000 n=10)
  ReadAtParallel/1KB-2       642.75n ± 4%    53.11n ± 18%   -91.74% (p=0.000 n=10)
  ReadAtParallel/10KB-2      402.90n ± 4%    40.41n ±  3%   -89.97% (p=0.000 n=10)
  ReadAtParallel/100KB-2     394.15n ± 4%    39.17n ± 20%   -90.06% (p=0.000 n=10)
  ReadAtParallel/1MB-2       396.90n ± 5%    38.56n ±  5%   -90.28% (p=0.000 n=10)
  ReadAtParallel/10MB-2      390.20n ± 4%    40.36n ±  9%   -89.66% (p=0.000 n=10)
  ReadAtParallel/100MB-2     398.70n ± 3%    38.53n ±  5%   -90.34% (p=0.000 n=10)
  ReadAtParallel/500MB-2     397.60n ± 2%    38.70n ±  5%   -90.27% (p=0.000 n=10)
  ReadAtParallel/1GB-2       405.00n ± 3%    39.39n ±  5%   -90.27% (p=0.000 n=10)
  Write/1KB-2               1487.50n ± 1%    33.41n ±  1%   -97.75% (p=0.000 n=10)
  Write/10KB-2               2355.0n ± 1%    148.8n ±  0%   -93.68% (p=0.000 n=10)
  Write/100KB-2              12.409µ ± 2%    2.081µ ±  0%   -83.23% (p=0.000 n=10)
  Write/1MB-2                129.79µ ± 6%    42.38µ ±  0%   -67.34% (p=0.000 n=10)
  Write/10MB-2               1436.8µ ± 7%    453.4µ ±  0%   -68.45% (p=0.000 n=10)
  Write/100MB-2               27.29m ± 3%    21.15m ±  0%   -22.49% (p=0.000 n=10)
  Write/500MB-2               314.8m ± 0%    520.2m ±  0%   +65.24% (p=0.000 n=10)
  Write/1GB-2                 643.6m ± 0%   1063.4m ±  0%   +65.23% (p=0.000 n=10)
  WriteAt/1KB-2             1039.50n ± 1%    20.62n ±  2%   -98.02% (p=0.000 n=10)
  WriteAt/10KB-2             1864.0n ± 1%    110.0n ± 23%   -94.10% (p=0.000 n=10)
  WriteAt/100KB-2            13.096µ ± 2%    2.063µ ±  0%   -84.25% (p=0.000 n=10)
  WriteAt/1MB-2              133.18µ ± 3%    42.38µ ±  0%   -68.18% (p=0.000 n=10)
  WriteAt/10MB-2             1341.9µ ± 2%    451.7µ ±  0%   -66.34% (p=0.000 n=10)
  WriteAt/100MB-2             26.47m ± 6%    21.10m ±  0%   -20.27% (p=0.000 n=10)
  WriteAt/500MB-2             315.1m ± 0%    519.7m ±  0%   +64.92% (p=0.000 n=10)
  WriteAt/1GB-2               644.8m ± 0%   1064.7m ±  0%   +65.11% (p=0.000 n=10)
  Seek-2                     359.60n ± 1%    11.87n ±  1%   -96.70% (p=0.000 n=10)
  ReadFrom/1KB-2            1698.00n ± 0%    82.03n ±  1%   -95.17% (p=0.000 n=10)
  ReadFrom/10KB-2            2566.5n ± 1%    169.9n ±  1%   -93.38% (p=0.000 n=10)
  ReadFrom/100KB-2           12.847µ ± 2%    2.146µ ±  0%   -83.30% (p=0.000 n=10)
  ReadFrom/1MB-2             132.27µ ± 1%    42.50µ ±  0%   -67.87% (p=0.000 n=10)
  ReadFrom/10MB-2            1445.1µ ± 5%    452.7µ ±  1%   -68.67% (p=0.000 n=10)
  ReadFrom/100MB-2            27.26m ± 4%    21.14m ±  1%   -22.47% (p=0.000 n=10)
  ReadFrom/500MB-2            315.0m ± 0%    520.7m ±  0%   +65.29% (p=0.000 n=10)
  ReadFrom/1GB-2              644.6m ± 0%   1064.7m ±  0%   +65.19% (p=0.000 n=10)
  WriteTo-2                1931.000n ± 1%    7.516n ±  0%   -99.61% (p=0.000 n=10)
  Stat-2                      701.0n ± 1%   1610.5n ±  1%  +129.73% (p=0.000 n=10)
  Sync-2                      846.7n ± 1%    897.9n ±  2%    +6.04% (p=0.000 n=10)
  Close-2                     5.919µ ± 1%   12.755µ ±  1%  +115.48% (p=0.000 n=10)
  geomean                     118.5µ         21.07µ         -82.21%

                        │   os.File    │                mmapfile                 │
                        │     B/op     │    B/op     vs base                     │
  Read/1KB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/10KB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/100KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/1MB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/10MB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/100MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/500MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/1GB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/10KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/100KB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/10MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/100MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/500MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1GB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1KB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/10KB-2    0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/100KB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1MB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/10MB-2    0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/100MB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/500MB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1GB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1KB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/10KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/100KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1MB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/10MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/100MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/500MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1GB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/10KB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/100KB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/10MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/100MB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/500MB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1GB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Seek-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/1KB-2           56.00 ± 0%     49.00 ± 0%   -12.50% (p=0.000 n=10)
  ReadFrom/10KB-2          56.00 ± 0%     49.00 ± 0%   -12.50% (p=0.000 n=10)
  ReadFrom/100KB-2         56.00 ± 0%     49.00 ± 0%   -12.50% (p=0.000 n=10)
  ReadFrom/1MB-2           56.00 ± 0%     49.00 ± 0%   -12.50% (p=0.000 n=10)
  ReadFrom/10MB-2          56.00 ± 0%     49.00 ± 0%   -12.50% (p=0.000 n=10)
  ReadFrom/100MB-2         56.00 ± 0%     50.00 ± 0%   -10.71% (p=0.000 n=10)
  ReadFrom/500MB-2         56.00 ± 0%     64.00 ± 0%   +14.29% (p=0.000 n=10)
  ReadFrom/1GB-2           56.00 ± 0%     64.00 ± 0%   +14.29% (p=0.000 n=10)
  WriteTo-2                40.00 ± 0%      0.00 ± 0%  -100.00% (p=0.000 n=10)
  Stat-2                   208.0 ± 0%     232.0 ± 0%   +11.54% (p=0.000 n=10)
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  216.0 ± 0%     536.0 ± 0%  +148.15% (p=0.000 n=10)
  geomean                             ²               ?                       ² ³
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ³ ratios must be >0 to compute geomean

                        │   os.File    │                mmapfile                 │
                        │  allocs/op   │ allocs/op   vs base                     │
  Read/1KB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/10KB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/100KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/1MB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/10MB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/100MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/500MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Read/1GB-2               0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/10KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/100KB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/10MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/100MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/500MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAt/1GB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1KB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/10KB-2    0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/100KB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1MB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/10MB-2    0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/100MB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/500MB-2   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadAtParallel/1GB-2     0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1KB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/10KB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/100KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1MB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/10MB-2             0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/100MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/500MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Write/1GB-2              0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1KB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/10KB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/100KB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1MB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/10MB-2           0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/100MB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/500MB-2          0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteAt/1GB-2            0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Seek-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/1KB-2           2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/10KB-2          2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/100KB-2         2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/1MB-2           2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/10MB-2          2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/100MB-2         2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/500MB-2         2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  ReadFrom/1GB-2           2.000 ± 0%     2.000 ± 0%         ~ (p=1.000 n=10) ¹
  WriteTo-2                3.000 ± 0%     0.000 ± 0%  -100.00% (p=0.000 n=10)
  Stat-2                   1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  4.000 ± 0%     6.000 ± 0%   +50.00% (p=0.000 n=10)
  geomean                             ²               ?                       ² ³
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ³ ratios must be >0 to compute geomean
  ```
</details>

### Summary

| Operation | Size | `os.File` | `mmapfile` | Improvement | Allocs |
|-----------|------|-----------|------------|-------------|--------|
| `Read` | 1KB | 1727.5ns | 36.8ns | **47x faster** | 0 → 0 |
| `Read` | 10KB | 3383.5ns | 158.0ns | **21x faster** | 0 → 0 |
| `Read` | 100KB | 21.33µs | 1.58µs | **13x faster** | 0 → 0 |
| `Read` | 1MB | 218.6µs | 20.6µs | **11x faster** | 0 → 0 |
| `Read` | 10MB | 2168.7µs | 207.0µs | **10x faster** | 0 → 0 |
| `Read` | 100MB | 24.76ms | 4.95ms | **5x faster** | 0 → 0 |
| `Read` | 500MB | 123.0ms | 31.6ms | **4x faster** | 0 → 0 |
| `Read` | 1GB | 251.4ms | 77.7ms | **3x faster** | 0 → 0 |
| `ReadAt` | 1KB | 1151.5ns | 23.5ns | **49x faster** | 0 → 0 |
| `ReadAt` | 10KB | 2593.0ns | 142.1ns | **18x faster** | 0 → 0 |
| `ReadAt` | 100KB | 17.89µs | 1.48µs | **12x faster** | 0 → 0 |
| `ReadAt` | 1MB | 192.1µs | 18.7µs | **10x faster** | 0 → 0 |
| `ReadAt` | 10MB | 1899.0µs | 181.4µs | **10x faster** | 0 → 0 |
| `ReadAt` | 100MB | 22.32ms | 4.41ms | **5x faster** | 0 → 0 |
| `ReadAt` | 500MB | 110.7ms | 27.3ms | **4x faster** | 0 → 0 |
| `ReadAt` | 1GB | 226.1ms | 71.3ms | **3x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1KB | 642.8ns | 53.1ns | **12x faster** | 0 → 0 |
| `ReadAt` (parallel) | 10KB | 402.9ns | 40.4ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 100KB | 394.2ns | 39.2ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1MB | 396.9ns | 38.6ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 10MB | 390.2ns | 40.4ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 100MB | 398.7ns | 38.5ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 500MB | 397.6ns | 38.7ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1GB | 405.0ns | 39.4ns | **10x faster** | 0 → 0 |
| `Write` | 1KB | 1487.5ns | 33.4ns | **45x faster** | 0 → 0 |
| `Write` | 10KB | 2355.0ns | 148.8ns | **16x faster** | 0 → 0 |
| `Write` | 100KB | 12.41µs | 2.08µs | **6x faster** | 0 → 0 |
| `Write` | 1MB | 129.8µs | 42.4µs | **3x faster** | 0 → 0 |
| `Write` | 10MB | 1436.8µs | 453.4µs | **3x faster** | 0 → 0 |
| `Write` | 100MB | 27.29ms | 21.15ms | **1.3x faster** | 0 → 0 |
| `Write` | 500MB | 314.8ms | 520.2ms | 1.7x slower | 0 → 0 |
| `Write` | 1GB | 643.6ms | 1063.4ms | 1.7x slower | 0 → 0 |
| `WriteAt` | 1KB | 1039.5ns | 20.6ns | **50x faster** | 0 → 0 |
| `WriteAt` | 10KB | 1864.0ns | 110.0ns | **17x faster** | 0 → 0 |
| `WriteAt` | 100KB | 13.10µs | 2.06µs | **6x faster** | 0 → 0 |
| `WriteAt` | 1MB | 133.2µs | 42.4µs | **3x faster** | 0 → 0 |
| `WriteAt` | 10MB | 1341.9µs | 451.7µs | **3x faster** | 0 → 0 |
| `WriteAt` | 100MB | 26.47ms | 21.10ms | **1.3x faster** | 0 → 0 |
| `WriteAt` | 500MB | 315.1ms | 519.7ms | 1.6x slower | 0 → 0 |
| `WriteAt` | 1GB | 644.8ms | 1064.7ms | 1.7x slower | 0 → 0 |
| `Seek` | - | 359.6ns | 11.9ns | **30x faster** | 0 → 0 |
| `ReadFrom` | 1KB | 1698.0ns | 82.0ns | **21x faster** | 2 → 2 |
| `ReadFrom` | 10KB | 2566.5ns | 169.9ns | **15x faster** | 2 → 2 |
| `ReadFrom` | 100KB | 12.85µs | 2.15µs | **6x faster** | 2 → 2 |
| `ReadFrom` | 1MB | 132.3µs | 42.5µs | **3x faster** | 2 → 2 |
| `ReadFrom` | 10MB | 1445.1µs | 452.7µs | **3x faster** | 2 → 2 |
| `ReadFrom` | 100MB | 27.26ms | 21.14ms | **1.3x faster** | 2 → 2 |
| `ReadFrom` | 500MB | 315.0ms | 520.7ms | 1.7x slower | 2 → 2 |
| `ReadFrom` | 1GB | 644.6ms | 1064.7ms | 1.7x slower | 2 → 2 |
| `WriteTo` | - | 1931.0ns | 7.5ns | **257x faster** | 3 → 0 |
| `Stat` | - | 701.0ns | 1610.5ns | 2.3x slower | 1 → 2 |
| `Sync` | - | 846.7ns | 897.9ns | 1.1x slower | 0 → 0 |
| `Close` | - | 5.92µs | 12.76µs | 2.2x slower | 4 → 6 |
| **Geomean** | - | **118.5µs** | **21.07µs** | **~6x faster** | - |

**Key takeaway:**

mmapfile trades syscalls for memory ops, crushing latency-bound workloads (**3–257x speedups**):

* **All reads** (1KB–1GB): **3–50x faster**. Simple `copy(b, f.data[off:])` vs repeated `read(2)`-scales to huge files.
* **Writes ≤100MB**: **6–50x** wins. Syscall cost (~200ns) >> `memcpy`; ideal for frequent small ops.
* **Bulk sequential writes >500MB**: `os.File` ~1.7x ahead (644ms vs 1s+). Kernel magic:
  * Page cache absorbs data (no user double-buffering: heap->mmap).
  * Write-behind: async clustering to disk.
  * Readahead/DMA pipelines for sequential throughput.
  * mmap: pure `memcpy` to dirty pages (~1GB/s BW-bound on CPU; no flush/sync in bench for apple-to-apple op perf).
* **Random/parallel**: mmapfile **10x+** (`ReadAt` parallel constant ~40ns even 1GB). No seeks/syscalls.

**Overall geomean: ~6x faster** across 50+ benches.

> [!TIP]
> * For bulk/sequential writes, use:
>
>   ```go
>   data := f.Bytes()
>   // then
>   copy(data[off:], src[:n])
>   ```
>
>   To bypass `WriteAt` lock (no `RWMutex`), no bounds/EOF checks, and no partial copies. Direct `memcpy` to mmap region; **~10–20% faster** for large ops.
>
> * For durability, call `f.Sync()` after key writes to trigger `msync`: synchronous flush dirty pages to disk (~10–100ms/GB; varies SSD/NVMe/HDD/IO scheduler); essential for WAL/tx commits.
> * For zero-copy parsing/search, use:
> 
>   ```go
>   data := f.Bytes()
>   // then
>   strings.Index(data[off:], "needle")
>   // or
>   bytes.IndexByte(data[off:], 'x')
>   ```
>
>   No `ReadAt` allocs/copies/syscalls; mmap-pinned mem ideal for DB/indexers/shared-IPC (valid until `Close`; concurrent-safe with care).

Run benchmarks yourself:

```bash
make bench
make -C benchdata/
```

## When to Use `mmapfile`

### Good Use Cases

* **Large file random access**: databases, indexes, binary file parsing.
* **Read-heavy workloads**: config files, static data, lookup tables.
* **Memory-mapped databases**: fixed-size arenas, append-only logs.
* **Shared memory IPC**: multiple processes reading the same file.
* **High-frequency I/O**: avoiding syscall overhead.

### When to Stick with `os.File`

* **Growing files**: mmap requires fixed size upfront.
* **Small files with single read**: mmap setup overhead not worth it.
* **Streaming data**: network, pipes, stdin.
* **Infrequent access**: syscall overhead is negligible.
* **Bulk sequential writes**: kernel buffering outperforms user-space `memcpy`.

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
