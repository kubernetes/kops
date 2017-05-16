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

	"github.com/golang/glog"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/apis/nodeup"
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
	if baseUrl == "" {
		version := strings.Replace(kops.Version, "+", "%2B", -1)
		kopsUrl := "/kops/" + version + "/"

		if spec.Assets != nil && spec.Assets.FileRepository != nil {
			repo := strings.TrimSuffix(*spec.Assets.FileRepository, "/")
			baseUrl = repo + kopsUrl
		} else {
			baseUrl = "https://kubeupv2.s3.amazonaws.com" + kopsUrl
		}
		glog.V(4).Infof("Using custom base url: %q", baseUrl)
	} else {
		glog.Warningf("Using base url from KOPS_BASE_URL env var: %q", baseUrl)
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
		glog.V(4).Infof("Using default nodeup location: %q", nodeUpLocation)
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

	proto, err := validation.GetContainerAndRepoAsString(spec, kops.DefaultProtokubeImageName())
	if err != nil {
		return nil, err
	}

	if proto != kops.DefaultProtokubeImageName() {

		// Assets.ContainerRepository is set
		if err != nil {
			return nil, err
		}

		protokubeImageSource = &nodeup.Image{
			Name:   proto,
			Source: proto,
		}
		glog.Infof("Using protokube location from assets container repository %q", *spec.Assets.ContainerRepository)

	} else if protokubeImageSourceEnv != "" {

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
		glog.V(4).Infof("Using default protokube location: %v", protokubeImageSource)

	}

	return protokubeImageSource, nil
}
