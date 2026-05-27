//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly && !windows

package mmapfile

import (
	"fmt"
	"io"
	"os"
)

func mapFile(file *os.File, size int, _ bool) ([]byte, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("mmapfile: read %q: %w", file.Name(), err)
	}

	return data, nil
}

func unmapFile(file *os.File, data []byte, writable bool) error {
	if !writable || len(data) == 0 {
		return nil
	}

	return writeFallbackData(file, data)
}

func syncFile(file *os.File, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if err := writeFallbackData(file, data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("mmapfile: sync %q: %w", file.Name(), err)
	}

	return nil
}

func writeFallbackData(file *os.File, data []byte) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("mmapfile: seek %q: %w", file.Name(), err)
	}
	if err := file.Truncate(int64(len(data))); err != nil {
		return fmt.Errorf("mmapfile: truncate %q: %w", file.Name(), err)
	}
	n, err := file.Write(data)
	if err != nil {
		return fmt.Errorf("mmapfile: write %q: %w", file.Name(), err)
	}
	if n != len(data) {
		return fmt.Errorf("mmapfile: write %q: %w", file.Name(), io.ErrShortWrite)
	}

	return nil
}
