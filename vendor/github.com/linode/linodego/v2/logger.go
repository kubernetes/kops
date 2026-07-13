package linodego

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Errorf(format string, v ...any)
	Warnf(format string, v ...any)
	Debugf(format string, v ...any)
}

type logger struct {
	l *log.Logger
}

func createLogger() *logger {
	l := &logger{l: log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)}
	return l
}

var _ Logger = (*logger)(nil)

func (l *logger) Errorf(format string, v ...any) {
	l.output("ERROR "+format, v...)
}

func (l *logger) Warnf(format string, v ...any) {
	l.output("WARN "+format, v...)
}

func (l *logger) Debugf(format string, v ...any) {
	l.output("DEBUG "+format, v...)
}

func (l *logger) output(format string, v ...any) { //nolint:goprintffuncname
	// Render the final message first, then sanitize control characters
	// to prevent log injection via both the format string and variadic args.
	var msg string
	if len(v) == 0 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	msg = strings.ReplaceAll(msg, "\r\n", "\\n")
	msg = strings.ReplaceAll(msg, "\r", "\\n")
	msg = strings.ReplaceAll(msg, "\n", "\\n")

	l.l.Print(msg)
}
