package semgrep

import "os"

func testOsCreate() error {
	// ruleid: go.mmapfile.create.use-instead-of.os.create
	f, err := os.Create("data.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
