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

package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"k8s.io/kops/util/pkg/hashing"
)

// findOCIHash resolves the exact file bytes, not the manifest or config digest.
func findOCIHash(location *url.URL) (*hashing.Hash, error) {
	if cached, found := downloadedFileHashes.Load(location.String()); found {
		return cached.(*hashing.Hash), nil
	}
	ref, err := name.NewTag(location.Host+location.Path, name.StrictValidation)
	if err != nil {
		return nil, fmt.Errorf("invalid OCI asset %q: %w", location, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	descriptor, err := remote.Get(ref,
		remote.WithContext(ctx),
		remote.WithAuth(authn.Anonymous),
		remote.WithTransport(&HTTPSOnlyTransport{Inner: remote.DefaultTransport}),
	)
	if err != nil {
		return nil, fmt.Errorf("reading staged OCI asset %q (run 'kops get assets --copy' before updating): %w", location, err)
	}
	var manifest v1.Manifest
	if err := json.Unmarshal(descriptor.Manifest, &manifest); err != nil {
		return nil, fmt.Errorf("decoding OCI asset %q: %w", location, err)
	}
	if len(manifest.Layers) != 1 {
		return nil, fmt.Errorf("OCI asset %q must contain exactly one file layer, found %d", location, len(manifest.Layers))
	}
	digest := manifest.Layers[0].Digest
	if digest.Algorithm != "sha256" {
		return nil, fmt.Errorf("OCI asset %q file layer must use SHA-256, found %q", location, digest)
	}
	hash, err := hashing.HashAlgorithmSHA256.FromString(digest.Hex)
	if err != nil {
		return nil, fmt.Errorf("invalid file layer digest for OCI asset %q: %w", location, err)
	}
	downloadedFileHashes.Store(location.String(), hash)
	return hash, nil
}

// HTTPSOnlyTransport rejects any request that is not HTTPS, including redirects. OCI registry
// traffic carries pull tokens, so it must never fall back to plain HTTP.
type HTTPSOnlyTransport struct {
	Inner http.RoundTripper
}

func (t *HTTPSOnlyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" {
		return nil, fmt.Errorf("OCI registry request must use HTTPS: %s", request.URL)
	}
	return t.Inner.RoundTrip(request)
}
