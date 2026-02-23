/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validators

import (
	"fmt"
	"io"
)

// OutputSink is an interface for writing output from the validators.
type OutputSink interface {
	// WriteText writes a text string to the output sink.
	WriteText(text string)

	// OnShellExec is called when a shell command is executed, with the command, its stdout, stderr, and any error that occurred.
	OnShellExec(command string, results *CommandResult)

	// Success indicates a successful check, allowing the output sink to format it accordingly.
	Success(text string)

	// Close closes the output sink and releases any resources.
	io.Closer
}

// Logf is like t.Logf, but also writes to the sinks.
func (h *ValidatorHarness) Logf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)

	h.output.WriteText(s)
	h.t.Logf(format, args...)
}

// Log is like t.Log, but also writes to the sinks.
func (h *ValidatorHarness) Log(s string) {
	h.output.WriteText(s)
	h.t.Log(s)
}

// Fatalf is like t.Fatalf, but also writes to the sinks.
func (h *ValidatorHarness) Fatalf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)

	h.output.WriteText("FATAL: " + s)
	h.t.Fatalf(format, args...)
}

// Success is like Logf, but indicates a successful check.
func (h *ValidatorHarness) Success(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	h.output.Success(s)
	h.t.Logf("SUCCESS: "+format, args...)
}
