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
	"os"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	k8s_api "k8s.io/client-go/pkg/api"
	k8s_apps "k8s.io/client-go/pkg/apis/apps"
	k8s_ext "k8s.io/client-go/pkg/apis/extensions"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	api_util "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	AssetContainer = "Container"
	AssetBinary    = "File"
)

type InventoryAsset struct {
	Type string
	Data string
	SHA  string
}

type Inventory struct {
}

// Build creates a slice of inventory assets that are the inventory of containers and binaries.
func (i *Inventory) Build(cluster *api.Cluster, ig []*api.InstanceGroup, clientset simple.Clientset) ([]*InventoryAsset, error) {

	applyClusterCmd, err := i.buildApplyCluster(cluster, ig, clientset)

	if err != nil {
		return nil, fmt.Errorf("error applying cluster build: %v", err)
	}

	nodeupConfigs, err := i.buildNodeUpConfigs(applyClusterCmd)

	if err != nil {
		return nil, fmt.Errorf("error building nodeup configs: %v", err)
	}

	a, err := i.buildInventoryAssets(applyClusterCmd.Cluster, nodeupConfigs)

	if err != nil {
		return nil, fmt.Errorf("error building inventory assests: %v", err)
	}

	return a, nil
}

// buildInventoryAssets builds a map of all unique inventory assets.
func (i *Inventory) buildInventoryAssets(cluster *api.Cluster, nodeupConfigs []*nodeup.NodeUpConfig) ([]*InventoryAsset, error) {

	inventoryMap := make(map[string]*InventoryAsset)

	spec := cluster.Spec

	// List of all containers in the API
	inventoryMap[spec.KubeAPIServer.Image] = &InventoryAsset{
		Data: spec.KubeAPIServer.Image,
		Type: AssetContainer,
	}
	inventoryMap[spec.KubeControllerManager.Image] = &InventoryAsset{
		Data: spec.KubeControllerManager.Image,
		Type: AssetContainer,
	}

	inventoryMap[spec.KubeControllerManager.Image] = &InventoryAsset{
		Data: spec.KubeControllerManager.Image,
		Type: AssetContainer,
	}
	inventoryMap[spec.KubeProxy.Image] = &InventoryAsset{
		Data: spec.KubeProxy.Image,
		Type: AssetContainer,
	}
	inventoryMap[spec.KubeScheduler.Image] = &InventoryAsset{
		Data: spec.KubeScheduler.Image,
		Type: AssetContainer,
	}

	// TODO Need to get this our of here hardcoded.
	// TODO This value is hardcoded in protokube.
	inventoryMap["gcr.io/google_containers/etcd:2.2.1"] = &InventoryAsset{
		Data: "gcr.io/google_containers/etcd:2.2.1",
		Type: AssetContainer,
	}

	inventoryMap["gcr.io/google_containers/pause-amd64:3.0"] = &InventoryAsset{
		Data: "gcr.io/google_containers/pause-amd64:3.0",
		Type: AssetContainer,
	}

	channel, err := api.ParseChannelLocation(cluster.Spec.Channel)

	if err != nil {
		return nil, fmt.Errorf("error getting channel location: %v", err)
	}

	inventoryMap[channel] = &InventoryAsset{
		Data: channel,
		Type: AssetBinary,
	}

	// nodeup items
	for _, n := range nodeupConfigs {

		// protokube
		inventoryMap[n.ProtokubeImage.Source] = &InventoryAsset{
			Data: n.ProtokubeImage.Source,
			Type: AssetContainer,
		}

		glog.V(8).Infof("%s\n", n.ProtokubeImage.Source)

		// binary assest
		for _, a := range n.Assets {
			asset := strings.Split(a, "@")
			glog.V(8).Infof("%q\n", asset[1])

			inventoryMap[asset[0]] = &InventoryAsset{
				Data: asset[1],
				Type: AssetBinary,
				SHA:  asset[0],
			}
		}
	}

	// bootstrap items
	bootstrapConfigs, err := i.getBootstrapChannel(cluster)

	if err != nil {
		return nil, fmt.Errorf("error building bootstrap images: %v", err)
	}

	for _, b := range bootstrapConfigs {
		inventoryMap[b] = &InventoryAsset{
			Data: b,
			Type: AssetContainer,
		}
	}

	// reduce map to a slice
	var a []*InventoryAsset
	for _, value := range inventoryMap {
		a = append(a, value)
	}

	return a, nil
}

// buildApplyCluster runs the build method in apply cluster.
func (i *Inventory) buildApplyCluster(cluster *api.Cluster, ig []*api.InstanceGroup, clientset simple.Clientset) (*ApplyClusterCmd, error) {
	applyClusterCmd := &ApplyClusterCmd{
		Clientset:      clientset,
		DryRun:         true,
		Cluster:        cluster,
		InstanceGroups: ig,
	}

	err := applyClusterCmd.Build()

	if err != nil {
		return nil, fmt.Errorf("error applying cluster build: %v", err)
	}

	return applyClusterCmd, nil
}

func (i *Inventory) getBootstrapChannel(cluster *api.Cluster) ([]string, error) {

	loader := &Loader{}
	loader.Init()
	loader.Cluster = cluster

	loader.AddTypes(map[string]interface{}{
		"keypair":     &fitasks.Keypair{},
		"secret":      &fitasks.Secret{},
		"managedFile": &fitasks.ManagedFile{},
	})

	loader.Builders = append(loader.Builders, &BootstrapChannelBuilder{cluster: cluster})

	tf := &TemplateFunctions{
		cluster: cluster,
		modelContext: &model.KopsModelContext{
			Cluster: cluster,
		},
	}

	tf.AddTo(loader.TemplateFunctions)

	modelStore, err := findModelStore()
	if err != nil {
		return nil, fmt.Errorf("error building model store: %v", err)
	}

	var fileModels []string
	fileModels = append(fileModels, "cloudup")

	taskMap, err := loader.BuildTasks(modelStore, fileModels)
	if err != nil {
		return nil, fmt.Errorf("error building tasks: %v", err)
	}

	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("error building config base: %v", err)
	}

	var target fi.Target

	target = fi.NewDryRunTarget(os.Stdout)

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return nil, err // err is already formatted
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return nil, err // err is already formatted
	}

	fiContext, err := fi.NewContext(target, nil, keyStore, secretStore, configBase, false, taskMap)
	if err != nil {
		return nil, fmt.Errorf("error building context: %v", err)
	}

	err = fiContext.RunTasks(42)
	if err != nil {
		return nil, fmt.Errorf("error running tasks: %v", err)
	}

	sv, err := api_util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q", cluster.Spec.KubernetesVersion)
	}

	optionsContext := &components.OptionsContext{
		ClusterName:       cluster.ObjectMeta.Name,
		KubernetesVersion: *sv,
	}

	delimiter := []byte("\n---\n")
	var containers []string
	for _, task := range fiContext.AllTasks() {
		switch m := task.(type) {
		case *fitasks.ManagedFile:
			glog.V(8).Infof("Location: %s", *m.Location)

			pre16 := strings.Contains(*m.Location, "pre-k8s-1.6")

			if pre16 && optionsContext.IsKubernetesGTE("1.6") {
				glog.V(2).Infof("skipping: %s", *m.Location)
				continue
			} else if !pre16 && optionsContext.IsKubernetesLT("1.6") {
				glog.V(2).Infof("skipping: %s", *m.Location)
				continue
			}

			s, err := m.Contents.AsBytes()
			if err != nil {
				glog.V(8).Infof("unable to parse file %q: %v", *m.Location, err)
				continue
			}

			sections := bytes.Split(s, delimiter)

			for _, section := range sections {
				o, err := runtime.Decode(k8s_api.Codecs.UniversalDecoder(), section)
				if err != nil {
					glog.V(10).Infof("unable to parse file %q: %v", *m.Location, err)
					continue
				}

				switch v := o.(type) {
				case *k8s_apps.StatefulSet:
					c := v.Spec.Template.Spec.Containers
					containers = getContainers(c, containers)
				case *k8s_ext.DaemonSet:
					c := v.Spec.Template.Spec.Containers
					containers = getContainers(c, containers)
				case *k8s_ext.Deployment:
					c := v.Spec.Template.Spec.Containers
					containers = getContainers(c, containers)
				default:
				}
			}
		default:
			glog.V(2).Infof("Type of object was %T", m)
		}

	}

	return containers, nil
}

func getContainers(c []k8s_api.Container, containers []string) []string {
	for _, cont := range c {
		containers = append(containers, cont.Image)
		glog.V(8).Infof("containers found %s", cont.Image)
	}
	return containers
}

// buildNodeUpConfigs gets the nodeup configurations from apply cluster.
func (i *Inventory) buildNodeUpConfigs(applyClusterCmd *ApplyClusterCmd) ([]*nodeup.NodeUpConfig, error) {
	path, err := applyClusterCmd.GetConfigBase()
	if err != nil {
		return nil, fmt.Errorf("error getting config base path: %v", err)
	}

	channels := applyClusterCmd.GetChannels(path)

	clusterTags, err := buildCloudupTags(applyClusterCmd.Cluster)
	if err != nil {
		return nil, fmt.Errorf("Unable to build cloudup tags %v", err)
	}

	tf := &TemplateFunctions{
		cluster:        applyClusterCmd.Cluster,
		instanceGroups: applyClusterCmd.InstanceGroups,
		tags:           clusterTags,
	}

	var configs []*nodeup.NodeUpConfig

	for _, ig := range applyClusterCmd.InstanceGroups {
		n, err := applyClusterCmd.RenderNodeConfig(ig, tf, channels, path)
		if err != nil {
			return nil, fmt.Errorf("Error rendering node config: %v", err)
		}

		configs = append(configs, n)
	}

	return configs, nil
}
