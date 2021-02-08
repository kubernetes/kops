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

package model

import (
	"fmt"
	"net"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/model/iam"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/blang/semver/v4"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog/v2"
)

const (
	clusterAutoscalerNodeTemplateTaint = "k8s.io/cluster-autoscaler/node-template/taint/"
)

// KopsModelContext is the kops model
type KopsModelContext struct {
	iam.IAMModelContext
	InstanceGroups     []*kops.InstanceGroup
	SecurityGroupRules map[string]*awstasks.SecurityGroupRule
	Region             string
	SSHPublicKeys      [][]byte
}

// GatherSubnets maps the subnet names in an InstanceGroup to the ClusterSubnetSpec objects (which are stored on the Cluster)
func (m *KopsModelContext) GatherSubnets(ig *kops.InstanceGroup) ([]*kops.ClusterSubnetSpec, error) {
	var subnets []*kops.ClusterSubnetSpec
	var subnetType kops.SubnetType

	for _, subnetName := range ig.Spec.Subnets {
		var matches []*kops.ClusterSubnetSpec
		for i := range m.Cluster.Spec.Subnets {
			clusterSubnet := &m.Cluster.Spec.Subnets[i]
			if clusterSubnet.Name == subnetName {
				matches = append(matches, clusterSubnet)
			}
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("subnet not found: %q", subnetName)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("found multiple subnets with name: %q", subnetName)
		}
		subnets = append(subnets, matches[0])

		// @step: check the instance is not cross subnet types
		switch subnetType {
		case "":
			subnetType = matches[0].Type
		default:
			if matches[0].Type != subnetType {
				return nil, fmt.Errorf("found subnets of different types: %v", strings.Join([]string{string(subnetType), string(matches[0].Type)}, ","))
			}
		}
	}

	return subnets, nil
}

// FindInstanceGroup returns the instance group with the matching Name (or nil if not found)
func (m *KopsModelContext) FindInstanceGroup(name string) *kops.InstanceGroup {
	for _, ig := range m.InstanceGroups {
		if ig.ObjectMeta.Name == name {
			return ig
		}
	}
	return nil
}

// FindSubnet returns the subnet with the matching Name (or nil if not found)
func (m *KopsModelContext) FindSubnet(name string) *kops.ClusterSubnetSpec {
	return model.FindSubnet(m.Cluster, name)
}

// FindZonesForInstanceGroup finds the zones for an InstanceGroup
func (m *KopsModelContext) FindZonesForInstanceGroup(ig *kops.InstanceGroup) ([]string, error) {
	return model.FindZonesForInstanceGroup(m.Cluster, ig)
}

// MasterInstanceGroups returns InstanceGroups with the master role
func (m *KopsModelContext) MasterInstanceGroups() []*kops.InstanceGroup {
	var groups []*kops.InstanceGroup
	for _, ig := range m.InstanceGroups {
		if !ig.IsMaster() {
			continue
		}
		groups = append(groups, ig)
	}
	return groups
}

// NodeInstanceGroups returns InstanceGroups with the node role
func (m *KopsModelContext) NodeInstanceGroups() []*kops.InstanceGroup {
	var groups []*kops.InstanceGroup
	for _, ig := range m.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleNode {
			continue
		}
		groups = append(groups, ig)
	}
	return groups
}

// CloudTagsForInstanceGroup computes the tags to apply to instances in the specified InstanceGroup
func (m *KopsModelContext) CloudTagsForInstanceGroup(ig *kops.InstanceGroup) (map[string]string, error) {
	labels := m.CloudTags(m.AutoscalingGroupName(ig), false)

	// Apply any user-specified global labels first so they can be overridden by IG-specific labels
	for k, v := range m.Cluster.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply labels for cluster autoscaler node labels
	for k, v := range nodelabels.BuildNodeLabels(m.Cluster, ig) {
		labels[nodeidentityaws.ClusterAutoscalerNodeTemplateLabel+k] = v
	}

	// Apply labels for cluster autoscaler node taints
	for _, v := range ig.Spec.Taints {
		splits := strings.SplitN(v, "=", 2)
		if len(splits) > 1 {
			labels[clusterAutoscalerNodeTemplateTaint+splits[0]] = splits[1]
		}
	}

	// The system tags take priority because the cluster likely breaks without them...

	if ig.Spec.Role == kops.InstanceGroupRoleMaster {
		labels[awstasks.CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleMaster))] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleNode {
		labels[awstasks.CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleNode))] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		labels[awstasks.CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleBastion))] = "1"
	}

	labels[nodeidentityaws.CloudTagInstanceGroupName] = ig.Name

	return labels, nil
}

// CloudTags computes the tags to apply to a normal cloud resource with the specified name
func (m *KopsModelContext) CloudTags(name string, shared bool) map[string]string {
	tags := make(map[string]string)

	switch kops.CloudProviderID(m.Cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		if shared {
			// If the resource is shared, we don't try to set the Name - we presume that is managed externally
			klog.V(4).Infof("Skipping Name tag for shared resource")
		} else {
			if name != "" {
				tags["Name"] = name
			}
		}

		// Kubernetes 1.6 introduced the shared ownership tag; that replaces TagClusterName
		setLegacyTag := true
		// For the moment, we only skip the legacy tag for shared resources
		// (other people may be using it)
		if shared {
			klog.V(4).Infof("Skipping %q tag for shared resource", awsup.TagClusterName)
			setLegacyTag = false
		}
		if setLegacyTag {
			tags[awsup.TagClusterName] = m.Cluster.ObjectMeta.Name
		}

		if shared {
			tags["kubernetes.io/cluster/"+m.Cluster.ObjectMeta.Name] = "shared"
		} else {
			tags["kubernetes.io/cluster/"+m.Cluster.ObjectMeta.Name] = "owned"
			for k, v := range m.Cluster.Spec.CloudLabels {
				tags[k] = v
			}
		}

	}
	return tags
}

// UseKopsControllerForNodeBootstrap checks if nodeup should use kops-controller to bootstrap.
func (m *KopsModelContext) UseKopsControllerForNodeBootstrap() bool {
	return model.UseKopsControllerForNodeBootstrap(m.Cluster)
}

// UseBootstrapTokens checks if bootstrap tokens are enabled
func (m *KopsModelContext) UseBootstrapTokens() bool {
	if m.Cluster.Spec.KubeAPIServer == nil || m.UseKopsControllerForNodeBootstrap() {
		return false
	}

	return fi.BoolValue(m.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken)
}

// UsesBastionDns checks if we should use a specific name for the bastion dns
func (m *KopsModelContext) UsesBastionDns() bool {
	if m.Cluster.Spec.Topology.Bastion != nil && m.Cluster.Spec.Topology.Bastion.BastionPublicName != "" {
		return true
	}
	return false
}

// UsesSSHBastion checks if we have a Bastion in the cluster
func (m *KopsModelContext) UsesSSHBastion() bool {
	for _, ig := range m.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			return true
		}
	}

	return false
}

// UseLoadBalancerForAPI checks if we are using a load balancer for the kubeapi
func (m *KopsModelContext) UseLoadBalancerForAPI() bool {
	if m.Cluster.Spec.API == nil {
		return false
	}
	return m.Cluster.Spec.API.LoadBalancer != nil
}

// UseLoadBalancerForInternalAPI check if true then we will use the created loadbalancer for internal kubelet
// connections.  The intention here is to make connections to apiserver more
// HA - see https://github.com/kubernetes/kops/issues/4252
func (m *KopsModelContext) UseLoadBalancerForInternalAPI() bool {
	return m.UseLoadBalancerForAPI() &&
		m.Cluster.Spec.API.LoadBalancer.UseForInternalApi
}

// APILoadBalancerClass returns which type of load balancer to use for the api
func (m *KopsModelContext) APILoadBalancerClass() kops.LoadBalancerClass {
	if m.Cluster.Spec.API != nil && m.Cluster.Spec.API.LoadBalancer != nil {
		return m.Cluster.Spec.API.LoadBalancer.Class
	}
	return kops.LoadBalancerClassClassic
}

// UsePrivateDNS checks if we are using private DNS
func (m *KopsModelContext) UsePrivateDNS() bool {
	topology := m.Cluster.Spec.Topology
	if topology != nil && topology.DNS != nil {
		switch topology.DNS.Type {
		case kops.DNSTypePublic:
			return false
		case kops.DNSTypePrivate:
			return true

		default:
			klog.Warningf("Unknown DNS type %q", topology.DNS.Type)
			return false
		}
	}

	return false
}

// UseClassicLoadBalancer checks if we are using Classic LoadBalancer
func (m *KopsModelContext) UseClassicLoadBalancer() bool {
	return m.Cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassClassic
}

// UseNetworkLoadBalancer checks if we are using Network LoadBalancer
func (m *KopsModelContext) UseNetworkLoadBalancer() bool {
	return m.Cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork
}

// UseEtcdManager checks to see if etcd manager is enabled
func (c *KopsModelContext) UseEtcdManager() bool {
	for _, x := range c.Cluster.Spec.EtcdClusters {
		if x.Provider == kops.EtcdProviderTypeManager {
			return true
		}
	}

	return false
}

// UseEtcdTLS checks to see if etcd tls is enabled
func (m *KopsModelContext) UseEtcdTLS() bool {
	for _, x := range m.Cluster.Spec.EtcdClusters {
		if x.EnableEtcdTLS {
			return true
		}
	}

	return false
}

// UseSSHKey returns true if SSHKeyName from the cluster spec is not set to an empty string (""). Setting SSHKeyName
// to an empty string indicates that an SSH key should not be set on instances.
func (m *KopsModelContext) UseSSHKey() bool {
	sshKeyName := m.Cluster.Spec.SSHKeyName
	return sshKeyName == nil || *sshKeyName != ""
}

// KubernetesVersion parses the semver version of kubernetes, from the cluster spec
func (m *KopsModelContext) KubernetesVersion() semver.Version {
	// TODO: Remove copy-pasting c.f. https://github.com/kubernetes/kops/blob/master/pkg/model/components/context.go#L32

	kubernetesVersion := m.Cluster.Spec.KubernetesVersion

	if kubernetesVersion == "" {
		klog.Fatalf("KubernetesVersion is required")
	}

	sv, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		klog.Fatalf("unable to determine kubernetes version from %q", kubernetesVersion)
	}

	return *sv
}

// IsKubernetesGTE checks if the kubernetes version is at least version, ignoring prereleases / patches
func (m *KopsModelContext) IsKubernetesGTE(version string) bool {
	return util.IsKubernetesGTE(version, m.KubernetesVersion())
}

// IsKubernetesLT checks if the kubernetes version is before the specified version, ignoring prereleases / patches
func (m *KopsModelContext) IsKubernetesLT(version string) bool {
	return !m.IsKubernetesGTE(version)
}

// WellKnownServiceIP returns a service ip with the service cidr
func (m *KopsModelContext) WellKnownServiceIP(id int) (net.IP, error) {
	return components.WellKnownServiceIP(&m.Cluster.Spec, id)
}

// NodePortRange returns the range of ports allocated to NodePorts
func (m *KopsModelContext) NodePortRange() (utilnet.PortRange, error) {
	// defaultServiceNodePortRange is the default port range for NodePort services.
	defaultServiceNodePortRange := utilnet.PortRange{Base: 30000, Size: 2768}

	kubeApiServer := m.Cluster.Spec.KubeAPIServer
	if kubeApiServer != nil && kubeApiServer.ServiceNodePortRange != "" {
		err := defaultServiceNodePortRange.Set(kubeApiServer.ServiceNodePortRange)
		if err != nil {
			return utilnet.PortRange{}, fmt.Errorf("error parsing ServiceNodePortRange %q", kubeApiServer.ServiceNodePortRange)
		}
	}

	return defaultServiceNodePortRange, nil
}

// UseServiceAccountIAM returns true if we are using service-account bound IAM roles.
func (m *KopsModelContext) UseServiceAccountIAM() bool {
	return featureflag.UseServiceAccountIAM.Enabled()
}
