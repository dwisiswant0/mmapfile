# mmapfile

[![tests](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml/badge.svg?branch=master)](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.dw1.io/mmapfile.svg)](https://pkg.go.dev/go.dw1.io/mmapfile)

An [`*os.File`](https://pkg.go.dev/os#File)-like type backed by memory-mapped I/O for Go.

**mmapfile** provides a drop-in replacement for `*os.File` in many contexts, offering significantly faster I/O operations by avoiding syscall overhead on every read/write.

## Features

* **[`*os.File`](https://pkg.go.dev/os#File)-compatible interface**: implements [`io.Reader`](https://pkg.go.dev/io#Reader), [`io.Writer`](https://pkg.go.dev/io#Writer), [`io.Seeker`](https://pkg.go.dev/io#Seeker), [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt), [`io.WriterAt`](https://pkg.go.dev/io#WriterAt), [`io.Closer`](https://pkg.go.dev/io#Closer), [`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom), [`io.WriterTo`](https://pkg.go.dev/io#WriterTo), and [`io.StringWriter`](https://pkg.go.dev/io#StringWriter).
* **Zero-copy reads**: direct access to file contents via [`Bytes()`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Bytes) method.
* **Cross-platform**: native mmap on Linux, Darwin, FreeBSD, OpenBSD, NetBSD, DragonFly, and Windows; fallback for other platforms.
* **Thread-safe**: concurrent [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt)/[`WriteAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.WriteAt) operations are safe.
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
// size is used when creating a new non-empty file or truncating a file.
f, err := mmapfile.OpenFile("file.txt", os.O_RDWR|os.O_CREATE, 0644, 1024*1024)
```

### Supported Flags

| Flag | Description |
|------|-------------|
| [`os.O_RDONLY`](https://pkg.go.dev/os#O_RDONLY) | Open for reading only |
| [`os.O_RDWR`](https://pkg.go.dev/os#O_RDWR) | Open for reading and writing |
| [`os.O_CREATE`](https://pkg.go.dev/os#O_CREATE) | Create if the file doesn't exist |
| [`os.O_EXCL`](https://pkg.go.dev/os#O_EXCL) | Require creating a new file when used with `os.O_CREATE` |
| [`os.O_SYNC`](https://pkg.go.dev/os#O_SYNC) | Open for synchronous I/O |
| [`os.O_TRUNC`](https://pkg.go.dev/os#O_TRUNC) | Truncate to specified size |

> [!NOTE]
> [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND) and [`os.O_WRONLY`](https://pkg.go.dev/os#O_WRONLY) are not supported.
> Truncating or creating a non-empty file requires [`os.O_RDWR`](https://pkg.go.dev/os#O_RDWR).

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
| `Close()` | Flush fallback mappings, unmap, and close the file |
| `Sync()` | Flush mapped changes to disk |
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
  cpu: AMD EPYC 9V74 80-Core Processor
                         │    os.File     │                mmapfile                │
                         │     sec/op     │    sec/op      vs base                 │
  Read/1KB-2                2226.00n ± 4%    43.54n ±  1%   -98.04% (p=0.000 n=10)
  Read/10KB-2                3926.5n ± 3%    186.3n ±  8%   -95.26% (p=0.000 n=10)
  Read/100KB-2               24.082µ ± 3%    1.708µ ±  9%   -92.91% (p=0.000 n=10)
  Read/1MB-2                 243.20µ ± 3%    20.17µ ±  4%   -91.71% (p=0.000 n=10)
  Read/10MB-2                2504.7µ ± 3%    211.4µ ±  2%   -91.56% (p=0.000 n=10)
  Read/100MB-2               25.476m ± 2%    3.250m ±  1%   -87.24% (p=0.000 n=10)
  Read/500MB-2               128.31m ± 5%    23.23m ± 12%   -81.90% (p=0.000 n=10)
  Read/1GB-2                 257.93m ± 1%    70.62m ±  1%   -72.62% (p=0.000 n=10)
  ReadAt/1KB-2              1513.00n ± 2%    25.89n ±  1%   -98.29% (p=0.000 n=10)
  ReadAt/10KB-2              3205.5n ± 5%    153.0n ±  0%   -95.23% (p=0.000 n=10)
  ReadAt/100KB-2             20.571µ ± 2%    1.694µ ±  1%   -91.77% (p=0.000 n=10)
  ReadAt/1MB-2               218.53µ ± 3%    17.61µ ±  0%   -91.94% (p=0.000 n=10)
  ReadAt/10MB-2              2274.1µ ± 3%    193.2µ ±  0%   -91.51% (p=0.000 n=10)
  ReadAt/100MB-2             24.726m ± 5%    3.127m ±  2%   -87.35% (p=0.000 n=10)
  ReadAt/500MB-2             117.83m ± 4%    23.15m ±  1%   -80.35% (p=0.000 n=10)
  ReadAt/1GB-2               239.03m ± 3%    70.65m ±  0%   -70.44% (p=0.000 n=10)
  ReadAtParallel/1KB-2       805.25n ± 3%    36.72n ±  1%   -95.44% (p=0.000 n=10)
  ReadAtParallel/10KB-2      445.45n ± 3%    42.57n ±  2%   -90.44% (p=0.000 n=10)
  ReadAtParallel/100KB-2     450.00n ± 2%    40.25n ±  0%   -91.05% (p=0.000 n=10)
  ReadAtParallel/1MB-2       452.55n ± 1%    40.21n ±  0%   -91.11% (p=0.000 n=10)
  ReadAtParallel/10MB-2      447.75n ± 3%    40.20n ±  0%   -91.02% (p=0.000 n=10)
  ReadAtParallel/100MB-2     448.40n ± 2%    40.19n ±  0%   -91.04% (p=0.000 n=10)
  ReadAtParallel/500MB-2     444.15n ± 5%    40.31n ±  0%   -90.92% (p=0.000 n=10)
  ReadAtParallel/1GB-2       450.55n ± 3%    40.22n ±  0%   -91.07% (p=0.000 n=10)
  Write/1KB-2               1761.00n ± 2%    36.56n ±  3%   -97.92% (p=0.000 n=10)
  Write/10KB-2               2258.0n ± 3%    135.5n ±  0%   -94.00% (p=0.000 n=10)
  Write/100KB-2               5.209µ ± 1%    2.343µ ±  0%   -55.02% (p=0.000 n=10)
  Write/1MB-2                 39.03µ ± 2%    37.87µ ±  1%    -2.96% (p=0.000 n=10)
  Write/10MB-2                388.1µ ± 4%    399.7µ ±  0%         ~ (p=0.075 n=10)
  Write/100MB-2               8.668m ± 2%   12.768m ±  1%   +47.29% (p=0.000 n=10)
  Write/500MB-2               73.41m ± 1%   434.73m ±  1%  +492.17% (p=0.000 n=10)
  Write/1GB-2                 220.8m ± 0%    883.7m ±  1%  +300.23% (p=0.000 n=10)
  WriteAt/1KB-2             1186.50n ± 6%    25.10n ±  0%   -97.88% (p=0.000 n=10)
  WriteAt/10KB-2             1673.0n ± 2%    126.2n ±  0%   -92.45% (p=0.000 n=10)
  WriteAt/100KB-2             4.633µ ± 1%    2.336µ ±  0%   -49.57% (p=0.000 n=10)
  WriteAt/1MB-2               38.86µ ± 1%    38.23µ ±  2%    -1.61% (p=0.015 n=10)
  WriteAt/10MB-2              388.8µ ± 2%    400.7µ ±  1%    +3.07% (p=0.000 n=10)
  WriteAt/100MB-2             8.257m ± 9%   12.773m ±  1%   +54.69% (p=0.000 n=10)
  WriteAt/500MB-2             72.72m ± 1%   436.63m ±  2%  +500.44% (p=0.000 n=10)
  WriteAt/1GB-2               221.9m ± 6%    889.8m ±  1%  +301.01% (p=0.000 n=10)
  Seek-2                     603.25n ± 4%    12.38n ±  0%   -97.95% (p=0.000 n=10)
  ReadFrom/1KB-2            2036.00n ± 2%    80.14n ±  1%   -96.06% (p=0.000 n=10)
  ReadFrom/10KB-2            2510.0n ± 4%    179.5n ±  2%   -92.85% (p=0.000 n=10)
  ReadFrom/100KB-2            5.519µ ± 1%    2.406µ ±  1%   -56.40% (p=0.000 n=10)
  ReadFrom/1MB-2              39.15µ ± 1%    38.86µ ±  0%    -0.74% (p=0.000 n=10)
  ReadFrom/10MB-2             384.7µ ± 1%    405.9µ ±  1%    +5.53% (p=0.000 n=10)
  ReadFrom/100MB-2            8.323m ± 4%   12.865m ±  2%   +54.57% (p=0.000 n=10)
  ReadFrom/500MB-2            73.45m ± 3%   432.96m ±  2%  +489.48% (p=0.000 n=10)
  ReadFrom/1GB-2              220.7m ± 1%    884.0m ±  0%  +300.59% (p=0.000 n=10)
  WriteTo-2                2266.000n ± 2%    9.766n ±  1%   -99.57% (p=0.000 n=10)
  Stat-2                      1.021µ ± 1%    1.041µ ±  2%    +1.96% (p=0.000 n=10)
  Sync-2                      1.190µ ± 3%    2.228µ ±  3%   +87.23% (p=0.000 n=10)
  Close-2                     6.725µ ± 3%   13.922µ ±  2%  +107.03% (p=0.000 n=10)
  geomean                     86.86µ         20.12µ         -76.84%

                         │   os.File    │               mmapfile                │
                         │     B/op     │    B/op     vs base                   │
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
  ReadFrom/1KB-2           48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/10KB-2          48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/100KB-2         48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/1MB-2           48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/10MB-2          48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/100MB-2         48.00 ± 0%     49.00 ± 0%    +2.08% (p=0.000 n=10)
  ReadFrom/500MB-2         48.00 ± 0%     64.00 ± 0%   +33.33% (p=0.000 n=10)
  ReadFrom/1GB-2           48.00 ± 0%     64.00 ± 0%   +33.33% (p=0.000 n=10)
  WriteTo-2                0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Stat-2                   208.0 ± 0%     208.0 ± 0%         ~ (p=1.000 n=10) ¹
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  216.0 ± 0%     536.0 ± 0%  +148.15% (p=0.000 n=10)
  geomean                             ²                 +3.08%                ²
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean

                         │   os.File    │               mmapfile                │
                         │  allocs/op   │ allocs/op   vs base                   │
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
  ReadFrom/1KB-2           1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/10KB-2          1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/100KB-2         1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/1MB-2           1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/10MB-2          1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/100MB-2         1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/500MB-2         1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  ReadFrom/1GB-2           1.000 ± 0%     2.000 ± 0%  +100.00% (p=0.000 n=10)
  WriteTo-2                0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Stat-2                   1.000 ± 0%     1.000 ± 0%         ~ (p=1.000 n=10) ¹
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  4.000 ± 0%     6.000 ± 0%   +50.00% (p=0.000 n=10)
  geomean                             ²                +11.88%                ²
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ```
</details>

### Summary

| Operation | Size | `os.File` | `mmapfile` | Result | Allocs |
|-----------|------|-----------|------------|--------|--------|
| `Read` | 1KB | 2226ns | 43.5ns | **51x faster** | 0 -> 0 |
| `Read` | 10KB | 3927ns | 186ns | **21x faster** | 0 -> 0 |
| `Read` | 100KB | 24.08µs | 1.71µs | **14x faster** | 0 -> 0 |
| `Read` | 1MB | 243µs | 20.2µs | **12x faster** | 0 -> 0 |
| `Read` | 10MB | 2505µs | 211µs | **12x faster** | 0 -> 0 |
| `Read` | 100MB | 25.48ms | 3.25ms | **7.8x faster** | 0 -> 0 |
| `Read` | 500MB | 128ms | 23.2ms | **5.5x faster** | 0 -> 0 |
| `Read` | 1GB | 258ms | 70.6ms | **3.7x faster** | 0 -> 0 |
| `ReadAt` | 1KB | 1513ns | 25.9ns | **58x faster** | 0 -> 0 |
| `ReadAt` | 10KB | 3206ns | 153ns | **21x faster** | 0 -> 0 |
| `ReadAt` | 100KB | 20.57µs | 1.69µs | **12x faster** | 0 -> 0 |
| `ReadAt` | 1MB | 219µs | 17.6µs | **12x faster** | 0 -> 0 |
| `ReadAt` | 10MB | 2274µs | 193µs | **12x faster** | 0 -> 0 |
| `ReadAt` | 100MB | 24.73ms | 3.13ms | **7.9x faster** | 0 -> 0 |
| `ReadAt` | 500MB | 118ms | 23.2ms | **5.1x faster** | 0 -> 0 |
| `ReadAt` | 1GB | 239ms | 70.7ms | **3.4x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 1KB | 805ns | 36.7ns | **22x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 10KB | 445ns | 42.6ns | **10x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 100KB | 450ns | 40.3ns | **11x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 1MB | 453ns | 40.2ns | **11x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 10MB | 448ns | 40.2ns | **11x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 100MB | 448ns | 40.2ns | **11x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 500MB | 444ns | 40.3ns | **11x faster** | 0 -> 0 |
| `ReadAt` (parallel, same offset) | 1GB | 451ns | 40.2ns | **11x faster** | 0 -> 0 |
| `Write` | 1KB | 1761ns | 36.6ns | **48x faster** | 0 -> 0 |
| `Write` | 10KB | 2258ns | 136ns | **17x faster** | 0 -> 0 |
| `Write` | 100KB | 5.21µs | 2.34µs | **2.2x faster** | 0 -> 0 |
| `Write` | 1MB | 39.0µs | 37.9µs | near parity | 0 -> 0 |
| `Write` | 10MB | 388µs | 400µs | near parity | 0 -> 0 |
| `Write` | 100MB | 8.67ms | 12.77ms | 1.5x slower | 0 -> 0 |
| `Write` | 500MB | 73.4ms | 435ms | 5.9x slower | 0 -> 0 |
| `Write` | 1GB | 221ms | 884ms | 4.0x slower | 0 -> 0 |
| `WriteAt` | 1KB | 1187ns | 25.1ns | **47x faster** | 0 -> 0 |
| `WriteAt` | 10KB | 1673ns | 126ns | **13x faster** | 0 -> 0 |
| `WriteAt` | 100KB | 4.63µs | 2.34µs | **2.0x faster** | 0 -> 0 |
| `WriteAt` | 1MB | 38.9µs | 38.2µs | near parity | 0 -> 0 |
| `WriteAt` | 10MB | 389µs | 401µs | near parity | 0 -> 0 |
| `WriteAt` | 100MB | 8.26ms | 12.77ms | 1.5x slower | 0 -> 0 |
| `WriteAt` | 500MB | 72.7ms | 437ms | 6.0x slower | 0 -> 0 |
| `WriteAt` | 1GB | 222ms | 890ms | 4.0x slower | 0 -> 0 |
| `Seek` | - | 603ns | 12.4ns | **49x faster** | 0 -> 0 |
| `ReadFrom` | 1KB | 2036ns | 80.1ns | **25x faster** | 1 -> 2 |
| `ReadFrom` | 10KB | 2510ns | 180ns | **14x faster** | 1 -> 2 |
| `ReadFrom` | 100KB | 5.52µs | 2.41µs | **2.3x faster** | 1 -> 2 |
| `ReadFrom` | 1MB | 39.2µs | 38.9µs | near parity | 1 -> 2 |
| `ReadFrom` | 10MB | 385µs | 406µs | near parity | 1 -> 2 |
| `ReadFrom` | 100MB | 8.32ms | 12.87ms | 1.5x slower | 1 -> 2 |
| `ReadFrom` | 500MB | 73.5ms | 433ms | 5.9x slower | 1 -> 2 |
| `ReadFrom` | 1GB | 221ms | 884ms | 4.0x slower | 1 -> 2 |
| `WriteTo` | - | 2266ns | 9.77ns | **232x faster** | 0 -> 0 |
| `Stat` | - | 1.02µs | 1.04µs | near parity | 1 -> 1 |
| `Sync` | - | 1.19µs | 2.23µs | 1.9x slower | 0 -> 0 |
| `Close` | - | 6.73µs | 13.92µs | 2.1x slower | 4 -> 6 |
| **Geomean** | - | **86.9µs** | **20.1µs** | **~4.3x faster** | - |

**Key takeaway:**

**`mmapfile`** trades syscalls for memory operations, which is strongest for read-heavy and latency-bound workloads:

* **Reads** (1KB-1GB): **3.4-58x faster** across `Read` and `ReadAt`, with zero allocations.
* **Parallel positional reads**: **10-22x faster** in this benchmark. The current benchmark repeatedly reads offset `0`, so it measures concurrent positional-read overhead rather than random offset distribution.
* **Small writes**: **2-48x faster** up to roughly 100KB. Around 1-10MB, mmapfile is near parity.
* **Large sequential writes**: [`*os.File`](https://pkg.go.dev/os#File) wins clearly from 100MB upward in this run, and is much faster at 500MB-1GB. Kernel buffering and writeback are better suited to bulk streaming writes.
* **Lifecycle operations**: `Sync` and `Close` are slower for mmapfile because they flush and unmap the mapping. These are not the primary hot path.

**Overall geomean: ~4.3x faster** across the benchmark suite.

> [!TIP]
> * For bulk/sequential writes, use [`*os.File`](https://pkg.go.dev/os#File) unless mmap semantics are required. For fixed-size in-place updates, mmapfile is best when writes are small or randomly positioned.
>
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
>   No [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt) allocs/copies/syscalls; mmap-pinned memory is useful for DB/indexers/shared-IPC (valid until [`Close`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Close); concurrent-safe with care).
>
> * For durability, call [`f.Sync()`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Sync) after key writes. Unix platforms that expose it use `msync`, Windows uses `FlushViewOfFile`, and NetBSD/fallback builds flush through the underlying file.

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

## Semgrep Rules

Use this [Semgrep](https://github.com/semgrep/semgrep) rules to automatically detect [`*os.File`](https://pkg.go.dev/os#File) usage and suggest `mmapfile` replacements.

Scan your codebase:

```bash
# Download rules
wget -q https://github.com/dwisiswant0/mmapfile/raw/refs/heads/master/extras/mmapfile-semgrep-rules.yaml
# Scan
semgrep scan --config mmapfile-semgrep-rules.yaml /path/to/your/go/workspace
# or Scan w/ autofix (REVIEW CHANGES!)
semgrep scan --autofix --config mmapfile-semgrep-rules.yaml /path/to/your/go/workspace
```

> [!WARNING]  
> The `--autofix` flag is only used for safe existing-file replacements with `size=0`. For <code>[os.O_CREATE](https://pkg.go.dev/os#O_CREATE)|[os.O_TRUNC](https://pkg.go.dev/os#O_TRUNC)</code>, choose the fixed size manually before replacing the call. `mmapfile` has fixed size (no growth/[`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND)).

Rules source: [extras/mmapfile-semgrep-rules.yaml](./extras/mmapfile-semgrep-rules.yaml).

## Limitations

1. **Fixed size**: Files cannot grow after opening. Use `size` with [`os.O_CREATE`](https://pkg.go.dev/os#O_CREATE) for new files and [`os.O_TRUNC`](https://pkg.go.dev/os#O_TRUNC) for existing files.
2. **No Truncate**: Changing file size requires closing and reopening.
3. **No [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND)**: Appending is not supported.
4. **Cursor operations are slower than positional**: Use [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt)/[`WriteAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.WriteAt) for best performance.

## Platform Support

| Platform | Implementation |
|----------|----------------|
| Linux | `mmap`/`munmap`/`msync` |
| Darwin (macOS) | `mmap`/`munmap`/`msync` |
| FreeBSD, OpenBSD, DragonFly | `mmap`/`munmap`/`msync` |
| NetBSD | `mmap`/`munmap`/`fsync` |
| Windows | `CreateFileMapping`/`MapViewOfFile`/`FlushViewOfFile` |
| Other | Fallback (reads file into memory, writes back on `Sync`/`Close`) |

## Thread Safety

- Methods synchronize with each other, including [`Close`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Close).
- [`Read`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Read), [`Write`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Write), and [`Seek`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Seek) share a cursor, so concurrent cursor operations may interleave.
- Slices returned by [`Bytes`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Bytes) are caller-managed and must not be used after [`Close`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Close).

## Status

> [!CAUTION]
> **`mmapfile`** is pre-v1 and does NOT provide a stable API; **use at your own risk**.

Occasional breaking changes may be introduced without notice until a post-v1 release.

## License

**mmapfile** is released with ♡ by [**@dwisiswant0**](https://github.com/dwisiswant0) under the Apache 2.0 license. See [LICENSE](/LICENSE).

## Acknowledgments

Inspired by [`golang.org/x/exp/mmap`](https://pkg.go.dev/golang.org/x/exp/mmap).
