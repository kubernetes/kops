/*
Copyright 2020 The Kubernetes Authors.

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

package builder

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

type BuildOptions struct {
	KopsRoot      string `flag:"-"`
	StageLocation string `flag:"-"`
}

// BuildResults describes the outcome of a successful build.
type BuildResults struct {
	KopsBaseURL string
}

// Build will build the kops artifacts and publish them to the stage location
func (b *BuildOptions) Build() (*BuildResults, error) {
	// We expect to upload to a subdirectory with a version identifier
	gcsLocation := b.StageLocation
	if !strings.HasSuffix(gcsLocation, "/") {
		gcsLocation += "/"
	}
	cmd := exec.Command("make", "gcs-publish-ci")
	cmd.SetEnv(
		fmt.Sprintf("HOME=%v", os.Getenv("HOME")),
		fmt.Sprintf("PATH=%v", os.Getenv("PATH")),
		fmt.Sprintf("GCS_LOCATION=%v", gcsLocation),
		fmt.Sprintf("GOPATH=%v", os.Getenv("GOPATH")),
	)
	cmd.SetDir(b.KopsRoot)
	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Get the full path (including subdirectory) that we uploaded to
	// It is written by gcs-publish-ci to .build/upload/latest-ci.txt
	latestPath := filepath.Join(b.KopsRoot, ".build", "upload", "latest-ci.txt")
	kopsBaseURL, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", latestPath, err)
	}
	u, err := url.Parse(strings.TrimSpace(string(kopsBaseURL)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %q from file %q: %w", string(kopsBaseURL), latestPath, err)
	}
	u.Path = strings.ReplaceAll(u.Path, "//", "/")
	results := &BuildResults{
		KopsBaseURL: u.String(),
	}

	// Write some meta files so that other tooling can know e.g. KOPS_BASE_URL
	metaDir := filepath.Join(b.KopsRoot, ".kubetest2")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to Mkdir(%q): %w", metaDir, err)
	}
	p := filepath.Join(metaDir, "kops-base-url")
	if err := os.WriteFile(p, []byte(results.KopsBaseURL), 0o644); err != nil {
		return nil, fmt.Errorf("failed to WriteFile(%q): %w", p, err)
	}
	klog.Infof("wrote file %q with %q", p, results.KopsBaseURL)

	return results, nil
}
