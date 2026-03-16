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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

// MarkdownOutput is an implementation of OutputSink that writes output in Markdown format to a file.
type MarkdownOutput struct {
	f *os.File
	t *testing.T

	// outputPath is the path to the markdown file being written, used for rendering HTML output alongside it.
	outputPath string
}

// WriteText writes the given plain text to the markdown file.
func (o *MarkdownOutput) WriteText(text string) {
	o.printf("\n")
	o.printf("%s", text)
}

// Success writes a success message to the markdown file, prefixed with a checkmark.
func (o *MarkdownOutput) Success(text string) {
	o.printf("&check; %s\n", text)
}

// Skip writes a skip message to the markdown file, prefixed with a warning symbol.
func (o *MarkdownOutput) Skip(message string) {
	o.printf("&warning; SKIPPED: %s\n", message)
}

// Close closes the underlying file and renders an HTML version alongside it.
func (o *MarkdownOutput) Close() error {
	if err := o.f.Close(); err != nil {
		return err
	}

	if err := o.renderHTML(); err != nil {
		o.t.Errorf("failed to render HTML: %v", err)
	}

	return nil
}

// renderHTML reads the markdown file and renders an HTML version alongside it.
func (o *MarkdownOutput) renderHTML() error {
	mdBytes, err := os.ReadFile(o.outputPath)
	if err != nil {
		return fmt.Errorf("reading markdown file: %w", err)
	}

	var htmlBody bytes.Buffer
	if err := goldmark.Convert(mdBytes, &htmlBody); err != nil {
		return fmt.Errorf("converting markdown to HTML: %w", err)
	}

	title := strings.TrimSuffix(filepath.Base(o.outputPath), ".md")

	var htmlOut bytes.Buffer
	htmlOut.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	htmlOut.WriteString("<meta charset=\"utf-8\">\n")
	fmt.Fprintf(&htmlOut, "<title>%s</title>\n", title)
	htmlOut.WriteString("<style>\n")
	htmlOut.WriteString("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif; max-width: 900px; margin: 40px auto; padding: 0 20px; line-height: 1.6; color: #24292e; }\n")
	htmlOut.WriteString("pre { background: #f6f8fa; padding: 16px; border-radius: 6px; overflow-x: auto; }\n")
	htmlOut.WriteString("code { background: #f6f8fa; padding: 2px 6px; border-radius: 3px; font-size: 85%; }\n")
	htmlOut.WriteString("pre code { background: none; padding: 0; }\n")
	htmlOut.WriteString("</style>\n")
	htmlOut.WriteString("</head>\n<body>\n")
	htmlOut.Write(htmlBody.Bytes())
	htmlOut.WriteString("</body>\n</html>\n")

	htmlPath := strings.TrimSuffix(o.outputPath, ".md") + ".html"
	if err := os.WriteFile(htmlPath, htmlOut.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing HTML file: %w", err)
	}

	return nil
}

// BeforeShellExec writes the executed shell command to the markdown file in a formatted code block.
func (o *MarkdownOutput) BeforeShellExec(command string) {
	o.printf("```bash\n> %s\n", command)
	o.printf("```\n")
}

// AfterShellExec writes the result of the executed shell command to the markdown file in a formatted code block.
func (o *MarkdownOutput) AfterShellExec(command string, results *CommandResult) {
	o.printf("```bash")
	o.printf("%s", results.Stdout())
	o.printf("%s", results.Stderr())
	o.printf("```\n")

	if results.Err() != nil {
		o.printf("Error:\n```\n%v\n```\n", results.Err())
	}
}

// printf is a helper method to write formatted text to the markdown file.
func (o *MarkdownOutput) printf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	_, err := fmt.Fprintln(o.f, s)
	if err != nil {
		o.t.Fatalf("failed to write to markdown file: %v", err)
	}
}

// createMarkdownOutput creates a MarkdownOutput that writes to a file based on the test's name.
func createMarkdownOutput(t *testing.T) OutputSink {
	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}

	testName := strings.ReplaceAll(t.Name(), "/", "_")
	outputPath := filepath.Join(artifactsDir, "tests", testName, "output.md")

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		t.Fatalf("failed to create output directory %v: %v", filepath.Dir(outputPath), err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("failed to create markdown file: %v", err)
	}
	output := &MarkdownOutput{f: outputFile, t: t, outputPath: outputPath}
	return output
}
