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
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/util/pkg/hashing"
)

const (
	defaultKopsBaseUrl = "https://kubeupv2.s3.amazonaws.com/kops/%s/"

	// defaultKopsMirrorBase will be detected and automatically set to pull from the defaultKopsMirrors
	defaultKopsMirrorBase = "https://kubeupv2.s3.amazonaws.com/kops/%s/"
)

// mirror holds the configuration for a mirror
type mirror struct {
	// URL is the base url
	URL string

	// Replace is a set of string replacements, so that we can follow the mirror's naming rules
	Replace map[string]string
}

// defaultKopsMirrors is a list of our well-known mirrors
// Note that we download in order
var defaultKopsMirrors = []mirror{
	{URL: "https://artifacts.k8s.io/binaries/kops/%s/"},
	{URL: "https://github.com/kubernetes/kops/releases/download/v%s/", Replace: map[string]string{"/": "-"}},
	// We do need to include defaultKopsMirrorBase - the list replaces the base url
	{URL: "https://kubeupv2.s3.amazonaws.com/kops/%s/"},
}

var kopsBaseUrl *url.URL

// nodeUpAsset caches the nodeup download urls/hash
var nodeUpAsset *MirroredAsset

// protokubeLocation caches the protokubeLocation url
var protokubeLocation *url.URL

// protokubeHash caches the hash for protokube
var protokubeHash *hashing.Hash

// BaseUrl returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseUrl() (*url.URL, error) {
	// returning cached value
	// Avoid repeated logging
	if kopsBaseUrl != nil {
		klog.V(8).Infof("Using cached kopsBaseUrl url: %q", kopsBaseUrl.String())
		return copyBaseURL(kopsBaseUrl)
	}

	baseUrlString := os.Getenv("KOPS_BASE_URL")
	var err error
	if baseUrlString == "" {
		baseUrlString = fmt.Sprintf(defaultKopsBaseUrl, kops.Version)
		klog.V(8).Infof("Using default base url: %q", baseUrlString)
		kopsBaseUrl, err = url.Parse(baseUrlString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %q as a url: %v", baseUrlString, err)
		}
	} else {
		kopsBaseUrl, err = url.Parse(baseUrlString)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var KOPS_BASE_URL %q as a url: %v", baseUrlString, err)
		}
		klog.Warningf("Using base url from KOPS_BASE_URL env var: %q", baseUrlString)
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

// NodeUpAsset returns the asset for where nodeup should be downloaded
func NodeUpAsset(assetsBuilder *assets.AssetBuilder) (*MirroredAsset, error) {
	// Avoid repeated logging
	if nodeUpAsset != nil {
		// Avoid repeated logging
		klog.V(8).Infof("Using cached nodeup location: %v", nodeUpAsset.Locations)
		return nodeUpAsset, nil
	}
	env := os.Getenv("NODEUP_URL")
	var err error
	var u *url.URL
	var hash *hashing.Hash
	if env == "" {
		u, hash, err = KopsFileUrl("linux/amd64/nodeup", assetsBuilder)
		if err != nil {
			return nil, err
		}
		klog.V(8).Infof("Using default nodeup location: %q", u.String())
	} else {
		u, err = url.Parse(env)
		if err != nil {
			return nil, fmt.Errorf("unable to parse env var NODEUP_URL %q as a url: %v", env, err)
		}

		u, hash, err = assetsBuilder.RemapFileAndSHA(u)
		if err != nil {
			return nil, err
		}
		klog.Warningf("Using nodeup location from NODEUP_URL env var: %q", u.String())
	}

	asset := BuildMirroredAsset(u, hash)

	nodeUpAsset = asset

	return asset, nil
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
		klog.V(8).Infof("Using cached protokube location: %q", protokubeLocation)
		return protokubeLocation, protokubeHash, nil
	}
	env := os.Getenv("PROTOKUBE_IMAGE")
	var err error
	if env == "" {
		protokubeLocation, protokubeHash, err = KopsFileUrl("images/protokube.tar.gz", assetsBuilder)
		if err != nil {
			return nil, nil, err
		}
		klog.V(8).Infof("Using default protokube location: %q", protokubeLocation)
	} else {
		protokubeImageSource, err := url.Parse(env)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse env var PROTOKUBE_IMAGE %q as a url: %v", env, err)
		}

		protokubeLocation, protokubeHash, err = assetsBuilder.RemapFileAndSHA(protokubeImageSource)
		if err != nil {
			return nil, nil, err
		}
		klog.Warningf("Using protokube location from PROTOKUBE_IMAGE env var: %q", protokubeLocation)
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

type MirroredAsset struct {
	Locations []string
	Hash      *hashing.Hash
}

// BuildMirroredAsset checks to see if this is a file under the standard base location, and if so constructs some mirror locations
func BuildMirroredAsset(u *url.URL, hash *hashing.Hash) *MirroredAsset {
	baseUrlString := fmt.Sprintf(defaultKopsMirrorBase, kops.Version)
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
			klog.Warningf("not using mirrors for asset %s as it does not have a known hash", u.String())
		} else {
			suffix := strings.TrimPrefix(urlString, baseUrlString)
			// This is under our base url - add our well-known mirrors
			a.Locations = []string{}
			for _, m := range defaultKopsMirrors {
				filename := suffix
				for k, v := range m.Replace {
					filename = strings.Replace(filename, k, v, -1)
				}
				base := fmt.Sprintf(m.URL, kops.Version)
				a.Locations = append(a.Locations, base+filename)
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
