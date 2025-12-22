//go:build windows

package mmapfile

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestOpenWindowsSharingViolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no_perm.txt")

	// Create file
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Hold exclusive write handle to deny read sharing
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatalf("UTF16PtrFromString failed: %v", err)
	}
	h, err := syscall.CreateFile(pathp,
		syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	defer syscall.CloseHandle(syscall.Handle(h))

	_, err = Open(path)
	if err == nil {
		t.Error("Open should fail due to sharing violation")
	}
}
