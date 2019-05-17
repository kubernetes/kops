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

package clusterapi

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kops "k8s.io/kops/pkg/apis/kops"
	kopsv1alpha2 "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/gcemodel"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type Builder struct {
	Clientset simple.Clientset
}

func (b *Builder) BuildMachineDeployment(clusterObj *kopsv1alpha2.Cluster, igObj *kopsv1alpha2.InstanceGroup) (*unstructured.Unstructured, error) {
	cluster := &kops.Cluster{}
	{
		if err := kopscodecs.Scheme.Convert(clusterObj, cluster, nil); err != nil {
			return nil, fmt.Errorf("error converting cluster to internal form: %v", err)
		}
	}

	ig := &kops.InstanceGroup{}
	{
		if err := kopscodecs.Scheme.Convert(igObj, ig, nil); err != nil {
			return nil, fmt.Errorf("error converting InstanceGroup to internal form: %v", err)
		}
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	gceCloud := cloud.(gce.GCECloud)
	region := gceCloud.Region()

	phase := cloudup.PhaseCluster
	assetBuilder := assets.NewAssetBuilder(cluster, string(phase))

	applyCmd := &cloudup.ApplyClusterCmd{
		Cluster:        cluster,
		Clientset:      b.Clientset,
		InstanceGroups: []*kops.InstanceGroup{ig},
		Phase:          phase,
	}

	if err := applyCmd.AddFileAssets(assetBuilder); err != nil {
		return nil, fmt.Errorf("error adding assets: %v", err)
	}

	nodeupConfig, err := applyCmd.BuildNodeUpConfig(assetBuilder, ig)
	if err != nil {
		return nil, fmt.Errorf("error building nodeup config: %v", err)
	}

	/*
		if !ig.IsMaster() {
			nodeupConfig.ProtokubeImage = nil
			nodeupConfig.Channels = nil
		}

			nodeupConfig.ConfigBase = fi.String("/etc/kubernetes/bootstrap")
	*/

	bootstrapScript := model.BootstrapScript{}

	nodeupLocation, nodeupHash, err := cloudup.NodeUpLocation(assetBuilder)
	if err != nil {
		return nil, err
	}
	bootstrapScript.NodeUpSource = nodeupLocation.String()
	bootstrapScript.NodeUpSourceHash = nodeupHash.Hex()
	bootstrapScript.NodeUpConfigBuilder = func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
		return nodeupConfig, err
	}

	script, err := bootstrapScript.ResourceNodeUp(ig, cluster)
	if err != nil {
		return nil, fmt.Errorf("error building bootstrap script: %v", err)
	}

	scriptString, err := script.AsString()
	if err != nil {
		return nil, fmt.Errorf("error building bootstrap script: %v", err)
	}

	/*
		file := &DataFile{}
		file.Header.Name = "bootstrap.sh"
		file.Header.Size = int64(len(scriptBytes))
		file.Header.Mode = 0755
		file.Data = scriptBytes
		files = append(files, file)
	*/

	//klog.Infof("script %s", string(scriptBytes))

	gce := &gcemodel.GCEModelContext{
		KopsModelContext: &model.KopsModelContext{},
	}
	gce.Cluster = cluster
	gce.Region = region

	volumeSize, err := gce.VolumeSize(ig)
	if err != nil {
		return nil, err
	}
	volumeType := gce.VolumeType(ig)

	// TODO: We should move this into the gcemodel package, and reuse the existing logic
	disks := []map[string]interface{}{
		{
			"initializeParams": map[string]interface{}{
				"diskSizeGb": volumeSize,
				"diskType":   volumeType,
			},
		},
	}

	// TODO	CanIPForward
	//t.CanIPForward = fi.Bool(false)

	subnetwork := gce.NameForIPAliasSubnet()
	subnetwork = "regions/" + region + "/subnetworks/" + subnetwork

	networkInterfaces := []map[string]interface{}{}

	if gce.UsesIPAliases() {
		ni := map[string]interface{}{
			"subnetwork": subnetwork,
		}
		var aliasIPRanges []map[string]interface{}
		for k, v := range gce.NodeAliasIPRanges() {
			r := make(map[string]interface{})
			r["ipCidrRange"] = v
			r["subnetworkRangeName"] = k
			aliasIPRanges = append(aliasIPRanges, r)
		}
		ni["aliasIpRanges"] = aliasIPRanges
		networkInterfaces = append(networkInterfaces, ni)
	}

	instanceMetadata := []map[string]interface{}{
		{
			"key":   "cluster-name",
			"value": cluster.Name,
		},
		{
			"key":   "startup-script",
			"value": scriptString,
		},
	}

	zones, err := gce.FindZonesForInstanceGroup(ig)
	if err != nil {
		return nil, err
	}
	zone := ""
	if len(zones) == 1 {
		zone = zones[0]
	} else if len(zones) < 1 {
		return nil, fmt.Errorf("must specify zone for GCE")
	} else {
		return nil, fmt.Errorf("cannot specify multiple zones for GCE")
	}

	machineType := ig.Spec.MachineType

	instanceTags := []string{}
	roles := []string{}
	switch ig.Spec.Role {
	case kops.InstanceGroupRoleNode:
		instanceTags = append(instanceTags, gce.GCETagForRole(kops.InstanceGroupRoleNode))
		roles = append(roles, "Node")

	default:
		return nil, fmt.Errorf("unsupported role %q", ig.Spec.Role)
	}

	email := "default"
	serviceAccounts := []map[string]interface{}{
		{"email": email},
	}

	image := ig.Spec.Image
	// Expand known short-forms
	{
		tokens := strings.Split(image, "/")
		if len(tokens) == 2 {
			image = "projects/" + tokens[0] + "/global/images/" + tokens[1]
		}
	}

	providerSpec := map[string]interface{}{
		"apiVersion":        "gceproviderconfig/v1alpha1",
		"kind":              "GCEProviderConfig",
		"roles":             roles,
		"zone":              zone,
		"machineType":       machineType,
		"networkInterfaces": networkInterfaces,
		"disks":             disks,
		"image":             image,
		"instanceTags":      instanceTags,
		"serviceAccounts":   serviceAccounts,
		"instanceMetadata":  instanceMetadata,
	}

	return buildMachineDeployment(cluster, ig, providerSpec)
}

func buildMachineDeployment(cluster *kops.Cluster, ig *kops.InstanceGroup, providerSpec map[string]interface{}) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}

	u.SetAPIVersion("cluster.k8s.io/v1alpha1")
	u.SetKind("MachineDeployment")
	u.SetName(ig.Name)
	u.SetNamespace(ig.Namespace)

	// TODO: We need MaxSize & MinSize?
	replicas := ig.Spec.MaxSize

	versions := map[string]interface{}{
		// Annoyingly, kubelet version is required by the schema.
		"kubelet": cluster.Spec.KubernetesVersion,
	}

	labels := map[string]string{
		"kops.k8s.io/instancegroup": ig.Name,
	}

	template := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
		"spec": map[string]interface{}{
			"versions": versions,
			"providerSpec": map[string]interface{}{
				"value": providerSpec,
			},
		},
	}

	spec := map[string]interface{}{
		"replicas": replicas,
		"selector": map[string]interface{}{
			"matchLabels": labels,
		},
		"template": template,
	}

	u.Object["spec"] = spec

	return u, nil
}
