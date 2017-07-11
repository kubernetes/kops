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

// This validation code is utilizing the docker reference funcs which parse and validate docker container
// syntax.  For instance gci.io/foo/mycontainer:42 passes, while gci.io//foo/mycontainer:42 fails.

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

// ParseContainer parses s and returns a syntactically valid Reference.
// If an error was encountered it is returned, along with a nil Reference.
func ParseContainer(s string) (*kops.ContainerAsset, error) {
	ref, err := reference.Parse(s)

	if err != nil {
		return nil, fmt.Errorf("unable to parse container %q: %v", s, err)
	}

	asset := &kops.ContainerAsset{}
	switch r := ref.(type) {
	case reference.NamedTagged:
		asset.Tag = r.Tag()
		asset.Domain, asset.Name = reference.SplitHostname(r)
	case reference.Canonical:
		asset.Domain, asset.Name = reference.SplitHostname(r)
		asset.Digest = r.Digest().String()
	case reference.Named:
		asset.Domain, asset.Name = reference.SplitHostname(r)
	case reference.Tagged:
		asset.Tag = r.Tag()
	case reference.Digested:
		asset.Digest = r.Digest().String()
	case reference.Reference:
		asset.String = r.String()
	default:
		glog.Errorf("We should not get here")
	}

	asset.String = ref.String()

	return asset, nil

}

// GetRegistryAsString get the asset container registry as a string if a cluster has an registry.
func GetRegistryAsString(clusterSpec *kops.ClusterSpec) (string, error) {

	if clusterSpec == nil {
		return "", fmt.Errorf("unable to parse assets container registry as cluster spec is nil")

	}

	asset, err := GetContainerRegistry(clusterSpec)

	if err != nil {
		return "", fmt.Errorf("unable to get container registry, asset: %v", err)
	}

	// registry is not set
	if asset == nil {
		return "", nil
	}

	return getRegistry(asset), nil

}

func getRegistry(asset *kops.ContainerAsset) string {

	if asset == nil {
		return ""
	}

	buf := new(bytes.Buffer)

	if asset.Domain != "" {
		buf.WriteString(asset.Domain)
		buf.WriteString("/")
	}

	if asset.Name != "" {
		buf.WriteString(asset.Name)
	}

	return buf.String()
}

func getRepoContainer(registry string, asset *kops.ContainerAsset) (string, error) {

	if asset == nil {
		return "", fmt.Errorf("asset cannot be nil")
	}
	buf := new(bytes.Buffer)

	if registry != "" {
		buf.WriteString(registry)
		buf.WriteString("/")
	} else if asset.Domain != "" {
		buf.WriteString(asset.Domain)
		buf.WriteString("/")
	}

	if asset.Name != "" {
		buf.WriteString(asset.Name)
	}

	if asset.Tag != "" {
		buf.WriteString(":")
		buf.WriteString(asset.Tag)
	}

	if asset.Digest != "" {
		buf.WriteString("@")
		buf.WriteString(asset.Digest)
	}

	return buf.String(), nil
}

// GetContainerRegistry returns a ContainerAsset get the asset container registry if a cluster has an registry.
func GetContainerRegistry(clusterSpec *kops.ClusterSpec) (*kops.ContainerAsset, error) {

	if clusterSpec == nil {
		return nil, fmt.Errorf("unable to parse assets container registry as cluster spec is nil")

	}

	if clusterSpec.Assets != nil && clusterSpec.Assets.ContainerRegistry != nil {
		registry := strings.TrimSuffix(*clusterSpec.Assets.ContainerRegistry, "/")
		asset, err := ParseContainer(registry)

		if err != nil {
			return nil, fmt.Errorf("unable to parse assets container registry api value: %v", registry)
		}

		return asset, nil
	}

	return nil, nil
}

// GetContainerAndRegistryAsString returns a full container string if a cluster has an asset container registry.
func GetContainerAndRegistryAsString(clusterSpec *kops.ClusterSpec, container string) (string, error) {

	if clusterSpec == nil {
		return "", fmt.Errorf("unable to parse assets container registry as cluster spec is nil")

	}

	registry, err := GetContainerRegistry(clusterSpec)

	if err != nil {
		return "", err
	}

	r := getRegistry(registry)

	asset, err := ParseContainer(container)

	if err != nil {
		return "", err
	}

	container, err = getRepoContainer(r, asset)

	if err != nil {
		return "", err
	}

	return container, nil
}

// GetContainerAsString returns a full parsed container string.
func GetContainerAsString(container string) (string, error) {
	asset, err := ParseContainer(container)

	if err != nil {
		return "", fmt.Errorf("unable to parse container: %+v: %v", container, err)
	}

	container, err = getRepoContainer("", asset)

	if err != nil {
		return "", fmt.Errorf("unable to parse container as asset: %v %v", asset, err)
	}

	return container, nil
}

// GetContainer provides a imageName normalized to use a cluster asset container registry if
// the cluster has a registry.
func GetContainer(clusterSpec *kops.ClusterSpec, imageName string) (string, error) {

	imageName, err := GetContainerAndRegistryAsString(clusterSpec, imageName)

	if err != nil {
		glog.Errorf("Unable to get container: %q: %v", imageName, err)
		return "", fmt.Errorf("unable to parse container %q", imageName)
	}

	return imageName, nil
}

const GCR_IO = "gcr.io/google_containers/"

// Cached string for the google container registry.  Usually set to `GCR_IO`.
var googleRepository *string

// GetGoogleImageRegistryContainer returns a container string, it is used for container that are
// typically hosted in the google registry.
func GetGoogleImageRegistryContainer(clusterSpec *kops.ClusterSpec, c string) (string, error) {

	c = strings.TrimPrefix(c, GCR_IO)
	c, err := GetContainerAsString(c)

	glog.V(2).Infof("container %s", c)
	if err != nil {
		return "", fmt.Errorf("Unable to get google image based container, container does not validate: %v", err)
	}

	if googleRepository != nil {
		glog.V(2).Infof("returning container %s", *googleRepository+c)
		return *googleRepository + c, nil
	}

	repo, err := GetRegistryAsString(clusterSpec)
	glog.V(2).Infof("registry %s", repo)

	if err != nil {
		return "", fmt.Errorf("Unable to get google image based container: %v", err)
	}

	if repo != "" {
		repo = repo + "/"
		glog.V(2).Infof("repo = %s", repo)
	} else {
		glog.V(2).Infof("repo = %s", GCR_IO)
		repo = GCR_IO
	}

	googleRepository = &repo
	glog.V(2).Infof("returning container %s", repo+c)

	return repo + c, nil
}

// Image returns the docker image name for the specified component
func Image(component string, clusterSpec *kops.ClusterSpec) (string, error) {

	// TODO figure out if we can add a parameter for version as well
	// TODO https://github.com/kubernetes/kops/pull/2573#discussion_r117329255

	if component == "kube-dns" {
		// TODO: Once we are shipping different versions, start to use them
		i, err := GetGoogleImageRegistryContainer(clusterSpec, "kubedns-amd64:1.3")

		if err != nil {
			return "", err
		}

		return i, nil

	}

	if !IsBaseURL(clusterSpec.KubernetesVersion) {
		c := component + ":" + "v" + clusterSpec.KubernetesVersion
		i, err := GetGoogleImageRegistryContainer(clusterSpec, c)

		if err != nil {
			return "", err
		}

		return i, nil

	}

	baseURL := clusterSpec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	tagURL := baseURL + "/bin/linux/amd64/" + component + ".docker_tag"
	glog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	glog.V(2).Infof("Found tag %q for %q", tag, component)
	c := component + ":" + tag

	i, err := GetGoogleImageRegistryContainer(clusterSpec, c)

	if err != nil {
		return "", err
	}

	return i, nil
}

func IsBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:")
}
