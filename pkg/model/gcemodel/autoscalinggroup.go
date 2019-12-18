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

package gcemodel

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/model/iam"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

const (
	DefaultVolumeType = "pd-standard"
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*GCEModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

// Build the GCE instance template object for an InstanceGroup
// We are then able to extract out the fields we need from here
func (b *AutoscalingGroupModelBuilder) buildInstanceTemplate(ig *kops.InstanceGroup) (*gcetasks.InstanceTemplate, error) {
	var err error

	name := b.SafeObjectName(ig.ObjectMeta.Name)

	startupScript, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return nil, err
	}

	volumeSize := fi.Int32Value(ig.Spec.RootVolumeSize)
	if volumeSize == 0 {
		volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
		if err != nil {
			return nil, err
		}
	}
	volumeType := fi.StringValue(ig.Spec.RootVolumeType)
	if volumeType == "" {
		volumeType = DefaultVolumeType
	}

	namePrefix := gce.LimitedLengthName(name, gcetasks.InstanceTemplateNamePrefixMaxLength)

	t := &gcetasks.InstanceTemplate{
		Name:           s(name),
		NamePrefix:     s(namePrefix),
		Lifecycle:      b.Lifecycle,
		Network:        b.LinkToNetwork(),
		MachineType:    s(ig.Spec.MachineType),
		BootDiskType:   s(volumeType),
		BootDiskSizeGB: i64(int64(volumeSize)),
		BootDiskImage:  s(ig.Spec.Image),

		// TODO: Support preemptible nodes?
		Preemptible: fi.Bool(false),

		Scopes: []string{
			"compute-rw",
			"monitoring",
			"logging-write",
		},

		Metadata: map[string]*fi.ResourceHolder{
			"startup-script": startupScript,
			//"config": resources/config.yaml $nodeset.Name
			"cluster-name": fi.WrapResource(fi.NewStringResource(b.ClusterName())),
			nodeidentitygce.MetadataKeyInstanceGroupName: fi.WrapResource(fi.NewStringResource(ig.Name)),
		},
	}

	storagePaths, err := iam.WriteableVFSPaths(b.Cluster, ig.Spec.Role)
	if err != nil {
		return nil, err
	}
	if len(storagePaths) == 0 {
		t.Scopes = append(t.Scopes, "storage-ro")
	} else {
		klog.Warningf("enabling storage-rw for etcd backups")
		t.Scopes = append(t.Scopes, "storage-rw")
	}

	if len(b.SSHPublicKeys) > 0 {
		var gFmtKeys []string
		for _, key := range b.SSHPublicKeys {
			gFmtKeys = append(gFmtKeys, fmt.Sprintf("%s: %s", fi.SecretNameSSHPrimary, key))
		}

		t.Metadata["ssh-keys"] = fi.WrapResource(fi.NewStringResource(strings.Join(gFmtKeys, "\n")))
	}

	switch ig.Spec.Role {
	case kops.InstanceGroupRoleMaster:
		// Grant DNS permissions
		t.Scopes = append(t.Scopes, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
		t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleMaster))

	case kops.InstanceGroupRoleNode:
		t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleNode))
	}

	if gce.UsesIPAliases(b.Cluster) {
		t.CanIPForward = fi.Bool(false)

		t.AliasIPRanges = map[string]string{
			b.NameForIPAliasRange("pods"): "/24",
		}
		t.Subnet = b.LinkToIPAliasSubnet()
	} else {
		t.CanIPForward = fi.Bool(true)
	}

	//labels, err := b.CloudTagsForInstanceGroup(ig)
	//if err != nil {
	//	return fmt.Errorf("error building cloud tags: %v", err)
	//}
	//t.Labels = labels

	return t, nil
}

func (b *AutoscalingGroupModelBuilder) splitToZones(ig *kops.InstanceGroup) (map[string]int, error) {
	zones, err := b.FindZonesForInstanceGroup(ig)
	if err != nil {
		return nil, err
	}

	// TODO: We should expose autoscaling with MinSize & MaxSize

	// TODO: Duplicated from aws - move to defaults?
	minSize := 1
	if ig.Spec.MinSize != nil {
		minSize = int(fi.Int32Value(ig.Spec.MinSize))
	} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
		minSize = 2
	}

	// We have to assign instances to the various zones
	// TODO: Switch to regional managed instance group
	// But we can't yet use RegionInstanceGroups:
	// 1) no support in terraform
	// 2) we can't steer to specific zones AFAICT, only to all zones in the region

	targetSizes := make([]int, len(zones))
	totalSize := 0
	for i := range zones {
		targetSizes[i] = minSize / len(zones)
		totalSize += targetSizes[i]
	}
	i := 0
	for {
		if totalSize >= minSize {
			break
		}
		targetSizes[i]++
		totalSize++

		i++
		if i > len(targetSizes) {
			i = 0
		}
	}

	instanceCountByZone := make(map[string]int)
	for i, zone := range zones {
		instanceCountByZone[zone] = targetSizes[i]
	}
	return instanceCountByZone, nil
}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		// InstanceTemplate
		instanceTemplate, err := b.buildInstanceTemplate(ig)
		if err != nil {
			return err
		}
		c.AddTask(instanceTemplate)

		instanceCountByZone, err := b.splitToZones(ig)
		if err != nil {
			return err
		}

		for zone, targetSize := range instanceCountByZone {
			name := gce.NameForInstanceGroupManager(b.Cluster, ig, zone)

			t := &gcetasks.InstanceGroupManager{
				Name:             s(name),
				Lifecycle:        b.Lifecycle,
				Zone:             s(zone),
				TargetSize:       fi.Int64(int64(targetSize)),
				BaseInstanceName: s(ig.ObjectMeta.Name),
				InstanceTemplate: instanceTemplate,
			}

			// Attach masters to load balancer if we're using one
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				if b.UseLoadBalancerForAPI() {
					t.TargetPools = append(t.TargetPools, b.LinkToTargetPool("api"))
				}
			}

			c.AddTask(t)
		}

		//{{ if HasTag "_master_lb" }}
		//# Attach ASG to ELB
		//loadBalancerAttachment/masters.{{ $m.Name }}.{{ SafeClusterName }}:
		//loadBalancer: loadBalancer/api.{{ ClusterName }}
		//autoscalingGroup: autoscalingGroup/{{ $m.Name }}.{{ ClusterName }}
		//{{ end }}

	}

	return nil
}

// MapToClusterAPI implements the cluster-api support
func (b *AutoscalingGroupModelBuilder) MapToClusterAPI(cluster *kops.Cluster, ig *kops.InstanceGroup) ([]*unstructured.Unstructured, error) {
	it, err := b.buildInstanceTemplate(ig)
	if err != nil {
		return nil, err
	}

	instanceCountByZone, err := b.splitToZones(ig)
	if err != nil {
		return nil, err
	}

	var objects []*unstructured.Unstructured

	for zone, replicas := range instanceCountByZone {

		image := fi.StringValue(it.BootDiskImage)
		// Expand known short-forms
		// TODO: Move this logic into GCP provider
		{
			tokens := strings.Split(image, "/")
			if len(tokens) == 2 {
				image = "projects/" + tokens[0] + "/global/images/" + tokens[1]
			}
		}

		email := "default"
		serviceAccount := map[string]interface{}{
			"email": email,
			// TODO: scopes?
		}

		// TODO: Should the provider encode this?  Not clear whether the data is supposed to be pre-encoded.
		var bootstrapData []byte

		additionalMetadata := make(map[string]string)
		for k, vr := range it.Metadata {
			if k == "startup-script" {
				b, err := vr.AsBytes()
				if err != nil {
					return nil, err
				}
				bootstrapData = b
				continue
			}

			v, err := vr.AsString()
			if err != nil {
				return nil, err
			}

			additionalMetadata[k] = v
		}

		// TODO: GCP provider requires this, but actually only requires it when we aren't using a custom image
		version := cluster.Spec.KubernetesVersion

		var machineTemplate *unstructured.Unstructured
		{
			u := &unstructured.Unstructured{}

			u.SetAPIVersion("infrastructure.cluster.x-k8s.io/v1alpha3")
			u.SetKind("GCPMachineTemplate")
			u.SetName(ig.Name)
			u.SetNamespace(ig.Namespace)

			// TODO: We should support instances on GCP without a public IP
			// (I think we need to set up a NAT gateway)
			publicIP := true

			template := map[string]interface{}{
				"spec": map[string]interface{}{
					"instanceType":          it.MachineType,
					"zone":                  zone,
					"image":                 image,
					"rootDeviceSize":        it.BootDiskSizeGB,
					"serviceAccount":        serviceAccount,
					"publicIP":              &publicIP,
					"additionalNetworkTags": it.Tags,
					"additionalMetadata":    additionalMetadata,
				},
			}

			spec := map[string]interface{}{
				"template": template,
			}

			u.Object["spec"] = spec
			machineTemplate = u
		}

		// We use a hash to ensure that we pick up new versions and avoid mutating in-place
		// TODO: Use hash from kustomize?
		hash := fmt.Sprintf("%x", sha256.Sum256(bootstrapData))[0:6]

		dataSecretName := ig.Name + "-bootstrap-" + hash

		var bootstrapSecret *unstructured.Unstructured
		{
			u := &unstructured.Unstructured{}

			u.SetAPIVersion("v1")
			u.SetKind("Secret")
			u.SetName(dataSecretName)
			u.SetNamespace(ig.Namespace)

			data := map[string]interface{}{
				"value": base64.StdEncoding.EncodeToString([]byte(bootstrapData)),
			}

			u.Object["data"] = data
			u.Object["type"] = "Opaque"

			bootstrapSecret = u
		}

		var machineDeployment *unstructured.Unstructured
		{
			u := &unstructured.Unstructured{}

			u.SetAPIVersion("cluster.x-k8s.io/v1alpha3")
			u.SetKind("MachineDeployment")
			u.SetName(ig.Name)
			u.SetNamespace(ig.Namespace)

			labels := map[string]string{
				"kops.k8s.io/instancegroup": ig.Name,
			}

			/*
				// For reasons not entirely clear, we resolve the cluster from the machine using this label
				// TODO: Use owner refs if available?
				labels["cluster.x-k8s.io/cluster-name"] = cluster.Name
			*/

			template := map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": labels,
				},
				"spec": map[string]interface{}{
					"clusterName": cluster.Name,
					"bootstrap": map[string]interface{}{
						"dataSecretName": dataSecretName,
					},
					"infrastructureRef": map[string]interface{}{
						"apiVersion": "infrastructure.cluster.x-k8s.io/v1alpha3",
						"kind":       "GCPMachineTemplate",
						"name":       ig.Name,
						"namespace":  ig.Namespace,
					},
					"version": version,
				},
			}

			// TODO: Should expose these options
			updateStrategy := map[string]interface{}{
				"type": "RollingUpdate",
				"rollingUpdate": map[string]interface{}{
					"maxSurge": "20%",
				},
			}
			minReadySeconds := 5

			spec := map[string]interface{}{
				"clusterName": cluster.Name,
				"replicas":    replicas,
				"selector": map[string]interface{}{
					"matchLabels": labels,
				},
				"template":        template,
				"strategy":        updateStrategy,
				"minReadySeconds": minReadySeconds,
			}

			u.Object["spec"] = spec

			machineDeployment = u
		}

		objects = append(objects, machineTemplate)
		objects = append(objects, bootstrapSecret)
		objects = append(objects, machineDeployment)

	}

	return objects, nil
}
