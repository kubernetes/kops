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
	"encoding/base32"
	"fmt"
	"hash/fnv"
	"net"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/components"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	"github.com/blang/semver"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/klog"
)

const (
	clusterAutoscalerNodeTemplateLabel = "k8s.io/cluster-autoscaler/node-template/label/"
	clusterAutoscalerNodeTemplateTaint = "k8s.io/cluster-autoscaler/node-template/taint/"
)

var UseLegacyELBName = featureflag.New("UseLegacyELBName", featureflag.Bool(false))

// KopsModelContext is the kops model
type KopsModelContext struct {
	Cluster        *kops.Cluster
	InstanceGroups []*kops.InstanceGroup
	Region         string
	SSHPublicKeys  [][]byte
}

// GetELBName32 will attempt to calculate a meaningful name for an ELB given a prefix
// Will never return a string longer than 32 chars
// Note this is _not_ the primary identifier for the ELB - we use the Name tag for that.
func (m *KopsModelContext) GetELBName32(prefix string) string {
	c := m.Cluster.ObjectMeta.Name

	if UseLegacyELBName.Enabled() {
		tokens := strings.Split(c, ".")
		s := fmt.Sprintf("%s-%s", prefix, tokens[0])
		if len(s) > 32 {
			s = s[:32]
		}
		klog.Infof("UseLegacyELBName feature-flag is set; built legacy name %q", s)
		return s
	}

	// The LoadBalancerName is exposed publicly as the DNS name for the load balancer.
	// So this will likely become visible in a CNAME record - this is potentially some
	// information leakage.
	// But... if a user can see the CNAME record, they can see the actual record also,
	// which will be the full cluster name.
	s := prefix + "-" + strings.Replace(c, ".", "-", -1)

	// We have a 32 character limit for ELB names
	// But we always compute the hash and add it, lest we trick users into assuming that we never do this
	h := fnv.New32a()
	if _, err := h.Write([]byte(s)); err != nil {
		klog.Fatalf("error hashing values: %v", err)
	}
	hashString := base32.HexEncoding.EncodeToString(h.Sum(nil))
	hashString = strings.ToLower(hashString)
	if len(hashString) > 6 {
		hashString = hashString[:6]
	}

	maxBaseLength := 32 - len(hashString) - 1
	if len(s) > maxBaseLength {
		s = s[:maxBaseLength]
	}
	s = s + "-" + hashString

	return s
}

// ClusterName returns the cluster name
func (m *KopsModelContext) ClusterName() string {
	return m.Cluster.ObjectMeta.Name
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
	labels := make(map[string]string)

	// Apply any user-specified global labels first so they can be overridden by IG-specific labels
	for k, v := range m.Cluster.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// Apply labels for cluster autoscaler node labels
	for k, v := range ig.Spec.NodeLabels {
		labels[clusterAutoscalerNodeTemplateLabel+k] = v
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
		if m.IsKubernetesGTE("1.6") {
			// For the moment, we only skip the legacy tag for shared resources
			// (other people may be using it)
			if shared {
				klog.V(4).Infof("Skipping %q tag for shared resource", awsup.TagClusterName)
				setLegacyTag = false
			}
		}
		if setLegacyTag {
			tags[awsup.TagClusterName] = m.Cluster.ObjectMeta.Name
		}

		if shared {
			tags["kubernetes.io/cluster/"+m.Cluster.ObjectMeta.Name] = "shared"
		} else {
			tags["kubernetes.io/cluster/"+m.Cluster.ObjectMeta.Name] = "owned"
		}

	}
	return tags
}

// UseBootstrapTokens checks if bootstrap tokens are enabled
func (m *KopsModelContext) UseBootstrapTokens() bool {
	if m.Cluster.Spec.KubeAPIServer == nil {
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
