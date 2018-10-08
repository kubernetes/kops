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
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops"
	"k8s.io/kops/pkg/assets"
)

const defaultKopsBaseUrl = "https://kubeupv2.s3.amazonaws.com/kops/%s/"

// defaultKopsMirrorBase will be detected and automatically set to pull from the defaultKopsMirrors
const defaultKopsMirrorBase = "https://kubeupv2.s3.amazonaws.com/kops/"

// defaultKopsMirrors is a list of our well-known mirrors
var defaultKopsMirrors = []string{
	"https://github.com/kubernetes/kops/releases/download/",
	// We do need to include defaultKopsMirrorBase - the list replaces the base url
	"https://kubeupv2.s3.amazonaws.com/kops/",
}

var kopsBaseUrl *url.URL

// nodeupAsset caches the nodeup Asset
var nodeupAsset *assets.FileAsset

// protokubeAsset caches the protokube Asset
var protokubeAsset *assets.FileAsset

// BaseURL returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseURL() (*url.URL, error) {
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
		_, err := KopsAsset(s, assetsBuilder)
		if err != nil {
			return err
		}
	}
	return nil
}

// NodeupAsset returns an asset representing how to download nodeup
func NodeupAsset(assetsBuilder *assets.AssetBuilder) (*assets.FileAsset, error) {
	// Avoid repeated logging
	if nodeupAsset != nil {
		// Avoid repeated logging
		glog.V(8).Infof("Using cached nodeup asset: %v", nodeupAsset)
		return nodeupAsset, nil
	}
	env := os.Getenv("NODEUP_URL")
	var err error
	if env == "" {
		nodeupAsset, err = KopsAsset("linux/amd64/nodeup", assetsBuilder)
		if err != nil {
			return nil, err
		}
		glog.V(8).Infof("Using default nodeup asset: %q", nodeupAsset)
	} else {
		nodeupURL, err := url.Parse(env)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var NODEUP_URL %q as a url: %v", env, err)
		}

		nodeupAsset, err = assetsBuilder.BuildAssetForURL(nodeupURL)
		if err != nil {
			return nil, err
		}
		glog.Warningf("Using nodeup location from NODEUP_URL env var: %q", nodeupURL.String())
	}

	return nodeupAsset, nil
}

// TODO make this a container when hosted assets
// TODO does this support a docker as well??
// FIXME comments says this works with a docker already ... need to check on that

// ProtokubeImageAsset returns the source for the docker image for protokube.
// Either a docker name (e.g. gcr.io/protokube:1.4), or a URL (https://...) in which case we download
// the contents of the url and docker load it
func ProtokubeImageAsset(assetsBuilder *assets.AssetBuilder) (*assets.FileAsset, error) {
	// Avoid repeated logging
	if protokubeAsset != nil {
		glog.V(8).Infof("Using cached protokube location: %v", protokubeAsset)
		return protokubeAsset, nil
	}
	env := os.Getenv("PROTOKUBE_IMAGE")
	var err error
	if env == "" {
		protokubeAsset, err = KopsAsset("images/protokube.tar.gz", assetsBuilder)
		if err != nil {
			return nil, err
		}
		glog.V(8).Infof("Using default protokube location: %q", protokubeAsset)
	} else {
		protokubeImageSource, err := url.Parse(env)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var PROTOKUBE_IMAGE %q as a url: %v", env, err)
		}

		protokubeAsset, err = assetsBuilder.BuildAssetForURL(protokubeImageSource)
		if err != nil {
			return nil, err
		}
		glog.Warningf("Using protokube from PROTOKUBE_IMAGE env var: %q", protokubeAsset)
	}

	return protokubeAsset, nil
}

// KopsAsset returns a FileAsset for a kops artifact, such as protokube or nodeup
func KopsAsset(file string, assetBuilder *assets.AssetBuilder) (*assets.FileAsset, error) {
	base, err := BaseURL()
	if err != nil {
		return nil, err
	}

	base.Path = path.Join(base.Path, file)

	asset, err := assetBuilder.BuildAssetForURL(base)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

type MirroredAsset struct {
	Locations []string
	Hash      *hashing.Hash
}

// BuildMirroredAsset checks to see if this is a file under the standard base location, and if so constructs some mirror locations
func BuildMirroredAsset(u *url.URL, hash *hashing.Hash) *MirroredAsset {
	baseUrlString := defaultKopsMirrorBase
	if !strings.HasSuffix(baseUrlString, "/") {
		baseUrlString += "/"
	}

	a := &MirroredAsset{
		Hash: hash,
	}

	urlString := u.String()
	a.Locations = []string{urlString}

	// Look at mirrors
	if strings.HasPrefix(urlString, baseUrlString) {
		if hash == nil {
			glog.Warningf("not using mirrors for asset %s as it does not have a known hash", u.String())
		} else {
			suffix := strings.TrimPrefix(urlString, baseUrlString)
			// This is under our base url - add our well-known mirrors
			a.Locations = []string{}
			for _, m := range defaultKopsMirrors {
				a.Locations = append(a.Locations, m+suffix)
			}
		}
	}

	return a
}

func (a *MirroredAsset) CompactString() string {
	var s string
	if a.Hash != nil {
		s = a.Hash.Hex()
	}
	s += "@" + strings.Join(a.Locations, ",")
	return s
}
