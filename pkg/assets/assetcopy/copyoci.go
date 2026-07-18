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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/vfs"
)

// CopyFileToOCI copies a file from a source file repository to an OCI registry, pushing it as an
// image with a single layer (a Docker schema 2 image manifest, which OCI registries accept). The
// layer blob digest is the sha256 hash of the file, so clients can download the blob directly by
// the file's hash without reading the manifest. Pushing authenticates with the local docker
// credentials.
type CopyFileToOCI struct {
	Name       string
	SourceFile string
	// TargetRef is the target location, in the form oci://<registry>/<repository>.
	TargetRef  string
	SHA        string
	VFSContext *vfs.VFSContext
}

func (e *CopyFileToOCI) Run() error {
	repository := strings.TrimPrefix(e.TargetRef, "oci://")

	// Tag the artifact with the file's hash; this makes the check for an already-pushed artifact a
	// single manifest fetch.
	ref, err := name.NewTag(repository + ":" + e.SHA)
	if err != nil {
		return fmt.Errorf("parsing reference for %q: %w", e.TargetRef, err)
	}

	// name.NewTag applies Docker-style normalization: a registry host without a dot or port is treated
	// as a Docker Hub repository, and docker.io aliases gain a library/ prefix. Nodes pull from the
	// URL literally, so refuse any reference that does not round-trip exactly.
	if ref.Context().Name() != repository {
		return fmt.Errorf("target %q parses as %q, which is not the location nodes download from; the registry host must be a fully qualified domain name", e.TargetRef, ref.Context().Name())
	}

	options := []remote.Option{remote.WithAuthFromKeychain(authn.DefaultKeychain)}

	if e.alreadyPushed(ref, options) {
		klog.Infof("no need to copy file from %v to %v", e.SourceFile, e.TargetRef)
		return nil
	}

	data, err := e.VFSContext.ReadFile(e.SourceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found %q: %w", e.SourceFile, err)
		}
		return fmt.Errorf("error downloading file %q: %w", e.SourceFile, err)
	}

	digest := sha256.Sum256(data)
	actualSHA := hex.EncodeToString(digest[:])
	if actualSHA != e.SHA {
		return fmt.Errorf("hash mismatch for %q: expected %q, got %q", e.SourceFile, e.SHA, actualSHA)
	}

	layer := static.NewLayer(data, types.MediaType("application/octet-stream"))
	image, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("building artifact for %q: %w", e.TargetRef, err)
	}

	klog.V(2).Infof("copying bits from %q to %q", e.SourceFile, e.TargetRef)

	if err := remote.Write(ref, image, options...); err != nil {
		return fmt.Errorf("unable to transfer %q to %q: %w", e.SourceFile, e.TargetRef, err)
	}

	return nil
}

// alreadyPushed reports whether the tag exists and its manifest references the expected layer blob.
// The tag alone is not proof: nodes download the blob directly by digest, so a stale or foreign
// manifest under this tag must be overwritten, not skipped.
func (e *CopyFileToOCI) alreadyPushed(ref name.Tag, options []remote.Option) bool {
	desc, err := remote.Get(ref, options...)
	if err != nil {
		return false
	}
	image, err := desc.Image()
	if err != nil {
		return false
	}
	manifest, err := image.Manifest()
	if err != nil {
		return false
	}
	for _, layer := range manifest.Layers {
		if layer.Digest.Hex == e.SHA {
			return true
		}
	}
	return false
}
