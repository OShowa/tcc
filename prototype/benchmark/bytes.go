package benchmark

import "fmt"

// Returns nil if the byte slices are identicals
func ByteOnByte(blob1, blob2 []byte) error {

	for i, b := range blob1 {
		if blob2[i] != b {
			return fmt.Errorf("mismatch on index %d: blob1 has '%d' while blob2 has '%d'", i, b, blob2[i])
		}
	}

	return nil
}
