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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/kops/tests/e2e/scenarios/ai-conformance/validators/conformance"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}
	reportPath := filepath.Join(artifactsDir, "ai-conformance.yaml")

	kubernetesVersion := ""
	kopsVersion := ""

	if kubernetesVersion == "" {
		v, err := getKubernetesVersion(ctx)
		if err != nil {
			return fmt.Errorf("getting kubernetes version: %v", err)
		}
		kubernetesVersion = v
	}

	if kopsVersion == "" {
		v, err := getKopsVersion(ctx)
		if err != nil {
			return fmt.Errorf("getting kOps version: %v", err)
		}
		kopsVersion = v
	}

	metadata := conformance.Metadata{
		KubernetesVersion:   kubernetesVersion,
		PlatformName:        "kOps",
		PlatformVersion:     kopsVersion,
		VendorName:          "kOps Project",
		WebsiteURL:          "https://kops.sigs.k8s.io/",
		RepoURL:             "https://github.com/kubernetes/kops",
		DocumentationURL:    "https://kops.sigs.k8s.io/",
		ProductLogoURL:      "https://raw.githubusercontent.com/kubernetes/kops/refs/heads/master/docs/img/logo.svg",
		Description:         "Kubernetes Operations (kOps) - Production Grade k8s Installation, Upgrades and Management",
		ContactEmailAddress: "sig-cluster-lifecycle@kubernetes.io",
	}

	if err := conformance.WriteReport(artifactsDir, metadata); err != nil {
		panic(fmt.Sprintf("Failed to write conformance report: %v", err))
	}

	fmt.Printf("Conformance report written to: %s\n", reportPath)
	return nil
}

// getKubernetesVersion retrieves the Kubernetes version by running "kubectl version" and parsing the output.
func getKubernetesVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute kubectl version: %v", err)
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		suffix, ok := strings.CutPrefix(line, "Server Version:")
		if ok {
			return strings.TrimSpace(suffix), nil
		}
	}
	return "", fmt.Errorf("server version not found in kubectl output: %q", string(output))
}

// getKopsVersion retrieves the kOps version by running "kops version --short".
func getKopsVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kops", "version", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute kops version: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
