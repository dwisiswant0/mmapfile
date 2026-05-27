//go:build netbsd

package mmapfile

func msync(_ []byte) error {
	return nil
}
