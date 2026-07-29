/*
Copyright 2021 The Kubernetes Authors.

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

package resources

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

func Test_NodeUpTabs(t *testing.T) {
	for i, line := range strings.Split(nodeUpTemplate, "\n") {
		if strings.Contains(line, "\t") {
			t.Errorf("NodeUpTemplate contains unexpected character %q on line %d: %q", "\t", i, line)
		}
	}
}

func Test_GCSDownload(t *testing.T) {
	// The authenticated GCS download is rendered only when the nodeup sources are gs:// URLs.
	gcsMarker := "Authorization: Bearer"
	standardMarker := "wget --compression=auto"

	nodeUpAssets := func(location string) map[architectures.Architecture]*assets.MirroredAsset {
		return map[architectures.Architecture]*assets.MirroredAsset{
			architectures.ArchitectureAmd64: {
				Locations: []string{location},
				Hash:      hashing.MustFromString("9acf6a83b249649354bb15b04250fabce0f8ff2377f2f0d4788e3fdda3f572a3"),
			},
		}
	}

	for _, tc := range []struct {
		name      string
		location  string
		expectGCS bool
	}{
		{
			name:      "gs source",
			location:  "gs://artifact-bucket/kops/1.34.0/linux/amd64/nodeup",
			expectGCS: true,
		},
		{
			name:      "default https source",
			location:  "https://artifacts.k8s.io/binaries/kops/1.34.0/linux/amd64/nodeup",
			expectGCS: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := &NodeUpScript{
				CloudProvider: "gce",
				NodeUpAssets:  nodeUpAssets(tc.location),
			}
			resource, err := script.Build()
			if err != nil {
				t.Fatalf("building nodeup script: %v", err)
			}
			rendered, err := fi.ResourceAsString(resource)
			if err != nil {
				t.Fatalf("rendering nodeup script: %v", err)
			}
			if got := strings.Contains(rendered, gcsMarker); got != tc.expectGCS {
				t.Errorf("authenticated GCS download rendered=%v, expected %v", got, tc.expectGCS)
			}
			if got := strings.Contains(rendered, standardMarker); got != !tc.expectGCS {
				t.Errorf("standard download commands rendered=%v, expected %v", got, !tc.expectGCS)
			}
			verifyShellSyntax(t, rendered)
		})
	}
}

// verifyShellSyntax parses the rendered script with bash, as the GCS download has no golden file.
func verifyShellSyntax(t *testing.T, script string) {
	t.Helper()

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH")
	}

	path := filepath.Join(t.TempDir(), "nodeup.sh")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatalf("writing rendered script: %v", err)
	}
	if output, err := exec.Command(bash, "-n", path).CombinedOutput(); err != nil {
		t.Errorf("rendered script is not valid bash: %v\n%s", err, output)
	}
}
