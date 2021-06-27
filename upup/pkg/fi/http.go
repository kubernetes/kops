/*
Copyright 2019 The Kubernetes Authors.

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

package fi

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

func DownloadURL(url string, dest string, hash *hashing.Hash) (*hashing.Hash, error) {
	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if match {
			return hash, nil
		}
	}

	dirMode := os.FileMode(0755)
	err := downloadURLAlways(url, dest, dirMode)
	if err != nil {
		return nil, err
	}

	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if !match {
			return nil, fmt.Errorf("downloaded from %q but hash did not match expected %q", url, hash)
		}
	} else {
		hash, err = hashing.HashAlgorithmSHA256.HashFile(dest)
		if err != nil {
			return nil, err
		}
	}

	return hash, nil
}

func downloadURLAlways(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	args := []string{"curl", "-v", "-f", "--compressed", "-Lo", destPath, "--connect-timeout", "20", "--retry", "6", "--retry-delay", "10", url}

	human := strings.Join(args, " ")
	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error downloading file '%s': %v: %s", human, err, string(output))
	}

	klog.Infof("output for %q:\n%s", url, string(output))
	return nil
}
