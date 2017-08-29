/*
Copyright 2017 The Kubernetes Authors.

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
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
)

// RewriteManifests controls whether we rewrite manifests
// Because manifest rewriting converts everything to and from YAML, we normalize everything by doing so
var RewriteManifests = featureflag.New("RewriteManifests", featureflag.Bool(true))

// AssetBuilder discovers and remaps assets
type AssetBuilder struct {
	ContainerAssets []*ContainerAsset
	FileAssets      []*FileAsset
	AssetsLocation  *kops.Assets
}

type ContainerAsset struct {
	// DockerImage will be the name of the docker image we should run, if this is a docker image
	DockerImage string

	// CanonicalLocation will be the source location of the image, if we should copy it to the actual location
	CanonicalLocation string
}

type FileAsset struct {
	// File will be the name of the file we should use
	File string

	// CanonicalLocation will be the source location of the file, if we should copy it to the actual location
	CanonicalLocation string
}

func NewAssetBuilder(assets *kops.Assets) *AssetBuilder {
	return &AssetBuilder{
		AssetsLocation: assets,
	}
}

// RemapManifest transforms a kubernetes manifest.
// Whenever we are building a Task that includes a manifest, we should pass it through RemapManifest first.
// This will:
// * rewrite the images if they are being redirected to a mirror, and ensure the image is uploaded
func (a *AssetBuilder) RemapManifest(data []byte) ([]byte, error) {
	if !RewriteManifests.Enabled() {
		return data, nil
	}
	manifests, err := kubemanifest.LoadManifestsFrom(data)
	if err != nil {
		return nil, err
	}

	var yamlSeparator = []byte("\n---\n\n")
	var remappedManifests [][]byte
	for _, manifest := range manifests {
		err := manifest.RemapImages(a.RemapImage)
		if err != nil {
			return nil, fmt.Errorf("error remapping images: %v", err)
		}
		y, err := manifest.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshalling manifest: %v", err)
		}

		remappedManifests = append(remappedManifests, y)
	}

	return bytes.Join(remappedManifests, yamlSeparator), nil
}

func (a *AssetBuilder) RemapImage(image string) (string, error) {
	asset := &ContainerAsset{}

	asset.DockerImage = image

	if strings.HasPrefix(image, "kope/dns-controller:") {
		// To use user-defined DNS Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
		// 2. export DNSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("DNSCONTROLLER_IMAGE")
		if override != "" {
			image = override
		}
	}

	if a.AssetsLocation != nil && a.AssetsLocation.ContainerRegistry != nil {
		registryMirror := *a.AssetsLocation.ContainerRegistry
		normalized := image

		// Remove the 'standard' kubernetes image prefix, just for sanity
		normalized = strings.TrimPrefix(normalized, "gcr.io/google_containers/")

		// We can't nest arbitrarily
		// Some risk of collisions, but also -- and __ in the names appear to be blocked by docker hub
		normalized = strings.Replace(normalized, "/", "-", -1)
		asset.DockerImage = registryMirror + "/" + normalized

		asset.CanonicalLocation = image

		// Run the new image
		image = asset.DockerImage
	}

	a.ContainerAssets = append(a.ContainerAssets, asset)

	return image, nil
}

// RemapFile sets a new url location for the file, if a AssetsLocation is defined.
func (a AssetBuilder) RemapFile(file string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("unable to remap an empty string")
	}

	fileAsset := &FileAsset{
		File:              file,
		CanonicalLocation: file,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		fileURL, err := url.Parse(file)
		if err != nil {
			return "", fmt.Errorf("unable to parse file url %q: %v", file, err)
		}

		fileRepo := strings.TrimSuffix(*a.AssetsLocation.FileRepository, "/")
		fileAsset.File = fileRepo + fileURL.Path
	}

	a.FileAssets = append(a.FileAssets, fileAsset)

	return fileAsset.File, nil
}
