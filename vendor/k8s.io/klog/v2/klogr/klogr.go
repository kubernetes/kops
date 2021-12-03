// Package klogr implements github.com/go-logr/logr.Logger in terms of
// k8s.io/klog.
package klogr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"sort"
	"strings"
)

// Option is a functional option that reconfigures the logger created with New.
type Option func(*klogger)

// Format defines how log output is produced.
type Format string

const (
	// FormatSerialize tells klogr to turn key/value pairs into text itself
	// before invoking klog.
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
func New() logr.Logger {
	return NewWithOptions(WithFormat(FormatSerialize))
}

// NewWithOptions returns a logr.Logger which serializes as determined
// by the WithFormat option and writes via klog. The default is
// FormatKlog.
func NewWithOptions(options ...Option) logr.Logger {
	l := klogger{
		level:  0,
		prefix: "",
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
	prefix    string
	values    []interface{}
	format    Format
}

func (l *klogger) Init(info logr.RuntimeInfo) {
	l.callDepth += info.CallDepth
}

// trimDuplicates will deduplicate elements provided in multiple KV tuple
// slices, whilst maintaining the distinction between where the items are
// contained.
func trimDuplicates(kvLists ...[]interface{}) [][]interface{} {
	// maintain a map of all seen keys
	seenKeys := map[interface{}]struct{}{}
	// build the same number of output slices as inputs
	outs := make([][]interface{}, len(kvLists))
	// iterate over the input slices backwards, as 'later' kv specifications
	// of the same key will take precedence over earlier ones
	for i := len(kvLists) - 1; i >= 0; i-- {
		// initialise this output slice
		outs[i] = []interface{}{}
		// obtain a reference to the kvList we are processing
		kvList := kvLists[i]

		// start iterating at len(kvList) - 2 (i.e. the 2nd last item) for
		// slices that have an even number of elements.
		// We add (len(kvList) % 2) here to handle the case where there is an
		// odd number of elements in a kvList.
		// If there is an odd number, then the last element in the slice will
		// have the value 'null'.
		for i2 := len(kvList) - 2 + (len(kvList) % 2); i2 >= 0; i2 -= 2 {
			k := kvList[i2]
			// if we have already seen this key, do not include it again
			if _, ok := seenKeys[k]; ok {
				continue
			}
			// make a note that we've observed a new key
			seenKeys[k] = struct{}{}
			// attempt to obtain the value of the key
			var v interface{}
			// i2+1 should only ever be out of bounds if we handling the first
			// iteration over a slice with an odd number of elements
			if i2+1 < len(kvList) {
				v = kvList[i2+1]
			}
			// add this KV tuple to the *start* of the output list to maintain
			// the original order as we are iterating over the slice backwards
			outs[i] = append([]interface{}{k, v}, outs[i]...)
		}
	}
	return outs
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
		keys = append(keys, k)
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
	encoder.Encode(value)
	return strings.TrimSpace(string(buffer.Bytes()))
}

func (l klogger) Info(level int, msg string, kvList ...interface{}) {
	switch l.format {
	case FormatSerialize:
		msgStr := flatten("msg", msg)
		trimmed := trimDuplicates(l.values, kvList)
		fixedStr := flatten(trimmed[0]...)
		userStr := flatten(trimmed[1]...)
		klog.InfoDepth(l.callDepth+1, l.prefix, " ", msgStr, " ", fixedStr, " ", userStr)
	case FormatKlog:
		trimmed := trimDuplicates(l.values, kvList)
		if l.prefix != "" {
			msg = l.prefix + ": " + msg
		}
		klog.InfoSDepth(l.callDepth+1, msg, append(trimmed[0], trimmed[1]...)...)
	}
}

func (l klogger) Enabled(level int) bool {
	return klog.V(klog.Level(level)).Enabled()
}

func (l klogger) Error(err error, msg string, kvList ...interface{}) {
	msgStr := flatten("msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	switch l.format {
	case FormatSerialize:
		errStr := flatten("error", loggableErr)
		trimmed := trimDuplicates(l.values, kvList)
		fixedStr := flatten(trimmed[0]...)
		userStr := flatten(trimmed[1]...)
		klog.ErrorDepth(l.callDepth+1, l.prefix, " ", msgStr, " ", errStr, " ", fixedStr, " ", userStr)
	case FormatKlog:
		trimmed := trimDuplicates(l.values, kvList)
		if l.prefix != "" {
			msg = l.prefix + ": " + msg
		}
		klog.ErrorSDepth(l.callDepth+1, err, msg, append(trimmed[0], trimmed[1]...)...)
	}
}

// WithName returns a new logr.Logger with the specified name appended.  klogr
// uses '/' characters to separate name elements.  Callers should not pass '/'
// in the provided name string, but this library does not actually enforce that.
func (l klogger) WithName(name string) logr.LogSink {
	if len(l.prefix) > 0 {
		l.prefix = l.prefix + "/"
	}
	l.prefix += name
	return &l
}

func (l klogger) WithValues(kvList ...interface{}) logr.LogSink {
	// Three slice args forces a copy.
	n := len(l.values)
	l.values = append(l.values[:n:n], kvList...)
	return &l
}

func (l klogger) WithCallDepth(depth int) logr.LogSink {
	l.callDepth += depth
	return &l
}

var _ logr.LogSink = &klogger{}
var _ logr.CallDepthLogSink = &klogger{}
