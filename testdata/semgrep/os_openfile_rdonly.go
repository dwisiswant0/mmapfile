package semgrep

import "os"

func testOsOpenFileRdonly() error {
	// ruleid: go.mmapfile.openfile.use-instead-of.os.openfile
	f, err := os.OpenFile("data.txt", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
