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
	"sort"
	"strconv"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
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

func (tf *TemplateFunctions) buildKarpenterNodePool(ig *kops.InstanceGroup) (*karpenterNodePool, error) {
	labels, err := nodelabels.BuildNodeLabels(tf.Cluster, ig)
	if err != nil {
		return nil, fmt.Errorf("building node labels for %q: %w", ig.Name, err)
	}
	labels = karpenterNodePoolTemplateLabels(labels)

	template := karpenterNodeClaimTemplate{
		Spec: karpenterNodeClaimSpec{
			Requirements: tf.karpenterRequirements(ig),
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

func (tf *TemplateFunctions) karpenterRequirements(ig *kops.InstanceGroup) []karpenterRequirement {
	requirements := []karpenterRequirement{
		{
			Key:      "kubernetes.io/os",
			Operator: "In",
			Values:   []string{"linux"},
		},
	}

	instanceTypes := karpenterInstanceTypes(ig)
	if len(instanceTypes) != 0 {
		requirements = append(requirements, karpenterRequirement{
			Key:      "node.kubernetes.io/instance-type",
			Operator: "In",
			Values:   instanceTypes,
		})
	}

	requirements = append(requirements, karpenterRequirement{
		Key:      karpenterCapacityTypeLabel,
		Operator: "In",
		Values:   karpenterCapacityTypes(ig),
	})

	return requirements
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
