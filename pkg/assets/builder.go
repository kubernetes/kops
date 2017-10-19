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

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

// RewriteManifests controls whether we rewrite manifests
// Because manifest rewriting converts everything to and from YAML, we normalize everything by doing so
var RewriteManifests = featureflag.New("RewriteManifests", featureflag.Bool(true))

// AssetBuilder discovers and remaps assets.
type AssetBuilder struct {
	ContainerAssets []*ContainerAsset
	FileAssets      []*FileAsset
	AssetsLocation  *kops.Assets
	// yay go cyclic dependency
	//Phase       cloudup.Phase
	Phase string
}

// ContainerAsset models a container's location.
type ContainerAsset struct {
	// DockerImage will be the name of the docker image we should run, if this is a docker image
	DockerImage string

	// CanonicalLocation will be the source location of the image, if we should copy it to the actual location
	CanonicalLocation string
}

// FileAsset models a file's location.
type FileAsset struct {
	// File will be the name of the file we should use
	File string

	// SHA will be the name of the sha file we should use
	SHA string

	// CanonicalLocation will be the source location of the file, if we should copy it to the actual location
	CanonicalLocation string

	// CanonicalSHALocation will be the source location of the sha file, if we should copy it to the actual location
	CononicalSHALocation string

	// SHAValue will be the value of the files SHA
	SHAValue string
}

// NewAssetBuilder creates a new AssetBuilder.
func NewAssetBuilder(assets *kops.Assets) *AssetBuilder {
	return &AssetBuilder{
		AssetsLocation: assets,
		Phase:          "",
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

// RemapImage normalizes a containers location if a user sets
// the AssetsLocation ContainerRegistry location.
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
func (a *AssetBuilder) RemapFileAndSHA(file string, sha string) (string, *hashing.Hash, error) {
	if file == "" {
		return "", nil, fmt.Errorf("unable to remap an empty string")
	}

	fileAsset := &FileAsset{
		File: file,
		SHA:  sha,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		fileAsset.CanonicalLocation = file
		fileAsset.CononicalSHALocation = sha

		file, err := a.normalizeURL(file)
		if err != nil {
			return "", nil, fmt.Errorf("unable to parse file url %q: %v", file, err)
		}

		fileAsset.File = file

		sha, err = a.normalizeURL(sha)
		if err != nil {
			return "", nil, fmt.Errorf("unable to parse sha url %q: %v", file, err)
		}

		fileAsset.SHA = sha
		glog.V(4).Infof("adding remapped file: %+v", fileAsset)
	}

	h, err := a.FindHash(fileAsset)
	if err != nil {
		return "", nil, err
	}

	fileAsset.SHAValue = h.Hex()
	a.FileAssets = append(a.FileAssets, fileAsset)

	glog.V(8).Infof("adding file: %+v", fileAsset)

	return fileAsset.File, h, nil
}

// RemapFileAndSHAValue is used exclusively to remap the cni tarball, as the tarball
// does not have a sha file in object storage.
func (a *AssetBuilder) RemapFileAndSHAValue(file string, shaValue string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("unable to remap an empty string")
	}

	fileAsset := &FileAsset{
		File:     file,
		SHAValue: shaValue,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		fileAsset.CanonicalLocation = file

		file, err := a.normalizeURL(file)
		if err != nil {
			return "", fmt.Errorf("unable to parse file url %q: %v", file, err)
		}

		fileAsset.File = file
		glog.V(4).Infof("adding remapped file: %q", fileAsset.File)
	}

	a.FileAssets = append(a.FileAssets, fileAsset)

	return fileAsset.File, nil
}

// FindHash returns the hash value from remove sha file via https.
func (a *AssetBuilder) FindHash(file *FileAsset) (*hashing.Hash, error) {
	u := file.File

	// FIXME ugly hack because dep loop with lifecycle
	// FIXME move lifecycle out of fi??
	if a.Phase == "assets" && file.CanonicalLocation != "" {
		u = file.CanonicalLocation
	}

	if u == "" {
		return nil, fmt.Errorf("file url is not defined")
	}

	for _, ext := range []string{".sha1"} {
		hashURL := u + ext
		b, err := vfs.Context.ReadFile(hashURL)
		if err != nil {
			glog.Infof("error reading hash file %q: %v", hashURL, err)
			continue
		}
		hashString := strings.TrimSpace(string(b))
		glog.V(2).Infof("Found hash %q for %q", hashString, u)

		return hashing.FromString(hashString)
	}
	return nil, fmt.Errorf("cannot determine hash for %q (have you specified a valid file location?)", u)
}

func (a *AssetBuilder) normalizeURL(file string) (string, error) {
	fileURL, err := url.Parse(file)
	if err != nil {
		return "", fmt.Errorf("unable to parse file url %q: %v", file, err)
	}

	fileRepo := strings.TrimSuffix(*a.AssetsLocation.FileRepository, "/")
	return fileRepo + fileURL.Path, nil
}
