package semgrep

import "os"

func testOsOpenFileWronly() error {
	// ok: go.mmapfile.openfile-existing.use-instead-of.os.openfile
	// ok: go.mmapfile.openfile-sized.use-instead-of.os.openfile
	f, err := os.OpenFile("data.txt", os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
