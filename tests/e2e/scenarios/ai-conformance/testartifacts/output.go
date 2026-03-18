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

package testartifacts

import (
	"os"
	"path/filepath"
)

// Testing is an interface that abstracts the testing.T type.
type Testing interface {
	// Name returns the name of the test.
	Name() string

	// Fatalf logs a formatted message and marks the test as failed.
	Fatalf(format string, args ...interface{})
}

// PathForTestArtifact defines the type for the functions-option pattern for PathForTestArtifact.
type PathForTestArtifactOption func(*pathForTestArtifactOptions)

// pathForTestArtifactOptions is the options built by the functions-option pattern for PathForTestArtifact.
type pathForTestArtifactOptions struct {
	// MkdirAll indicates that the directories leading to the artifact should be created if they do not exist.
	MkdirAll bool

	// RelativeToArtifactsDir indicates that the file path should be relative to the artifacts directory, rather than the test-specific subdirectory.
	RelativeToArtifactsDir bool
}

// WithMkdirAll is a PathForTestArtifactOption that indicates that the directories leading to the artifact should be created if they do not exist.
func WithMkdirAll() PathForTestArtifactOption {
	return func(o *pathForTestArtifactOptions) {
		o.MkdirAll = true
	}
}

// RelativeToArtifactsDir is a PathForTestArtifactOption that indicates that the file path should be relative to the artifacts directory, rather than the test-specific subdirectory.
func RelativeToArtifactsDir() PathForTestArtifactOption {
	return func(o *pathForTestArtifactOptions) {
		o.RelativeToArtifactsDir = true
	}
}

// PathForTestArtifact returns the file path for a test artifact based on the test name and the provided artifact name.
func PathForTestArtifact(t Testing, fileName string, opts ...PathForTestArtifactOption) string {
	var options pathForTestArtifactOptions
	for _, o := range opts {
		o(&options)
	}

	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}

	testName := t.Name()
	relativePath := filepath.Join("tests", testName, fileName)
	p := filepath.Join(artifactsDir, relativePath)

	if options.MkdirAll {
		d := filepath.Dir(p)
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("failed to create directory for test artifact %q: %v", d, err)
		}
	}

	if options.RelativeToArtifactsDir {
		return relativePath
	}
	return p
}

// WriteTestArtifact writes the provided content to a file path determined by PathForTestArtifact, creating any necessary directories.
func WriteTestArtifact(t Testing, fileName string, content []byte) {
	outputFile := PathForTestArtifact(t, fileName, WithMkdirAll())

	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		t.Fatalf("failed to write attestation file %q: %v", outputFile, err)
		return
	}
}
