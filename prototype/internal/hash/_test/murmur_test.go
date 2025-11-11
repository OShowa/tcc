package hash_test

import (
	"fmt"
	"sis/internal/hash"
	"testing"
)

func TestMurmur3(t *testing.T) {
	// original:
	digest := hash.Murmur3([]byte("Hello, world!"))
	t.Log(fmt.Sprintf("%x", digest))
	// 02c07f89768dac56a2567445945a6ed8
}
