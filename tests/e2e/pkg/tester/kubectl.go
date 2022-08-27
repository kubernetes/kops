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
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/artifacts"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// AcquireKubectl obtains kubectl and places it in rundir
// If a kubectl already exists in rundir, it will be reused.
func (t *Tester) AcquireKubectl() error {
	if _, err := os.Stat(KubectlPath()); errors.Is(err, os.ErrNotExist) {
		klog.Infof("found kubectl in %s. Will reuse it.", KubectlPath())
	}

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

	clientTar := fmt.Sprintf("kubernetes-client-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	downloadDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get user cache directory: %w", err)
	}

	downloadPath := filepath.Join(downloadDir, clientTar)

	if err := t.ensureClientTar(downloadPath, clientTar); err != nil {
		return err
	}

	return t.extractBinaries(downloadPath)
}

func (t *Tester) extractBinaries(downloadPath string) error {
	// finally, search for the client package and extract it
	f, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded tar at %s: %w", downloadPath, err)
	}
	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	tarReader := tar.NewReader(gzf)

	// this is the expected path of the package inside the tar
	// it will be extracted to kubectlDir in the loop
	kubectlPackagePath := "kubernetes/client/bin/kubectl"
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error during tar read: %w", err)
		}
		if header.Name == kubectlPackagePath {
			kubectlPath := KubectlPath()
			outFile, err := os.Create(kubectlPath)
			if err != nil {
				return fmt.Errorf("error creating file at %s: %w", kubectlPath, err)
			}
			defer outFile.Close()

			if err := outFile.Chmod(0o700); err != nil {
				return fmt.Errorf("failed to make %s executable: %w", kubectlPath, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("error reading data from tar with header name %s: %w", header.Name, err)
			}
			return nil
		}
	}
	return fmt.Errorf("failed to find %s in %s", kubectlPackagePath, downloadPath)
}

// ensureClientTar checks if the kubernetes client tarball already exists
// and verifies the hashes
// else downloads it from GCS
func (t *Tester) ensureClientTar(downloadPath, clientTar string) error {
	if _, err := os.Stat(downloadPath); err == nil {
		klog.V(0).Infof("Found existing tar at %v", downloadPath)
		if err := t.compareSHA(downloadPath, clientTar); err == nil {
			klog.V(0).Infof("Validated hash for existing tar at %v", downloadPath)
			return nil
		}
		klog.Warning(err)
	}

	args := []string{
		"gsutil", "cp",
		fmt.Sprintf(
			"gs://%s/%s/%s/%s",
			t.TestPackageBucket,
			t.TestPackageDir,
			t.TestPackageVersion,
			clientTar,
		),
		downloadPath,
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	exec.InheritOutput(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download release tar %s for release %s: %s", clientTar, t.TestPackageVersion, err)
	}
	return nil
}

func (t *Tester) compareSHA(downloadPath string, clientTar string) error {
	cmd := exec.Command("gsutil", "cat",
		fmt.Sprintf(
			"gs://%s/%s/%s/%s",
			t.TestPackageBucket,
			t.TestPackageDir,
			t.TestPackageVersion,
			clientTar+".sha256",
		),
	)
	expectedSHABytes, err := exec.Output(cmd)
	if err != nil {
		return fmt.Errorf("failed to get sha256 for release tar %s for release %s: %s", clientTar, t.TestPackageVersion, err)
	}
	expectedSHA := strings.TrimSuffix(string(expectedSHABytes), "\n")
	actualSHA, err := sha256sum(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to compute sha256 for %q: %v", downloadPath, err)
	}
	if actualSHA != expectedSHA {
		return fmt.Errorf("sha256 does not match")
	}
	return nil
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
