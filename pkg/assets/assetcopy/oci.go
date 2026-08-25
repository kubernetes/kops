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

package assetcopy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/kops/util/pkg/vfs"
)

var ociGet = remote.Get
var ociWrite = remote.Write

type httpsOnlyRegistryTransport struct {
	inner http.RoundTripper
}

func (t *httpsOnlyRegistryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" {
		return nil, fmt.Errorf("OCI registry request must use HTTPS: %s", request.URL)
	}
	return t.inner.RoundTrip(request)
}

type copyOCIAsset struct {
	source string
	sha256 string
	target string
	vfs    *vfs.VFSContext
}

func (c *copyOCIAsset) Run() error {
	location, err := url.Parse(c.target)
	if err != nil || location.Scheme != "oci" || location.Host == "" || location.User != nil || location.RawQuery != "" || location.Fragment != "" {
		return fmt.Errorf("invalid OCI target %q", c.target)
	}
	ref, err := name.NewTag(location.Host+location.Path, name.StrictValidation)
	if err != nil {
		return fmt.Errorf("invalid OCI target %q: %w", c.target, err)
	}

	options := []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
		remote.WithTransport(&httpsOnlyRegistryTransport{inner: remote.DefaultTransport}),
	}
	if c.sha256 == "" {
		return fmt.Errorf("OCI target %q has no source files", c.target)
	}

	present, err := c.tagContainsBlob(ref, c.sha256, options)
	if err != nil {
		return err
	}
	if present {
		return nil
	}

	data, err := c.vfs.ReadFile(c.source)
	if err != nil {
		return fmt.Errorf("reading %q: %w", c.source, err)
	}
	digest, _, err := v1.SHA256(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if digest.Hex != c.sha256 {
		return fmt.Errorf("source %q has sha256:%s, expected sha256:%s", c.source, digest.Hex, c.sha256)
	}
	layer := &fileBlob{data: data, digest: digest}

	image, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("creating OCI manifest: %w", err)
	}
	image = mutate.MediaType(image, types.OCIManifestSchema1)
	image = mutate.ConfigMediaType(image, types.OCIConfigJSON)
	return c.ensureTag(ref, image, c.sha256, options)
}

// ensureTag pushes the image and re-checks the tag. This is best-effort, not compare-and-swap:
// concurrent staging processes must not publish different content for the same tag.
func (c *copyOCIAsset) ensureTag(ref name.Tag, image v1.Image, digest string, options []remote.Option) error {
	if err := ociWrite(ref, image, options...); err != nil {
		return fmt.Errorf("pushing OCI tag %q: %w", ref.Name(), err)
	}
	present, err := c.tagContainsBlob(ref, digest, options)
	if err != nil {
		return err
	}
	if !present {
		return fmt.Errorf("OCI tag %q did not retain sha256:%s", ref.Name(), digest)
	}
	return nil
}

func (c *copyOCIAsset) tagContainsBlob(ref name.Reference, digest string, options []remote.Option) (bool, error) {
	descriptor, err := ociGet(ref, options...)
	if err != nil {
		var registryError *transport.Error
		if errors.As(err, &registryError) && registryError.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("reading OCI tag %q: %w", ref.Name(), err)
	}
	var manifest v1.Manifest
	if err := json.Unmarshal(descriptor.Manifest, &manifest); err != nil {
		return false, fmt.Errorf("decoding OCI tag %q: %w", ref.Name(), err)
	}
	existingDigests := make([]string, 0, len(manifest.Layers))
	for _, layer := range manifest.Layers {
		existingDigests = append(existingDigests, layer.Digest.String())
	}
	expected := "sha256:" + digest
	if len(existingDigests) == 1 && existingDigests[0] == expected {
		return true, nil
	}
	const remediation = "publish changed bytes under a distinct asset version, or remove the conflicting tag only after verifying that it is no longer needed"
	if len(manifest.Layers) == 0 {
		return false, fmt.Errorf(
			"OCI tag %q references media type %q with no file layer, expected %s; %s",
			ref.Name(), descriptor.MediaType, expected, remediation)
	}
	return false, fmt.Errorf(
		"OCI tag %q references layer digest(s) %s, expected %s; %s",
		ref.Name(), strings.Join(existingDigests, ", "), expected, remediation)
}

// fileBlob exposes the source bytes unchanged as an OCI layer. The digest is precomputed
// because go-containerregistry requests it several times per push.
type fileBlob struct {
	data   []byte
	digest v1.Hash
}

func (l *fileBlob) Digest() (v1.Hash, error) {
	return l.digest, nil
}

func (l *fileBlob) DiffID() (v1.Hash, error) {
	return l.digest, nil
}

func (l *fileBlob) Compressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.data)), nil
}

func (l *fileBlob) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.data)), nil
}

func (l *fileBlob) Size() (int64, error) {
	return int64(len(l.data)), nil
}

func (l *fileBlob) MediaType() (types.MediaType, error) {
	return types.MediaType("application/octet-stream"), nil
}
