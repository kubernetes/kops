// +build !linux,!darwin

package tpmutil

import (
	"os"
)

// Not implemented on Windows.
func poll(f *os.File) error { return nil }
