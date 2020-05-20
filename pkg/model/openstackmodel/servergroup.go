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
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// ServerGroupModelBuilder configures server group objects
type ServerGroupModelBuilder struct {
	*OpenstackModelContext
	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &ServerGroupModelBuilder{}

func (b *ServerGroupModelBuilder) buildInstances(c *fi.ModelBuilderContext, sg *openstacktasks.ServerGroup, ig *kops.InstanceGroup) error {

	sshKeyNameFull, err := b.SSHKeyName()
	if err != nil {
		return err
	}

	sshKeyName := strings.Replace(sshKeyNameFull, ":", "_", -1)

	clusterTag := "KubernetesCluster:" + strings.Replace(b.ClusterName(), ".", "-", -1)

	var igUserData *string
	igMeta := make(map[string]string)

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
	igMeta["KopsInstanceGroup"] = ig.Name
	igMeta["KopsRole"] = string(ig.Spec.Role)
	igMeta[openstack.INSTANCE_GROUP_GENERATION] = fmt.Sprintf("%d", ig.GetGeneration())
	igMeta[openstack.CLUSTER_GENERATION] = fmt.Sprintf("%d", b.Cluster.GetGeneration())

	if e, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_FROM_VOLUME]; ok {
		igMeta[openstack.BOOT_FROM_VOLUME] = e
	}

	if v, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_VOLUME_SIZE]; ok {
		igMeta[openstack.BOOT_VOLUME_SIZE] = v
	}

	startupScript, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return fmt.Errorf("could not create startup script for instance group %s: %v", ig.Name, err)
	}
	if startupScript != nil {
		// var userData bytes.Buffer
		startupStr, err := startupScript.AsString()
		if err != nil {
			return fmt.Errorf("could not create startup script for instance group %s: %v", ig.Name, err)
		}
		igUserData = fi.String(startupStr)
	}

	var securityGroups []*openstacktasks.SecurityGroup
	securityGroupName := b.SecurityGroupName(ig.Spec.Role)
	securityGroups = append(securityGroups, b.LinkToSecurityGroup(securityGroupName))

	if b.Cluster.Spec.CloudConfig.Openstack.Loadbalancer == nil && ig.Spec.Role == kops.InstanceGroupRoleMaster {
		securityGroups = append(securityGroups, b.LinkToSecurityGroup(b.Cluster.Spec.MasterPublicName))
	}

	// In the future, OpenStack will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of servers
	for i := int32(0); i < *ig.Spec.MinSize; i++ {
		// FIXME: Must ensure 63 or less characters
		// replace all dots and _ with -, this is needed to get external cloudprovider working
		iName := strings.Replace(strings.ToLower(fmt.Sprintf("%s-%d.%s", ig.Name, i+1, b.ClusterName())), "_", "-", -1)
		instanceName := fi.String(strings.Replace(iName, ".", "-", -1))

		var az *string
		var subnets []*openstacktasks.Subnet
		if len(ig.Spec.Subnets) > 0 {
			subnet := ig.Spec.Subnets[int(i)%len(ig.Spec.Subnets)]
			// bastion subnet name might contain a "utility-" prefix
			if ig.Spec.Role == kops.InstanceGroupRoleBastion {
				az = fi.String(strings.Replace(subnet, "utility-", "", 1))
			} else {
				az = fi.String(subnet)
			}

			subnetName, err := b.findSubnetClusterSpec(subnet)
			if err != nil {
				return err
			}
			subnets = append(subnets, b.LinkToSubnet(s(subnetName)))
		}
		if len(ig.Spec.Zones) > 0 {
			zone := ig.Spec.Zones[int(i)%len(ig.Spec.Zones)]
			az = fi.String(zone)
		}
		// Create instance port task
		portTask := &openstacktasks.Port{
			Name:                     fi.String(fmt.Sprintf("%s-%s", "port", *instanceName)),
			Network:                  b.LinkToNetwork(),
			Tag:                      s(b.ClusterName()),
			SecurityGroups:           securityGroups,
			AdditionalSecurityGroups: ig.Spec.AdditionalSecurityGroups,
			Subnets:                  subnets,
			Lifecycle:                b.Lifecycle,
		}
		c.AddTask(portTask)

		instanceTask := &openstacktasks.Instance{
			Name:             instanceName,
			Region:           fi.String(b.Cluster.Spec.Subnets[0].Region),
			Flavor:           fi.String(ig.Spec.MachineType),
			Image:            fi.String(ig.Spec.Image),
			SSHKey:           fi.String(sshKeyName),
			ServerGroup:      sg,
			Tags:             []string{clusterTag},
			Role:             fi.String(string(ig.Spec.Role)),
			Port:             portTask,
			Metadata:         igMeta,
			SecurityGroups:   ig.Spec.AdditionalSecurityGroups,
			AvailabilityZone: az,
		}
		if igUserData != nil {
			instanceTask.UserData = igUserData
		}
		c.AddTask(instanceTask)

		// Associate a floating IP to the master and bastion always if we have external network in router
		// associate it to a node if bastion is not used
		if b.Cluster.Spec.CloudConfig.Openstack != nil && b.Cluster.Spec.CloudConfig.Openstack.Router != nil {
			if ig.Spec.AssociatePublicIP != nil && !fi.BoolValue(ig.Spec.AssociatePublicIP) {
				if ig.Spec.Role == kops.InstanceGroupRoleMaster {
					b.associateFixedIPToKeypair(c, instanceTask)
				}
				continue
			}
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleBastion:
				t := &openstacktasks.FloatingIP{
					Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
					Server:    instanceTask,
					Lifecycle: b.Lifecycle,
				}
				c.AddTask(t)
			case kops.InstanceGroupRoleMaster:
				if b.Cluster.Spec.CloudConfig.Openstack.Loadbalancer == nil {
					t := &openstacktasks.FloatingIP{
						Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
						Server:    instanceTask,
						Lifecycle: b.Lifecycle,
					}
					c.AddTask(t)
					b.associateFIPToKeypair(c, t)
				}
			default:
				if !b.UsesSSHBastion() {
					t := &openstacktasks.FloatingIP{
						Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
						Server:    instanceTask,
						Lifecycle: b.Lifecycle,
					}
					c.AddTask(t)
				}
			}
		} else if b.Cluster.Spec.CloudConfig.Openstack != nil && b.Cluster.Spec.CloudConfig.Openstack.Router == nil {
			// No external router, but we need to add master fixed ips to certificates
			if ig.Spec.Role == kops.InstanceGroupRoleMaster {
				b.associateFixedIPToKeypair(c, instanceTask)
			}
		}
	}

	return nil
}

func (b *ServerGroupModelBuilder) associateFixedIPToKeypair(c *fi.ModelBuilderContext, fipTask *openstacktasks.Instance) error {
	// Ensure the floating IP is included in the TLS certificate,
	// if we're not going to use an alias for it
	// TODO: I don't love this technique for finding the task by name & modifying it
	masterKeypairTask, found := c.Tasks["Keypair/master"]
	if !found {
		return fmt.Errorf("keypair/master task not found")
	}
	masterKeypair := masterKeypairTask.(*fitasks.Keypair)
	masterKeypair.AlternateNameTasks = append(masterKeypair.AlternateNameTasks, fipTask)
	return nil
}

func (b *ServerGroupModelBuilder) associateFIPToKeypair(c *fi.ModelBuilderContext, fipTask *openstacktasks.FloatingIP) error {
	// Ensure the floating IP is included in the TLS certificate,
	// if we're not going to use an alias for it
	// TODO: I don't love this technique for finding the task by name & modifying it
	masterKeypairTask, found := c.Tasks["Keypair/master"]
	if !found {
		return fmt.Errorf("keypair/master task not found")
	}
	masterKeypair := masterKeypairTask.(*fitasks.Keypair)
	masterKeypair.AlternateNameTasks = append(masterKeypair.AlternateNameTasks, fipTask)
	return nil
}

func (b *ServerGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()

	var masters []*openstacktasks.ServerGroup
	for _, ig := range b.InstanceGroups {
		klog.V(2).Infof("Found instance group with name %s and role %v.", ig.Name, ig.Spec.Role)
		sgTask := &openstacktasks.ServerGroup{
			Name:        s(fmt.Sprintf("%s-%s", clusterName, ig.Name)),
			ClusterName: s(clusterName),
			IGName:      s(ig.Name),
			Policies:    []string{"anti-affinity"},
			Lifecycle:   b.Lifecycle,
			MaxSize:     ig.Spec.MaxSize,
		}
		c.AddTask(sgTask)

		err := b.buildInstances(c, sgTask, ig)
		if err != nil {
			return err
		}

		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			masters = append(masters, sgTask)
		}
	}

	if b.Cluster.Spec.CloudConfig.Openstack.Loadbalancer != nil {
		var lbSubnetName string
		var err error
		for _, sp := range b.Cluster.Spec.Subnets {
			if sp.Type == kops.SubnetTypePrivate {
				lbSubnetName, err = b.findSubnetNameByID(sp.ProviderID, sp.Name)
				if err != nil {
					return err
				}
				break
			}
		}
		if lbSubnetName == "" {
			return fmt.Errorf("could not find subnet for master loadbalancer")
		}
		lbTask := &openstacktasks.LB{
			Name:      fi.String(b.Cluster.Spec.MasterPublicName),
			Subnet:    fi.String(lbSubnetName),
			Lifecycle: b.Lifecycle,
		}

		useVIPACL := b.UseVIPACL()
		if !useVIPACL {
			lbTask.SecurityGroup = b.LinkToSecurityGroup(b.Cluster.Spec.MasterPublicName)
		}

		c.AddTask(lbTask)

		lbfipTask := &openstacktasks.FloatingIP{
			Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *lbTask.Name)),
			LB:        lbTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(lbfipTask)

		if dns.IsGossipHostname(b.Cluster.Name) || b.UsePrivateDNS() {
			b.associateFIPToKeypair(c, lbfipTask)
		}

		poolTask := &openstacktasks.LBPool{
			Name:         fi.String(fmt.Sprintf("%s-https", fi.StringValue(lbTask.Name))),
			Loadbalancer: lbTask,
			Lifecycle:    b.Lifecycle,
		}
		c.AddTask(poolTask)

		listenerTask := &openstacktasks.LBListener{
			Name:      lbTask.Name,
			Lifecycle: b.Lifecycle,
			Pool:      poolTask,
		}
		if useVIPACL {
			listenerTask.AllowedCIDRs = b.Cluster.Spec.KubernetesAPIAccess
		}
		c.AddTask(listenerTask)

		ifName, err := b.GetNetworkName()
		if err != nil {
			return err
		}
		for _, mastersg := range masters {
			associateTask := &openstacktasks.PoolAssociation{
				Name:          mastersg.Name,
				Pool:          poolTask,
				ServerGroup:   mastersg,
				InterfaceName: fi.String(ifName),
				ProtocolPort:  fi.Int(443),
				Lifecycle:     b.Lifecycle,
			}

			c.AddTask(associateTask)
		}

	}

	return nil
}
