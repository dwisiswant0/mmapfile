package semgrep

import "os"

func testOsOpen() error {
	// ruleid: go.mmapfile.open.use-instead-of.os.open
	f, err := os.Open("data.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
