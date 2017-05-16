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

package cloudup

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"

	api_util "k8s.io/kops/pkg/apis/kops/util"

	"k8s.io/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

type ApplyInventory struct {
	Cluster             *api.Cluster
	InstanceGroups      []*api.InstanceGroup
	NodeUpConfigBuilder func(ig *api.InstanceGroup) (*nodeup.NodeUpConfig, error)
	BootstrapContainers *sets.String
	TaskMap             map[string]fi.Task
	NodeupLocation      string
}

// buildInventoryAssets builds a map of all unique inventory assets.
func (i *ApplyInventory) BuildInventoryAssets() (*api.Inventory, error) {

	if i.Cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}
	if i.InstanceGroups == nil {
		return nil, fmt.Errorf("instance groups cannot be nil")
	}
	if i.NodeUpConfigBuilder == nil {
		return nil, fmt.Errorf("builder function cannot be nil")
	}

	inv := &api.Inventory{
		Spec: api.InventorySpec{
			CompressedFileAssets: make([]*api.CompressedFileAsset, 0),
			ContainerAssets:      make([]*api.ContainerAsset, 0),
			ExecutableFileAsset:  make([]*api.ExecutableFileAsset, 0),
			HostAssets:           make([]*api.HostAsset, 0),
			Cluster:              &i.Cluster.Spec,
			KopsVersion:          &kops.Version,
		},
	}

	inventoryBinaryMap := make(map[string]*api.ExecutableFileAsset)
	inventoryCompressedMap := make(map[string]*api.CompressedFileAsset)

	spec := i.Cluster.Spec

	// nodeup binary
	inventoryBinaryMap["nodeup"] = &api.ExecutableFileAsset{
		Location: nodeUpLocation,
		Name:     "nodeup",
	}

	// Build Kubernetes Base Containers
	{
		etcd, err := components.GetGoogleImageRepositoryContainer(&spec, "etcd:2.2.1")

		if err != nil {
			return nil, err
		}

		pause, err := components.GetGoogleImageRepositoryContainer(&spec, "pause-amd64:3.0")

		if err != nil {
			return nil, err
		}

		k8sContainers := map[string]string{
			spec.KubeAPIServer.Image:         "KubeAPIServer",
			spec.KubeControllerManager.Image: "KubeControllerManager",
			spec.KubeProxy.Image:             "KubeProxy",
			spec.KubeScheduler.Image:         "KubeScheduler",
			// TODO This value is hardcoded in protokube.
			etcd:  "Etcd",
			pause: "pause",
		}

		for k, v := range k8sContainers {
			c, err := parseContainer(k, v)
			if err != nil {
				return nil, err
			}
			inv.Spec.ContainerAssets = append(inv.Spec.ContainerAssets, c)
		}

	}

	channelLocation, err := api.ParseChannelLocation(i.Cluster.Spec.Channel)

	if err != nil {
		return nil, fmt.Errorf("unable to get channel location: %v", err)
	}

	inv.Spec.ChannelAsset = &api.ChannelAsset{
		Location: channelLocation,
	}

	if strings.HasSuffix(channelLocation, "alpha") {
		inv.Spec.ChannelAsset.Name = "alpha"
	} else {
		inv.Spec.ChannelAsset.Name = "stable"
	}

	for _, ig := range i.InstanceGroups {
		n, err := i.NodeUpConfigBuilder(ig)
		if err != nil {
			return nil, fmt.Errorf("unable to render node config: %v", err)
		}

		host := &api.HostAsset{
			Name:          ig.Spec.Image,
			Cloud:         i.Cluster.Spec.CloudProvider,
			Role:          string(ig.Spec.Role),
			InstanceGroup: ig.ObjectMeta.Name,
		}

		inv.Spec.HostAssets = append(inv.Spec.HostAssets, host)

		// binary asset
		for _, a := range n.Assets {

			asset := strings.Split(a, "@")
			url := asset[1]
			name := strings.Split(url, "/")
			fileName := name[len(name)-1]

			if strings.HasSuffix(fileName, "gz") {
				inventoryCompressedMap[asset[0]] = &api.CompressedFileAsset{
					Location: url,
					Name:     fileName,
				}

			} else {

				inventoryBinaryMap[asset[0]] = &api.ExecutableFileAsset{
					Location: url,
					Name:     fileName,
					SHA:      url + ".sha",
				}
			}
		}
	}

	c := &api.ContainerAsset{
		Name: protokubeImageSource.Name,
		Tag:  strings.Replace(kops.Version, "+", "-", -1),
	}

	if protokubeImageSource.Hash != "" {
		c.Location = protokubeImageSource.Source
		c.Hash = protokubeImageSource.Hash
		c.SHA = protokubeImageSource.Source + ".sha"
	}

	inv.Spec.ContainerAssets = append(inv.Spec.ContainerAssets, c)
	// if the ManagedFile tasks have not runs
	sv, err := api_util.ParseKubernetesVersion(spec.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q", spec.KubernetesVersion)
	}

	optionsContext := &components.OptionsContext{
		KubernetesVersion: *sv,
	}

	isVersionGTE1_6 := optionsContext.IsKubernetesGTE("1.6")
	isVersionLT1_6 := optionsContext.IsKubernetesLT("1.6")

	v := sv.String()
	inv.Spec.KubernetesVersion = &v

	for _, task := range i.TaskMap {
		switch m := task.(type) {
		case *fitasks.ManagedFile:

			pre16 := strings.Contains(*m.Location, "pre-k8s-1.6")

			if pre16 && isVersionGTE1_6 {
				continue
			} else if !pre16 && isVersionLT1_6 {
				continue
			}

			m.Contents.AsBytes()
		default:

		}

	}

	for _, b := range i.BootstrapContainers.List() {
		c, err := validation.ParseContainer(b)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse container: %v", err)
		}
		inv.Spec.ContainerAssets = append(inv.Spec.ContainerAssets, c)
	}

	// reduce map to a slice
	for _, value := range inventoryBinaryMap {
		inv.Spec.ExecutableFileAsset = append(inv.Spec.ExecutableFileAsset, value)
	}

	// reduce map to a slice
	for _, value := range inventoryCompressedMap {

		if !strings.HasPrefix(value.Location, "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-") {
			value.SHA = value.Location + ".sha1"
		}

		inv.Spec.CompressedFileAssets = append(inv.Spec.CompressedFileAssets, value)
	}

	glog.V(2).Infof("data %+v", inv)

	return inv, nil
}

func parseContainer(image string, name string) (*api.ContainerAsset, error) {
	// List of all containers in the API
	c, err := validation.ParseContainer(image)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse %s container %q: %v", name, image, err)
	}
	c.Name = name
	return c, nil
}
