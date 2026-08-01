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
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"sigs.k8s.io/yaml"
)

const (
	karpenterAWSAPIGroup       = "karpenter.k8s.aws"
	karpenterNodePoolAPIGroup  = "karpenter.sh"
	karpenterNodePoolLabel     = "karpenter.sh/nodepool"
	karpenterCapacityTypeLabel = "karpenter.sh/capacity-type"

	karpenterOSLabel             = "kubernetes.io/os"
	karpenterInstanceTypeLabel   = "node.kubernetes.io/instance-type"
	karpenterInstanceCPULabel    = "karpenter.k8s.aws/instance-cpu"
	karpenterInstanceMemoryLabel = "karpenter.k8s.aws/instance-memory"
	karpenterInstanceFamilyLabel = "karpenter.k8s.aws/instance-family"
)

type karpenterObjectMeta struct {
	Name string `json:"name"`
}

type karpenterEC2NodeClass struct {
	APIVersion string                    `json:"apiVersion"`
	Kind       string                    `json:"kind"`
	Metadata   karpenterObjectMeta       `json:"metadata"`
	Spec       karpenterEC2NodeClassSpec `json:"spec"`
}

type karpenterEC2NodeClassSpec struct {
	AMIFamily                string                         `json:"amiFamily"`
	AMISelectorTerms         []karpenterAMITerm             `json:"amiSelectorTerms"`
	AssociatePublicIPAddress *bool                          `json:"associatePublicIPAddress,omitempty"`
	BlockDeviceMappings      []karpenterBlockDeviceMapping  `json:"blockDeviceMappings,omitempty"`
	Tags                     map[string]string              `json:"tags,omitempty"`
	SubnetSelectorTerms      []karpenterSelectorTerm        `json:"subnetSelectorTerms"`
	SecurityGroupTerms       []karpenterSelectorTerm        `json:"securityGroupSelectorTerms"`
	InstanceProfile          string                         `json:"instanceProfile"`
	UserData                 string                         `json:"userData"`
	Kubelet                  *karpenterKubeletConfiguration `json:"kubelet,omitempty"`
}

// karpenterKubeletConfiguration maps to EC2NodeClass.spec.kubelet. Surfacing kubelet
// settings to Karpenter (rather than only configuring them via the nodeup bootstrap
// script) lets Karpenter compute node allocatable capacity correctly when binpacking.
type karpenterKubeletConfiguration struct {
	MaxPods        *int32            `json:"maxPods,omitempty"`
	SystemReserved map[string]string `json:"systemReserved,omitempty"`
	KubeReserved   map[string]string `json:"kubeReserved,omitempty"`
}

// karpenterBlockDeviceMapping maps to EC2NodeClass.spec.blockDeviceMappings[]. kOps
// emits exactly one mapping, for the root volume, built from the InstanceGroup's
// spec.rootVolume with the same defaults the ASG launch template applies.
type karpenterBlockDeviceMapping struct {
	DeviceName string                   `json:"deviceName,omitempty"`
	EBS        *karpenterBlockDeviceEBS `json:"ebs,omitempty"`
	// RootVolume tells Karpenter which mapping backs the kubelet root dir. With
	// amiFamily: Custom Karpenter cannot infer it, and without the flag it computes
	// node ephemeral-storage capacity from its own default rather than this volume.
	RootVolume *bool `json:"rootVolume,omitempty"`
}

type karpenterBlockDeviceEBS struct {
	DeleteOnTermination *bool   `json:"deleteOnTermination,omitempty"`
	Encrypted           *bool   `json:"encrypted,omitempty"`
	IOPS                *int64  `json:"iops,omitempty"`
	KMSKeyID            *string `json:"kmsKeyID,omitempty"`
	Throughput          *int64  `json:"throughput,omitempty"`
	VolumeSize          string  `json:"volumeSize,omitempty"`
	VolumeType          string  `json:"volumeType,omitempty"`
}

type karpenterAMITerm struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Owner        string `json:"owner,omitempty"`
	SSMParameter string `json:"ssmParameter,omitempty"`
}

type karpenterSelectorTerm struct {
	ID   string            `json:"id,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

type karpenterNodePool struct {
	APIVersion string                `json:"apiVersion"`
	Kind       string                `json:"kind"`
	Metadata   karpenterObjectMeta   `json:"metadata"`
	Spec       karpenterNodePoolSpec `json:"spec"`
}

type karpenterNodePoolSpec struct {
	Template karpenterNodeClaimTemplate `json:"template"`
	Replicas *int64                     `json:"replicas,omitempty"`
	Limits   *karpenterNodePoolLimits   `json:"limits,omitempty"`
}

type karpenterNodePoolLimits struct {
	Nodes string `json:"nodes,omitempty"`
}

type karpenterNodeClaimTemplate struct {
	Metadata *karpenterNodeClaimMetadata `json:"metadata,omitempty"`
	Spec     karpenterNodeClaimSpec      `json:"spec"`
}

type karpenterNodeClaimMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

type karpenterNodeClaimSpec struct {
	Requirements []karpenterRequirement `json:"requirements,omitempty"`
	Taints       []karpenterTaint       `json:"taints,omitempty"`
	NodeClassRef karpenterNodeClassRef  `json:"nodeClassRef"`
}

type karpenterRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

type karpenterTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"`
}

type karpenterNodeClassRef struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
}

func (tf *TemplateFunctions) KarpenterInstanceGroups() []*kops.InstanceGroup {
	if tf.tasks == nil {
		return nil
	}
	if tf.Cluster == nil || tf.Cluster.GetCloudProvider() != kops.CloudProviderAWS {
		return nil
	}
	if tf.Cluster.Spec.Karpenter == nil || !tf.Cluster.Spec.Karpenter.Enabled {
		return nil
	}

	var instanceGroups []*kops.InstanceGroup
	for _, ig := range tf.InstanceGroups {
		if ig != nil && ig.IsKarpenterManaged() {
			instanceGroups = append(instanceGroups, ig)
		}
	}
	return instanceGroups
}

func (tf *TemplateFunctions) KarpenterEC2NodeClass(ig *kops.InstanceGroup) (string, error) {
	nodeClass, err := tf.buildKarpenterEC2NodeClass(ig)
	if err != nil {
		return "", err
	}
	return marshalKarpenterResource(ig, nodeClass)
}

func (tf *TemplateFunctions) KarpenterNodePool(ig *kops.InstanceGroup) (string, error) {
	nodePool, err := tf.buildKarpenterNodePool(ig)
	if err != nil {
		return "", err
	}
	return marshalKarpenterResource(ig, nodePool)
}

func marshalKarpenterResource(ig *kops.InstanceGroup, object interface{}) (string, error) {
	data, err := yaml.Marshal(object)
	if err != nil {
		return "", fmt.Errorf("marshaling Karpenter resource for %q: %w", ig.Name, err)
	}
	return strings.TrimSpace(string(data)), nil
}

func (tf *TemplateFunctions) buildKarpenterEC2NodeClass(ig *kops.InstanceGroup) (*karpenterEC2NodeClass, error) {
	amiSelectorTerms, err := buildKarpenterAMITerms(ig.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("building amiSelectorTerms for %q: %w", ig.Name, err)
	}

	instanceProfile, err := tf.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return nil, fmt.Errorf("building instance profile for %q: %w", ig.Name, err)
	}

	tags, err := tf.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("building tags for %q: %w", ig.Name, err)
	}
	tags = karpenterEC2NodeClassTags(tags)
	associatePublicIP, err := tf.karpenterAssociatePublicIP(ig)
	if err != nil {
		return nil, err
	}
	userData, err := tf.managedFileContents("nodeupscript-" + ig.Name)
	if err != nil {
		return nil, fmt.Errorf("reading userData for %q: %w", ig.Name, err)
	}
	rootDeviceName, err := tf.karpenterRootDeviceName(ig.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("resolving root device for %q: %w", ig.Name, err)
	}
	blockDeviceMappings, err := buildKarpenterBlockDeviceMappings(ig, rootDeviceName)
	if err != nil {
		return nil, fmt.Errorf("building blockDeviceMappings for %q: %w", ig.Name, err)
	}

	subnetTerms := []karpenterSelectorTerm{
		{
			Tags: map[string]string{
				"KubernetesCluster":                     tf.ClusterName(),
				"kops.k8s.io/instance-group/" + ig.Name: "true",
			},
		},
	}

	securityGroupTerms := []karpenterSelectorTerm{}
	if ig.Spec.SecurityGroupOverride != nil {
		securityGroupTerms = append(securityGroupTerms, karpenterSelectorTerm{ID: fi.ValueOf(ig.Spec.SecurityGroupOverride)})
	} else {
		securityGroupTerms = append(securityGroupTerms, karpenterSelectorTerm{
			Tags: map[string]string{
				"KubernetesCluster": tf.ClusterName(),
				"Name":              fi.ValueOf(tf.LinkToSecurityGroup(ig.Spec.Role).Name),
			},
		})
	}
	for _, id := range ig.Spec.AdditionalSecurityGroups {
		securityGroupTerms = append(securityGroupTerms, karpenterSelectorTerm{ID: id})
	}

	return &karpenterEC2NodeClass{
		APIVersion: karpenterAWSAPIGroup + "/v1",
		Kind:       "EC2NodeClass",
		Metadata: karpenterObjectMeta{
			Name: ig.Name,
		},
		Spec: karpenterEC2NodeClassSpec{
			AMIFamily:                "Custom",
			AMISelectorTerms:         amiSelectorTerms,
			AssociatePublicIPAddress: associatePublicIP,
			BlockDeviceMappings:      blockDeviceMappings,
			Tags:                     tags,
			SubnetSelectorTerms:      subnetTerms,
			SecurityGroupTerms:       securityGroupTerms,
			InstanceProfile:          fi.ValueOf(instanceProfile.Name),
			UserData:                 userData,
			Kubelet:                  buildKarpenterKubeletConfiguration(ig),
		},
	}, nil
}

func buildKarpenterKubeletConfiguration(ig *kops.InstanceGroup) *karpenterKubeletConfiguration {
	if ig.Spec.Kubelet == nil {
		return nil
	}
	kubelet := &karpenterKubeletConfiguration{
		MaxPods:        ig.Spec.Kubelet.MaxPods,
		SystemReserved: ig.Spec.Kubelet.SystemReserved,
		KubeReserved:   ig.Spec.Kubelet.KubeReserved,
	}
	if kubelet.MaxPods == nil && len(kubelet.SystemReserved) == 0 && len(kubelet.KubeReserved) == 0 {
		return nil
	}
	return kubelet
}

func buildKarpenterBlockDeviceMappings(ig *kops.InstanceGroup, rootDeviceName string) ([]karpenterBlockDeviceMapping, error) {
	volumeSize, err := defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
	if err != nil {
		return nil, err
	}
	var volumeType ec2types.VolumeType
	deleteOnTermination := awsmodel.DefaultVolumeDeleteOnTermination
	encryption := awsmodel.DefaultVolumeEncryption
	var encryptionKey *string
	var iops, throughput *int32

	if rootVolume := ig.Spec.RootVolume; rootVolume != nil {
		if fi.ValueOf(rootVolume.Size) > 0 {
			volumeSize = fi.ValueOf(rootVolume.Size)
		}
		volumeType = ec2types.VolumeType(fi.ValueOf(rootVolume.Type))
		if rootVolume.Encryption != nil {
			encryption = fi.ValueOf(rootVolume.Encryption)
		}
		// As in the ASG path, the key is only honoured when encryption is set explicitly, not when
		// it is merely defaulted to true, and an empty key is dropped rather than sent to EC2.
		if fi.ValueOf(rootVolume.Encryption) && fi.ValueOf(rootVolume.EncryptionKey) != "" {
			encryptionKey = rootVolume.EncryptionKey
		}
		iops = rootVolume.IOPS
		throughput = rootVolume.Throughput
	}
	if volumeType == "" {
		volumeType = awsmodel.DefaultVolumeType
	}

	// IOPS and throughput are only meaningful for some volume types, and each has a
	// minimum below which EC2 rejects the request.
	switch volumeType {
	case ec2types.VolumeTypeIo1, ec2types.VolumeTypeIo2:
		if fi.ValueOf(iops) < awsmodel.DefaultVolumeIonIops {
			iops = new(int32(awsmodel.DefaultVolumeIonIops))
		}
		throughput = nil
	case ec2types.VolumeTypeGp3:
		if fi.ValueOf(iops) < awsmodel.DefaultVolumeGp3Iops {
			iops = new(int32(awsmodel.DefaultVolumeGp3Iops))
		}
		if fi.ValueOf(throughput) < awsmodel.DefaultVolumeGp3Throughput {
			throughput = new(int32(awsmodel.DefaultVolumeGp3Throughput))
		}
	default:
		iops = nil
		throughput = nil
	}

	ebs := &karpenterBlockDeviceEBS{
		DeleteOnTermination: new(deleteOnTermination),
		Encrypted:           new(encryption),
		KMSKeyID:            encryptionKey,
		VolumeSize:          fmt.Sprintf("%dGi", volumeSize),
		VolumeType:          string(volumeType),
	}
	if iops != nil {
		ebs.IOPS = new(int64(fi.ValueOf(iops)))
	}
	if throughput != nil {
		ebs.Throughput = new(int64(fi.ValueOf(throughput)))
	}

	return []karpenterBlockDeviceMapping{
		{
			DeviceName: rootDeviceName,
			EBS:        ebs,
			RootVolume: new(true),
		},
	}, nil
}

// karpenterRootDeviceName resolves the root device name of the InstanceGroup image, so
// that the generated block device mapping overrides the image's root volume rather than
// attaching an additional one. The name varies between images (/dev/xvda, /dev/sda1),
// so it has to come from the image itself.
func (tf *TemplateFunctions) karpenterRootDeviceName(image string) (string, error) {
	cloud, ok := tf.cloud.(awsup.AWSCloud)
	if !ok {
		return "", fmt.Errorf("expected an AWS cloud, got %T", tf.cloud)
	}
	resolved, err := cloud.ResolveImage(image)
	if err != nil {
		return "", fmt.Errorf("unable to resolve image %q: %w", image, err)
	}
	if resolved == nil {
		return "", fmt.Errorf("unable to resolve image %q: not found", image)
	}
	rootDeviceName := fi.ValueOf(resolved.RootDeviceName)
	if rootDeviceName == "" {
		return "", fmt.Errorf("image %q has no root device name", image)
	}
	return rootDeviceName, nil
}

func (tf *TemplateFunctions) buildKarpenterNodePool(ig *kops.InstanceGroup) (*karpenterNodePool, error) {
	labels, err := nodelabels.BuildNodeLabels(tf.Cluster, ig)
	if err != nil {
		return nil, fmt.Errorf("building node labels for %q: %w", ig.Name, err)
	}
	labels = karpenterNodePoolTemplateLabels(labels)

	requirements, err := tf.karpenterRequirements(ig)
	if err != nil {
		return nil, fmt.Errorf("building requirements for %q: %w", ig.Name, err)
	}

	template := karpenterNodeClaimTemplate{
		Spec: karpenterNodeClaimSpec{
			Requirements: requirements,
			NodeClassRef: karpenterNodeClassRef{
				Group: karpenterAWSAPIGroup,
				Kind:  "EC2NodeClass",
				Name:  ig.Name,
			},
		},
	}
	if len(labels) != 0 {
		template.Metadata = &karpenterNodeClaimMetadata{Labels: labels}
	}
	for _, taintSpec := range ig.Spec.Taints {
		taint, err := kopsutil.ParseTaint(taintSpec)
		if err != nil {
			return nil, fmt.Errorf("parsing taint %q for %q: %w", taintSpec, ig.Name, err)
		}
		template.Spec.Taints = append(template.Spec.Taints, karpenterTaint{
			Key:    taint["key"],
			Value:  taint["value"],
			Effect: taint["effect"],
		})
	}

	spec := karpenterNodePoolSpec{
		Template: template,
	}
	if ig.Spec.MinSize != nil && *ig.Spec.MinSize > 0 {
		spec.Replicas = new(int64(*ig.Spec.MinSize))
	}
	if ig.Spec.MaxSize != nil {
		spec.Limits = &karpenterNodePoolLimits{Nodes: strconv.FormatInt(int64(*ig.Spec.MaxSize), 10)}
	}

	return &karpenterNodePool{
		APIVersion: karpenterNodePoolAPIGroup + "/v1",
		Kind:       "NodePool",
		Metadata: karpenterObjectMeta{
			Name: ig.Name,
		},
		Spec: spec,
	}, nil
}

func buildKarpenterAMITerms(image string) ([]karpenterAMITerm, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return nil, fmt.Errorf("image is required")
	}
	if strings.Contains(image, "://") {
		return nil, fmt.Errorf("image %q must be ami-*, ssm:<parameter>, <name>, or <owner>/<name>", image)
	}
	if strings.HasPrefix(image, "ami-") {
		return []karpenterAMITerm{{ID: image}}, nil
	}
	if strings.HasPrefix(image, "ssm:") {
		parameter := strings.TrimPrefix(image, "ssm:")
		if parameter == "" {
			return nil, fmt.Errorf("ssm image parameter is required")
		}
		return []karpenterAMITerm{{SSMParameter: parameter}}, nil
	}

	tokens := strings.SplitN(image, "/", 2)
	if len(tokens) == 1 {
		return []karpenterAMITerm{{Name: image, Owner: "self"}}, nil
	}
	if tokens[0] == "" || tokens[1] == "" {
		return nil, fmt.Errorf("image %q must be ami-*, ssm:<parameter>, <name>, or <owner>/<name>", image)
	}
	return []karpenterAMITerm{{Owner: awsup.ResolveImageOwnerAlias(tokens[0]), Name: tokens[1]}}, nil
}

func (tf *TemplateFunctions) karpenterAssociatePublicIP(ig *kops.InstanceGroup) (*bool, error) {
	subnets, err := tf.GatherSubnets(ig)
	if err != nil {
		return nil, err
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("could not determine any subnets for InstanceGroup %q; subnets was %s", ig.Name, ig.Spec.Subnets)
	}

	switch subnets[0].Type {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		if ig.Spec.AssociatePublicIP != nil {
			return ig.Spec.AssociatePublicIP, nil
		}
		return new(true), nil
	case kops.SubnetTypeDualStack, kops.SubnetTypePrivate:
		return new(false), nil
	default:
		return nil, fmt.Errorf("unknown subnet type %q for InstanceGroup %q", subnets[0].Type, ig.Name)
	}
}

func (tf *TemplateFunctions) karpenterRequirements(ig *kops.InstanceGroup) ([]karpenterRequirement, error) {
	requirements := []karpenterRequirement{
		{
			Key:      karpenterOSLabel,
			Operator: "In",
			Values:   []string{"linux"},
		},
	}

	var instanceRequirements *kops.InstanceRequirementsSpec
	if ig.Spec.MixedInstancesPolicy != nil {
		instanceRequirements = ig.Spec.MixedInstancesPolicy.InstanceRequirements
	}

	if instanceRequirements != nil {
		converted, err := karpenterInstanceRequirements(instanceRequirements)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, converted...)
	} else if instanceTypes := karpenterInstanceTypes(ig); len(instanceTypes) != 0 {
		requirements = append(requirements, karpenterRequirement{
			Key:      karpenterInstanceTypeLabel,
			Operator: "In",
			Values:   instanceTypes,
		})
	}

	requirements = append(requirements, karpenterRequirement{
		Key:      karpenterCapacityTypeLabel,
		Operator: "In",
		Values:   karpenterCapacityTypes(ig),
	})

	return requirements, nil
}

func karpenterInstanceRequirements(spec *kops.InstanceRequirementsSpec) ([]karpenterRequirement, error) {
	var requirements []karpenterRequirement

	// Karpenter's Gt/Lt/Gte/Lte operators take exactly one value, interpreted as an integer.
	addBound := func(key string, quantity *resource.Quantity, operator string, toValue func(*resource.Quantity) (int64, error)) error {
		if quantity == nil {
			return nil
		}
		value, err := toValue(quantity)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		requirements = append(requirements, karpenterRequirement{
			Key:      key,
			Operator: operator,
			Values:   []string{strconv.FormatInt(value, 10)},
		})
		return nil
	}

	if cpu := spec.CPU; cpu != nil {
		if err := addBound(karpenterInstanceCPULabel, cpu.Min, "Gte", quantityToCount); err != nil {
			return nil, err
		}
		if err := addBound(karpenterInstanceCPULabel, cpu.Max, "Lte", quantityToCount); err != nil {
			return nil, err
		}
	}

	if memory := spec.Memory; memory != nil {
		// Round the lower bound up and the upper bound down, so the generated envelope
		// never admits an instance outside the requested range.
		if err := addBound(karpenterInstanceMemoryLabel, memory.Min, "Gte", quantityToMiBRoundUp); err != nil {
			return nil, err
		}
		if err := addBound(karpenterInstanceMemoryLabel, memory.Max, "Lte", quantityToMiBRoundDown); err != nil {
			return nil, err
		}
	}

	excluded, err := karpenterExcludedTypeRequirements(spec.ExcludedInstanceTypes)
	if err != nil {
		return nil, err
	}
	return append(requirements, excluded...), nil
}

// karpenterExcludedInstanceFamily matches the "<family>.*" wildcard form of
// excludedInstanceTypes, which maps to an instance family rather than an instance type.
var karpenterExcludedInstanceFamily = regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)\.\*$`)

func karpenterExcludedTypeRequirements(excluded []string) ([]karpenterRequirement, error) {
	var families, instanceTypes []string
	for _, entry := range excluded {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if match := karpenterExcludedInstanceFamily.FindStringSubmatch(entry); match != nil {
			families = append(families, match[1])
			continue
		}
		if strings.Contains(entry, "*") {
			return nil, fmt.Errorf("excludedInstanceTypes entry %q is not supported for Karpenter; only an instance type or a %q family wildcard can be expressed as a NodePool requirement", entry, "<family>.*")
		}
		instanceTypes = append(instanceTypes, entry)
	}

	var requirements []karpenterRequirement
	if len(families) != 0 {
		sort.Strings(families)
		requirements = append(requirements, karpenterRequirement{
			Key:      karpenterInstanceFamilyLabel,
			Operator: "NotIn",
			Values:   families,
		})
	}
	if len(instanceTypes) != 0 {
		sort.Strings(instanceTypes)
		requirements = append(requirements, karpenterRequirement{
			Key:      karpenterInstanceTypeLabel,
			Operator: "NotIn",
			Values:   instanceTypes,
		})
	}
	return requirements, nil
}

// quantityToCount converts a Quantity to a whole count, for labels such as instance-cpu.
func quantityToCount(quantity *resource.Quantity) (int64, error) {
	value, ok := quantity.AsInt64()
	if !ok {
		return 0, fmt.Errorf("value %q is not a whole number", quantity.String())
	}
	return checkKarpenterBound(quantity, value)
}

// quantityToMiBRoundUp and quantityToMiBRoundDown convert a Quantity to MiB, the unit of
// the karpenter.k8s.aws/instance-memory label. Note this differs from the ASG path, which
// scales by 10^6 into a field AWS defines as MiB; see kubernetes/kops#15305.
func quantityToMiBRoundUp(quantity *resource.Quantity) (int64, error) {
	bytes := quantity.Value()
	return checkKarpenterBound(quantity, (bytes+mib-1)/mib)
}

func quantityToMiBRoundDown(quantity *resource.Quantity) (int64, error) {
	return checkKarpenterBound(quantity, quantity.Value()/mib)
}

const mib = 1024 * 1024

func checkKarpenterBound(quantity *resource.Quantity, value int64) (int64, error) {
	if value < 0 {
		return 0, fmt.Errorf("value %q must not be negative", quantity.String())
	}
	if value > math.MaxInt32 {
		return 0, fmt.Errorf("value %q is too large", quantity.String())
	}
	return value, nil
}

func karpenterInstanceTypes(ig *kops.InstanceGroup) []string {
	seen := make(map[string]bool)
	var values []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		values = append(values, value)
	}

	if ig.Spec.MixedInstancesPolicy != nil {
		for _, instanceType := range ig.Spec.MixedInstancesPolicy.Instances {
			add(instanceType)
		}
	}
	for _, instanceType := range strings.Split(ig.Spec.MachineType, ",") {
		add(instanceType)
	}

	sort.Strings(values)
	return values
}

func karpenterCapacityTypes(ig *kops.InstanceGroup) []string {
	if ig.Spec.MaxPrice != nil || ig.Spec.SpotDurationInMinutes != nil {
		return []string{"spot"}
	}
	if ig.Spec.MixedInstancesPolicy != nil {
		spec := ig.Spec.MixedInstancesPolicy
		if spec.OnDemandAboveBase != nil && fi.ValueOf(spec.OnDemandAboveBase) < 100 {
			return []string{"on-demand", "spot"}
		}
	}
	return []string{"on-demand"}
}

func karpenterEC2NodeClassTags(tags map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range tags {
		if k != "" && !isKarpenterEC2NodeClassReservedTag(k) {
			filtered[k] = v
		}
	}
	return filtered
}

func isKarpenterEC2NodeClassReservedTag(key string) bool {
	if strings.HasPrefix(key, "kubernetes.io/cluster") {
		return true
	}
	switch key {
	case "eks:eks-cluster-name", karpenterNodePoolLabel, "karpenter.sh/nodeclaim", "karpenter.k8s.aws/ec2nodeclass":
		return true
	}
	return false
}

func karpenterNodePoolTemplateLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range labels {
		if !isKarpenterNodePoolTemplateReservedLabel(k) {
			filtered[k] = v
		}
	}
	return filtered
}

func isKarpenterNodePoolTemplateReservedLabel(key string) bool {
	if key == karpenterCapacityTypeLabel {
		return false
	}
	if key == "kubernetes.io/hostname" {
		return true
	}
	domain, _, found := strings.Cut(key, "/")
	if !found {
		return false
	}
	return domain == "karpenter.sh" || strings.HasSuffix(domain, ".karpenter.sh") ||
		domain == "karpenter.k8s.aws" || strings.HasSuffix(domain, ".karpenter.k8s.aws")
}

func (tf *TemplateFunctions) managedFileContents(name string) (string, error) {
	task, err := tf.Task("ManagedFile", name)
	if err != nil {
		return "", err
	}
	managedFile, ok := task.(*fitasks.ManagedFile)
	if !ok {
		return "", fmt.Errorf("task %q is %T, expected *fitasks.ManagedFile", name, task)
	}
	data, err := fi.ResourceAsBytes(managedFile.Contents)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
