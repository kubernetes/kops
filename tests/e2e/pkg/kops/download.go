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

package kops

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"k8s.io/kops/tests/e2e/pkg/util"
)

// DownloadKops will download the kops binary from the version marker URL
// KOPS_BASE_URL will be set to the contents of marker + marker URL
// Example markerURL: https://storage.googleapis.com/k8s-staging-kops/kops/latest.txt
func DownloadKops(markerURL, downloadPath, kopsVersion string) (string, error) {
	var b bytes.Buffer
	var kopsBaseURL string
	if markerURL == "" && kopsVersion != "" {
		kopsBaseURL = fmt.Sprintf("https://artifacts.k8s.io/binaries/kops/%s", kopsVersion)
	}
	if markerURL != "" && kopsVersion == "" {
		if err := util.HTTPGETWithHeaders(markerURL, nil, &b); err != nil {
			return "", err
		}
		baseURL := markerURL
		if slash := strings.LastIndex(markerURL, "/"); slash != -1 {
			baseURL = markerURL[:slash]
		}
		kopsBaseURL = baseURL + "/" + strings.TrimSpace(b.String())
	}

	kopsFile, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}

	kopsURL := fmt.Sprintf("%v/%v/%v/kops", kopsBaseURL, runtime.GOOS, runtime.GOARCH)
	if err := util.HTTPGETWithHeaders(kopsURL, nil, kopsFile); err != nil {
		return "", err
	}
	if err := kopsFile.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(kopsFile.Name(), 0o755); err != nil {
		return "", err
	}
	return kopsBaseURL, nil
}
