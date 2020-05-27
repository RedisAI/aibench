package fraud

import (
	"bytes"
	"testing"
)

func testIfInByteStringSlice(t *testing.T, arr [][]byte, choice []byte) {
	for _, x := range arr {
		if bytes.Equal(x, choice) {
			return
		}
	}
	t.Errorf("could known find choice in array: %s", choice)
}

func testIfInInt64Slice(t *testing.T, arr []int64, choice int64) {
	for _, x := range arr {
		if x == choice {
			return
		}
	}
	t.Errorf("could known find choice in array: %d", choice)
}
