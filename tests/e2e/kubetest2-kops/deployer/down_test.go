/*
Copyright 2026 The Kubernetes Authors.

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

package deployer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"
)

// Up() calls Down() to delete a cluster leaked by an earlier run, at a point where the job may
// not have exported KOPS_STATE_STORE. The cluster lookup must use the state store the deployer
// resolved, otherwise it fails and Down() returns as if there was no cluster to delete.
func TestDownLooksUpClusterInDeployerStateStore(t *testing.T) {
	t.Setenv("KOPS_STATE_STORE", "")

	// Stand-in for kops that records its environment and exits non-zero, as kops does when the
	// cluster is not found, so that Down() returns before it deletes anything.
	dir := t.TempDir()
	envFile := filepath.Join(dir, "env")
	kopsBinary := filepath.Join(dir, "kops")
	script := fmt.Sprintf("#!/bin/sh\nenv > %q\nexit 1\n", envFile)
	if err := os.WriteFile(kopsBinary, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(kopsBinary, 0o700); err != nil {
		t.Fatal(err)
	}

	d := &deployer{
		CloudProvider:  "azure",
		ClusterName:    "test-down.k8s.local",
		KopsBinaryPath: kopsBinary,
		BuildOptions:   &builder.BuildOptions{},
		ArtifactsDir:   dir,
	}
	// initialize() needs cloud credentials and job settings that do not matter for the lookup.
	d.doInit.Do(func() {})

	if err := d.Down(); err != nil {
		t.Fatalf("Down() = %v, expected nil when the cluster does not exist", err)
	}

	env, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("kops was not run: %v", err)
	}
	// Report only the state store: the rest of the environment may hold credentials.
	var actual []string
	for line := range strings.Lines(string(env)) {
		if value, ok := strings.CutPrefix(strings.TrimSuffix(line, "\n"), "KOPS_STATE_STORE="); ok {
			actual = append(actual, value)
		}
	}
	expected := []string{"azureblob://stkopsstatestore/cluster-state"}
	if !slices.Equal(actual, expected) {
		t.Errorf("cluster lookup KOPS_STATE_STORE = %q, expected %q", actual, expected)
	}
}
