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

package tester

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/artifacts"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// AcquireKubectl obtains kubectl and places it in rundir
// If a kubectl already exists in rundir, it will be reused.
func (t *Tester) AcquireKubectl() error {

	// first, get the name of the latest release (e.g. v1.20.0-alpha.0)
	if t.TestPackageVersion == "" {
		cmd := exec.Command(
			"gsutil",
			"cat",
			fmt.Sprintf("gs://%s/%s/%v", t.TestPackageBucket, t.TestPackageDir, t.TestPackageMarker),
		)
		lines, err := exec.OutputLines(cmd)
		if err != nil {
			return fmt.Errorf("failed to get latest release name: %w", err)
		}
		if len(lines) == 0 {
			return fmt.Errorf("getting latest release name had no output")
		}
		t.TestPackageVersion = lines[0]

		klog.Infof("Kubectl package version was not specified. Defaulting to version from %s: %s", t.TestPackageMarker, t.TestPackageVersion)
	}

	if err := t.ensureKubectl(); err != nil {
		return err
	}
	return nil
}

// ensureKubectl checks if the kubectl binary already exists
// and verifies the hashes else downloads it from GCS
func (t *Tester) ensureKubectl() error {
	if _, err := os.Stat(KubectlPath()); err == nil {
		klog.V(0).Infof("Found existing kubectl at %s", KubectlPath())
		if err := t.compareSHA(); err == nil {
			klog.V(0).Infof("Validated hash for existing kubectl binary at %v", KubectlPath())
			return nil
		}
		klog.Warning(err)
	}

	args := []string{
		"gsutil", "cp",
		t.kubectlGSLocation(),
		KubectlPath(),
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download kubectl binary for release %s: %s", t.TestPackageVersion, err)
	}
	os.Chmod(KubectlPath(), os.FileMode(0o700))
	return nil
}

func (t *Tester) compareSHA() error {
	cmd := exec.Command("gsutil", "cat", t.kubectlGSLocation()+".sha256")
	expectedSHABytes, err := exec.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to get sha256 for kubectl binary for release %s: %s", t.TestPackageVersion, err)
	}
	expectedSHA := strings.TrimSuffix(string(expectedSHABytes), "\n")
	actualSHA, err := sha256sum(KubectlPath())
	if err != nil {
		return fmt.Errorf("failed to compute sha256 for %q: %v", KubectlPath(), err)
	}
	if actualSHA != expectedSHA {
		return fmt.Errorf("sha256 does not match")
	}
	return nil
}

func (t *Tester) kubectlGSLocation() string {
	return fmt.Sprintf(
		"gs://%s/%s/%s/bin/%s/%s/kubectl",
		t.TestPackageBucket,
		t.TestPackageDir,
		t.TestPackageVersion,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func sha256sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func KubectlPath() string {
	return artifacts.RunDir() + "/kubectl"
}
