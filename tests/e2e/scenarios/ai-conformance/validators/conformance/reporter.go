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

package conformance

import (
	"os"
	"path"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// Reporter provides methods for recording conformance test results and logging messages.
type Reporter struct {
	t Testing
}

// Testing is a minimal interface for test contexts, allowing the client to report failures and log messages.
type Testing interface {
	// Name returns the name of the test, used for organizing attestation files.
	Name() string

	// Logf logs a formatted message to the test output.
	Logf(format string, args ...interface{})

	// Errorf reports a formatted error message to the test output and marks the test as failed, but continues execution.
	Errorf(format string, args ...interface{})
}

// NewReporter creates a new Reporter.
func NewReporter(t Testing) *Reporter {
	return &Reporter{
		t: t,
	}
}

// RecordConformanceOption implements the functional options pattern for RecordConformance.
type RecordConformanceOption func(*Info)

// Attestation represents a single conformance test attestation, written to disk as a YAML file.
type Attestation struct {
	Section string `json:"section"`
	Info    Info   `json:"info"`
}

// RecordConformance records that a specific conformance test was passed by writing an attestation file.
func (h *Reporter) RecordConformance(section string, test string, opt ...RecordConformanceOption) {
	info := Info{
		ID:     test,
		Status: "Implemented",
	}

	for _, o := range opt {
		o(&info)
	}

	evidencePath := path.Join("tests", h.t.Name(), "output.html")
	info.Evidence = append(info.Evidence, evidencePath)

	h.t.Logf("Conformance %v/%q passed: %+v", section, info.ID, info)

	attestation := Attestation{
		Section: section,
		Info:    info,
	}

	outputFile := attestationFile(h.t)
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		h.t.Errorf("failed to create attestation directory %q: %v", outputDir, err)
		return
	}

	b, err := yaml.Marshal(attestation)
	if err != nil {
		h.t.Errorf("failed to marshal attestation: %v", err)
		return
	}

	if err := os.WriteFile(outputFile, b, 0644); err != nil {
		h.t.Errorf("failed to write attestation file %q: %v", outputFile, err)
		return
	}
}

// attestationFile returns the path to write the ai conformance attestaion file.
func attestationFile(t Testing) string {
	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}
	testName := t.Name()
	return filepath.Join(artifactsDir, "tests", testName, "ai-conformance.yaml")
}

// Info represents the details of a conformance test result, including its ID, status, evidence, and any additional notes.
type Info struct {
	ID       string   `json:"id"`
	Status   string   `json:"status,omitempty"`   // Implemented, Not Implemented, Partially Implemented, N/A
	Evidence []string `json:"evidence,omitempty"` // List of URLs or references to documentation/test results
	Notes    string   `json:"notes,omitempty"`    // Must provide a justification when status is N/A
}
