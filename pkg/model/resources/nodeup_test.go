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
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	for _, appendedSuffix := range []string{`${url}.xz`, `${url#gs://}.xz`, `${url#s3://}.xz`, `${rest}.xz`} {
		if strings.Contains(rendered, appendedSuffix) {
			t.Errorf("rendered script appends an xz suffix to an asset URL: %s", appendedSuffix)
		}
	}
	return rendered
}

func Test_GCSDownload(t *testing.T) {
	// The authenticated GCS download is rendered only when the nodeup sources are gs:// URLs.
	gcsMarker := "Authorization: Bearer"
	standardMarker := `--retry-delay 10 "${url}"`

	for _, tc := range []struct {
		name      string
		location  string
		expectGCS bool
	}{
		{
			name:      "gs source",
			location:  "gs://artifact-bucket/kops/1.37.0/linux/amd64/nodeup.xz",
			expectGCS: true,
		},
		{
			name:      "default https source",
			location:  "https://artifacts.k8s.io/binaries/kops/1.37.0/linux/amd64/nodeup.xz",
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
			location: "s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup.xz",
			expected: "s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup.xz",
		},
		{
			name:     "plus sign",
			location: "s3://artifact-bucket/kops/1.37.0+abcdef/linux/amd64/nodeup.xz",
			expected: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
		},
		{
			name:     "existing escape",
			location: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
			expected: "s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
		},
		{
			name:     "reserved and unicode characters",
			location: "s3://artifact-bucket/space here/100%25/%3F%23/雪,nodeup.xz",
			expected: "s3://artifact-bucket/space%20here/100%25/%3F%23/%E9%9B%AA%2Cnodeup.xz",
		},
		{
			name:     "path separators",
			location: "s3://artifact-bucket/a//b/nodeup.xz",
			expected: "s3://artifact-bucket/a//b/nodeup.xz",
		},
		{
			name:      "invalid escape",
			location:  "s3://artifact-bucket/100%/nodeup.xz",
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
	standardMarker := `--retry-delay 10 "${url}"`

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
			location:       "s3://artifact-bucket/kops/1.37.0+abcdef/linux/amd64/nodeup.xz",
			expectedSource: "NODEUP_URL_AMD64=s3://artifact-bucket/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
			expectS3:       true,
		},
		{
			name:          "default https source",
			cloudProvider: "aws",
			location:      "https://artifacts.k8s.io/binaries/kops/1.37.0/linux/amd64/nodeup.xz",
			expectS3:      false,
		},
		{
			name:          "s3 source on another cloud provider",
			cloudProvider: "hetzner",
			location:      "s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup.xz",
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
			location: "azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup.xz",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup.xz",
		},
		{
			name:     "plus sign",
			location: "azureblob://exampleaccount/assets/kops/1.37.0+abcdef/linux/amd64/nodeup.xz",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
		},
		{
			name:     "existing escape",
			location: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
			expected: "azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
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
	standardMarker := `--retry-delay 10 "${url}"`

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
			location:        "azureblob://exampleaccount/assets/kops/1.37.0+abcdef/linux/amd64/nodeup.xz",
			expectedSource:  "NODEUP_URL_AMD64=azureblob://exampleaccount/assets/kops/1.37.0%2Babcdef/linux/amd64/nodeup.xz",
			expectAzureBlob: true,
		},
		{
			name:            "default https source",
			cloudProvider:   "azure",
			location:        "https://artifacts.k8s.io/binaries/kops/1.37.0/linux/amd64/nodeup.xz",
			expectAzureBlob: false,
		},
		{
			name:            "azureblob source on another cloud provider",
			cloudProvider:   "hetzner",
			location:        "azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup.xz",
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
		NodeUpAssets:  singleNodeUpAsset("azureblob://exampleaccount/assets/kops/1.37.0/linux/amd64/nodeup.xz"),
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
		NodeUpAssets:  singleNodeUpAsset("s3://artifact-bucket/kops/1.37.0/linux/amd64/nodeup.xz"),
	}
	if _, err := script.Build(); err == nil {
		t.Errorf("expected an error building an s3:// nodeup script without a resolved region")
	}
}

func TestDownloadOrBustUsesTransientArchive(t *testing.T) {
	requireNodeUpDownloadCommands(t)
	downloadFunctions := renderNodeUpDownloadFunctions(t)

	sourceDir := t.TempDir()
	workDir := t.TempDir()
	archive, hash := buildNodeUpArchive(t, "nodeup contents")
	sourcePath := filepath.Join(sourceDir, "nodeup.xz")
	if err := os.WriteFile(sourcePath, archive, 0o600); err != nil {
		t.Fatal(err)
	}
	sourceURL := (&url.URL{Scheme: "file", Path: sourcePath}).String()

	runNodeUpDownload(t, downloadFunctions, workDir, hash, sourceURL)
	assertInstalledNodeUp(t, workDir, "nodeup contents")

	if err := os.WriteFile(filepath.Join(workDir, "nodeup"), []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}

	runNodeUpDownload(t, downloadFunctions, workDir, hash, sourceURL)
	assertInstalledNodeUp(t, workDir, "nodeup contents")
}

func TestDownloadOrBustValidatesBeforeDecompression(t *testing.T) {
	xz := requireNodeUpDownloadCommands(t)
	downloadFunctions := renderNodeUpDownloadFunctions(t)

	sourceDir := t.TempDir()
	workDir := t.TempDir()
	badPath := filepath.Join(sourceDir, "bad.xz")
	if err := os.WriteFile(badPath, []byte("not the expected archive"), 0o600); err != nil {
		t.Fatal(err)
	}
	archive, hash := buildNodeUpArchive(t, "nodeup contents")
	goodPath := filepath.Join(sourceDir, "good.xz")
	if err := os.WriteFile(goodPath, archive, 0o600); err != nil {
		t.Fatal(err)
	}

	fakeBin := t.TempDir()
	xzLog := filepath.Join(t.TempDir(), "xz.log")
	xzWrapper := filepath.Join(fakeBin, "xz")
	wrapper := fmt.Sprintf("#!/bin/sh\nprintf 'xz\\n' >> %q\nexec %q \"$@\"\n", xzLog, xz)
	if err := os.WriteFile(xzWrapper, []byte(wrapper), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(xzWrapper, 0o700); err != nil {
		t.Fatal(err)
	}

	urls := strings.Join([]string{
		(&url.URL{Scheme: "file", Path: badPath}).String(),
		(&url.URL{Scheme: "file", Path: goodPath}).String(),
	}, ",")
	runNodeUpDownloadWithEnv(t, downloadFunctions, workDir, hash, urls, "PATH="+fakeBin+":"+os.Getenv("PATH"))
	assertInstalledNodeUp(t, workDir, "nodeup contents")

	logData, err := os.ReadFile(xzLog)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(logData), "xz\n"); got != 1 {
		t.Errorf("decompression attempts = %d, want 1", got)
	}
}

func renderNodeUpDownloadFunctions(t *testing.T) string {
	t.Helper()
	rendered := renderNodeUpScript(t, &NodeUpScript{
		NodeUpAssets: singleNodeUpAsset("file:///nodeup.xz"),
	})
	start := strings.Index(rendered, "download-or-bust() {")
	end := strings.Index(rendered, "function download-release() {")
	if start == -1 || end == -1 || start >= end {
		t.Fatal("download functions not found in rendered nodeup script")
	}
	return rendered[start:end]
}

func buildNodeUpArchive(t *testing.T, contents string) ([]byte, string) {
	t.Helper()
	xz := requireNodeUpDownloadCommands(t)
	cmd := exec.Command(xz, "-6", "-T1", "-c")
	cmd.Stdin = strings.NewReader(contents)
	archive, err := cmd.Output()
	if err != nil {
		t.Fatalf("creating xz archive: %v", err)
	}
	hash := sha256.Sum256(archive)
	return archive, fmt.Sprintf("%x", hash)
}

func requireNodeUpDownloadCommands(t *testing.T) string {
	t.Helper()
	for _, command := range []string{"bash", "curl", "xz"} {
		path, err := exec.LookPath(command)
		if err != nil {
			t.Skipf("%s not found in PATH", command)
		}
		if command == "xz" {
			return path
		}
	}
	panic("unreachable")
}

func runNodeUpDownload(t *testing.T, functions, workDir, hash, urls string) {
	t.Helper()
	runNodeUpDownloadWithEnv(t, functions, workDir, hash, urls)
}

func runNodeUpDownloadWithEnv(t *testing.T, functions, workDir, hash, urls string, env ...string) {
	t.Helper()
	script := "#!/bin/bash\nset -o errexit\nset -o nounset\nset -o pipefail\n\n" + functions + `
cd "$1"
download-or-bust nodeup "$2" "$3"
`
	scriptPath := filepath.Join(t.TempDir(), "nodeup-download.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(scriptPath, 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, scriptPath, workDir, hash, urls)
	cmd.Env = environmentWith(env...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("running nodeup download functions: %v\n%s", err, output)
	}
}

func environmentWith(overrides ...string) []string {
	env := os.Environ()
	for _, override := range overrides {
		key, _, _ := strings.Cut(override, "=")
		filtered := env[:0]
		for _, entry := range env {
			entryKey, _, _ := strings.Cut(entry, "=")
			if entryKey != key {
				filtered = append(filtered, entry)
			}
		}
		env = filtered
	}
	return append(env, overrides...)
}

func assertInstalledNodeUp(t *testing.T, workDir, contents string) {
	t.Helper()
	nodeup, err := os.ReadFile(filepath.Join(workDir, "nodeup"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(nodeup); got != contents {
		t.Errorf("installed nodeup = %q, want %q", got, contents)
	}
	info, err := os.Stat(filepath.Join(workDir, "nodeup"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("installed nodeup is not executable")
	}
	if _, err := os.Stat(filepath.Join(workDir, "nodeup.xz")); !os.IsNotExist(err) {
		t.Errorf("compressed archive was not removed")
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
