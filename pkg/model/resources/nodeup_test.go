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
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/nodeup"
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

func Test_NodeUpScriptOCIAssetRegistry(t *testing.T) {
	renderScript := func(cloudProvider, baseURL string) string {
		script := &NodeUpScript{
			NodeUpAssets: map[architectures.Architecture]*assets.MirroredAsset{
				architectures.ArchitectureAmd64: {
					Locations: []string{baseURL + "/binaries/kops/1.35.0/linux/amd64/nodeup"},
					Hash:      hashing.MustFromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a"),
				},
				architectures.ArchitectureArm64: {
					Locations: []string{baseURL + "/binaries/kops/1.35.0/linux/arm64/nodeup"},
					Hash:      hashing.MustFromString("e525c28a65ff0ce4f95f9e730195b4e67fdcb15ceb1f36b5ad6921a8a4490c71"),
				},
			},
			BootConfig:    &nodeup.BootConfig{},
			CloudProvider: cloudProvider,
		}
		resource, err := script.Build()
		if err != nil {
			t.Fatalf("building nodeup script: %v", err)
		}
		rendered, err := fi.ResourceAsString(resource)
		if err != nil {
			t.Fatalf("rendering nodeup script: %v", err)
		}
		return rendered
	}

	rendered := renderScript("hetzner", "oci://registry.example.com/assets")
	for _, expected := range []string{
		"download-oci()",
		`if ! download-oci "${file}" "${hash}" "${url}"; then`,
		`"https://${registry}/v2/${repository}/blobs/sha256:${hash}"`,
		// Anonymous pulls may need a token from the challenge endpoint.
		`realm=$(`,
		// curl merges the token parameters into any query the realm already carries.
		`--get --data-urlencode "service=${service}"`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("expected the nodeup script to contain %q", expected)
		}
	}
	// All file assets are remapped to the OCI registry; the generic download commands are not needed.
	if strings.Contains(rendered, "commands=(") {
		t.Errorf("expected the nodeup script to not contain the generic download commands")
	}

	// The OCI download also replaces the GCE download commands.
	rendered = renderScript("gce", "oci://registry.example.com/assets")
	if !strings.Contains(rendered, "download-oci()") {
		t.Errorf("expected the nodeup script to contain the OCI download function")
	}
	if strings.Contains(rendered, "gcloud storage cp") {
		t.Errorf("expected the nodeup script to not contain the GCE download commands")
	}

	// Without an oci:// fileRepository, the OCI download function is not emitted.
	rendered = renderScript("gce", "https://artifacts.k8s.io")
	if strings.Contains(rendered, "download-oci") {
		t.Errorf("expected the nodeup script to not contain the OCI download function")
	}
	if !strings.Contains(rendered, "gcloud storage cp") {
		t.Errorf("expected the nodeup script to contain the GCE download commands")
	}
}
