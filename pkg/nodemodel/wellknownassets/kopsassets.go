/*
Copyright 2019 The Kubernetes Authors.

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

package wellknownassets

import (
	"fmt"
	"net/url"
	"os"
	"path"

	"k8s.io/klog/v2"
	"k8s.io/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
)

const (
	defaultKopsBaseURL = "https://artifacts.k8s.io/binaries/kops/%s/"
)

var kopsBaseURL *url.URL

// BaseURL returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseURL() (*url.URL, error) {
	// returning cached value
	// Avoid repeated logging
	if kopsBaseURL != nil {
		klog.V(8).Infof("Using cached kopsBaseUrl url: %q", kopsBaseURL.String())
		return copyBaseURL(kopsBaseURL)
	}

	baseURLString := os.Getenv("KOPS_BASE_URL")
	var err error
	if baseURLString == "" {
		baseURLString = fmt.Sprintf(defaultKopsBaseURL, kops.Version)
		klog.V(8).Infof("Using default base url: %q", baseURLString)
		kopsBaseURL, err = url.Parse(baseURLString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %q as a url: %v", baseURLString, err)
		}
	} else {
		kopsBaseURL, err = url.Parse(baseURLString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var KOPS_BASE_URL %q as a url: %v", baseURLString, err)
		}
		klog.Warningf("Using base url from env var: KOPS_BASE_URL=%q", baseURLString)

		// The last path component of KOPS_BASE_URL is the artifact version.
		// Override kops.Version so image tags in manifests match the sideloaded images.
		if v := path.Base(kopsBaseURL.Path); v != "" && v != "." && v != "/" && v != kops.Version {
			klog.Infof("Overriding kops version from KOPS_BASE_URL: %q -> %q", kops.Version, v)
			kops.Version = v
		}
	}

	return copyBaseURL(kopsBaseURL)
}

// copyBaseURL makes a copy of the base url or the path.Joins can append stuff to this URL
func copyBaseURL(base *url.URL) (*url.URL, error) {
	u, err := url.Parse(base.String())
	if err != nil {
		return nil, err
	}
	return u, nil
}

// NodeUpAsset returns nodeup after registering it with assetsBuilder. The result is not cached
// because its locations depend on the builder's repository mapping. Hashes are cached separately.
func NodeUpAsset(assetsBuilder *assets.AssetBuilder, arch architectures.Architecture) (*assets.MirroredAsset, error) {
	asset, err := KopsFileURL(fmt.Sprintf("linux/%s/nodeup.xz", arch), assetsBuilder)
	if err != nil {
		return nil, err
	}
	klog.V(8).Infof("Using default nodeup location for %s: %q", arch, asset.DownloadURL.String())

	return assets.BuildMirroredAsset(asset), nil
}

// KopsFileURL returns the base url for the distribution of kops - in particular for nodeup & docker images
func KopsFileURL(file string, assetBuilder *assets.AssetBuilder) (*assets.FileAsset, error) {
	base, err := BaseURL()
	if err != nil {
		return nil, err
	}

	base.Path = path.Join(base.Path, file)

	asset, err := assetBuilder.RemapFile(base, nil)
	if err != nil {
		return nil, err
	}

	return asset, nil
}
