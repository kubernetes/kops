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
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"

	"k8s.io/kops/tests/e2e/pkg/util"
)

// DownloadKops will download the kops binary from the version marker URL
// Returning the URL to use for KOPS_BASE_URL
// Markers contain either an absolute artifact URL or a version relative to the marker's directory.
// Example markerURL: https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt
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
		baseURL, err := kopsBaseURLFromMarker(markerURL, b.String())
		if err != nil {
			return "", err
		}
		kopsBaseURL = baseURL
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

// kopsBaseURLFromMarker preserves legacy absolute URLs and resolves version-only markers relative
// to their directory.
func kopsBaseURLFromMarker(markerURL, contents string) (string, error) {
	contents = strings.TrimSpace(contents)
	marker, err := url.Parse(contents)
	if err != nil {
		return "", fmt.Errorf("invalid kops marker %q: %w", markerURL, err)
	}
	if marker.IsAbs() {
		return contents, nil
	}
	if _, err := semver.Parse(strings.TrimPrefix(contents, "v")); err != nil {
		return "", fmt.Errorf("invalid kops version in marker %q: %w", markerURL, err)
	}
	baseURL, err := url.Parse(markerURL)
	if err != nil {
		return "", fmt.Errorf("invalid kops marker URL %q: %w", markerURL, err)
	}
	// Resolve against the requested URL so CDN users also fetch artifacts through the CDN.
	return baseURL.ResolveReference(marker).String(), nil
}
