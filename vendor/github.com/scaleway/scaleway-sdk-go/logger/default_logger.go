package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var DefaultLogger = newLogger(os.Stderr, LogLevelWarning)
var logger Logger = DefaultLogger

// loggerT is the default logger used by scaleway-sdk-go.
type loggerT struct {
	m [4]*log.Logger
	v LogLevel
}

// Init create a new default logger.
// Not mutex-protected, should be called before any scaleway-sdk-go functions.
func (g *loggerT) Init(w io.Writer, level LogLevel) {
	g.m = newLogger(w, level).m
	g.v = level
}

// Debugf logs to the DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) { logger.Debugf(format, args...) }
func (g *loggerT) Debugf(format string, args ...interface{}) {
	g.m[LogLevelDebug].Printf(format, args...)
}

// Infof logs to the INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(format string, args ...interface{}) { logger.Infof(format, args...) }
func (g *loggerT) Infof(format string, args ...interface{}) {
	g.m[LogLevelInfo].Printf(format, args...)
}

// Warningf logs to the WARNING log. Arguments are handled in the manner of fmt.Printf.
func Warningf(format string, args ...interface{}) { logger.Warningf(format, args...) }
func (g *loggerT) Warningf(format string, args ...interface{}) {
	g.m[LogLevelWarning].Printf(format, args...)
}

// Errorf logs to the ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) { logger.Errorf(format, args...) }
func (g *loggerT) Errorf(format string, args ...interface{}) {
	g.m[LogLevelError].Printf(format, args...)
}

// ShouldLog reports whether verbosity level l is at least the requested verbose level.
func ShouldLog(level LogLevel) bool { return logger.ShouldLog(level) }
func (g *loggerT) ShouldLog(level LogLevel) bool {
	return level >= g.v
}

func isEnabled(envKey string) bool {
	env, exist := os.LookupEnv(envKey)
	if !exist {
		return false
	}

	value, err := strconv.ParseBool(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: environment variable %s has invalid boolean value\n", envKey)
	}

	return value
}

// newLogger creates a logger to be used as default logger.
// All logs are written to w.
func newLogger(w io.Writer, level LogLevel) *loggerT {
	errorW := ioutil.Discard
	warningW := ioutil.Discard
	infoW := ioutil.Discard
	debugW := ioutil.Discard
	if isEnabled(DebugEnv) {
		level = LogLevelDebug
	}
	switch level {
	case LogLevelDebug:
		debugW = w
	case LogLevelInfo:
		infoW = w
	case LogLevelWarning:
		warningW = w
	case LogLevelError:
		errorW = w
	}

	// Error logs will be written to errorW, warningW, infoW and debugW.
	// Warning logs will be written to warningW, infoW and debugW.
	// Info logs will be written to infoW and debugW.
	// Debug logs will be written to debugW.
	var m [4]*log.Logger

	m[LogLevelError] = log.New(io.MultiWriter(debugW, infoW, warningW, errorW),
		severityName[LogLevelError]+": ", log.LstdFlags)

	m[LogLevelWarning] = log.New(io.MultiWriter(debugW, infoW, warningW),
		severityName[LogLevelWarning]+": ", log.LstdFlags)

	m[LogLevelInfo] = log.New(io.MultiWriter(debugW, infoW),
		severityName[LogLevelInfo]+": ", log.LstdFlags)

	m[LogLevelDebug] = log.New(debugW,
		severityName[LogLevelDebug]+": ", log.LstdFlags)

	return &loggerT{m: m, v: level}
}
