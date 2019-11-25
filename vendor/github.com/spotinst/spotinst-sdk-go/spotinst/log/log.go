package log

import (
	"log"
	"os"
)

// DefaultStdLogger represents the default logger which will write log messages
// to stdout, and use same formatting runes as the stdlib log.Logger.
var DefaultStdLogger Logger = log.New(os.Stderr, "", log.LstdFlags)

// Logger specifies the interface for all log operations.
type Logger interface {
	Printf(format string, args ...interface{})
}

// The LoggerFunc type is an adapter to allow the use of ordinary functions as
// Logger. If f is a function with the appropriate signature, LoggerFunc(f) is
// a Logger that calls f.
type LoggerFunc func(format string, args ...interface{})

// Printf calls f(format, args).
func (f LoggerFunc) Printf(format string, args ...interface{}) {
	f(format, args...)
}
