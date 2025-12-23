# mmapfile

[![tests](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml/badge.svg?branch=master)](https://github.com/dwisiswant0/mmapfile/actions/workflows/tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.dw1.io/mmapfile.svg)](https://pkg.go.dev/go.dw1.io/mmapfile)

An [`*os.File`](https://pkg.go.dev/os#File)-like type backed by memory-mapped I/O for Go.

**mmapfile** provides a drop-in replacement for `*os.File` in many contexts, offering significantly faster I/O operations by avoiding syscall overhead on every read/write.

## Features

* **[`*os.File`](https://pkg.go.dev/os#File)-compatible interface**: implements [`io.Reader`](https://pkg.go.dev/io#Reader), [`io.Writer`](https://pkg.go.dev/io#Writer), [`io.Seeker`](https://pkg.go.dev/io#Seeker), [`io.ReaderAt`](https://pkg.go.dev/io#ReaderAt), [`io.WriterAt`](https://pkg.go.dev/io#WriterAt), [`io.Closer`](https://pkg.go.dev/io#Closer), [`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom), [`io.WriterTo`](https://pkg.go.dev/io#WriterTo), and [`io.StringWriter`](https://pkg.go.dev/io#StringWriter).
* **Zero-copy reads**: direct access to file contents via [`Bytes()`](https://github.com/semgrep/semgrep) method.
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
  Read/1KB-2                1722.50n ± 2%    36.87n ±  2%   -97.86% (p=0.000 n=10)
  Read/10KB-2                3370.0n ± 2%    159.4n ±  1%   -95.27% (p=0.000 n=10)
  Read/100KB-2               21.241µ ± 1%    1.613µ ±  7%   -92.41% (p=0.000 n=10)
  Read/1MB-2                 218.43µ ± 2%    20.45µ ±  2%   -90.64% (p=0.000 n=10)
  Read/10MB-2                2170.0µ ± 0%    207.2µ ±  1%   -90.45% (p=0.000 n=10)
  Read/100MB-2               24.642m ± 1%    5.006m ±  1%   -79.68% (p=0.000 n=10)
  Read/500MB-2               123.26m ± 1%    32.99m ±  0%   -73.24% (p=0.000 n=10)
  Read/1GB-2                 252.52m ± 2%    80.08m ±  1%   -68.29% (p=0.000 n=10)
  ReadAt/1KB-2              1145.00n ± 1%    23.40n ±  0%   -97.96% (p=0.000 n=10)
  ReadAt/10KB-2              2603.0n ± 1%    157.3n ± 24%   -93.96% (p=0.000 n=10)
  ReadAt/100KB-2             17.915µ ± 0%    1.471µ ±  0%   -91.79% (p=0.000 n=10)
  ReadAt/1MB-2               192.15µ ± 1%    18.96µ ±  2%   -90.13% (p=0.000 n=10)
  ReadAt/10MB-2              1955.3µ ± 3%    190.7µ ±  2%   -90.25% (p=0.000 n=10)
  ReadAt/100MB-2             22.607m ± 1%    4.439m ±  3%   -80.36% (p=0.000 n=10)
  ReadAt/500MB-2             112.24m ± 1%    27.72m ±  1%   -75.31% (p=0.000 n=10)
  ReadAt/1GB-2               227.84m ± 1%    72.76m ±  6%   -68.07% (p=0.000 n=10)
  ReadAtParallel/1KB-2       633.00n ± 3%    51.95n ±  9%   -91.79% (p=0.000 n=10)
  ReadAtParallel/10KB-2      395.75n ± 5%    36.84n ±  0%   -90.69% (p=0.000 n=10)
  ReadAtParallel/100KB-2     388.85n ± 8%    36.86n ±  0%   -90.52% (p=0.000 n=10)
  ReadAtParallel/1MB-2       388.80n ± 4%    36.84n ±  0%   -90.52% (p=0.000 n=10)
  ReadAtParallel/10MB-2      393.10n ± 5%    36.86n ±  0%   -90.62% (p=0.000 n=10)
  ReadAtParallel/100MB-2     400.00n ± 6%    36.88n ±  0%   -90.78% (p=0.000 n=10)
  ReadAtParallel/500MB-2     401.35n ± 4%    38.72n ±  2%   -90.35% (p=0.000 n=10)
  ReadAtParallel/1GB-2       399.30n ± 5%    36.94n ±  2%   -90.75% (p=0.000 n=10)
  Write/1KB-2               1497.00n ± 1%    33.63n ± 13%   -97.75% (p=0.000 n=10)
  Write/10KB-2               2374.0n ± 1%    121.9n ±  0%   -94.87% (p=0.000 n=10)
  Write/100KB-2              12.552µ ± 2%    2.080µ ±  0%   -83.42% (p=0.000 n=10)
  Write/1MB-2                133.08µ ± 6%    46.01µ ±  0%   -65.43% (p=0.000 n=10)
  Write/10MB-2               1352.5µ ± 3%    491.2µ ±  0%   -63.68% (p=0.000 n=10)
  Write/100MB-2               26.64m ± 1%    21.41m ±  0%   -19.63% (p=0.000 n=10)
  Write/500MB-2               317.6m ± 0%    524.6m ±  1%   +65.19% (p=0.000 n=10)
  Write/1GB-2                 649.7m ± 1%   1072.1m ±  0%   +65.01% (p=0.000 n=10)
  WriteAt/1KB-2             1037.50n ± 1%    20.30n ±  0%   -98.04% (p=0.000 n=10)
  WriteAt/10KB-2             1867.0n ± 1%    110.0n ±  1%   -94.11% (p=0.000 n=10)
  WriteAt/100KB-2            12.216µ ± 3%    2.062µ ±  0%   -83.12% (p=0.000 n=10)
  WriteAt/1MB-2              139.24µ ± 5%    45.99µ ±  0%   -66.97% (p=0.000 n=10)
  WriteAt/10MB-2             1409.0µ ± 7%    490.4µ ±  0%   -65.20% (p=0.000 n=10)
  WriteAt/100MB-2             26.94m ± 3%    21.45m ±  1%   -20.35% (p=0.000 n=10)
  WriteAt/500MB-2             316.6m ± 0%    523.7m ±  0%   +65.40% (p=0.000 n=10)
  WriteAt/1GB-2               648.5m ± 1%   1072.0m ±  0%   +65.31% (p=0.000 n=10)
  Seek-2                     355.15n ± 1%    12.22n ±  1%   -96.56% (p=0.000 n=10)
  ReadFrom/1KB-2            1711.50n ± 1%    84.14n ±  1%   -95.08% (p=0.000 n=10)
  ReadFrom/10KB-2            2562.0n ± 1%    167.7n ±  1%   -93.45% (p=0.000 n=10)
  ReadFrom/100KB-2           12.763µ ± 2%    2.148µ ±  1%   -83.17% (p=0.000 n=10)
  ReadFrom/1MB-2             138.24µ ± 3%    46.11µ ±  0%   -66.65% (p=0.000 n=10)
  ReadFrom/10MB-2            1373.8µ ± 3%    501.2µ ±  2%   -63.52% (p=0.000 n=10)
  ReadFrom/100MB-2            26.24m ± 5%    21.42m ±  1%   -18.39% (p=0.000 n=10)
  ReadFrom/500MB-2            319.6m ± 1%    523.8m ±  0%   +63.91% (p=0.000 n=10)
  ReadFrom/1GB-2              649.7m ± 1%   1082.6m ±  1%   +66.65% (p=0.000 n=10)
  WriteTo-2                1926.500n ± 1%    7.604n ±  0%   -99.61% (p=0.000 n=10)
  Stat-2                      699.7n ± 1%    713.1n ±  1%    +1.91% (p=0.000 n=10)
  Sync-2                      844.0n ± 1%    877.8n ±  1%    +4.00% (p=0.000 n=10)
  Close-2                     5.901µ ± 1%   12.707µ ±  1%  +115.34% (p=0.000 n=10)
  geomean                     118.4µ         20.87µ         -82.38%

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
  Stat-2                   208.0 ± 0%     208.0 ± 0%         ~ (p=1.000 n=10) ¹
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  216.0 ± 7%     544.0 ± 0%  +151.85% (p=0.000 n=10)
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
  Stat-2                   1.000 ± 0%     1.000 ± 0%         ~ (p=1.000 n=10) ¹
  Sync-2                   0.000 ± 0%     0.000 ± 0%         ~ (p=1.000 n=10) ¹
  Close-2                  4.000 ± 0%     7.000 ± 0%   +75.00% (p=0.000 n=10)
  geomean                             ²               ?                       ² ³
  ¹ all samples are equal
  ² summaries must be >0 to compute geomean
  ³ ratios must be >0 to compute geomean
  ```
</details>

### Summary

| Operation | Size | `os.File` | `mmapfile` | Improvement | Allocs |
|-----------|------|-----------|------------|-------------|--------|
| `Read` | 1KB | 1722.5ns | 36.9ns | **47x faster** | 0 → 0 |
| `Read` | 10KB | 3370ns | 159ns | **21x faster** | 0 → 0 |
| `Read` | 100KB | 21.24µs | 1.61µs | **13x faster** | 0 → 0 |
| `Read` | 1MB | 218.4µs | 20.5µs | **11x faster** | 0 → 0 |
| `Read` | 10MB | 2170µs | 207µs | **10x faster** | 0 → 0 |
| `Read` | 100MB | 24.64ms | 5.01ms | **5x faster** | 0 → 0 |
| `Read` | 500MB | 123.3ms | 33.0ms | **4x faster** | 0 → 0 |
| `Read` | 1GB | 252.5ms | 80.1ms | **3x faster** | 0 → 0 |
| `ReadAt` | 1KB | 1145ns | 23.4ns | **49x faster** | 0 → 0 |
| `ReadAt` | 10KB | 2603ns | 157ns | **17x faster** | 0 → 0 |
| `ReadAt` | 100KB | 17.92µs | 1.47µs | **12x faster** | 0 → 0 |
| `ReadAt` | 1MB | 192.2µs | 19.0µs | **10x faster** | 0 → 0 |
| `ReadAt` | 10MB | 1955µs | 191µs | **10x faster** | 0 → 0 |
| `ReadAt` | 100MB | 22.61ms | 4.44ms | **5x faster** | 0 → 0 |
| `ReadAt` | 500MB | 112.2ms | 27.7ms | **4x faster** | 0 → 0 |
| `ReadAt` | 1GB | 227.8ms | 72.8ms | **3x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1KB | 633ns | 52ns | **12x faster** | 0 → 0 |
| `ReadAt` (parallel) | 10KB | 396ns | 37ns | **11x faster** | 0 → 0 |
| `ReadAt` (parallel) | 100KB | 389ns | 37ns | **11x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1MB | 389ns | 37ns | **11x faster** | 0 → 0 |
| `ReadAt` (parallel) | 10MB | 393ns | 37ns | **11x faster** | 0 → 0 |
| `ReadAt` (parallel) | 100MB | 400ns | 37ns | **11x faster** | 0 → 0 |
| `ReadAt` (parallel) | 500MB | 401ns | 39ns | **10x faster** | 0 → 0 |
| `ReadAt` (parallel) | 1GB | 399ns | 37ns | **11x faster** | 0 → 0 |
| `Write` | 1KB | 1497ns | 33.6ns | **45x faster** | 0 → 0 |
| `Write` | 10KB | 2374ns | 122ns | **19x faster** | 0 → 0 |
| `Write` | 100KB | 12.55µs | 2.08µs | **6x faster** | 0 → 0 |
| `Write` | 1MB | 133.1µs | 46.0µs | **3x faster** | 0 → 0 |
| `Write` | 10MB | 1353µs | 491µs | **3x faster** | 0 → 0 |
| `Write` | 100MB | 26.64ms | 21.41ms | **1.2x faster** | 0 → 0 |
| `Write` | 500MB | 318ms | 525ms | 1.7x slower | 0 → 0 |
| `Write` | 1GB | 650ms | 1072ms | 1.7x slower | 0 → 0 |
| `WriteAt` | 1KB | 1038ns | 20.3ns | **51x faster** | 0 → 0 |
| `WriteAt` | 10KB | 1867ns | 110ns | **17x faster** | 0 → 0 |
| `WriteAt` | 100KB | 12.22µs | 2.06µs | **6x faster** | 0 → 0 |
| `WriteAt` | 1MB | 139.2µs | 46.0µs | **3x faster** | 0 → 0 |
| `WriteAt` | 10MB | 1409µs | 490µs | **3x faster** | 0 → 0 |
| `WriteAt` | 100MB | 26.94ms | 21.45ms | **1.3x faster** | 0 → 0 |
| `WriteAt` | 500MB | 317ms | 524ms | 1.7x slower | 0 → 0 |
| `WriteAt` | 1GB | 649ms | 1072ms | 1.7x slower | 0 → 0 |
| `Seek` | - | 355ns | 12.2ns | **29x faster** | 0 → 0 |
| `ReadFrom` | 1KB | 1712ns | 84ns | **20x faster** | 2 → 2 |
| `ReadFrom` | 10KB | 2562ns | 168ns | **15x faster** | 2 → 2 |
| `ReadFrom` | 100KB | 12.76µs | 2.15µs | **6x faster** | 2 → 2 |
| `ReadFrom` | 1MB | 138.2µs | 46.1µs | **3x faster** | 2 → 2 |
| `ReadFrom` | 10MB | 1374µs | 501µs | **3x faster** | 2 → 2 |
| `ReadFrom` | 100MB | 26.24ms | 21.42ms | **1.2x faster** | 2 → 2 |
| `ReadFrom` | 500MB | 320ms | 524ms | 1.6x slower | 2 → 2 |
| `ReadFrom` | 1GB | 650ms | 1083ms | 1.7x slower | 2 → 2 |
| `WriteTo` | - | 1927ns | 7.6ns | **254x faster** | 3 → 0 |
| `Stat` | - | 700ns | 713ns | 1.0x slower | 1 → 1 |
| `Sync` | - | 844ns | 878ns | 1.0x slower | 0 → 0 |
| `Close` | - | 5.90µs | 12.71µs | 2.2x slower | 4 → 7 |
| **Geomean** | - | **118µs** | **20.9µs** | **~6x faster** | - |

**Key takeaway:**

mmapfile trades syscalls for memory ops, crushing latency-bound workloads:

* **All reads** (1KB–1GB): **3–50x faster**. Simple `copy(b, f.data[off:])` vs repeated `read(2)`-scales to huge files.
* **Writes ≤100MB**: **6–50x** wins. Syscall cost (~200ns) >> `memcpy`; ideal for frequent small ops.
* **Bulk sequential writes >500MB**: [`*os.File`](https://pkg.go.dev/os#File) ~1.7x ahead (644ms vs 1s+). Kernel magic:
  * Page cache absorbs data (no user double-buffering: heap->mmap).
  * Write-behind: async clustering to disk.
  * Readahead/DMA pipelines for sequential throughput.
  * mmap: pure `memcpy` to dirty pages (~1GB/s BW-bound on CPU; no flush/sync in bench for apple-to-apple op perf).
* **Random/parallel**: mmapfile **10x+** ([`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt) parallel constant ~40ns even 1GB). No seeks/syscalls.

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
>   To bypass [`WriteAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.WriteAt) lock (no [`*sync.RWMutex`](https://pkg.go.dev/sync#RWMutex)), no bounds/EOF checks, and no partial copies. Direct `memcpy` to mmap region; **~10–20% faster** for large ops.
>
> * For durability, call [`f.Sync()`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Sync) after key writes to trigger `msync`: synchronous flush dirty pages to disk (~10–100ms/GB; varies SSD/NVMe/HDD/IO scheduler); essential for WAL/tx commits.
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
>   No [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt) allocs/copies/syscalls; mmap-pinned mem ideal for DB/indexers/shared-IPC (valid until [`Close`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Close); concurrent-safe with care).

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
> The `--autofix` flag inserts `size=0` (safe for existing files). For <code>[os.O_CREATE](https://pkg.go.dev/os#O_CREATE)|[os.O_TRUNC](https://pkg.go.dev/os#O_TRUNC)</code>, **manually set `size > 0`** to your expected file size. `mmapfile` has fixed size (no growth/[`os.O_APPEND`](https://pkg.go.dev/os#O_CREATE)).

Rules source: [extras/mmapfile-semgrep-rules.yaml](./extras/mmapfile-semgrep-rules.yaml).

## Limitations

1. **Fixed size**: Files cannot grow after opening. Use `size` parameter with [`os.O_CREATE`](https://pkg.go.dev/os#O_CREATE).
2. **No Truncate**: Changing file size requires closing and reopening.
3. **No [`os.O_APPEND`](https://pkg.go.dev/os#O_APPEND)**: Appending is not supported.
4. **Cursor operations are slower than positional**: Use [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt)/[`WriteAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.WriteAt) for best performance.

## Platform Support

| Platform | Implementation |
|----------|----------------|
| Linux | `mmap`/`munmap`/`msync` |
| Darwin (macOS) | `mmap`/`munmap`/`msync` |
| FreeBSD, OpenBSD, NetBSD, DragonFly | `mmap`/`munmap`/`msync` |
| Windows | `CreateFileMapping`/`MapViewOfFile`/`FlushViewOfFile` |
| Other | Fallback (reads file into memory) |

## Thread Safety

- [`ReadAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.ReadAt) and [`WriteAt`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.WriteAt) are safe for concurrent use.
- [`Read`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Read), [`Write`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Write), and [`Seek`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Seek) share a cursor, concurrent use will interleave unpredictably.
- [`Close`](https://pkg.go.dev/go.dw1.io/mmapfile#MmapFile.Close) should not be called concurrently with other operations.

## Status

> [!CAUTION]
> **`mmapfile`** is pre-v1 and does NOT provide a stable API; **use at your own risk**.

Occasional breaking changes may be introduced without notice until a post-v1 release.

## License

**mmapfile** is released with ♡ by [**@dwisiswant0**](https://github.com/dwisiswant0) under the Apache 2.0 license. See [LICENSE](/LICENSE).

## Acknowledgments

Inspired by [`golang.org/x/exp/mmap`](https://pkg.go.dev/golang.org/x/exp/mmap).
