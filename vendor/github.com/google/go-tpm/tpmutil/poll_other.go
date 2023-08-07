//go:build !linux && !darwin

package tpmutil

import (
	"os"
)

// Not implemented on Windows.
func poll(_ *os.File) error { return nil }
