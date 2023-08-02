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

package openstackmodel

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
	"k8s.io/utils/net"
)

// ServerGroupModelBuilder configures server group objects
type ServerGroupModelBuilder struct {
	*OpenstackModelContext
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &ServerGroupModelBuilder{}

// See https://specs.openstack.org/openstack/nova-specs/specs/newton/approved/lowercase-metadata-keys.html for details
var instanceMetadataNotAllowedCharacters = regexp.MustCompile("[^a-zA-Z0-9-_:. ]")

// Constants for truncating Tags
const MAX_TAG_LENGTH_OPENSTACK = 60

var TRUNCATE_OPT = truncate.TruncateStringOptions{
	MaxLength:     MAX_TAG_LENGTH_OPENSTACK,
	AlwaysAddHash: false,
	HashLength:    6,
}

func (b *ServerGroupModelBuilder) buildAllowedAddressPairs(annotations map[string]string) []ports.AddressPair {
	keyPrefix := openstack.OS_ANNOTATION + openstack.ALLOWED_ADDRESS_PAIR + "/"

	var allowedAddressPairs []ports.AddressPair
	for key := range annotations {
		if strings.HasPrefix(key, keyPrefix) {
			ipAddress, macAddress, _ := strings.Cut(annotations[key], ",")

			allowedAddressPair := ports.AddressPair{
				IPAddress: ipAddress,
			}
			if macAddress != "" {
				allowedAddressPair.MACAddress = macAddress
			}

			allowedAddressPairs = append(allowedAddressPairs, allowedAddressPair)
		}
	}

	sort.Slice(allowedAddressPairs, func(i, j int) bool {
		return allowedAddressPairs[i].IPAddress < allowedAddressPairs[j].IPAddress
	})

	return allowedAddressPairs
}

func (b *ServerGroupModelBuilder) buildInstances(c *fi.CloudupModelBuilderContext, sg *openstacktasks.ServerGroup, ig *kops.InstanceGroup) error {
	sshKeyNameFull, err := b.SSHKeyName()
	if err != nil {
		return err
	}

	sshKeyName := strings.Replace(sshKeyNameFull, ":", "_", -1)

	igMeta := make(map[string]string)
	cloudTags, err := b.KopsModelContext.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return fmt.Errorf("could not get cloud tags for instance group %s: %v", ig.Name, err)
	}
	for label, labelVal := range cloudTags {
		sanitizedLabel := strings.ToLower(
			instanceMetadataNotAllowedCharacters.ReplaceAllLiteralString(label, "_"),
		)
		igMeta[sanitizedLabel] = labelVal
	}
	if ig.Spec.Role != kops.InstanceGroupRoleBastion {
		// Bastion does not belong to the cluster and will not be running protokube.

		igMeta[openstack.TagClusterName] = b.ClusterName()
	}
	igMeta["k8s"] = b.ClusterName()
	netName, err := b.GetNetworkName()
	if err != nil {
		return err
	}
	igMeta[openstack.TagKopsNetwork] = netName
	igMeta[openstack.TagKopsInstanceGroup] = ig.Name
	igMeta[openstack.TagKopsRole] = string(ig.Spec.Role)
	igMeta[openstack.INSTANCE_GROUP_GENERATION] = fmt.Sprintf("%d", ig.GetGeneration())
	igMeta[openstack.CLUSTER_GENERATION] = fmt.Sprintf("%d", b.Cluster.GetGeneration())

	if e, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_FROM_VOLUME]; ok {
		igMeta[openstack.BOOT_FROM_VOLUME] = e
	}

	if v, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_VOLUME_SIZE]; ok {
		igMeta[openstack.BOOT_VOLUME_SIZE] = v
	}

	startupScript, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
	if err != nil {
		return fmt.Errorf("could not create startup script for instance group %s: %v", ig.Name, err)
	}

	var securityGroups []*openstacktasks.SecurityGroup
	securityGroupName := b.SecurityGroupName(ig.Spec.Role)
	securityGroups = append(securityGroups, b.LinkToSecurityGroup(securityGroupName))

	if b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer == nil && ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
		securityGroups = append(securityGroups, b.LinkToSecurityGroup(b.APIResourceName()))
	}

	r := strings.NewReplacer("_", "-", ".", "-")
	groupName := r.Replace(strings.ToLower(ig.Name))
	// In the future, OpenStack will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of servers
	for i := int32(0); i < *ig.Spec.MinSize; i++ {
		// FIXME: Must ensure 63 or less characters
		// replace all dots and _ with -, this is needed to get external cloudprovider working
		iName := strings.Replace(strings.ToLower(fmt.Sprintf("%s-%d.%s", ig.Name, i+1, b.ClusterName())), "_", "-", -1)
		instanceName := fi.PtrTo(strings.Replace(iName, ".", "-", -1))

		var az *string
		var subnets []*openstacktasks.Subnet
		havePublicSubnet := false
		if len(ig.Spec.Subnets) > 0 {
			subnet := ig.Spec.Subnets[int(i)%len(ig.Spec.Subnets)]
			// bastion subnet name might contain a "utility-" prefix
			if ig.Spec.Role == kops.InstanceGroupRoleBastion {
				az = fi.PtrTo(strings.Replace(subnet, "utility-", "", 1))
			} else {
				az = fi.PtrTo(subnet)
			}

			subnetName, subnetType, err := b.findSubnetClusterSpec(subnet)
			if err != nil {
				return err
			}
			subnets = append(subnets, b.LinkToSubnet(s(subnetName)))
			if subnetType == kops.SubnetTypePublic || subnetType == kops.SubnetTypeUtility {
				havePublicSubnet = true
			}
		}
		if len(ig.Spec.Zones) > 0 {
			zone := ig.Spec.Zones[int(i)%len(ig.Spec.Zones)]
			az = fi.PtrTo(zone)
		}
		// Create instance port task
		portName := fmt.Sprintf("%s-%s", "port", *instanceName)
		portTagKopsName := strings.Replace(
			strings.Replace(
				strings.ToLower(
					fmt.Sprintf("port-%s-%d", ig.Name, i+1),
				),
				"_", "-", -1,
			), ".", "-", -1,
		)
		portTask := &openstacktasks.Port{
			Name:              fi.PtrTo(portName),
			InstanceGroupName: &groupName,
			Network:           b.LinkToNetwork(),
			Tags: []string{
				truncate.TruncateString(fmt.Sprintf("%s=%s", openstack.TagKopsInstanceGroup, groupName), TRUNCATE_OPT),
				truncate.TruncateString(fmt.Sprintf("%s=%s", openstack.TagKopsName, portTagKopsName), TRUNCATE_OPT),
				truncate.TruncateString(fmt.Sprintf("%s=%s", openstack.TagClusterName, b.ClusterName()), TRUNCATE_OPT),
			},
			SecurityGroups:           securityGroups,
			AdditionalSecurityGroups: ig.Spec.AdditionalSecurityGroups,
			Subnets:                  subnets,
			AllowedAddressPairs:      b.buildAllowedAddressPairs(ig.ObjectMeta.Annotations),
			Lifecycle:                b.Lifecycle,
		}
		c.AddTask(portTask)

		if b.Cluster.UsesNoneDNS() && ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
			portTask.ForAPIServer = true
		}

		metaWithName := make(map[string]string)
		for k, v := range igMeta {
			metaWithName[k] = v
		}
		metaWithName[openstack.TagKopsName] = fi.ValueOf(instanceName)
		instanceTask := &openstacktasks.Instance{
			Name:             instanceName,
			Lifecycle:        b.Lifecycle,
			GroupName:        s(groupName),
			Region:           fi.PtrTo(b.Cluster.Spec.Networking.Subnets[0].Region),
			Flavor:           fi.PtrTo(ig.Spec.MachineType),
			Image:            fi.PtrTo(ig.Spec.Image),
			SSHKey:           fi.PtrTo(sshKeyName),
			ServerGroup:      sg,
			Role:             fi.PtrTo(string(ig.Spec.Role)),
			Port:             portTask,
			UserData:         startupScript,
			Metadata:         metaWithName,
			SecurityGroups:   ig.Spec.AdditionalSecurityGroups,
			AvailabilityZone: az,
			ConfigDrive:      b.Cluster.Spec.CloudProvider.Openstack.Metadata.ConfigDrive,
		}
		c.AddTask(instanceTask)

		// Associate a floating IP to the instances if we have external network in router
		// and respective subnet is "Public" or "Utility".
		if b.Cluster.Spec.CloudProvider.Openstack.Router != nil {
			if ig.Spec.AssociatePublicIP != nil && !fi.ValueOf(ig.Spec.AssociatePublicIP) {
				continue
			}
			if havePublicSubnet || ig.Spec.Role == kops.InstanceGroupRoleBastion {
				t := &openstacktasks.FloatingIP{
					Name:      fi.PtrTo(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
					Lifecycle: b.Lifecycle,
				}
				c.AddTask(t)
				if ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
					b.associateFIPToKeypair(t)
				}
				instanceTask.FloatingIP = t
			}
		}
	}

	return nil
}

func (b *ServerGroupModelBuilder) associateFIPToKeypair(fipTask *openstacktasks.FloatingIP) {
	// Ensure the floating IP is included in the TLS certificate,
	// if we're not going to use an alias for it
	fipTask.ForAPIServer = true
}

func (b *ServerGroupModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	clusterName := b.ClusterName()

	sgs := make(map[string]*openstacktasks.ServerGroup)
	for _, ig := range b.InstanceGroups {
		klog.V(2).Infof("Found instance group with name %s and role %v.", ig.Name, ig.Spec.Role)
		affinityPolicies := []string{}
		if v, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.SERVER_GROUP_AFFINITY]; ok {
			affinityPolicies = append(affinityPolicies, v)
		} else {
			affinityPolicies = append(affinityPolicies, "anti-affinity")
		}

		sgName := fmt.Sprintf("%s-%s", clusterName, ig.Name)
		if name, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.SERVER_GROUP_NAME]; ok {
			sgName = fmt.Sprintf("%s-%s", clusterName, name)
		}

		sgTask, ok := sgs[sgName]
		if !ok {
			igMap := make(map[string]*int32)
			igMap[ig.Name] = ig.Spec.MaxSize
			sgTask = &openstacktasks.ServerGroup{
				Name:        s(sgName),
				ClusterName: s(clusterName),
				IGMap:       igMap,
				Policies:    affinityPolicies,
				Lifecycle:   b.Lifecycle,
			}
			sgs[sgName] = sgTask
		} else {
			sgTask.IGMap[ig.Name] = ig.Spec.MaxSize
		}

		err := b.buildInstances(c, sgTask, ig)
		if err != nil {
			return err
		}
	}

	for _, s := range sgs {
		c.AddTask(s)
	}

	if b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer != nil {
		var lbSubnetName string
		var err error
		for _, sp := range b.Cluster.Spec.Networking.Subnets {
			if sp.Type == kops.SubnetTypeDualStack || sp.Type == kops.SubnetTypePrivate {
				lbSubnetName, err = b.findSubnetNameByID(sp.ID, sp.Name)
				if err != nil {
					return err
				}
				break
			}
		}
		if lbSubnetName == "" {
			return fmt.Errorf("could not find subnet for Kubernetes API loadbalancer")
		}

		lbTask := &openstacktasks.LB{
			Name:      fi.PtrTo(b.APIResourceName()),
			Subnet:    fi.PtrTo(lbSubnetName),
			Lifecycle: b.Lifecycle,
		}

		if b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer.FlavorID != nil {
			lbTask.FlavorID = b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer.FlavorID
		}

		useVIPACL := b.UseVIPACL()
		if !useVIPACL {
			lbTask.SecurityGroup = b.LinkToSecurityGroup(b.APIResourceName())
		}

		c.AddTask(lbTask)

		lbfipTask := &openstacktasks.FloatingIP{
			Name:      fi.PtrTo(fmt.Sprintf("%s-%s", "fip", *lbTask.Name)),
			LB:        lbTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(lbfipTask)

		if b.Cluster.UsesLegacyGossip() || b.Cluster.UsesPrivateDNS() || b.Cluster.UsesNoneDNS() {
			b.associateFIPToKeypair(lbfipTask)
		}

		poolTask := &openstacktasks.LBPool{
			Name:         fi.PtrTo(fmt.Sprintf("%s-https", fi.ValueOf(lbTask.Name))),
			Loadbalancer: lbTask,
			Lifecycle:    b.Lifecycle,
		}
		c.AddTask(poolTask)

		nameForResource := fi.ValueOf(lbTask.Name)
		listenerTask := &openstacktasks.LBListener{
			Name:      fi.PtrTo(nameForResource),
			Port:      fi.PtrTo(wellknownports.KubeAPIServer),
			Lifecycle: b.Lifecycle,
			Pool:      poolTask,
		}
		if useVIPACL {
			var AllowedCIDRs []string
			// currently kOps openstack supports only ipv4 addresses
			for _, CIDR := range b.Cluster.Spec.API.Access {
				if net.IsIPv4CIDRString(CIDR) {
					AllowedCIDRs = append(AllowedCIDRs, CIDR)
				}
			}
			sort.Strings(AllowedCIDRs)
			listenerTask.AllowedCIDRs = AllowedCIDRs
		}
		c.AddTask(listenerTask)

		monitorTask := &openstacktasks.PoolMonitor{
			Name:      fi.PtrTo(nameForResource),
			Pool:      poolTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(monitorTask)

		ifName, err := b.GetNetworkName()
		if err != nil {
			return err
		}

		for _, ig := range b.InstanceGroups {
			if ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
				associateTask := &openstacktasks.PoolAssociation{
					Name:          fi.PtrTo(fmt.Sprintf("%s-%s", clusterName, ig.Name)),
					ServerPrefix:  fi.PtrTo(ig.Name),
					Pool:          poolTask,
					InterfaceName: fi.PtrTo(ifName),
					ProtocolPort:  fi.PtrTo(wellknownports.KubeAPIServer),
					Lifecycle:     b.Lifecycle,
					Weight:        fi.PtrTo(1),
				}
				c.AddTask(associateTask)
			}
		}

	}

	return nil
}
