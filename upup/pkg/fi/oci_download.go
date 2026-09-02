/*
Copyright 2026 The Kubernetes Authors.

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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/kops/util/pkg/hashing"
)

func openOCIAsset(ctx context.Context, location *url.URL, expected *hashing.Hash) (io.ReadCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	reader, err := openOCIAssetWithTransport(ctx, newDownloadHTTPClient().Transport, location, expected)
	if err != nil {
		cancel()
		return nil, err
	}
	return &cancelOnCloseReadCloser{ReadCloser: reader, cancel: cancel}, nil
}

func openOCIAssetWithTransport(ctx context.Context, transport http.RoundTripper, location *url.URL, expected *hashing.Hash) (io.ReadCloser, error) {
	if expected == nil || expected.Algorithm != hashing.HashAlgorithmSHA256 {
		return nil, fmt.Errorf("OCI asset %q requires a SHA-256", location)
	}
	repository, err := ociRepository(location)
	if err != nil {
		return nil, err
	}
	ref, err := name.NewDigest(location.Host+"/"+repository+"@sha256:"+expected.Hex(), name.StrictValidation)
	if err != nil {
		return nil, fmt.Errorf("invalid OCI asset %q: %w", location, err)
	}
	layer, err := remote.Layer(ref,
		remote.WithContext(ctx),
		remote.WithAuth(authn.Anonymous),
		remote.WithTransport(&httpsOnlyOCITransport{inner: transport}),
	)
	if err != nil {
		return nil, fmt.Errorf("opening OCI asset %q: %w", location, err)
	}
	reader, err := layer.Compressed()
	if err != nil {
		return nil, fmt.Errorf("downloading OCI asset %q: %w", location, err)
	}
	return reader, nil
}

type httpsOnlyOCITransport struct {
	inner http.RoundTripper
}

func (t *httpsOnlyOCITransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" {
		return nil, fmt.Errorf("OCI registry request must use HTTPS: %s", request.URL)
	}
	return t.inner.RoundTrip(request)
}

func ociRepository(location *url.URL) (string, error) {
	if location.Scheme != "oci" || location.Host == "" || location.User != nil || location.RawQuery != "" || location.Fragment != "" {
		return "", fmt.Errorf("invalid OCI asset location %q", location)
	}
	reference := strings.Trim(location.Path, "/")
	separator := strings.LastIndex(reference, ":")
	if separator <= strings.LastIndex(reference, "/") || separator == len(reference)-1 {
		return "", fmt.Errorf("OCI asset location %q must include a tag", location)
	}
	return reference[:separator], nil
}

// OCIAssetFamily returns the final repository component, which identifies the asset family.
func OCIAssetFamily(location *url.URL) (string, error) {
	repository, err := ociRepository(location)
	if err != nil {
		return "", err
	}
	family := path.Base(repository)
	if family == "." || family == "/" {
		return "", fmt.Errorf("OCI asset location %q has no repository", location)
	}
	return family, nil
}
