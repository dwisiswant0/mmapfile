package semgrep

import "os"

func testOsOpenFileTrunc() error {
	// ruleid: go.mmapfile.openfile.use-instead-of.os.openfile
	f, err := os.OpenFile("data.txt", os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
