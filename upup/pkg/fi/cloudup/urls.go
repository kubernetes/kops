/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"net/url"
	"os"

	"path"

	"github.com/golang/glog"
	"k8s.io/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/hashing"
)

const defaultKopsBaseUrl = "https://kubeupv2.s3.amazonaws.com/kops/%s/"

var kopsBaseUrl *url.URL

// nodeUpLocation caches the nodeUpLocation url
var nodeUpLocation *url.URL

// nodeUpHash caches the hash for nodeup
var nodeUpHash *hashing.Hash

// protokubeLocation caches the protokubeLocation url
var protokubeLocation *url.URL

// protokubeHash caches the hash for protokube
var protokubeHash *hashing.Hash

// BaseUrl returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseUrl() (*url.URL, error) {
	// returning cached value
	// Avoid repeated logging
	if kopsBaseUrl != nil {
		glog.V(8).Infof("Using cached kopsBaseUrl url: %q", kopsBaseUrl.String())
		return copyBaseURL(kopsBaseUrl)
	}

	baseUrlString := os.Getenv("KOPS_BASE_URL")
	var err error
	if baseUrlString == "" {
		baseUrlString = fmt.Sprintf(defaultKopsBaseUrl, kops.Version)
		glog.V(8).Infof("Using default base url: %q", baseUrlString)
		kopsBaseUrl, err = url.Parse(baseUrlString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %q as a url: %v", baseUrlString, err)
		}
	} else {
		kopsBaseUrl, err = url.Parse(baseUrlString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var KOPS_BASE_URL %q as a url: %v", baseUrlString, err)
		}
		glog.Warningf("Using base url from KOPS_BASE_URL env var: %q", baseUrlString)
	}

	return copyBaseURL(kopsBaseUrl)
}

// copyBaseURL makes a copy of the base url or the path.Joins can append stuff to this URL
func copyBaseURL(base *url.URL) (*url.URL, error) {
	u, err := url.Parse(base.String())
	if err != nil {
		return nil, err
	}
	return u, nil
}

// SetKopsAssetsLocations sets the kops assets locations
// This func adds kops binary to the list of file assets, and stages the binary in the assets file repository
func SetKopsAssetsLocations(assetsBuilder *assets.AssetBuilder) error {
	for _, s := range []string{
		"linux/amd64/kops", "darwin/amd64/kops",
	} {
		_, _, err := KopsFileUrl(s, assetsBuilder)
		if err != nil {
			return err
		}
	}
	return nil
}

// NodeUpLocation returns the URL where nodeup should be downloaded
func NodeUpLocation(assetsBuilder *assets.AssetBuilder) (*url.URL, *hashing.Hash, error) {
	// Avoid repeated logging
	if nodeUpLocation != nil && nodeUpHash != nil {
		// Avoid repeated logging
		glog.V(8).Infof("Using cached nodeup location: %q", nodeUpLocation.String())
		return nodeUpLocation, nodeUpHash, nil
	}
	env := os.Getenv("NODEUP_URL")
	var err error
	if env == "" {
		nodeUpLocation, nodeUpHash, err = KopsFileUrl("linux/amd64/nodeup", assetsBuilder)
		if err != nil {
			return nil, nil, err
		}
		glog.V(8).Infof("Using default nodeup location: %q", nodeUpLocation.String())
	} else {
		nodeUpLocation, err = url.Parse(env)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse env var NODEUP_URL %q as a url: %v", env, err)
		}

		nodeUpLocation, nodeUpHash, err = assetsBuilder.RemapFileAndSHA(nodeUpLocation)
		if err != nil {
			return nil, nil, err
		}
		glog.Warningf("Using nodeup location from NODEUP_URL env var: %q", nodeUpLocation.String())
	}

	return nodeUpLocation, nodeUpHash, nil
}

// TODO make this a container when hosted assets
// TODO does this support a docker as well??
// FIXME comments says this works with a docker already ... need to check on that

// ProtokubeImageSource returns the source for the docker image for protokube.
// Either a docker name (e.g. gcr.io/protokube:1.4), or a URL (https://...) in which case we download
// the contents of the url and docker load it
func ProtokubeImageSource(assetsBuilder *assets.AssetBuilder) (*url.URL, *hashing.Hash, error) {
	// Avoid repeated logging
	if protokubeLocation != nil && protokubeHash != nil {
		glog.V(8).Infof("Using cached protokube location: %q", protokubeLocation)
		return protokubeLocation, protokubeHash, nil
	}
	env := os.Getenv("PROTOKUBE_IMAGE")
	var err error
	if env == "" {
		protokubeLocation, protokubeHash, err = KopsFileUrl("images/protokube.tar.gz", assetsBuilder)
		if err != nil {
			return nil, nil, err
		}
		glog.V(8).Infof("Using default protokube location: %q", protokubeLocation)
	} else {
		protokubeImageSource, err := url.Parse(env)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse env var PROTOKUBE_IMAGE %q as a url: %v", env, err)
		}

		protokubeLocation, protokubeHash, err = assetsBuilder.RemapFileAndSHA(protokubeImageSource)
		if err != nil {
			return nil, nil, err
		}
		glog.Warningf("Using protokube location from PROTOKUBE_IMAGE env var: %q", protokubeLocation)
	}

	return protokubeLocation, protokubeHash, nil
}

// KopsFileUrl returns the base url for the distribution of kops - in particular for nodeup & docker images
func KopsFileUrl(file string, assetBuilder *assets.AssetBuilder) (*url.URL, *hashing.Hash, error) {
	base, err := BaseUrl()
	if err != nil {
		return nil, nil, err
	}

	base.Path = path.Join(base.Path, file)

	fileUrl, hash, err := assetBuilder.RemapFileAndSHA(base)
	if err != nil {
		return nil, nil, err
	}

	return fileUrl, hash, nil
}
