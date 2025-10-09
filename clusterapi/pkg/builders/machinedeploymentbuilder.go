/*
Copyright 2025 The Kubernetes Authors.

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

package builders

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
)

type MachineDeploymentBuilder struct {
	ClusterName string

	// Name is the name of the MachineDeployment (and other objects) to create
	Name string
	// Namespace is the namespace for the MachineDeployment (and other objects) to create
	Namespace string

	// Replicas is the number of replicas for the MachineDeployment
	Replicas int

	// Zones is the list of the zones for the MachineDeployment (also called failureDomain)
	// Because each MachineDeployment can only have one zone, we create one per zone
	// (and divide the replicas between them)
	Zones []string

	// MachineType is the instance type for the MachineDeployment
	MachineType string

	// Subnet is the subnet for the MachineDeployment
	Subnet string

	// Image is the image for the MachineDeployment
	Image string

	// Role is the kops InstanceGroupRole for the MachineDeployment
	Role kops.InstanceGroupRole

	// KubernetesVersion is the kubernetes version for the MachineDeployment
	KubernetesVersion string

	// AdditionalMetadata is additional metadata to add to the instances
	AdditionalMetadata map[string]string

	// capiMachineDeployments holds the MachineDeployment objects,
	// we create one per zone
	capiMachineDeployments map[string]*unstructured.Unstructured

	capiInfra          *unstructured.Unstructured
	capiConfigTemplate *unstructured.Unstructured
}

func (b *MachineDeploymentBuilder) buildMachineDeployments() error {
	b.capiMachineDeployments = make(map[string]*unstructured.Unstructured)

	// TODO: Sync up with main replicas logic
	zoneCount := len(b.Zones)
	replicasByZone := make([]int, zoneCount)
	for i := range replicasByZone {
		replicasByZone[i] = b.Replicas / zoneCount
	}
	totalReplicas := replicasByZone[0] * zoneCount
	nextZone := 0
	for totalReplicas < b.Replicas {
		replicasByZone[nextZone]++
		totalReplicas++
		nextZone = (nextZone + 1) % zoneCount
	}

	for i, zone := range b.Zones {
		name := fmt.Sprintf("%s-%s", b.Name, zone)

		// This is because of network tags in cloud-provider-gcp
		// TODO: cloud-provider-gcp should not assume cluster name is a valid prefix
		escapedClusterName := gce.SafeClusterName(b.ClusterName)
		kubernetesVersion := b.KubernetesVersion

		configRef := makeRef(b.capiConfigTemplate)
		infrastructureRef := makeRef(b.capiInfra)

		spec := map[string]any{
			"clusterName":   escapedClusterName,
			"version":       kubernetesVersion,
			"failureDomain": zone,
			"bootstrap": map[string]any{
				"configRef": configRef,
			},
			"infrastructureRef": infrastructureRef,
		}

		obj := map[string]any{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "MachineDeployment",
			"metadata": map[string]any{
				"name":      name,
				"namespace": b.Namespace,
			},
			"spec": map[string]any{
				"clusterName": escapedClusterName,
				"replicas":    replicasByZone[i],
				"template": map[string]any{
					"spec": spec,
				},
			},
		}

		u := &unstructured.Unstructured{Object: obj}

		b.capiMachineDeployments[zone] = u
	}
	return nil

}

func (b *MachineDeploymentBuilder) createKopsConfigTemplate(ctx context.Context) error {
	templateSpec := map[string]any{}

	obj := map[string]any{
		"apiVersion": "bootstrap.cluster.x-k8s.io/v1beta1",
		"kind":       "KopsConfigTemplate",
		"metadata": map[string]any{
			"name":      b.Name,
			"namespace": b.Namespace,
		},
		"spec": map[string]any{
			"template": map[string]any{
				"spec": templateSpec,
			},
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	b.capiConfigTemplate = u
	return nil
}

func (b *MachineDeploymentBuilder) createGCPMachineTemplate(ctx context.Context) error {
	metadata := map[string]string{
		gcemetadata.MetadataKeyClusterName: b.ClusterName,
	}
	for k, v := range b.AdditionalMetadata {
		metadata[k] = v
	}

	additionalNetworkTags := []string{
		gce.TagForRole(b.ClusterName, kops.InstanceGroupRole(b.Role)),
	}

	gceSubnet := gce.ClusterSuffixedName(b.Subnet, b.ClusterName, 63)

	imageSpec := b.Image
	tokens := strings.Split(imageSpec, "/")
	// TODO: Sync with GCE logic
	if len(tokens) == 2 {
		imageSpec = fmt.Sprintf("projects/%s/global/images/%s", tokens[0], tokens[1])
	}

	templateSpec := map[string]any{}
	templateSpec["instanceType"] = b.MachineType
	templateSpec["image"] = imageSpec
	templateSpec["subnet"] = gceSubnet
	templateSpec["additionalNetworkTags"] = additionalNetworkTags
	templateSpec["publicIP"] = true
	additionalMetadata := []map[string]string{}
	for k, v := range metadata {
		additionalMetadata = append(additionalMetadata, map[string]string{
			"key":   k,
			"value": v,
		})
	}

	templateSpec["additionalMetadata"] = additionalMetadata

	obj := map[string]any{
		"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
		"kind":       "GCPMachineTemplate",
		"metadata": map[string]any{
			"name":      b.Name,
			"namespace": b.Namespace,
		},
		"spec": map[string]any{
			"template": map[string]any{
				"spec": templateSpec,
			},
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	b.capiInfra = u
	return nil
}

func (b *MachineDeploymentBuilder) BuildObjects(ctx context.Context) ([]*unstructured.Unstructured, error) {
	if err := b.createKopsConfigTemplate(ctx); err != nil {
		return nil, fmt.Errorf("error creating kops config template: %w", err)
	}

	if err := b.createGCPMachineTemplate(ctx); err != nil {
		return nil, fmt.Errorf("error creating gcp machine template: %w", err)
	}

	if err := b.buildMachineDeployments(); err != nil {
		return nil, fmt.Errorf("error building machine deployments: %w", err)
	}

	var objects []*unstructured.Unstructured
	objects = append(objects, b.capiConfigTemplate)
	objects = append(objects, b.capiInfra)

	var machineDeployments []*unstructured.Unstructured
	for _, md := range b.capiMachineDeployments {
		machineDeployments = append(machineDeployments, md)
	}
	sort.Slice(machineDeployments, func(i, j int) bool {
		return machineDeployments[i].GetName() < machineDeployments[j].GetName()
	})

	objects = append(objects, machineDeployments...)

	return objects, nil
}
