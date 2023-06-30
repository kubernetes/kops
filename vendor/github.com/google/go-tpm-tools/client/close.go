package client

import (
	"io"
	"testing"

	"github.com/google/go-tpm/legacy/tpm2"
)

// CheckedClose closes the simulator and asserts that there were no leaked handles.
func CheckedClose(tb testing.TB, rwc io.ReadWriteCloser) {
	for _, t := range []tpm2.HandleType{
		tpm2.HandleTypeLoadedSession,
		tpm2.HandleTypeSavedSession,
		tpm2.HandleTypeTransient,
	} {
		handles, err := Handles(rwc, t)
		if err != nil {
			tb.Errorf("failed to fetch handles of type %v: %v", t, err)
		}
		if len(handles) != 0 {
			tb.Errorf("tests leaked handles: %v", handles)
		}
	}

	if err := rwc.Close(); err != nil {
		tb.Errorf("when closing simulator: %v", err)
	}
}
