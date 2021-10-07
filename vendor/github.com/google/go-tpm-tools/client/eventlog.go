package client

import "io"

// GetEventLog grabs the crypto-agile TCG event log for the system. The TPM can
// override this implementation by implementing EventLogGetter.
func GetEventLog(rw io.ReadWriter) ([]byte, error) {
	if elg, ok := rw.(EventLogGetter); ok {
		return elg.EventLog()
	}
	return getRealEventLog()
}

// EventLogGetter allows a TPM (io.ReadWriter) to specify a particular
// implementation for GetEventLog(). This is useful for testing and necessary
// for Windows Event Log support (which requires a handle to the TPM).
type EventLogGetter interface {
	EventLog() ([]byte, error)
}
