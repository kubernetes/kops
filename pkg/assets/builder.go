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
	"path"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/values"
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
	// TODO we'd like to use cloudup.Phase here, but that introduces a go cyclic dependency
	Phase string

	// KubernetesVersion is the version of kubernetes we are installing
	KubernetesVersion semver.Version
}

// ContainerAsset models a container's location.
type ContainerAsset struct {
	// DockerImage will be the name of the container we should run.
	// This is used to copy a container to a ContainerRegistry.
	DockerImage string
	// CanonicalLocation will be the source location of the container.
	CanonicalLocation string
}

// FileAsset models a file's location.
type FileAsset struct {
	// FileURL is the URL of a file that is accessed by a Kubernetes cluster.
	FileURL *url.URL
	// CanonicalFileURL is the source URL of a file. This is used to copy a file to a FileRepository.
	CanonicalFileURL *url.URL
	// SHAValue is the SHA hash of the FileAsset.
	SHAValue string
}

// NewAssetBuilder creates a new AssetBuilder.
func NewAssetBuilder(cluster *kops.Cluster, phase string) *AssetBuilder {
	a := &AssetBuilder{
		AssetsLocation: cluster.Spec.Assets,
		Phase:          phase,
	}

	version, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		// This should have already been validated
		glog.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", cluster.Spec.KubernetesVersion, err)
	}
	a.KubernetesVersion = *version

	return a
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
		if err := manifest.RemapImages(a.RemapImage); err != nil {
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

// RemapImage normalizes a containers location if a user sets the AssetsLocation ContainerRegistry location.
func (a *AssetBuilder) RemapImage(image string) (string, error) {
	asset := &ContainerAsset{}

	asset.DockerImage = image

	// The k8s.gcr.io prefix is an alias, but for CI builds we run from a docker load,
	// and we only double-tag from 1.10 onwards.
	// For versions prior to 1.10, remap k8s.gcr.io to the old name.
	// This also means that we won't start using the aliased names on existing clusters,
	// which could otherwise be surprising to users.
	if !util.IsKubernetesGTE("1.10", a.KubernetesVersion) && strings.HasPrefix(image, "k8s.gcr.io/") {
		image = "gcr.io/google_containers/" + strings.TrimPrefix(image, "k8s.gcr.io/")
	}

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

	if a.AssetsLocation != nil && a.AssetsLocation.ContainerProxy != nil {
		containerProxy := strings.TrimRight(*a.AssetsLocation.ContainerProxy, "/")
		normalized := image

		// If the image name contains only a single / we need to determine if the image is located on docker-hub or if it's using a convenient URL like k8s.gcr.io/<image-name>
		// In case of a hub image it should be sufficient to just prepend the proxy url, producing eg docker-proxy.example.com/weaveworks/weave-kube
		if strings.Count(normalized, "/") <= 1 && !strings.ContainsAny(strings.Split(normalized, "/")[0], ".:") {
			normalized = containerProxy + "/" + normalized
		} else {
			var re = regexp.MustCompile(`^[^/]+`)
			normalized = re.ReplaceAllString(normalized, containerProxy)
		}

		asset.DockerImage = normalized
		asset.CanonicalLocation = image

		// Run the new image
		image = asset.DockerImage
	}

	if a.AssetsLocation != nil && a.AssetsLocation.ContainerRegistry != nil {
		registryMirror := *a.AssetsLocation.ContainerRegistry
		normalized := image

		// Remove the 'standard' kubernetes image prefix, just for sanity
		if !util.IsKubernetesGTE("1.10", a.KubernetesVersion) && strings.HasPrefix(normalized, "gcr.io/google_containers/") {
			normalized = strings.TrimPrefix(normalized, "gcr.io/google_containers/")
		} else {
			normalized = strings.TrimPrefix(normalized, "k8s.gcr.io/")
		}

		// When assembling the cluster spec, kops may call the option more then once until the config converges
		// This means that this function may me called more than once on the same image
		// It this is pass is the second one, the image will already have been normalized with the containerRegistry settings
		// If this is the case, passing though the process again will re-prepend the container registry again
		// and again, causing the spec to never converge and the config build to fail.
		if !strings.HasPrefix(normalized, registryMirror+"/") {
			// We can't nest arbitrarily
			// Some risk of collisions, but also -- and __ in the names appear to be blocked by docker hub
			normalized = strings.Replace(normalized, "/", "-", -1)
			asset.DockerImage = registryMirror + "/" + normalized
		}

		asset.CanonicalLocation = image

		// Run the new image
		image = asset.DockerImage
	}

	a.ContainerAssets = append(a.ContainerAssets, asset)
	return image, nil
}

// RemapFileAndSHA returns a remapped url for the file, if AssetsLocation is defined.
// It also returns the SHA hash of the file.
func (a *AssetBuilder) RemapFileAndSHA(fileURL *url.URL) (*url.URL, *hashing.Hash, error) {
	if fileURL == nil {
		return nil, nil, fmt.Errorf("unable to remap a nil URL")
	}

	fileAsset := &FileAsset{
		FileURL: fileURL,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		fileAsset.CanonicalFileURL = fileURL

		normalizedFileURL, err := a.normalizeURL(fileURL)
		if err != nil {
			return nil, nil, err
		}

		fileAsset.FileURL = normalizedFileURL

		glog.V(4).Infof("adding remapped file: %+v", fileAsset)
	}

	h, err := a.findHash(fileAsset)
	if err != nil {
		return nil, nil, err
	}
	fileAsset.SHAValue = h.Hex()

	a.FileAssets = append(a.FileAssets, fileAsset)
	glog.V(8).Infof("adding file: %+v", fileAsset)

	return fileAsset.FileURL, h, nil
}

// TODO - remove this method as CNI does now have a SHA file

// RemapFileAndSHAValue is used exclusively to remap the cni tarball, as the tarball does not have a sha file in object storage.
func (a *AssetBuilder) RemapFileAndSHAValue(fileURL *url.URL, shaValue string) (*url.URL, error) {
	if fileURL == nil {
		return nil, fmt.Errorf("unable to remap a nil URL")
	}

	fileAsset := &FileAsset{
		FileURL:  fileURL,
		SHAValue: shaValue,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		fileAsset.CanonicalFileURL = fileURL

		normalizedFile, err := a.normalizeURL(fileURL)
		if err != nil {
			return nil, err
		}

		fileAsset.FileURL = normalizedFile
		glog.V(4).Infof("adding remapped file: %q", fileAsset.FileURL.String())
	}

	a.FileAssets = append(a.FileAssets, fileAsset)

	return fileAsset.FileURL, nil
}

// FindHash returns the hash value of a FileAsset.
func (a *AssetBuilder) findHash(file *FileAsset) (*hashing.Hash, error) {

	// If the phase is "assets" we use the CanonicalFileURL,
	// but during other phases we use the hash from the FileRepository or the base kops path.
	// We do not want to just test for CanonicalFileURL as it is defined in
	// other phases, but is not used to test for the SHA.
	// This prevents a chicken and egg problem where the file is not yet in the FileRepository.
	//
	// assets phase -> get the sha file from the source / CanonicalFileURL
	// any other phase -> get the sha file from the kops base location or the FileRepository
	//
	// TLDR; we use the file.CanonicalFileURL during assets phase, and use file.FileUrl the
	// rest of the time. If not we get a chicken and the egg problem where we are reading the sha file
	// before it exists.
	u := file.FileURL
	if a.Phase == "assets" && file.CanonicalFileURL != nil {
		u = file.CanonicalFileURL
	}

	if u == nil {
		return nil, fmt.Errorf("file url is not defined")
	}

	for _, ext := range []string{".sha1"} {
		hashURL := u.String() + ext
		b, err := vfs.Context.ReadFile(hashURL)
		if err != nil {
			glog.Infof("error reading hash file %q: %v", hashURL, err)
			continue
		}
		hashString := strings.TrimSpace(string(b))
		glog.V(2).Infof("Found hash %q for %q", hashString, u)

		// Accept a hash string that is `<hash> <filename>`
		fields := strings.Fields(hashString)
		return hashing.FromString(fields[0])
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		return nil, fmt.Errorf("you may have not staged your files correctly, please execute kops update cluster using the assets phase")
	}
	return nil, fmt.Errorf("cannot determine hash for %q (have you specified a valid file location?)", u)
}

func (a *AssetBuilder) normalizeURL(file *url.URL) (*url.URL, error) {

	if a.AssetsLocation == nil || a.AssetsLocation.FileRepository == nil {
		return nil, fmt.Errorf("assetLocation and fileRepository cannot be nil to normalize an file asset URL")
	}

	f := values.StringValue(a.AssetsLocation.FileRepository)

	if f == "" {
		return nil, fmt.Errorf("assetsLocation fileRepository cannot be an empty string")
	}

	fileRepo, err := url.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file repository URL %q: %v", values.StringValue(a.AssetsLocation.FileRepository), err)
	}

	fileRepo.Path = path.Join(fileRepo.Path, file.Path)

	return fileRepo, nil
}
