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

package assets

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/klog/v2"
)

// CopyImage copies a docker image from a source registry, to a target registry,
// typically used for highly secure clusters.
type CopyImage struct {
	Name        string
	SourceImage string
	TargetImage string
}

func (e *CopyImage) Run(ctx context.Context) error {
	source := e.SourceImage
	target := e.TargetImage

	sourceRef, err := name.ParseReference(source)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", source, err)
	}

	targetRef, err := name.ParseReference(target)
	if err != nil {
		return fmt.Errorf("parsing reference for %q: %v", target, err)
	}

	options := []remote.Option{remote.WithAuthFromKeychain(authn.DefaultKeychain)}
	options = append(options, remote.WithContext(ctx))

	desc, err := remote.Get(sourceRef, options...)
	if err != nil {
		return fmt.Errorf("fetching %q: %v", source, err)
	}

	targetDesc, err := remote.Get(targetRef, options...)
	if err == nil && desc.Digest.String() == targetDesc.Digest.String() {
		klog.Infof("no need to copy image from %v to %v", sourceRef, targetRef)
		return nil
	}

	switch desc.MediaType {
	case types.OCIImageIndex, types.DockerManifestList:
		// Handle indexes separately.
		if err := copyIndex(desc, sourceRef, targetRef, options...); err != nil {
			return fmt.Errorf("failed to copy index: %v", err)
		}
	default:
		// Assume anything else is an image, since some registries don't set mediaTypes properly.
		if err := copyImage(desc, sourceRef, targetRef, options...); err != nil {
			return fmt.Errorf("failed to copy image: %v", err)
		}
	}

	return nil
}

func copyImage(desc *remote.Descriptor, sourceRef name.Reference, targetRef name.Reference, options ...remote.Option) error {
	klog.Infof("copying image from %v to %v", sourceRef, targetRef)

	img, err := desc.Image()
	if err != nil {
		return err
	}
	return remote.Write(targetRef, img, options...)
}

func copyIndex(desc *remote.Descriptor, sourceRef name.Reference, targetRef name.Reference, options ...remote.Option) error {
	klog.Infof("copying image index from %v to %v", sourceRef, targetRef)

	idx, err := desc.ImageIndex()
	if err != nil {
		return err
	}
	return remote.WriteIndex(targetRef, idx, options...)
}
