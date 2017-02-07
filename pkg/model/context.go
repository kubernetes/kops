/*
Copyright 2016 The Kubernetes Authors.

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
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
)

type KopsModelContext struct {
	Region         string
	Cluster        *kops.Cluster
	InstanceGroups []*kops.InstanceGroup

	SSHPublicKeys [][]byte
}

// Will attempt to calculate a meaningful name for an ELB given a prefix
// Will never return a string longer than 32 chars
func (m *KopsModelContext) GetELBName32(prefix string) (string, error) {
	var returnString string
	c := m.Cluster.ObjectMeta.Name
	s := strings.Split(c, ".")

	// TODO: We used to have this...
	//master-{{ replace .ClusterName "." "-" }}
	// TODO: strings.Split cannot return empty
	if len(s) > 0 {
		returnString = fmt.Sprintf("%s-%s", prefix, s[0])
	} else {
		returnString = fmt.Sprintf("%s-%s", prefix, c)
	}
	if len(returnString) > 32 {
		returnString = returnString[:32]
	}
	return returnString, nil
}

func (m *KopsModelContext) ClusterName() string {
	return m.Cluster.ObjectMeta.Name
}

// GatherSubnets maps the subnet names in an InstanceGroup to the ClusterSubnetSpec objects (which are stored on the Cluster)
func (m *KopsModelContext) GatherSubnets(ig *kops.InstanceGroup) ([]*kops.ClusterSubnetSpec, error) {
	var subnets []*kops.ClusterSubnetSpec
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
	for i := range m.Cluster.Spec.Subnets {
		s := &m.Cluster.Spec.Subnets[i]
		if s.Name == name {
			return s
		}
	}
	return nil
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

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// The system tags take priority because the cluster likely breaks without them...

	if ig.Spec.Role == kops.InstanceGroupRoleMaster {
		labels["k8s.io/role/master"] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleNode {
		labels["k8s.io/role/node"] = "1"
	}

	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		labels["k8s.io/role/bastion"] = "1"
	}

	return labels, nil
}

func (m *KopsModelContext) UsesBastionDns() bool {
	if m.Cluster.Spec.Topology.Bastion != nil && m.Cluster.Spec.Topology.Bastion.BastionPublicName != "" {
		return true
	}
	return false
}

func (m *KopsModelContext) UsesSSHBastion() bool {
	for _, ig := range m.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			return true
		}
	}

	return false
}

func (m *KopsModelContext) UseLoadBalancerForAPI() bool {
	if m.Cluster.Spec.API == nil {
		return false
	}
	return m.Cluster.Spec.API.LoadBalancer != nil
}

func (m *KopsModelContext) UsePrivateDNS() bool {
	topology := m.Cluster.Spec.Topology
	if topology != nil && topology.DNS != nil {
		switch topology.DNS.Type {
		case kops.DNSTypePublic:
			return false
		case kops.DNSTypePrivate:
			return true

		default:
			glog.Warningf("Unknown DNS type %q", topology.DNS.Type)
			return false
		}
	}

	return false
}

// KubernetesVersion parses the semver version of kubernetes, from the cluster spec
func (c *KopsModelContext) KubernetesVersion() (semver.Version, error) {
	// TODO: Remove copy-pasting c.f. https://github.com/kubernetes/kops/blob/master/pkg/model/components/context.go#L32

	kubernetesVersion := c.Cluster.Spec.KubernetesVersion

	if kubernetesVersion == "" {
		return semver.Version{}, fmt.Errorf("KubernetesVersion is required")
	}

	sv, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		return semver.Version{}, fmt.Errorf("unable to determine kubernetes version from %q", kubernetesVersion)
	}

	return *sv, nil
}

// VersionGTE is a simplified semver comparison
func VersionGTE(version semver.Version, major uint64, minor uint64) bool {
	if version.Major > major {
		return true
	}
	if version.Major == major && version.Minor >= minor {
		return true
	}
	return false
}
