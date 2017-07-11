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
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"

	"k8s.io/kubernetes/pkg/kubectl/resource"
)

type ApplyInventory struct {
	Cluster             *api.Cluster
	InstanceGroups      []*api.InstanceGroup
	NodeUpConfigBuilder func(ig *api.InstanceGroup) (*nodeup.NodeUpConfig, error)
	TaskMap             map[string]fi.Task
	NodeupLocation      string
	AssetBuilder        *assets.AssetBuilder
}

// BuildInventoryAssets populates the Inventory of a kops Kubernetes cluster.  This func is only
// accessible when running an ApplyClusterCmd.Run() with a target of TargetInventory.
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
	inventoryContainerMap := make(map[string]*api.ContainerAsset)
	spec := i.Cluster.Spec

	// nodeup binary
	inventoryBinaryMap["nodeup"] = &api.ExecutableFileAsset{
		Location: nodeUpLocation,
		Name:     "nodeup",
		SHA:      nodeUpLocation + ".sha1",
	}

	// kops binary
	kopsLocation := strings.TrimSuffix(nodeUpLocation, "nodeup") + "kops"
	inventoryBinaryMap["kops"] = &api.ExecutableFileAsset{
		Location: kopsLocation,
		Name:     "kops",
		SHA:      kopsLocation + ".sha1",
	}

	// Build Kubernetes Base Containers
	{
		etcd, err := assets.GetGoogleImageRegistryContainer(&spec, "etcd:2.2.1")

		if err != nil {
			return nil, err
		}

		pause, err := assets.GetGoogleImageRegistryContainer(&spec, "pause-amd64:3.0")

		if err != nil {
			return nil, err
		}

		k8sContainers := map[string]string{
			spec.KubeAPIServer.Image:         "KubeAPIServer",
			spec.KubeControllerManager.Image: "KubeControllerManager",
			spec.KubeProxy.Image:             "KubeProxy",
			spec.KubeScheduler.Image:         "KubeScheduler",
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

	// Build Inventory channel
	{
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
	}

	// Build host information and assets created by nodeup.  I choose to parse
	// the assets to allow dynamic creation and deletion of assets.  Kubernetes assets
	// are very stable, but these assets are growing in kops.
	{
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

				glog.V(2).Infof("Asset %s", a)
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

					a := &api.ExecutableFileAsset{
						Location: url,
						Name:     fileName,
						SHA:      url + ".sha1",
					}

					if !strings.HasPrefix(a.Location, "https://storage.googleapis.com/kubernetes-release/network-plugins/cni-") {
						a.SHA = a.Location + ".sha1"
					}

					inventoryBinaryMap[asset[0]] = a
				}
			}
		}
	}

	// protokube
	{
		c := &api.ContainerAsset{
			Name: protokubeImageSource.Name,
			Tag:  strings.Replace(kops.Version, "+", "-", -1),
		}

		if protokubeImageSource.Hash != "" {
			c.Location = protokubeImageSource.Source
			c.Hash = protokubeImageSource.Hash
			c.SHA = protokubeImageSource.Source + ".sha1"
		} else {
			c.Location = protokubeImageSource.Name
		}
		inv.Spec.ContainerAssets = append(inv.Spec.ContainerAssets, c)
	}

	// assets like cni provider
	{
		for _, a := range i.AssetBuilder.Assets {
			glog.V(2).Infof("  %s %s", a.Origin, a.Mirror)
			c, err := assets.ParseContainer(a.Mirror)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse container: %v", err)
			}
			inventoryContainerMap[a.Mirror] = c
		}
	}

	// Normalize the data.

	// reduce map to a slice
	for _, value := range inventoryContainerMap {
		inv.Spec.ContainerAssets = append(inv.Spec.ContainerAssets, value)
	}

	// reduce map to a slice
	for _, value := range inventoryBinaryMap {
		inv.Spec.ExecutableFileAsset = append(inv.Spec.ExecutableFileAsset, value)
	}

	// reduce map to a slice
	for _, value := range inventoryCompressedMap {
		inv.Spec.CompressedFileAssets = append(inv.Spec.CompressedFileAssets, value)
	}

	return inv, nil
}

func parseContainer(image string, name string) (*api.ContainerAsset, error) {
	// List of all containers in the API
	c, err := assets.ParseContainer(image)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse %s container %q: %v", name, image, err)
	}
	if c.Name == "" {
		c.Name = name
	}
	return c, nil
}

type ExtractInventory struct {
	Clientset simple.Clientset
	resource.FilenameOptions
	ClusterName       string
	Channel           string
	KubernetesVersion string
	SSHPublicKey      string
}

func (e *ExtractInventory) ExtractAssets() (*api.Inventory, string, error) {
	var cluster *api.Cluster
	var ig []*api.InstanceGroup
	var err error

	if len(e.Filenames) != 0 {
		cluster, ig, err = e.readFiles()
		if err != nil {
			return nil, "", fmt.Errorf("Error loading file(s) %q, %v", e.Filenames, err)
		}

		e.ClusterName = cluster.ObjectMeta.Name
	} else if e.ClusterName != "" {

		cluster, err = e.Clientset.GetCluster(e.ClusterName)

		if err != nil {
			return nil, "", fmt.Errorf("Error getting cluster  %q", err)
		}

		if cluster == nil {
			return nil, "", fmt.Errorf("cluster not found %q", e.ClusterName)
		}

	}

	if cluster == nil {
		return nil, "", fmt.Errorf("error getting cluster")
	}

	if cluster.Spec.KubernetesVersion == "" {
		channel, err := api.LoadChannel(e.Channel)
		if err != nil {
			return nil, "", fmt.Errorf("Unable to load channel %q", err)
		}
		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
			e.KubernetesVersion = cluster.Spec.KubernetesVersion
		} else {

			return nil, "", fmt.Errorf("Unable to find kubernetes version")
		}

	}

	// TODO check if the cluster has a key?
	// TODO how do we get this out of here??
	// TODO - Reference: https://github.com/kubernetes/kops/issues/2659
	sshPublicKeys := make(map[string][]byte)
	if e.SSHPublicKey != "" {
		e.SSHPublicKey = utils.ExpandPath(e.SSHPublicKey)
		authorized, err := ioutil.ReadFile(e.SSHPublicKey)
		if err != nil {
			return nil, "", fmt.Errorf("error reading SSH key file %q: %v", e.SSHPublicKey, err)
		}
		sshPublicKeys[fi.SecretNameSSHPrimary] = authorized

	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return nil, "", err
	}

	for k, data := range sshPublicKeys {
		err = keyStore.AddSSHPublicKey(k, data)
		if err != nil {
			return nil, "", fmt.Errorf("error addding SSH public key: %v", err)
		}
	}

	applyClusterCmd := &ApplyClusterCmd{
		Clientset:      e.Clientset,
		Cluster:        cluster,
		InstanceGroups: ig,
		TargetName:     TargetInventory,
		Models:         []string{"config", "proto", "cloudup"},
	}

	err = applyClusterCmd.Run()

	if err != nil {
		return nil, "", fmt.Errorf("error applying cluster build: %v", err)
	}

	a := applyClusterCmd.Inventory

	clusterExists, err := e.Clientset.GetCluster(e.ClusterName)

	// Need to delete tmp ssh key
	// If we can get ApplyClusterCmd to run w/o ssh key we will not have to create it
	if err != nil || clusterExists == nil {

		glog.V(2).Infof("Deleting cluster resources for %q", e.ClusterName)
		configBase, err := registry.ConfigBase(cluster)

		if err != nil {
			return nil, "", err
		}

		err = registry.DeleteAllClusterState(configBase)
		if err != nil {
			return nil, "", fmt.Errorf("error removing cluster from state store: %v", err)
		}

	}

	return a, e.ClusterName, nil
}

// readFiles inputs and marshalls YAML files.
func (e *ExtractInventory) readFiles() (*api.Cluster, []*api.InstanceGroup, error) {

	codec := api.Codecs.UniversalDecoder(api.SchemeGroupVersion)

	var cluster *api.Cluster
	var ig []*api.InstanceGroup

	for _, f := range e.Filenames {
		var sb bytes.Buffer
		fmt.Fprintf(&sb, "\n")

		contents, err := vfs.Context.ReadFile(f)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading file %q: %v", f, err)
		}

		sections := bytes.Split(contents, []byte("\n---\n"))
		for _, section := range sections {
			defaults := &schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
			}
			o, _, err := codec.Decode(section, defaults, nil)
			if err != nil {
				return nil, nil, fmt.Errorf("error parsing file %q: %v", f, err)
			}

			switch v := o.(type) {
			case *api.Cluster:
				err = PerformAssignments(v)
				if err != nil {
					return nil, nil, fmt.Errorf("error populating configuration: %v", err)
				}
				cluster = v
			case *api.InstanceGroup:
				ig = append(ig, v)
			default:
				glog.V(8).Infof("Type of object was %T", v)
			}
		}
	}

	return cluster, ig, nil
}
