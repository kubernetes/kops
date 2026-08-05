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

func singleNodeUpAsset(location string) map[architectures.Architecture]*assets.MirroredAsset {
	return map[architectures.Architecture]*assets.MirroredAsset{
		architectures.ArchitectureAmd64: {
			Locations: []string{location},
			Hash:      hashing.MustFromString("9acf6a83b249649354bb15b04250fabce0f8ff2377f2f0d4788e3fdda3f572a3"),
		},
	}
}

func renderNodeUpScript(t *testing.T, script *NodeUpScript) string {
	t.Helper()
	resource, err := script.Build()
	if err != nil {
		t.Fatalf("building nodeup script: %v", err)
	}
	rendered, err := fi.ResourceAsString(resource)
	if err != nil {
		t.Fatalf("rendering nodeup script: %v", err)
	}
	verifyShellSyntax(t, rendered)
	return rendered
}

func Test_GCSDownload(t *testing.T) {
	// The authenticated GCS download is rendered only when the nodeup sources are gs:// URLs.
	gcsMarker := "Authorization: Bearer"
	standardMarker := "wget --compression=auto"

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
				NodeUpAssets:  singleNodeUpAsset(tc.location),
			}
			rendered := renderNodeUpScript(t, script)
			if got := strings.Contains(rendered, gcsMarker); got != tc.expectGCS {
				t.Errorf("authenticated GCS download rendered=%v, expected %v", got, tc.expectGCS)
			}
			if got := strings.Contains(rendered, standardMarker); got != !tc.expectGCS {
				t.Errorf("standard download commands rendered=%v, expected %v", got, !tc.expectGCS)
			}
		})
	}
}

func TestEscapeS3Location(t *testing.T) {
	for _, tc := range []struct {
		name      string
		location  string
		expected  string
		expectErr bool
	}{
		{
			name:     "plain path",
			location: "s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup",
			expected: "s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup",
		},
		{
			name:     "plus sign",
			location: "s3://artifact-bucket/kops/1.37.0+abcdef/linux/amd64/nodeup",
			expected: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
		},
		{
			name:     "existing escape",
			location: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
			expected: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
		},
		{
			name:     "reserved and unicode characters",
			location: "s3://artifact-bucket/space here/100%25/%3F%23/雪,nodeup",
			expected: "s3://artifact-bucket/space%20here/100%25/%3F%23/%E9%9B%AA%2Cnodeup",
		},
		{
			name:     "path separators",
			location: "s3://artifact-bucket/a//b/nodeup",
			expected: "s3://artifact-bucket/a//b/nodeup",
		},
		{
			name:      "invalid escape",
			location:  "s3://artifact-bucket/100%/nodeup",
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := escapeS3Location(tc.location)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected an error escaping %q", tc.location)
				}
				return
			}
			if err != nil {
				t.Fatalf("escaping %q: %v", tc.location, err)
			}
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func Test_S3Download(t *testing.T) {
	s3Marker := "--aws-sigv4"
	standardMarker := "wget --compression=auto"

	for _, tc := range []struct {
		name           string
		cloudProvider  string
		location       string
		expectedSource string
		expectS3       bool
	}{
		{
			name:           "s3 source",
			cloudProvider:  "aws",
			location:       "s3://artifact-bucket/kops/1.34.0+abcdef/linux/amd64/nodeup",
			expectedSource: "NODEUP_URL_AMD64=s3://artifact-bucket/kops/1.34.0%2Babcdef/linux/amd64/nodeup",
			expectS3:       true,
		},
		{
			name:          "default https source",
			cloudProvider: "aws",
			location:      "https://artifacts.k8s.io/binaries/kops/1.34.0/linux/amd64/nodeup",
			expectS3:      false,
		},
		{
			name:          "s3 source on another cloud provider",
			cloudProvider: "hetzner",
			location:      "s3://artifact-bucket/kops/1.34.0/linux/amd64/nodeup",
			expectS3:      false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := &NodeUpScript{
				CloudProvider: tc.cloudProvider,
				NodeUpAssets:  singleNodeUpAsset(tc.location),
				S3Region:      "us-east-2",
			}
			rendered := renderNodeUpScript(t, script)
			if got := strings.Contains(rendered, s3Marker); got != tc.expectS3 {
				t.Errorf("authenticated S3 download rendered=%v, expected %v", got, tc.expectS3)
			}
			if tc.expectS3 && !strings.Contains(rendered, "https://s3.us-east-2.amazonaws.com/") {
				t.Errorf("authenticated S3 download does not use the bucket region")
			}
			if tc.expectedSource != "" && !strings.Contains(rendered, tc.expectedSource) {
				t.Errorf("rendered script does not contain escaped source %q", tc.expectedSource)
			}
			if got := strings.Contains(rendered, standardMarker); got != !tc.expectS3 {
				t.Errorf("standard download commands rendered=%v, expected %v", got, !tc.expectS3)
			}
		})
	}
}

func TestEscapeAzureBlobLocation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		location  string
		expected  string
		expectErr bool
	}{
		{
			name:     "plain path",
			location: "azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup",
		},
		{
			name:     "plus sign",
			location: "azureblob://exampleaccount/assets/kops/1.37.0+abcdef/linux/amd64/nodeup",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
		},
		{
			name:     "existing escape",
			location: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup",
		},
		{
			name:      "missing account",
			location:  "azureblob:///assets/kops/nodeup",
			expectErr: true,
		},
		{
			name:      "missing container",
			location:  "azureblob://exampleaccount",
			expectErr: true,
		},
		{
			name:      "missing key",
			location:  "azureblob://exampleaccount/assets",
			expectErr: true,
		},
		{
			name:      "invalid escape",
			location:  "azureblob://exampleaccount/assets/100%/nodeup",
			expectErr: true,
		},
		{
			name:      "port in host",
			location:  "azureblob://exampleaccount:443/assets/kops/nodeup",
			expectErr: true,
		},
		{
			name:      "userinfo",
			location:  "azureblob://user@exampleaccount/assets/kops/nodeup",
			expectErr: true,
		},
		{
			name:      "query string",
			location:  "azureblob://exampleaccount/assets/kops/nodeup?sig=secret",
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := escapeBlobLocation(tc.location)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected an error escaping %q", tc.location)
				}
				return
			}
			if err != nil {
				t.Fatalf("escaping %q: %v", tc.location, err)
			}
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func Test_AzureBlobDownload(t *testing.T) {
	azureBlobMarker := "blob.core.windows.net"
	standardMarker := "wget --compression=auto"

	for _, tc := range []struct {
		name            string
		cloudProvider   string
		location        string
		expectedSource  string
		expectAzureBlob bool
	}{
		{
			name:            "azureblob source",
			cloudProvider:   "azure",
			location:        "azureblob://exampleaccount/assets/kops/1.34.0+abcdef/linux/amd64/nodeup",
			expectedSource:  "NODEUP_URL_AMD64=azureblob://exampleaccount/assets/kops/1.34.0%2Babcdef/linux/amd64/nodeup",
			expectAzureBlob: true,
		},
		{
			name:            "default https source",
			cloudProvider:   "azure",
			location:        "https://artifacts.k8s.io/binaries/kops/1.34.0/linux/amd64/nodeup",
			expectAzureBlob: false,
		},
		{
			name:            "azureblob source on another cloud provider",
			cloudProvider:   "hetzner",
			location:        "azureblob://exampleaccount/assets/kops/1.34.0/linux/amd64/nodeup",
			expectAzureBlob: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := &NodeUpScript{
				CloudProvider: tc.cloudProvider,
				NodeUpAssets:  singleNodeUpAsset(tc.location),
			}
			rendered := renderNodeUpScript(t, script)
			if got := strings.Contains(rendered, azureBlobMarker); got != tc.expectAzureBlob {
				t.Errorf("authenticated Azure Blob download rendered=%v, expected %v", got, tc.expectAzureBlob)
			}
			if tc.expectedSource != "" && !strings.Contains(rendered, tc.expectedSource) {
				t.Errorf("rendered script does not contain escaped source %q", tc.expectedSource)
			}
			if got := strings.Contains(rendered, standardMarker); got != !tc.expectAzureBlob {
				t.Errorf("standard download commands rendered=%v, expected %v", got, !tc.expectAzureBlob)
			}
		})
	}
}

func Test_AzureBlobDownloadRequiresPublicCloud(t *testing.T) {
	script := &NodeUpScript{
		CloudProvider: "azure",
		NodeUpAssets:  singleNodeUpAsset("azureblob://exampleaccount/assets/kops/1.34.0/linux/amd64/nodeup"),
	}

	// Environment names are case-insensitive, and AzureCloud is the CLI name of the public cloud.
	for _, publicCloud := range []string{"AzurePublicCloud", "azurepubliccloud", "AzureCloud"} {
		t.Setenv("AZURE_ENVIRONMENT", publicCloud)
		if _, err := script.Build(); err != nil {
			t.Errorf("building an azureblob:// nodeup script with AZURE_ENVIRONMENT=%s: %v", publicCloud, err)
		}
	}

	t.Setenv("AZURE_ENVIRONMENT", "AzureChinaCloud")
	if _, err := script.Build(); err == nil {
		t.Errorf("expected an error building an azureblob:// nodeup script in a non-public cloud")
	}
}

func Test_S3DownloadRequiresRegion(t *testing.T) {
	script := &NodeUpScript{
		CloudProvider: "aws",
		NodeUpAssets:  singleNodeUpAsset("s3://artifact-bucket/kops/1.34.0/linux/amd64/nodeup"),
	}
	if _, err := script.Build(); err == nil {
		t.Errorf("expected an error building an s3:// nodeup script without a resolved region")
	}
}

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
