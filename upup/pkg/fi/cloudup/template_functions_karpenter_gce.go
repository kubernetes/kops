/*
Copyright 2026 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/model/gcemodel"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

const karpenterGCPAPIGroup = "karpenter.k8s.gcp"

// defaultKarpenterGCEMaxPods matches the kubelet default; the GCENodeClass requires an explicit
// maxPods so Karpenter can compute allocatable capacity.
const defaultKarpenterGCEMaxPods = int32(110)

type karpenterGCENodeClass struct {
	APIVersion string                    `json:"apiVersion"`
	Kind       string                    `json:"kind"`
	Metadata   karpenterObjectMeta       `json:"metadata"`
	Spec       karpenterGCENodeClassSpec `json:"spec"`
}

type karpenterGCENodeClassSpec struct {
	ImageSelectorTerms     []karpenterGCEImageTerm             `json:"imageSelectorTerms"`
	Disks                  []karpenterGCEDisk                  `json:"disks"`
	KubeletConfiguration   *karpenterGCEKubeletConfiguration   `json:"kubeletConfiguration,omitempty"`
	Labels                 map[string]string                   `json:"labels,omitempty"`
	Metadata               map[string]string                   `json:"metadata,omitempty"`
	NetworkConfig          *karpenterGCENetworkConfig          `json:"networkConfig,omitempty"`
	NetworkTags            []string                            `json:"networkTags,omitempty"`
	ServiceAccount         string                              `json:"serviceAccount,omitempty"`
	ShieldedInstanceConfig *karpenterGCEShieldedInstanceConfig `json:"shieldedInstanceConfig,omitempty"`
	StartupScript          string                              `json:"startupScript,omitempty"`
}

type karpenterGCEImageTerm struct {
	ID string `json:"id,omitempty"`
}

type karpenterGCEDisk struct {
	SizeGiB               int32  `json:"sizeGiB"`
	Category              string `json:"category,omitempty"`
	Boot                  bool   `json:"boot"`
	ProvisionedIOPS       *int64 `json:"provisionedIOPS,omitempty"`
	ProvisionedThroughput *int64 `json:"provisionedThroughput,omitempty"`
}

// karpenterGCEKubeletConfiguration maps to GCENodeClass.spec.kubeletConfiguration. As on AWS,
// surfacing kubelet settings lets Karpenter compute node allocatable capacity correctly when
// binpacking; maxPods is required by the GCENodeClass validation rules for self-hosted clusters.
type karpenterGCEKubeletConfiguration struct {
	MaxPods        *int32            `json:"maxPods,omitempty"`
	SystemReserved map[string]string `json:"systemReserved,omitempty"`
	KubeReserved   map[string]string `json:"kubeReserved,omitempty"`
}

type karpenterGCENetworkConfig struct {
	EnablePrivateNodes *bool  `json:"enablePrivateNodes,omitempty"`
	Subnetwork         string `json:"subnetwork,omitempty"`
}

type karpenterGCEShieldedInstanceConfig struct {
	EnableVtpm *bool `json:"enableVtpm,omitempty"`
}

func (tf *TemplateFunctions) KarpenterGCENodeClass(ig *kops.InstanceGroup) (string, error) {
	nodeClass, err := tf.buildKarpenterGCENodeClass(ig)
	if err != nil {
		return "", err
	}
	return marshalKarpenterResource(ig, nodeClass)
}

func (tf *TemplateFunctions) buildKarpenterGCENodeClass(ig *kops.InstanceGroup) (*karpenterGCENodeClass, error) {
	projectID, err := tf.gceProjectID()
	if err != nil {
		return nil, err
	}

	imageID, err := karpenterGCEImageID(ig.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("building imageSelectorTerms for %q: %w", ig.Name, err)
	}

	bootDisk, err := karpenterGCEBootDisk(ig)
	if err != nil {
		return nil, fmt.Errorf("building disks for %q: %w", ig.Name, err)
	}

	subnets, err := tf.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) != 1 {
		return nil, fmt.Errorf("expected exactly one subnet for InstanceGroup %q, got %d", ig.Name, len(subnets))
	}
	subnet := subnets[0]

	subnetName := subnet.ID
	if subnetName == "" {
		subnetName = gce.ClusterSuffixedName(subnet.Name, tf.Cluster.ObjectMeta.Name, 63)
	}
	networkConfig := &karpenterGCENetworkConfig{
		Subnetwork: fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", projectID, tf.Region, subnetName),
		// Public subnets get external IPs; private subnets rely on the cluster's NAT.
		EnablePrivateNodes: new(subnet.Type == kops.SubnetTypePrivate),
	}

	// The same ownership labels the MIG instance template would carry; these are also how
	// GetCloudGroups finds Karpenter instances.
	labels, err := tf.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("building labels for %q: %w", ig.Name, err)
	}

	gceModel := &gcemodel.GCEModelContext{
		ProjectID:        projectID,
		KopsModelContext: &tf.KopsModelContext,
	}
	serviceAccount := fi.ValueOf(gceModel.LinkToServiceAccount(ig).Email)

	startupScript, err := tf.managedFileContents("nodeupscript-" + ig.Name)
	if err != nil {
		return nil, fmt.Errorf("reading startupScript for %q: %w", ig.Name, err)
	}

	return &karpenterGCENodeClass{
		APIVersion: karpenterGCPAPIGroup + "/v1alpha1",
		Kind:       "GCENodeClass",
		Metadata: karpenterObjectMeta{
			Name: ig.Name,
		},
		Spec: karpenterGCENodeClassSpec{
			ImageSelectorTerms:   []karpenterGCEImageTerm{{ID: imageID}},
			Disks:                []karpenterGCEDisk{*bootDisk},
			KubeletConfiguration: buildKarpenterGCEKubeletConfiguration(ig),
			Labels:               labels,
			Metadata: map[string]string{
				// The bootstrap TPM verifier and the kops-controller node identifier derive
				// instance group ownership from this key.
				nodeidentitygce.MetadataKeyInstanceGroupName: ig.Name,
			},
			NetworkConfig:  networkConfig,
			NetworkTags:    []string{gce.TagForRole(tf.Cluster.ObjectMeta.Name, ig.Spec.Role)},
			ServiceAccount: serviceAccount,
			// The bootstrap TPM verifier requires a vTPM attestation.
			ShieldedInstanceConfig: &karpenterGCEShieldedInstanceConfig{EnableVtpm: new(true)},
			StartupScript:          startupScript,
		},
	}, nil
}

func (tf *TemplateFunctions) gceProjectID() (string, error) {
	if tf.Cluster.Spec.CloudProvider.GCE != nil && tf.Cluster.Spec.CloudProvider.GCE.Project != "" {
		return tf.Cluster.Spec.CloudProvider.GCE.Project, nil
	}
	if gceCloud, ok := tf.cloud.(gce.GCECloud); ok {
		return gceCloud.Project(), nil
	}
	return "", fmt.Errorf("could not determine GCE project")
}

// karpenterGCEImageID resolves the kOps GCE image spec (<project>/<name>) to the image id form
// used by GCENodeClass imageSelectorTerms.
func karpenterGCEImageID(image string) (string, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return "", fmt.Errorf("image is required")
	}
	tokens := strings.Split(image, "/")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return "", fmt.Errorf("image %q must be <project>/<name>", image)
	}
	return fmt.Sprintf("projects/%s/global/images/%s", tokens[0], tokens[1]), nil
}

// karpenterGCEBootDisk builds the boot disk from the instance group's root volume settings,
// applying the same defaults as the MIG instance template. The GCENodeClass requires a boot disk;
// instance creation fails without one.
func karpenterGCEBootDisk(ig *kops.InstanceGroup) (*karpenterGCEDisk, error) {
	var volumeSize int32
	var volumeType string
	var volumeIOPS, volumeThroughput int32
	if ig.Spec.RootVolume != nil {
		volumeSize = fi.ValueOf(ig.Spec.RootVolume.Size)
		volumeType = fi.ValueOf(ig.Spec.RootVolume.Type)
		volumeIOPS = fi.ValueOf(ig.Spec.RootVolume.IOPS)
		volumeThroughput = fi.ValueOf(ig.Spec.RootVolume.Throughput)
	}
	if volumeSize == 0 {
		var err error
		volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
		if err != nil {
			return nil, err
		}
	}
	if volumeType == "" {
		volumeType = gcemodel.DefaultVolumeType
	}

	disk := &karpenterGCEDisk{
		SizeGiB:  volumeSize,
		Category: volumeType,
		Boot:     true,
	}
	if volumeIOPS > 0 {
		disk.ProvisionedIOPS = new(int64(volumeIOPS))
	}
	if volumeThroughput > 0 {
		disk.ProvisionedThroughput = new(int64(volumeThroughput))
	}
	return disk, nil
}

func buildKarpenterGCEKubeletConfiguration(ig *kops.InstanceGroup) *karpenterGCEKubeletConfiguration {
	kubelet := &karpenterGCEKubeletConfiguration{
		// maxPods is required so Karpenter binpacks with the same value the kubelet uses;
		// default to the kubelet default.
		MaxPods: new(defaultKarpenterGCEMaxPods),
	}
	if ig.Spec.Kubelet != nil {
		if ig.Spec.Kubelet.MaxPods != nil {
			kubelet.MaxPods = ig.Spec.Kubelet.MaxPods
		}
		kubelet.SystemReserved = ig.Spec.Kubelet.SystemReserved
		kubelet.KubeReserved = ig.Spec.Kubelet.KubeReserved
	}
	return kubelet
}
