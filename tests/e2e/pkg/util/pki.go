/*
Copyright 2022 The Kubernetes Authors.

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

package util

import (
	"errors"
	"os"
	"path/filepath"

	"sigs.k8s.io/kubetest2/pkg/exec"
)

// CreateSSHKeyPair creates a key pair in a temp directory
// and returns the paths to the private and public keys respectively.
// The file paths are deterministic from the clusterName.
func CreateSSHKeyPair(clusterName string) (string, string, error) {
	tmp := os.TempDir()
	dir := filepath.Join(tmp, "kops", clusterName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", "", err
	}
	privateKeyPath := filepath.Join(dir, "id_ed25519")
	publicKeyPath := filepath.Join(dir, "id_ed25519.pub")

	if _, err := os.Stat(privateKeyPath); errors.Is(err, os.ErrNotExist) {
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-q", "-f", privateKeyPath)
		exec.InheritOutput(cmd)
		if err := cmd.Run(); err != nil {
			return "", "", err
		}
	}

	return publicKeyPath, privateKeyPath, nil
}
