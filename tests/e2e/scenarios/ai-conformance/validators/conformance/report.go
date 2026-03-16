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
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

// Report is the conformance report data, including metadata and test results.
type Report struct {
	Metadata Metadata          `json:"metadata"`
	Spec     map[string][]Info `json:"spec"`
}

// Metadata is the header for the conformance report, including Kubernetes version, platform information, vendor details, and contact information.
type Metadata struct {
	KubernetesVersion   string `json:"kubernetesVersion,omitempty"`
	PlatformName        string `json:"platformName,omitempty"`
	PlatformVersion     string `json:"platformVersion,omitempty"`
	VendorName          string `json:"vendorName,omitempty"`
	WebsiteURL          string `json:"websiteUrl,omitempty"`
	RepoURL             string `json:"repoUrl,omitempty"`
	DocumentationURL    string `json:"documentationUrl,omitempty"`
	ProductLogoURL      string `json:"productLogoUrl,omitempty"`
	Description         string `json:"description,omitempty"`
	ContactEmailAddress string `json:"contactEmailAddress,omitempty"`
}

// WriteReport scans the artifacts directory for attestation files and writes
// the final conformance report to the artifacts directory.
func WriteReport(artifactsDir string, metadata Metadata) error {
	report := &Report{
		Metadata: metadata,
		Spec:     make(map[string][]Info),
	}

	// Walk the artifactsDir tree, looking for "ai-conformance.yaml"
	if err := filepath.WalkDir(artifactsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking artifacts directory %q: %w", artifactsDir, err)
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "ai-conformance.yaml" {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading attestation file %q: %w", path, err)
		}

		var attestation Attestation
		if err := yaml.Unmarshal(b, &attestation); err != nil {
			return fmt.Errorf("parsing attestation file %q: %w", path, err)
		}

		report.Spec[attestation.Section] = append(report.Spec[attestation.Section], attestation.Info)
		return nil
	}); err != nil {
		return fmt.Errorf("scanning artifacts directory for attestations: %w", err)
	}

	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("creating artifacts directory %q: %w", artifactsDir, err)
	}

	reportPath := filepath.Join(artifactsDir, "ai-conformance.yaml")
	b, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshaling conformance report: %w", err)
	}
	if err := os.WriteFile(reportPath, b, 0644); err != nil {
		return fmt.Errorf("writing conformance report to %q: %w", reportPath, err)
	}

	return nil
}
