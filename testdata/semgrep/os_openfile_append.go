package semgrep

import "os"

func testOsOpenFileAppend() error {
	// ok: go.mmapfile.openfile.use-instead-of.os.openfile
	f, err := os.OpenFile("data.txt", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
