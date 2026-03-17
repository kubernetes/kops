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
	"testing"
)

// OutputSink is an interface for writing output from the validators.
type OutputSink interface {
	// WriteText writes a text string to the output sink.
	WriteText(text string)

	// BeforeShellExec is called when a shell command is about to be executed, with the command string.
	BeforeShellExec(command string)
	// AfterShellExec is called when a shell command has been executed, with the command, its stdout, stderr, and any error that occurred.
	AfterShellExec(command string, results *CommandResult)

	// Success indicates a successful check, allowing the output sink to format it accordingly.
	Success(text string)

	// Skip indicates that a test was skipped, allowing the output sink to format it accordingly.
	Skip(message string)

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

// Errorf is like t.Errorf, but also writes to the sinks.
func (h *ValidatorHarness) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)

	h.output.WriteText("ERROR: " + s)
	h.t.Errorf(format, args...)
}

// Run is like t.Run, but creates a sub-harness that shares the output.
func (h *ValidatorHarness) Run(name string, testFunc func(h *ValidatorHarness)) {
	h.t.Run(name, func(t *testing.T) {
		subHarness := &ValidatorHarness{
			t: t,

			// Share most of the state with the parent harness, but use the sub-test's *testing.T and a new context.
			output:        h.output,
			dynamicClient: h.dynamicClient,
			restConfig:    h.restConfig,
		}
		testFunc(subHarness)
	})
}

// AllPassed returns true if all sub-tests have passed so far. This can be used to conditionally record conformance only if all checks passed.
func (h *ValidatorHarness) AllPassed() bool {
	return !h.t.Failed() && !h.t.Skipped()
}

// Success is like Logf, but indicates a successful check.
func (h *ValidatorHarness) Success(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	h.output.Success(s)
	h.t.Logf("SUCCESS: "+format, args...)
}

// RecordConformance records that a specific conformance test was passed.
func (h *ValidatorHarness) RecordConformance(testName string) {
	// We should gather these in a structured way, but for now we'll just log them.
	h.Logf("Conformance %q passed", testName)
}
