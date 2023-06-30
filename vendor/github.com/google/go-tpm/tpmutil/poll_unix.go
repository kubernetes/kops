//go:build linux || darwin

package tpmutil

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// poll blocks until the file descriptor is ready for reading or an error occurs.
func poll(f *os.File) error {
	var (
		fds = []unix.PollFd{{
			Fd:     int32(f.Fd()),
			Events: 0x1, // POLLIN
		}}
		timeout = -1 // Indefinite timeout
	)

	if _, err := unix.Poll(fds, timeout); err != nil {
		return err
	}

	// Revents is filled in by the kernel.
	// If the expected event happened, Revents should match Events.
	if fds[0].Revents != fds[0].Events {
		return fmt.Errorf("unexpected poll Revents 0x%x", fds[0].Revents)
	}
	return nil
}
