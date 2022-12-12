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

package cloudup

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"k8s.io/klog/v2"
	"k8s.io/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
	"k8s.io/kops/util/pkg/mirrors"
)

const (
	defaultKopsBaseURL = "https://artifacts.k8s.io/binaries/kops/%s/"
)

var kopsBaseURL *url.URL

// nodeUpAsset caches the nodeup binary download url/hash
var nodeUpAsset map[architectures.Architecture]*mirrors.MirroredAsset

// protokubeAsset caches the protokube binary download url/hash
var protokubeAsset map[architectures.Architecture]*mirrors.MirroredAsset

// channelsAsset caches the channels binary download url/hash
var channelsAsset map[architectures.Architecture]*mirrors.MirroredAsset

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

// NodeUpAsset returns the asset for where nodeup should be downloaded
func NodeUpAsset(ctx context.Context, assetsBuilder *assets.AssetBuilder, arch architectures.Architecture) (*mirrors.MirroredAsset, error) {
	if nodeUpAsset == nil {
		nodeUpAsset = make(map[architectures.Architecture]*mirrors.MirroredAsset)
	}
	if nodeUpAsset[arch] != nil {
		// Avoid repeated logging
		klog.V(8).Infof("Using cached nodeup location for %s: %v", arch, nodeUpAsset[arch].Locations)
		return nodeUpAsset[arch], nil
	}

	u, hash, err := KopsFileURL(ctx, fmt.Sprintf("linux/%s/nodeup", arch), assetsBuilder)
	if err != nil {
		return nil, err
	}
	nodeUpAsset[arch] = mirrors.BuildMirroredAsset(u, hash)
	klog.V(8).Infof("Using default nodeup location for %s: %q", arch, u.String())

	return nodeUpAsset[arch], nil
}

// ProtokubeAsset returns the url and hash of the protokube binary
func ProtokubeAsset(ctx context.Context, assetsBuilder *assets.AssetBuilder, arch architectures.Architecture) (*mirrors.MirroredAsset, error) {
	if protokubeAsset == nil {
		protokubeAsset = make(map[architectures.Architecture]*mirrors.MirroredAsset)
	}
	if protokubeAsset[arch] != nil {
		klog.V(8).Infof("Using cached protokube binary location for %s: %v", arch, protokubeAsset[arch].Locations)
		return protokubeAsset[arch], nil
	}

	u, hash, err := KopsFileURL(ctx, fmt.Sprintf("linux/%s/protokube", arch), assetsBuilder)
	if err != nil {
		return nil, err
	}
	protokubeAsset[arch] = mirrors.BuildMirroredAsset(u, hash)
	klog.V(8).Infof("Using default protokube location for %s: %q", arch, u.String())

	return protokubeAsset[arch], nil
}

// ChannelsAsset returns the url and hash of the channels binary
func ChannelsAsset(ctx context.Context, assetsBuilder *assets.AssetBuilder, arch architectures.Architecture) (*mirrors.MirroredAsset, error) {
	if channelsAsset == nil {
		channelsAsset = make(map[architectures.Architecture]*mirrors.MirroredAsset)
	}
	if channelsAsset[arch] != nil {
		klog.V(8).Infof("Using cached channels binary location for %s: %v", arch, channelsAsset[arch].Locations)
		return channelsAsset[arch], nil
	}

	u, hash, err := KopsFileURL(ctx, fmt.Sprintf("linux/%s/channels", arch), assetsBuilder)
	if err != nil {
		return nil, err
	}
	channelsAsset[arch] = mirrors.BuildMirroredAsset(u, hash)
	klog.V(8).Infof("Using default channels location for %s: %q", arch, u.String())

	return channelsAsset[arch], nil
}

// KopsFileURL returns the base url for the distribution of kops - in particular for nodeup & docker images
func KopsFileURL(ctx context.Context, file string, assetBuilder *assets.AssetBuilder) (*url.URL, *hashing.Hash, error) {
	base, err := BaseURL()
	if err != nil {
		return nil, nil, err
	}

	base.Path = path.Join(base.Path, file)

	fileURL, hash, err := assetBuilder.RemapFileAndSHA(ctx, base)
	if err != nil {
		return nil, nil, err
	}

	return fileURL, hash, nil
}
