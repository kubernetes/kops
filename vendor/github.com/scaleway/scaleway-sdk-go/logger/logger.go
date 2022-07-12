package logger

import "os"

type LogLevel int

const DebugEnv = "SCW_DEBUG"

const (
	// LogLevelDebug indicates Debug severity.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo indicates Info severity.
	LogLevelInfo
	// LogLevelWarning indicates Warning severity.
	LogLevelWarning
	// LogLevelError indicates Error severity.
	LogLevelError
)

// severityName contains the string representation of each severity.
var severityName = []string{
	LogLevelDebug:   "DEBUG",
	LogLevelInfo:    "INFO",
	LogLevelWarning: "WARNING",
	LogLevelError:   "ERROR",
}

// Logger does underlying logging work for scaleway-sdk-go.
type Logger interface {
	// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
	Debugf(format string, args ...interface{})
	// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
	Infof(format string, args ...interface{})
	// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
	Warningf(format string, args ...interface{})
	// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
	Errorf(format string, args ...interface{})
	// ShouldLog reports whether verbosity level l is at least the requested verbose level.
	ShouldLog(level LogLevel) bool
}

// SetLogger sets logger that is used in by the SDK.
// Not mutex-protected, should be called before any scaleway-sdk-go functions.
func SetLogger(l Logger) {
	logger = l
}

// EnableDebugMode enable LogLevelDebug on the default logger.
// If a custom logger was provided with SetLogger this method has no effect.
func EnableDebugMode() {
	DefaultLogger.Init(os.Stderr, LogLevelDebug)
}
