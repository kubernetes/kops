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
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets/assetdata"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/values"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

// AssetBuilder discovers and remaps assets.
type AssetBuilder struct {
	vfsContext     *vfs.VFSContext
	ImageAssets    []*ImageAsset
	FileAssets     []*FileAsset
	AssetsLocation *kops.AssetsSpec
	GetAssets      bool

	// KubernetesVersion is the version of kubernetes we are installing
	KubernetesVersion semver.Version

	// StaticManifests records manifests used by nodeup:
	// * e.g. sidecar manifests for static pods run by kubelet
	StaticManifests []*StaticManifest

	// StaticFiles records static files:
	// * Configuration files supporting static pods
	StaticFiles []*StaticFile
}

type StaticFile struct {
	// Path is the path to the manifest.
	Path string

	// Content holds the desired file contents.
	Content string

	// The static manifest will only be applied to instances matching the specified role
	Roles []kops.InstanceGroupRole
}

type StaticManifest struct {
	// Key is the unique identifier of the manifest
	Key string

	// Path is the path to the manifest.
	Path string

	// The static manifest will only be applied to instances matching the specified role
	Roles []kops.InstanceGroupRole

	// Contents is the contents of the manifest, which may be easier than fetching it from Path
	Contents []byte
}

func (m *StaticManifest) AppliesToRole(role kops.InstanceGroupRole) bool {
	for _, r := range m.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ImageAsset models an image's location.
type ImageAsset struct {
	// DownloadLocation will be the name of the image we should run.
	// This is used to copy an image to a ContainerRegistry.
	DownloadLocation string
	// CanonicalLocation will be the source location of the image.
	CanonicalLocation string
}

// FileAsset models a file's location.
type FileAsset struct {
	// DownloadURL is the URL from which the cluster should download the asset.
	DownloadURL *url.URL
	// CanonicalURL is the canonical location of the asset, for example as distributed by the kops project
	CanonicalURL *url.URL
	// SHAValue is the SHA hash of the FileAsset.
	SHAValue *hashing.Hash
}

// NewAssetBuilder creates a new AssetBuilder.
func NewAssetBuilder(vfsContext *vfs.VFSContext, assets *kops.AssetsSpec, kubernetesVersion string, getAssets bool) *AssetBuilder {
	a := &AssetBuilder{
		vfsContext:     vfsContext,
		AssetsLocation: assets,
		GetAssets:      getAssets,
	}

	version, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		// This should have already been validated
		klog.Fatalf("unexpected error from ParseKubernetesVersion %s: %v", kubernetesVersion, err)
	}
	a.KubernetesVersion = *version

	return a
}

// RemapManifest transforms a kubernetes manifest.
// Whenever we are building a Task that includes a manifest, we should pass it through RemapManifest first.
// This will:
// * rewrite the images if they are being redirected to a mirror, and ensure the image is uploaded
func (a *AssetBuilder) RemapManifest(data []byte) ([]byte, error) {
	objects, err := kubemanifest.LoadObjectsFrom(data)
	if err != nil {
		return nil, err
	}

	for _, object := range objects {
		if err := object.RemapImages(a.RemapImage); err != nil {
			return nil, fmt.Errorf("error remapping images: %v", err)
		}
	}

	return objects.ToYAML()
}

// RemapImage normalizes a containers location if a user sets the AssetsLocation ContainerRegistry location.
func (a *AssetBuilder) RemapImage(image string) (string, error) {
	asset := &ImageAsset{
		DownloadLocation:  image,
		CanonicalLocation: image,
	}

	if strings.HasPrefix(image, "registry.k8s.io/kops/dns-controller:") {
		// To use user-defined DNS Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
		// 2. export DNSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("DNSCONTROLLER_IMAGE")
		if override != "" {
			image = override
		}
	}

	if strings.HasPrefix(image, "k8s.gcr.io/kops/kops-controller:") || strings.HasPrefix(image, "registry.k8s.io/kops/kops-controller:") {
		// To use user-defined kops Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make kops-controller-push
		// 2. export KOPSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("KOPSCONTROLLER_IMAGE")
		if override != "" {
			image = override
		}
	}

	if strings.HasPrefix(image, "registry.k8s.io/kops/kube-apiserver-healthcheck:") {
		override := os.Getenv("KUBE_APISERVER_HEALTHCHECK_IMAGE")
		if override != "" {
			image = override
		}
	}

	if a.AssetsLocation != nil && a.AssetsLocation.ContainerProxy != nil {
		containerProxy := strings.TrimSuffix(*a.AssetsLocation.ContainerProxy, "/")
		normalized := image

		// If the image name contains only a single / we need to determine if the image is located on docker-hub or if it's using a convenient URL,
		// like registry.k8s.io/<image-name> or registry.k8s.io/<image-name>
		// In case of a hub image it should be sufficient to just prepend the proxy url, producing eg docker-proxy.example.com/weaveworks/weave-kube
		if strings.Count(normalized, "/") <= 1 && !strings.ContainsAny(strings.Split(normalized, "/")[0], ".:") {
			normalized = containerProxy + "/" + normalized
		} else {
			re := regexp.MustCompile(`^[^/]+`)
			normalized = re.ReplaceAllString(normalized, containerProxy)
		}

		asset.DownloadLocation = normalized

		// Run the new image
		image = asset.DownloadLocation
	}

	if a.AssetsLocation != nil && a.AssetsLocation.ContainerRegistry != nil {
		registryMirror := *a.AssetsLocation.ContainerRegistry
		normalized := image

		// Remove the 'standard' kubernetes image prefixes, just for sanity
		normalized = strings.TrimPrefix(normalized, "registry.k8s.io/")

		// When assembling the cluster spec, kops may call the option more then once until the config converges
		// This means that this function may me called more than once on the same image
		// It this is pass is the second one, the image will already have been normalized with the containerRegistry settings
		// If this is the case, passing though the process again will re-prepend the container registry again
		// and again, causing the spec to never converge and the config build to fail.
		if !strings.HasPrefix(normalized, registryMirror+"/") {
			// We can't nest arbitrarily
			// Some risk of collisions, but also -- and __ in the names appear to be blocked by docker hub
			normalized = strings.Replace(normalized, "/", "-", -1)
			asset.DownloadLocation = registryMirror + "/" + normalized
		}

		// Run the new image
		image = asset.DownloadLocation
	}

	a.ImageAssets = append(a.ImageAssets, asset)

	if !featureflag.ImageDigest.Enabled() || os.Getenv("KOPS_BASE_URL") != "" {
		return image, nil
	}

	if strings.Contains(image, "@") {
		return image, nil
	}

	digest, err := crane.Digest(image, crane.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		klog.Warningf("failed to digest image %q: %s", image, err)
		return image, nil
	}

	return image + "@" + digest, nil
}

// RemapFile returns a remapped URL for the file, if AssetsLocation is defined.
// It is returns in a FileAsset, alongside the SHA hash of the file.
// The SHA hash is is knownHash is provided, and otherwise will be found first by
// checking the canonical URL against our well-known hashes, and failing that via download.
func (a *AssetBuilder) RemapFile(canonicalURL *url.URL, knownHash *hashing.Hash) (*FileAsset, error) {
	if canonicalURL == nil {
		return nil, fmt.Errorf("unable to remap a nil URL")
	}

	fileAsset := &FileAsset{
		DownloadURL:  canonicalURL,
		CanonicalURL: canonicalURL,
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		normalizedFile, err := a.remapURL(canonicalURL)
		if err != nil {
			return nil, err
		}

		if canonicalURL.Host != normalizedFile.Host {
			fileAsset.DownloadURL = normalizedFile
			klog.V(4).Infof("adding remapped file: %q", fileAsset.DownloadURL.String())
		}
	}

	if knownHash == nil {
		h, err := a.findHash(fileAsset)
		if err != nil {
			return nil, err
		}
		knownHash = h
	}

	fileAsset.SHAValue = knownHash

	klog.V(8).Infof("adding file: %+v", fileAsset)
	a.FileAssets = append(a.FileAssets, fileAsset)

	return fileAsset, nil
}

// findHash returns the hash value of a FileAsset.
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
	u := file.DownloadURL
	if a.GetAssets {
		u = file.CanonicalURL
	}

	if u == nil {
		return nil, fmt.Errorf("file url is not defined")
	}

	knownHash, found, err := assetdata.GetHash(file.CanonicalURL)
	if err != nil {
		return nil, err
	}
	if found {
		return knownHash, nil
	}

	klog.Infof("asset %q is not well-known, downloading hash", file.CanonicalURL)

	// We now prefer sha256 hashes
	for backoffSteps := 1; backoffSteps <= 3; backoffSteps++ {
		// We try first with a short backoff, so we don't
		// waste too much time looking for files that don't
		// exist before trying the next one
		backoff := wait.Backoff{
			Duration: 500 * time.Millisecond,
			Factor:   2,
			Steps:    backoffSteps,
		}

		for _, ext := range []string{".sha256", ".sha256sum"} {
			for _, mirror := range FindURLMirrors(u.String()) {
				hashURL := mirror + ext
				klog.V(3).Infof("Trying to read hash file: %q", hashURL)
				b, err := a.vfsContext.ReadFile(hashURL, vfs.WithBackoff(backoff))
				if err != nil {
					// Try to log without being too alarming - issue #7550
					klog.V(2).Infof("Unable to read hash file %q: %v", hashURL, err)
					continue
				}
				hashString := strings.TrimSpace(string(b))
				klog.V(2).Infof("Found hash %q for %q", hashString, u)

				// Accept a hash string that is `<hash> <filename>`
				fields := strings.Fields(hashString)
				if len(fields) == 0 {
					klog.Infof("Hash file was empty %q", hashURL)
					continue
				}
				return hashing.FromString(fields[0])
			}
			if ext == ".sha256" {
				klog.V(2).Infof("Unable to read new sha256 hash file (is this an older/unsupported kubernetes release?)")
			}
		}
	}

	if a.AssetsLocation != nil && a.AssetsLocation.FileRepository != nil {
		return nil, fmt.Errorf("you might have not staged your files correctly, please execute 'kops get assets --copy'")
	}
	return nil, fmt.Errorf("cannot determine hash for %q (have you specified a valid file location?)", u)
}

func (a *AssetBuilder) remapURL(canonicalURL *url.URL) (*url.URL, error) {
	f := ""
	if a.AssetsLocation != nil {
		f = values.StringValue(a.AssetsLocation.FileRepository)
	}
	if f == "" {
		return nil, fmt.Errorf("assetsLocation.fileRepository must be set to remap asset %v", canonicalURL)
	}

	fileRepo, err := url.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse assetsLocation.fileRepository %q: %v", f, err)
	}

	fileRepo.Path = path.Join(fileRepo.Path, canonicalURL.Path)

	return fileRepo, nil
}
