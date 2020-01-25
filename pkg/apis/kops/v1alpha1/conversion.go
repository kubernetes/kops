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

package v1alpha1

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kops/pkg/apis/kops"
)

func Convert_v1alpha1_BastionSpec_To_kops_BastionSpec(in *BastionSpec, out *kops.BastionSpec, s conversion.Scope) error {
	out.BastionPublicName = in.PublicName
	out.IdleTimeoutSeconds = in.IdleTimeout

	if !in.Enable {
		out.BastionPublicName = ""
		out.IdleTimeoutSeconds = nil
	}

	return nil
}

func Convert_kops_KubeSchedulerConfig_To_v1alpha1_KubeSchedulerConfig(in *kops.KubeSchedulerConfig, out *KubeSchedulerConfig, s conversion.Scope) error {
	return autoConvert_kops_KubeSchedulerConfig_To_v1alpha1_KubeSchedulerConfig(in, out, s)
}

func Convert_kops_BastionSpec_To_v1alpha1_BastionSpec(in *kops.BastionSpec, out *BastionSpec, s conversion.Scope) error {
	out.PublicName = in.BastionPublicName
	out.IdleTimeout = in.IdleTimeoutSeconds

	out.Enable = true
	out.MachineType = ""

	return nil
}

func Convert_v1alpha1_ClusterSpec_To_kops_ClusterSpec(in *ClusterSpec, out *kops.ClusterSpec, s conversion.Scope) error {
	topologyPrivate := false
	if in.Topology != nil && in.Topology.Masters == TopologyPrivate {
		topologyPrivate = true
	}

	if in.Zones != nil {
		for _, z := range in.Zones {
			if topologyPrivate {
				// A private zone is mapped to a private- and a utility- subnet
				if z.PrivateCIDR != "" {
					out.Subnets = append(out.Subnets, kops.ClusterSubnetSpec{
						Name:       z.Name,
						CIDR:       z.PrivateCIDR,
						ProviderID: z.ProviderID,
						Zone:       z.Name,
						Type:       kops.SubnetTypePrivate,
						Egress:     z.Egress,
					})
				}

				if z.CIDR != "" {
					out.Subnets = append(out.Subnets, kops.ClusterSubnetSpec{
						Name:   "utility-" + z.Name,
						CIDR:   z.CIDR,
						Zone:   z.Name,
						Type:   kops.SubnetTypeUtility,
						Egress: z.Egress,
					})
				}
			} else {
				out.Subnets = append(out.Subnets, kops.ClusterSubnetSpec{
					Name:       z.Name,
					CIDR:       z.CIDR,
					ProviderID: z.ProviderID,
					Zone:       z.Name,
					Type:       kops.SubnetTypePublic,
					Egress:     z.Egress,
				})
			}
		}
	} else {
		out.Subnets = nil
	}

	adminAccess := in.AdminAccess
	if len(adminAccess) == 0 {
		// The default in v1alpha1 was 0.0.0.0/0
		adminAccess = []string{"0.0.0.0/0"}
	}
	out.SSHAccess = adminAccess
	out.KubernetesAPIAccess = adminAccess

	return autoConvert_v1alpha1_ClusterSpec_To_kops_ClusterSpec(in, out, s)
}

// ByName implements sort.Interface for []*ClusterZoneSpec on the Name field.
type ByName []*ClusterZoneSpec

func (a ByName) Len() int {
	return len(a)
}
func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

func Convert_kops_ClusterSpec_To_v1alpha1_ClusterSpec(in *kops.ClusterSpec, out *ClusterSpec, s conversion.Scope) error {
	topologyPrivate := false
	if in.Topology != nil && in.Topology.Masters == TopologyPrivate {
		topologyPrivate = true
	}

	if in.Subnets != nil {
		zoneMap := make(map[string]*ClusterZoneSpec)

		for _, s := range in.Subnets {
			zoneName := s.Name
			if s.Type == kops.SubnetTypeUtility {
				if !strings.HasPrefix(zoneName, "utility-") {
					return fmt.Errorf("cannot convert subnet to v1alpha1 when subnet with Type=utility does not have name starting with utility-: %q", zoneName)
				}
				zoneName = strings.TrimPrefix(zoneName, "utility-")
			}
			if s.Zone != zoneName {
				return fmt.Errorf("cannot convert to v1alpha1 when subnet Zone != Name: %q != %q", s.Zone, s.Name)
			}

			zone := zoneMap[zoneName]
			if zone == nil {
				zone = &ClusterZoneSpec{
					Name: s.Zone,
				}
				zoneMap[zoneName] = zone
			}

			if topologyPrivate {
				subnetType := s.Type
				if subnetType == "" {
					subnetType = kops.SubnetTypePrivate
				}
				switch subnetType {
				case kops.SubnetTypePrivate:
					if zone.PrivateCIDR != "" || zone.ProviderID != "" {
						return fmt.Errorf("cannot convert to v1alpha1: duplicate zone: %v", zone)
					}
					zone.PrivateCIDR = s.CIDR
					zone.Egress = s.Egress
					zone.ProviderID = s.ProviderID

				case kops.SubnetTypeUtility:
					if zone.CIDR != "" {
						return fmt.Errorf("cannot convert to v1alpha1: duplicate zone: %v", zone)
					}
					zone.CIDR = s.CIDR

					// We simple can't express this in v1alpha1
					if s.ProviderID != "" {
						return fmt.Errorf("cannot convert to v1alpha1: utility subnet had ProviderID %v", s.Name)
					}

				case kops.SubnetTypePublic:
					return fmt.Errorf("cannot convert to v1alpha1 when subnet type is public")

				default:
					return fmt.Errorf("unknown SubnetType: %v", subnetType)
				}
			} else {
				if zone.CIDR != "" || zone.ProviderID != "" {
					return fmt.Errorf("cannot convert to v1alpha1: duplicate zone: %v", zone)
				}
				zone.CIDR = s.CIDR
				zone.Egress = s.Egress
				zone.ProviderID = s.ProviderID
			}
		}

		for _, z := range zoneMap {
			out.Zones = append(out.Zones, z)
		}

		sort.Sort(ByName(out.Zones))
	} else {
		out.Zones = nil
	}

	if !reflect.DeepEqual(in.SSHAccess, in.KubernetesAPIAccess) {
		return fmt.Errorf("cannot convert to v1alpha1: SSHAccess != KubernetesAPIAccess")
	}
	out.AdminAccess = in.SSHAccess

	return autoConvert_kops_ClusterSpec_To_v1alpha1_ClusterSpec(in, out, s)
}

func Convert_v1alpha1_EtcdMemberSpec_To_kops_EtcdMemberSpec(in *EtcdMemberSpec, out *kops.EtcdMemberSpec, s conversion.Scope) error {
	if in.Zone != nil {
		instanceGroup := "master-" + *in.Zone
		out.InstanceGroup = &instanceGroup
	} else {
		out.InstanceGroup = nil
	}

	return autoConvert_v1alpha1_EtcdMemberSpec_To_kops_EtcdMemberSpec(in, out, s)
}

func Convert_kops_EtcdMemberSpec_To_v1alpha1_EtcdMemberSpec(in *kops.EtcdMemberSpec, out *EtcdMemberSpec, s conversion.Scope) error {
	err := autoConvert_kops_EtcdMemberSpec_To_v1alpha1_EtcdMemberSpec(in, out, s)
	if err != nil {
		return err
	}

	if in.InstanceGroup != nil {
		zone := *in.InstanceGroup
		if !strings.HasPrefix(zone, "master-") {
			return fmt.Errorf("cannot convert etc instance group name %q to v1alpha1: need master- prefix", zone)
		}
		zone = strings.TrimPrefix(zone, "master-")
		out.Zone = &zone
	} else {
		out.Zone = nil
	}

	return nil
}

func Convert_v1alpha1_InstanceGroupSpec_To_kops_InstanceGroupSpec(in *InstanceGroupSpec, out *kops.InstanceGroupSpec, s conversion.Scope) error {
	err := autoConvert_v1alpha1_InstanceGroupSpec_To_kops_InstanceGroupSpec(in, out, s)
	if err != nil {
		return err
	}

	out.Subnets = in.Zones
	out.Zones = nil // Those zones are not the same as v1alpha1 zones

	return nil
}

func Convert_kops_InstanceGroupSpec_To_v1alpha1_InstanceGroupSpec(in *kops.InstanceGroupSpec, out *InstanceGroupSpec, s conversion.Scope) error {
	err := autoConvert_kops_InstanceGroupSpec_To_v1alpha1_InstanceGroupSpec(in, out, s)
	if err != nil {
		return err
	}

	out.Zones = in.Subnets

	return nil
}

func Convert_v1alpha1_TopologySpec_To_kops_TopologySpec(in *TopologySpec, out *kops.TopologySpec, s conversion.Scope) error {
	out.Masters = in.Masters
	out.Nodes = in.Nodes
	if in.Bastion != nil && in.Bastion.Enable {
		out.Bastion = new(kops.BastionSpec)
		if err := Convert_v1alpha1_BastionSpec_To_kops_BastionSpec(in.Bastion, out.Bastion, s); err != nil {
			return err
		}
	} else {
		out.Bastion = nil
	}
	if in.DNS != nil {
		out.DNS = new(kops.DNSSpec)
		if err := Convert_v1alpha1_DNSSpec_To_kops_DNSSpec(in.DNS, out.DNS, s); err != nil {
			return err
		}
	} else {
		out.DNS = nil
	}
	return nil
}

func Convert_kops_TopologySpec_To_v1alpha1_TopologySpec(in *kops.TopologySpec, out *TopologySpec, s conversion.Scope) error {
	out.Masters = in.Masters
	out.Nodes = in.Nodes
	if in.Bastion != nil {
		out.Bastion = new(BastionSpec)
		if err := Convert_kops_BastionSpec_To_v1alpha1_BastionSpec(in.Bastion, out.Bastion, s); err != nil {
			return err
		}
	} else {
		out.Bastion = nil
	}
	if in.DNS != nil {
		out.DNS = new(DNSSpec)
		if err := Convert_kops_DNSSpec_To_v1alpha1_DNSSpec(in.DNS, out.DNS, s); err != nil {
			return err
		}
	} else {
		out.DNS = nil
	}
	return nil
}
