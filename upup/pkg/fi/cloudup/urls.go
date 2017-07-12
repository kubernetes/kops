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
	"os"
	"strings"

	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
)

// baseUrl caches the BaseUrl value
var baseUrl string

// BaseUrl returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseUrl(spec *api.ClusterSpec) string {
	if baseUrl != "" {
		// Avoid repeated logging
		return baseUrl
	}

	baseUrl = os.Getenv("KOPS_BASE_URL")
	if baseUrl != "" {
		glog.Warningf("Using base url from KOPS_BASE_URL env var: %q", baseUrl)
	} else {
		version := strings.Replace(kops.Version, "+", "%2B", -1)
		kopsUrl := "/kops/" + version + "/"

		if spec.Assets != nil && spec.Assets.FileRepository != nil {
			repo := strings.TrimSuffix(*spec.Assets.FileRepository, "/")
			baseUrl = repo + kopsUrl
			glog.Warningf("Using custom base url: %q", baseUrl)
		} else {
			baseUrl = "https://kubeupv2.s3.amazonaws.com" + kopsUrl
			glog.V(4).Infof("Using default url: %q", baseUrl)
		}
	}

	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}

	return baseUrl
}

// nodeUpLocation caches the NodeUpLocation value
var nodeUpLocation string

// NodeUpLocation returns the URL where nodeup should be downloaded
func NodeUpLocation(spec *api.ClusterSpec) string {
	if nodeUpLocation != "" {
		// Avoid repeated logging
		return nodeUpLocation
	}
	nodeUpLocation = os.Getenv("NODEUP_URL")
	if nodeUpLocation == "" {
		nodeUpLocation = BaseUrl(spec) + "linux/amd64/nodeup"
		glog.V(4).Infof("Using default nodeup location set via kops base url: %q", nodeUpLocation)
	} else {
		glog.Warningf("Using nodeup location from NODEUP_URL env var: %q", nodeUpLocation)
	}
	return nodeUpLocation
}

// protokubeImageSource caches the ProtokubeImageSource value
var protokubeImageSource *nodeup.Image

// ProtokubeImageSource returns the source for the docker image for protokube.
// Either a docker name (e.g. gcr.io/protokube:1.4), or a URL (https://...) in which case we download
// the contents of the url and docker load it
func ProtokubeImageSource(spec *api.ClusterSpec) (*nodeup.Image, error) {

	protokubeImageSourceEnv := os.Getenv("PROTOKUBE_IMAGE")

	// return cached
	if protokubeImageSource != nil {
		return protokubeImageSource, nil
	}

	var proto string
	var err error

	// only do this test it the env variable is not set
	if protokubeImageSourceEnv == "" {
		// A cluster asset container registry value can the default protokube name.
		// Use the container registry value.
		proto, err = assets.GetContainerAndRegistryAsString(spec, kops.DefaultProtokubeImageName())
		if err != nil {
			return nil, fmt.Errorf("unable to get protokube container name: %v", err)
		}
	}

	if protokubeImageSourceEnv != "" {

		// use env variable
		hash, err := findHash(protokubeImageSourceEnv)
		if err != nil {
			return nil, err
		}

		protokubeImageSource = &nodeup.Image{
			Name:   kops.DefaultProtokubeImageName(),
			Source: protokubeImageSourceEnv,
			Hash:   hash.Hex(),
		}
		glog.Warningf("Using protokube location from PROTOKUBE_IMAGE env var: %q", protokubeImageSource)

	} else if proto != kops.DefaultProtokubeImageName() {

		// Assets.ContainerRegistry is set
		if err != nil {
			return nil, err
		}

		protokubeImageSource = &nodeup.Image{
			Name:   proto,
			Source: proto,
		}
		glog.Infof("Using protokube location from assets container registry %q", proto)
	} else {

		//use default from url
		source := BaseUrl(spec) + "images/protokube.tar.gz"
		hash, err := findHash(source)
		if err != nil {
			return nil, err
		}

		protokubeImageSource = &nodeup.Image{
			Name:   kops.DefaultProtokubeImageName(),
			Source: source,
			Hash:   hash.Hex(),
		}
		glog.V(4).Infof("Using default protokube location set via kops base url: %v", protokubeImageSource)

	}

	return protokubeImageSource, nil
}
