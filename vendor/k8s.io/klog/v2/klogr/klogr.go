// Package klogr implements github.com/go-logr/logr.Logger in terms of
// k8s.io/klog.
package klogr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/internal/serialize"
)

const (
	// nameKey is used to log the `WithName` values as an additional attribute.
	nameKey = "logger"
)

// Option is a functional option that reconfigures the logger created with New.
type Option func(*klogger)

// Format defines how log output is produced.
type Format string

const (
	// FormatSerialize tells klogr to turn key/value pairs into text itself
	// before invoking klog. Key/value pairs are sorted by key.
	FormatSerialize Format = "Serialize"

	// FormatKlog tells klogr to pass all text messages and key/value pairs
	// directly to klog. Klog itself then serializes in a human-readable
	// format and optionally passes on to a structure logging backend.
	FormatKlog Format = "Klog"
)

// WithFormat selects the output format.
func WithFormat(format Format) Option {
	return func(l *klogger) {
		l.format = format
	}
}

// New returns a logr.Logger which serializes output itself
// and writes it via klog.
//
// Deprecated: this uses a custom, out-dated output format. Use textlogger.NewLogger instead.
func New() logr.Logger {
	return NewWithOptions(WithFormat(FormatSerialize))
}

// NewWithOptions returns a logr.Logger which serializes as determined
// by the WithFormat option and writes via klog. The default is
// FormatKlog.
//
// Deprecated: FormatSerialize is out-dated. For FormatKlog, use textlogger.NewLogger instead.
func NewWithOptions(options ...Option) logr.Logger {
	l := klogger{
		level:  0,
		values: nil,
		format: FormatKlog,
	}
	for _, option := range options {
		option(&l)
	}
	return logr.New(&l)
}

type klogger struct {
	level     int
	callDepth int

	// hasPrefix is true if the first entry in values is the special
	// nameKey key/value. Such an entry gets added and later updated in
	// WithName.
	hasPrefix bool

	values []interface{}
	format Format
}

func (l *klogger) Init(info logr.RuntimeInfo) {
	l.callDepth += info.CallDepth
}

func flatten(kvList ...interface{}) string {
	keys := make([]string, 0, len(kvList))
	vals := make(map[string]interface{}, len(kvList))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			panic(fmt.Sprintf("key is not a string: %s", pretty(kvList[i])))
		}
		var v interface{}
		if i+1 < len(kvList) {
			v = kvList[i+1]
		}
		// Only print each key once...
		if _, seen := vals[k]; !seen {
			keys = append(keys, k)
		}
		// ... with the latest value.
		vals[k] = v
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for i, k := range keys {
		v := vals[k]
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(pretty(k))
		buf.WriteString("=")
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

func pretty(value interface{}) string {
	if err, ok := value.(error); ok {
		if _, ok := value.(json.Marshaler); !ok {
			value = err.Error()
		}
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return fmt.Sprintf("<<error: %v>>", err)
	}
	return strings.TrimSpace(buffer.String())
}

func (l *klogger) Info(level int, msg string, kvList ...interface{}) {
	switch l.format {
	case FormatSerialize:
		msgStr := flatten("msg", msg)
		merged := serialize.MergeKVs(l.values, kvList)
		kvStr := flatten(merged...)
		klog.VDepth(l.callDepth+1, klog.Level(level)).InfoDepth(l.callDepth+1, msgStr, " ", kvStr)
	case FormatKlog:
		merged := serialize.MergeKVs(l.values, kvList)
		klog.VDepth(l.callDepth+1, klog.Level(level)).InfoSDepth(l.callDepth+1, msg, merged...)
	}
}

func (l *klogger) Enabled(level int) bool {
	return klog.VDepth(l.callDepth+1, klog.Level(level)).Enabled()
}

func (l *klogger) Error(err error, msg string, kvList ...interface{}) {
	msgStr := flatten("msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = serialize.ErrorToString(err)
	}
	switch l.format {
	case FormatSerialize:
		errStr := flatten("error", loggableErr)
		merged := serialize.MergeKVs(l.values, kvList)
		kvStr := flatten(merged...)
		klog.ErrorDepth(l.callDepth+1, msgStr, " ", errStr, " ", kvStr)
	case FormatKlog:
		merged := serialize.MergeKVs(l.values, kvList)
		klog.ErrorSDepth(l.callDepth+1, err, msg, merged...)
	}
}

// WithName returns a new logr.Logger with the specified name appended.  klogr
// uses '.' characters to separate name elements.  Callers should not pass '.'
// in the provided name string, but this library does not actually enforce that.
func (l klogger) WithName(name string) logr.LogSink {
	if l.hasPrefix {
		// Copy slice and modify value. No length checks and type
		// assertions are needed because hasPrefix is only true if the
		// first two elements exist and are key/value strings.
		v := make([]interface{}, 0, len(l.values))
		v = append(v, l.values...)
		prefix, _ := v[1].(string)
		prefix = prefix + "." + name
		v[1] = prefix
		l.values = v
	} else {
		// Preprend new key/value pair.
		v := make([]interface{}, 0, 2+len(l.values))
		v = append(v, nameKey, name)
		v = append(v, l.values...)
		l.values = v
		l.hasPrefix = true
	}
	return &l
}

func (l klogger) WithValues(kvList ...interface{}) logr.LogSink {
	l.values = serialize.WithValues(l.values, kvList)
	return &l
}

func (l klogger) WithCallDepth(depth int) logr.LogSink {
	l.callDepth += depth
	return &l
}

var _ logr.LogSink = &klogger{}
var _ logr.CallDepthLogSink = &klogger{}
