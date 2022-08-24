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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/kubemanifest"
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
	InstanceGroups []*kops.InstanceGroup
	Region         string
	SSHPublicKeys  [][]byte

	// AdditionalObjects holds cluster-asssociated configuration objects, other than the Cluster and InstanceGroups.
	AdditionalObjects kubemanifest.ObjectList
}

// GatherSubnets maps the subnet names in an InstanceGroup to the ClusterSubnetSpec objects (which are stored on the Cluster)
func (b *KopsModelContext) GatherSubnets(ig *kops.InstanceGroup) ([]*kops.ClusterSubnetSpec, error) {
	var subnets []*kops.ClusterSubnetSpec
	var subnetType kops.SubnetType

	for _, subnetName := range ig.Spec.Subnets {
		var matches []*kops.ClusterSubnetSpec
		for i := range b.Cluster.Spec.Subnets {
			clusterSubnet := &b.Cluster.Spec.Subnets[i]
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
func (b *KopsModelContext) FindInstanceGroup(name string) *kops.InstanceGroup {
	for _, ig := range b.InstanceGroups {
		if ig.ObjectMeta.Name == name {
			return ig
		}
	}
	return nil
}

// FindSubnet returns the subnet with the matching Name (or nil if not found)
func (b *KopsModelContext) FindSubnet(name string) *kops.ClusterSubnetSpec {
	return model.FindSubnet(b.Cluster, name)
}

// FindZonesForInstanceGroup finds the zones for an InstanceGroup
func (b *KopsModelContext) FindZonesForInstanceGroup(ig *kops.InstanceGroup) ([]string, error) {
	return model.FindZonesForInstanceGroup(b.Cluster, ig)
}

// MasterInstanceGroups returns InstanceGroups with the master role
func (b *KopsModelContext) MasterInstanceGroups() []*kops.InstanceGroup {
	var groups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if !ig.IsMaster() {
			continue
		}
		groups = append(groups, ig)
	}
	return groups
}

// NodeInstanceGroups returns InstanceGroups with the node role
func (b *KopsModelContext) NodeInstanceGroups() []*kops.InstanceGroup {
	var groups []*kops.InstanceGroup
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleNode {
			continue
		}
		groups = append(groups, ig)
	}
	return groups
}

// CloudTagsForInstanceGroup computes the tags to apply to instances in the specified InstanceGroup
func (b *KopsModelContext) CloudTagsForInstanceGroup(ig *kops.InstanceGroup) (map[string]string, error) {
	labels := b.CloudTags(b.AutoscalingGroupName(ig), false)

	// Apply any user-specified global labels first so they can be overridden by IG-specific labels
	for k, v := range b.Cluster.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply NTH Labels
	nth := b.Cluster.Spec.NodeTerminationHandler
	if nth != nil && fi.BoolValue(nth.Enabled) && fi.BoolValue(nth.EnableSQSTerminationDraining) {
		labels[fi.StringValue(nth.ManagedASGTag)] = ""
	}

	// Apply labels for cluster autoscaler node labels
	for k, v := range nodelabels.BuildNodeLabels(b.Cluster, ig) {
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

	if ig.Spec.Role == kops.InstanceGroupRoleAPIServer {
		labels[awstasks.CloudTagInstanceGroupRolePrefix+strings.ToLower(string(kops.InstanceGroupRoleAPIServer))] = "1"
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

func (b *KopsModelContext) CloudTagsForServiceAccount(name string, sa types.NamespacedName) map[string]string {
	tags := b.CloudTags(name, false)
	tags[awstasks.CloudTagServiceAccountName] = sa.Name
	tags[awstasks.CloudTagServiceAccountNamespace] = sa.Namespace
	return tags
}

// CloudTags computes the tags to apply to a normal cloud resource with the specified name
func (b *KopsModelContext) CloudTags(name string, shared bool) map[string]string {
	tags := make(map[string]string)

	switch b.Cluster.Spec.GetCloudProvider() {
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
			tags[awsup.TagClusterName] = b.Cluster.ObjectMeta.Name
		}

		if shared {
			tags["kubernetes.io/cluster/"+b.Cluster.ObjectMeta.Name] = "shared"
		} else {
			tags["kubernetes.io/cluster/"+b.Cluster.ObjectMeta.Name] = "owned"
			for k, v := range b.Cluster.Spec.CloudLabels {
				tags[k] = v
			}
		}
	}
	return tags
}

// UseKopsControllerForNodeBootstrap checks if nodeup should use kops-controller to bootstrap.
func (b *KopsModelContext) UseKopsControllerForNodeBootstrap() bool {
	return model.UseKopsControllerForNodeBootstrap(b.Cluster)
}

// UseBootstrapTokens checks if bootstrap tokens are enabled
func (b *KopsModelContext) UseBootstrapTokens() bool {
	if b.Cluster.Spec.KubeAPIServer == nil || b.UseKopsControllerForNodeBootstrap() {
		return false
	}

	return fi.BoolValue(b.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken)
}

// UsesBastionDns checks if we should use a specific name for the bastion dns
func (b *KopsModelContext) UsesBastionDns() bool {
	if b.Cluster.Spec.Topology.Bastion != nil && b.Cluster.Spec.Topology.Bastion.PublicName != "" {
		return true
	}
	return false
}

// UsesSSHBastion checks if we have a Bastion in the cluster
func (b *KopsModelContext) UsesSSHBastion() bool {
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			return true
		}
	}

	return false
}

// UseLoadBalancerForAPI checks if we are using a load balancer for the kubeapi
func (b *KopsModelContext) UseLoadBalancerForAPI() bool {
	if b.Cluster.Spec.API == nil {
		return false
	}
	return b.Cluster.Spec.API.LoadBalancer != nil
}

// UseLoadBalancerForInternalAPI check if true then we will use the created loadbalancer for internal kubelet
// connections.  The intention here is to make connections to apiserver more
// HA - see https://github.com/kubernetes/kops/issues/4252
func (b *KopsModelContext) UseLoadBalancerForInternalAPI() bool {
	return b.UseLoadBalancerForAPI() &&
		b.Cluster.Spec.API.LoadBalancer.UseForInternalAPI
}

// APILoadBalancerClass returns which type of load balancer to use for the api
func (b *KopsModelContext) APILoadBalancerClass() kops.LoadBalancerClass {
	if b.Cluster.Spec.API != nil && b.Cluster.Spec.API.LoadBalancer != nil {
		return b.Cluster.Spec.API.LoadBalancer.Class
	}
	return kops.LoadBalancerClassClassic
}

// UsePrivateDNS checks if we are using private DNS
func (b *KopsModelContext) UsePrivateDNS() bool {
	topology := b.Cluster.Spec.Topology
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
func (b *KopsModelContext) UseClassicLoadBalancer() bool {
	return b.Cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassClassic
}

// UseNetworkLoadBalancer checks if we are using Network LoadBalancer
func (b *KopsModelContext) UseNetworkLoadBalancer() bool {
	return b.Cluster.Spec.API.LoadBalancer.Class == kops.LoadBalancerClassNetwork
}

// UseSSHKey returns true if SSHKeyName from the cluster spec is set to a nonempty string
// or there is an SSH public key provisioned in the key store.
func (b *KopsModelContext) UseSSHKey() bool {
	sshKeyName := b.Cluster.Spec.SSHKeyName
	if sshKeyName == nil {
		return len(b.SSHPublicKeys) > 0
	}
	return *sshKeyName != ""
}

// KubernetesVersion parses the semver version of kubernetes, from the cluster spec
func (b *KopsModelContext) KubernetesVersion() semver.Version {
	// TODO: Remove copy-pasting c.f. https://github.com/kubernetes/kops/blob/master/pkg/model/components/context.go#L32

	kubernetesVersion := b.Cluster.Spec.KubernetesVersion

	if kubernetesVersion == "" {
		klog.Fatalf("KubernetesVersion is required")
	}

	sv, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil || sv == nil {
		klog.Fatalf("unable to determine kubernetes version from %q: %v", kubernetesVersion, err)
	}
	return *sv
}

// IsKubernetesGTE checks if the kubernetes version is at least version, ignoring prereleases / patches
func (b *KopsModelContext) IsKubernetesGTE(version string) bool {
	return util.IsKubernetesGTE(version, b.KubernetesVersion())
}

// IsKubernetesLT checks if the kubernetes version is before the specified version, ignoring prereleases / patches
func (b *KopsModelContext) IsKubernetesLT(version string) bool {
	return !b.IsKubernetesGTE(version)
}

func (b *KopsModelContext) IsIPv6Only() bool {
	return b.Cluster.Spec.IsIPv6Only()
}

func (b *KopsModelContext) UseIPv6ForAPI() bool {
	for _, ig := range b.InstanceGroups {
		if ig.Spec.Role != kops.InstanceGroupRoleMaster && ig.Spec.Role != kops.InstanceGroupRoleAPIServer {
			break
		}
		for _, igSubnetName := range ig.Spec.Subnets {
			for _, clusterSubnet := range b.Cluster.Spec.Subnets {
				if igSubnetName != clusterSubnet.Name {
					continue
				}
				if clusterSubnet.IPv6CIDR != "" {
					return true
				}
			}
		}
	}
	return false
}

// WellKnownServiceIP returns a service ip with the service cidr
func (b *KopsModelContext) WellKnownServiceIP(id int) (net.IP, error) {
	return components.WellKnownServiceIP(&b.Cluster.Spec, id)
}

// NodePortRange returns the range of ports allocated to NodePorts
func (b *KopsModelContext) NodePortRange() (utilnet.PortRange, error) {
	// defaultServiceNodePortRange is the default port range for NodePort services.
	defaultServiceNodePortRange := utilnet.PortRange{Base: 30000, Size: 2768}

	kubeApiServer := b.Cluster.Spec.KubeAPIServer
	if kubeApiServer != nil && kubeApiServer.ServiceNodePortRange != "" {
		err := defaultServiceNodePortRange.Set(kubeApiServer.ServiceNodePortRange)
		if err != nil {
			return utilnet.PortRange{}, fmt.Errorf("error parsing ServiceNodePortRange %q", kubeApiServer.ServiceNodePortRange)
		}
	}

	return defaultServiceNodePortRange, nil
}

// UseServiceAccountExternalPermissions returns true if we are using service-account bound IAM roles.
func (b *KopsModelContext) UseServiceAccountExternalPermissions() bool {
	return b.Cluster.Spec.IAM != nil &&
		fi.BoolValue(b.Cluster.Spec.IAM.UseServiceAccountExternalPermissions)
}

// NetworkingIsCalico returns true if we are using calico networking
func (b *KopsModelContext) NetworkingIsCalico() bool {
	return b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Calico != nil
}

// NetworkingIsCilium returns true if we are using cilium networking
func (b *KopsModelContext) NetworkingIsCilium() bool {
	return b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Cilium != nil
}

// IsGossip returns true if we are using gossip instead of "real" DNS
func (b *KopsModelContext) IsGossip() bool {
	return dns.IsGossipHostname(b.Cluster.Name)
}
