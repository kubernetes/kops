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
	"sort"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets/assetdata"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/values"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/vfs"
)

// ImageDigestResolver looks up the manifest digest for an image, returning it in the form
// "sha256:...".
type ImageDigestResolver func(image string) (string, error)

// imageDigestResolver is set (during startup) only by binaries that should resolve image digests
// by querying container registries, i.e. the kops CLI. Runtime binaries (kops-controller, nodeup)
// leave it unset, both because digest resolution is a cluster-configuration concern and so that
// they do not link the registry client libraries it requires.
var imageDigestResolver ImageDigestResolver

// SetImageDigestResolver installs the function RemapImage uses to resolve image digests. When no
// resolver is set, images are not pinned by digest.
func SetImageDigestResolver(resolver ImageDigestResolver) {
	imageDigestResolver = resolver
}

// AssetBuilder discovers and remaps assets.
type AssetBuilder struct {
	mu          sync.RWMutex
	imageAssets []*ImageAsset
	fileAssets  []*FileAsset

	// The following fields are immutable after construction via NewAssetBuilder
	// and are safe to read without holding mu.
	vfsContext     *vfs.VFSContext
	assetsLocation *kops.AssetsSpec
	getAssets      bool

	// KubeletSupportedVersion is the max version of kubelet that we are currently allowed to run on worker nodes.
	// This is used to avoid violating the kubelet supported version skew policy,
	// (we are not allowed to run a newer kubelet on a worker node than the control plane)
	KubeletSupportedVersion string

	// StaticManifests records manifests used by nodeup:
	// * e.g. sidecar manifests for static pods run by kubelet
	staticManifests []*StaticManifest

	// StaticFiles records static files:
	// * Configuration files supporting static pods
	staticFiles []*StaticFile
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
func NewAssetBuilder(vfsContext *vfs.VFSContext, assets *kops.AssetsSpec, getAssets bool) *AssetBuilder {
	a := &AssetBuilder{
		vfsContext:     vfsContext,
		assetsLocation: assets,
		getAssets:      getAssets,
	}

	return a
}

// VFSContext returns the VFS context used to read assets.
func (a *AssetBuilder) VFSContext() *vfs.VFSContext {
	return a.vfsContext
}

// SetAssetsLocation updates the assets location used for remapping.
// The AssetBuilder is created from the cluster spec before defaulting has run;
// cluster completion calls this once the assets locations are known.
func (a *AssetBuilder) SetAssetsLocation(assets *kops.AssetsSpec) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.assetsLocation = assets
}

func (a *AssetBuilder) addImageAsset(asset *ImageAsset) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.imageAssets = append(a.imageAssets, asset)
}

func (a *AssetBuilder) addFileAsset(asset *FileAsset) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.fileAssets = append(a.fileAssets, asset)
}

// AddStaticManifest records a nodeup static manifest.
func (a *AssetBuilder) AddStaticManifest(manifest *StaticManifest) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.staticManifests = append(a.staticManifests, manifest)
}

// AddStaticFile records a nodeup static file.
func (a *AssetBuilder) AddStaticFile(file *StaticFile) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.staticFiles = append(a.staticFiles, file)
}

// ImageAssets returns a sorted copy of the collected image assets.
func (a *AssetBuilder) ImageAssets() []*ImageAsset {
	a.mu.RLock()
	snapshot := append([]*ImageAsset(nil), a.imageAssets...)
	a.mu.RUnlock()

	// CanonicalLocation identifies the source image, so use it to make snapshots deterministic.
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].CanonicalLocation < snapshot[j].CanonicalLocation
	})

	return snapshot
}

// FileAssets returns a sorted copy of the collected file assets.
func (a *AssetBuilder) FileAssets() []*FileAsset {
	a.mu.RLock()
	snapshot := append([]*FileAsset(nil), a.fileAssets...)
	a.mu.RUnlock()

	// CanonicalURL identifies the source file, so use it to make snapshots deterministic.
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].CanonicalURL.String() < snapshot[j].CanonicalURL.String()
	})

	return snapshot
}

// StaticManifests returns a sorted copy of the collected static manifests.
func (a *AssetBuilder) StaticManifests() []*StaticManifest {
	a.mu.RLock()
	snapshot := append([]*StaticManifest(nil), a.staticManifests...)
	a.mu.RUnlock()

	// Key already identifies the static manifest, so use it to make snapshots deterministic.
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].Key < snapshot[j].Key
	})

	return snapshot
}

// StaticFiles returns a sorted copy of the collected static files.
func (a *AssetBuilder) StaticFiles() []*StaticFile {
	a.mu.RLock()
	snapshot := append([]*StaticFile(nil), a.staticFiles...)
	a.mu.RUnlock()

	// Path already identifies the static file on disk, so use it to make snapshots deterministic.
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].Path < snapshot[j].Path
	})

	return snapshot
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
func (a *AssetBuilder) RemapImage(image string) string {
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

	if os.Getenv("KOPS_BASE_URL") != "" && strings.HasPrefix(image, "registry.k8s.io/kops/") {
		// The kops images of a development build are not published to a registry;
		// they are sideloaded from KOPS_BASE_URL tarballs with these names.
		asset.DownloadLocation = image
		a.addImageAsset(asset)
		return image
	}

	normalized := NormalizeImage(a, image)
	image = normalized
	asset.DownloadLocation = normalized

	a.addImageAsset(asset)

	if imageDigestResolver == nil || !featureflag.ImageDigest.Enabled() || os.Getenv("KOPS_BASE_URL") != "" {
		return image
	}

	if strings.Contains(image, "@") {
		return image
	}

	digest, err := imageDigestResolver(image)
	if err != nil {
		klog.Warningf("failed to digest image %q: %s", image, err)
		return image
	}

	return image + "@" + digest
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

	if a.assetsLocation != nil && a.assetsLocation.FileRepository != nil {
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
	a.addFileAsset(fileAsset)

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
	if a.getAssets {
		u = file.CanonicalURL
	}

	if u != nil && u.Scheme == "oci" {
		// OCI artifacts are addressed by digest (the sha256 of the content), so the
		// hash must come from the canonical source; the registry has no sidecar hash files.
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

	klog.V(2).Infof("asset %q is not well-known, downloading hash", file.CanonicalURL)

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

	if a.assetsLocation != nil && a.assetsLocation.FileRepository != nil {
		return nil, fmt.Errorf("you might have not staged your files correctly, please execute 'kops get assets --copy'")
	}
	return nil, fmt.Errorf("cannot determine hash for %q (have you specified a valid file location?)", u)
}

func (a *AssetBuilder) remapURL(canonicalURL *url.URL) (*url.URL, error) {
	f := ""
	if a.assetsLocation != nil {
		f = values.StringValue(a.assetsLocation.FileRepository)
	}
	if f == "" {
		return nil, fmt.Errorf("assetsLocation.fileRepository must be set to remap asset %v", canonicalURL)
	}

	fileRepo, err := url.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse assetsLocation.fileRepository %q: %v", f, err)
	}

	assetPath := canonicalURL.Path
	if fileRepo.Scheme == "oci" {
		// The path becomes an OCI repository name, which has a restricted character set.
		assetPath = sanitizeOCIRepository(assetPath)
	}
	fileRepo.Path = path.Join(fileRepo.Path, assetPath)

	return fileRepo, nil
}

// ociRepositoryIllegalCharacters matches the characters that are not allowed in
// OCI repository names, which can only contain lowercase alphanumerics, `.`, `_`,
// `-` and `/`.
var ociRepositoryIllegalCharacters = regexp.MustCompile(`[^a-z0-9._/-]`)

// sanitizeOCIRepository maps an asset path to a valid OCI repository name;
// for example, CI builds are staged under paths containing `+`.
func sanitizeOCIRepository(assetPath string) string {
	return ociRepositoryIllegalCharacters.ReplaceAllString(strings.ToLower(assetPath), "_")
}

func NormalizeImage(a *AssetBuilder, image string) string {
	if a.assetsLocation != nil && a.assetsLocation.ContainerProxy != nil {
		containerProxy := strings.TrimSuffix(*a.assetsLocation.ContainerProxy, "/")
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

		// Run the new image
		image = normalized
	}

	if a.assetsLocation != nil && a.assetsLocation.ContainerRegistry != nil {
		registryMirror := *a.assetsLocation.ContainerRegistry
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
			normalized = strings.ReplaceAll(normalized, "/", "-")
			normalized = registryMirror + "/" + normalized
		}
		image = normalized
	}
	// Run the new image
	return image
}
