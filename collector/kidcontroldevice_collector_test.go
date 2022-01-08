package collector

import (
	"testing"
)

func TestParseBytes(t *testing.T) {

	bytesString := []struct {
		u string
		v float64
	}{
		{"0", 0},
		{"12", 12},
		{"260.0KiB", 260 * 1024},
		{"260.0MiB", 260 * 1024000},
		{"260.2GiB", 260.2 * 1024000000},
	}

	for _, bytes := range bytesString {
		b := parseBytes(bytes.u)
		if b == -1 {
			t.Error("Unable to parse value ")
		}
		if b != bytes.v {
			t.Errorf("bytes : %f != v : %f\n", b, bytes.v)
		}
	}
}
