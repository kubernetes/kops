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
	"bytes"
	"os/exec"
)

// ShellExec executes the given shell command and returns the result. If the command fails, it fails the test.
func (h *ValidatorHarness) ShellExec(shellCommand string) *CommandResult {
	ctx := h.Context()
	cmd := exec.CommandContext(ctx, "sh", "-c", shellCommand)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	h.Logf("ShellExec(%q)", shellCommand)
	err := cmd.Run()

	result := &CommandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
	}

	h.output.OnShellExec(shellCommand, result)

	if err != nil {
		h.Logf("Command failed: %q", shellCommand)
		h.Logf("Stdout: %s", result.Stdout())
		h.Logf("Stderr: %s", result.Stderr())
		h.Fatalf("Command execution %q failed with error: %v", shellCommand, err)
	}

	return result
}

// CommandResult encapsulates the result of executing a shell command, including stdout, stderr, and any error that occurred.
type CommandResult struct {
	stdout string
	stderr string
	err    error
}

// Stdout returns the standard output of the command execution.
func (r *CommandResult) Stdout() string {
	return r.stdout
}

// Stderr returns the standard error output of the command execution.
func (r *CommandResult) Stderr() string {
	return r.stderr
}

// Err returns the error that occurred during command execution, if any.
func (r *CommandResult) Err() error {
	return r.err
}
