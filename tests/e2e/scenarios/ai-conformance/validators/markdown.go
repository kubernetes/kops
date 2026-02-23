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
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// MarkdownOutput is an implementation of OutputSink that writes output in Markdown format to a file.
type MarkdownOutput struct {
	f *os.File
	t *testing.T
}

// WriteText writes the given plain text to the markdown file.
func (o *MarkdownOutput) WriteText(text string) {
	o.printf("%s", text)
}

// Success writes a success message to the markdown file, prefixed with a checkmark.
func (o *MarkdownOutput) Success(text string) {
	o.printf("&check; %s\n", text)
}

// Close closes the underlying file. It should be called when all output is done.
func (o *MarkdownOutput) Close() error {
	return o.f.Close()
}

// OnShellExec writes the executed shell command and its output to the markdown file in a formatted code block.
func (o *MarkdownOutput) OnShellExec(command string, results *CommandResult) {
	o.printf("```bash\n> %s\n", command)
	o.printf("%s", results.Stdout())
	o.printf("%s", results.Stderr())

	if results.Err() != nil {
		o.printf("Error:\n```\n%v\n```\n", results.Err())
	}

	o.printf("```\n")
}

// printf is a helper method to write formatted text to the markdown file.
func (o *MarkdownOutput) printf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_, err := fmt.Fprintln(o.f, s)
	if err != nil {
		o.t.Fatalf("failed to write to markdown file: %v", err)
	}
}

// createMarkdownOutput creates a MarkdownOutput that writes to a file based on the test's name and location.
func createMarkdownOutput(t *testing.T) OutputSink {
	// Get file path and other info from the current caller frame (depth 0)
	_, testFilename, _, ok := runtime.Caller(2)
	if !ok {
		t.Fatal("Could not get test caller")
	}
	_, baseFilename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("Could not get test caller")
	}

	baseDir := filepath.Dir(baseFilename)

	testRelativePath, err := filepath.Rel(baseDir, testFilename)
	if err != nil {
		t.Fatalf("failed to get relative path: %v", err)
	}

	testRelativeDir := filepath.Dir(testRelativePath)

	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}
	outputBase := filepath.Join(artifactsDir, "ai-conformance")

	outputPath := filepath.Join(outputBase, testRelativeDir, t.Name()+".md")

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("failed to create output directory %v: %v", filepath.Dir(outputPath), err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("failed to create markdown file: %v", err)
	}
	output := &MarkdownOutput{f: outputFile, t: t}
	return output
}
