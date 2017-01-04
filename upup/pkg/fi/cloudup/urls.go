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
)

// baseUrl caches the BaseUrl value
var baseUrl string

// BaseUrl returns the base url for the distribution of kops - in particular for nodeup & docker images
func BaseUrl() string {
	if baseUrl != "" {
		// Avoid repeated logging
		return baseUrl
	}

	baseUrl = os.Getenv("KOPS_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://kubeupv2.s3.amazonaws.com/kops/" + kops.Version + "/"
		glog.V(4).Infof("Using default base url: %q", baseUrl)
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
func NodeUpLocation() string {
	if nodeUpLocation != "" {
		// Avoid repeated logging
		return nodeUpLocation
	}
	nodeUpLocation = os.Getenv("NODEUP_URL")
	if nodeUpLocation == "" {
		nodeUpLocation = BaseUrl() + "linux/amd64/nodeup"
		glog.V(4).Infof("Using default nodeup location: %q", nodeUpLocation)
	} else {
		glog.Warningf("Using nodeup location from NODEUP_URL env var: %q", nodeUpLocation)
	}
	return nodeUpLocation
}

// protokubeImageSource caches the ProtokubeImageSource value
var protokubeImageSource string

// ProtokubeImageSource returns the source for the docker image for protokube.
// Either a docker name (e.g. gcr.io/protokube:1.4), or a URL (https://...) in which case we download
// the contents of the url and docker load it
func ProtokubeImageSource() string {
	if protokubeImageSource != "" {
		// Avoid repeated logging
		return protokubeImageSource
	}
	protokubeImageSource = os.Getenv("PROTOKUBE_IMAGE")
	if protokubeImageSource == "" {
		protokubeImageSource = BaseUrl() + "images/protokube.tar.gz"
		glog.V(4).Infof("Using default protokube location: %q", protokubeImageSource)
	} else {
		glog.Warningf("Using protokube location from PROTOKUBE_IMAGE env var: %q", protokubeImageSource)
	}
	return protokubeImageSource
}
