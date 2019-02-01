/*
Copyright 2018 The Kubernetes Authors.

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

	"github.com/golang/glog"
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

	startupScript, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
	if err != nil {
		return fmt.Errorf("Could not create startup script for instance group %s: %v", ig.Name, err)
	}
	if startupScript != nil {
		// var userData bytes.Buffer
		startupStr, err := startupScript.AsString()
		if err != nil {
			return fmt.Errorf("Could not create startup script for instance group %s: %v", ig.Name, err)
		}
		igUserData = fi.String(startupStr)
	}

	// In the future, OpenStack will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of servers
	for i := int32(0); i < *ig.Spec.MinSize; i++ {
		if err != nil {
			return fmt.Errorf("Failed to create UUID for instance: %v", err)
		}
		// FIXME: Must ensure 63 or less characters
		instanceName := fi.String(
			strings.ToLower(
				fmt.Sprintf("%s-%d", *sg.Name, i+1),
			),
		)
		securityGroupName := b.SecurityGroupName(ig.Spec.Role)
		securityGroup := b.LinkToSecurityGroup(securityGroupName)

		// Create instance port task
		portTask := &openstacktasks.Port{
			Name:           fi.String(fmt.Sprintf("%s-%s", "port", *instanceName)),
			Network:        b.LinkToNetwork(),
			SecurityGroups: append([]*openstacktasks.SecurityGroup{}, securityGroup),
			Lifecycle:      b.Lifecycle,
		}
		c.AddTask(portTask)

		instanceTask := &openstacktasks.Instance{
			Name:        instanceName,
			Region:      fi.String(b.Cluster.Spec.Subnets[0].Region),
			Flavor:      fi.String(ig.Spec.MachineType),
			Image:       fi.String(ig.Spec.Image),
			SSHKey:      fi.String(sshKeyName),
			ServerGroup: sg,
			Tags:        []string{clusterTag},
			Role:        fi.String(string(ig.Spec.Role)),
			Port:        portTask,
			Metadata:    igMeta,
		}
		if igUserData != nil {
			instanceTask.UserData = igUserData
		}
		c.AddTask(instanceTask)

		// Associate a floating IP to the master and bastion always, associate it to a node if bastion is not used
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleBastion:
			t := &openstacktasks.FloatingIP{
				Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
				Server:    instanceTask,
				Lifecycle: b.Lifecycle,
			}
			c.AddTask(t)
		case kops.InstanceGroupRoleMaster:
			if !b.UseLoadBalancerForAPI() {
				t := &openstacktasks.FloatingIP{
					Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
					Server:    instanceTask,
					Lifecycle: b.Lifecycle,
				}
				c.AddTask(t)
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
	}

	return nil
}

func (b *ServerGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()

	var masters []*openstacktasks.ServerGroup
	for _, ig := range b.InstanceGroups {
		glog.V(2).Infof("Found instance group with name %s and role %v.", ig.Name, ig.Spec.Role)
		sgTask := &openstacktasks.ServerGroup{
			Name:      s(fmt.Sprintf("%s-%s", clusterName, ig.Name)),
			Policies:  []string{"anti-affinity"},
			Lifecycle: b.Lifecycle,
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

	if b.UseLoadBalancerForAPI() {
		lbSubnetName := b.MasterInstanceGroups()[0].Spec.Subnets[0]
		lbTask := &openstacktasks.LB{
			Name:      fi.String(b.Cluster.Spec.MasterPublicName),
			Subnet:    fi.String(lbSubnetName + "." + b.ClusterName()),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(lbTask)

		lbfipTask := &openstacktasks.FloatingIP{
			Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *lbTask.Name)),
			LB:        lbTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(lbfipTask)

		if dns.IsGossipHostname(b.Cluster.Name) || b.UsePrivateDNS() {
			// Ensure the floating IP is included in the TLS certificate,
			// if we're not going to use an alias for it
			// TODO: I don't love this technique for finding the task by name & modifying it
			masterKeypairTask, found := c.Tasks["Keypair/master"]
			if !found {
				return fmt.Errorf("keypair/master task not found")
			}
			masterKeypair := masterKeypairTask.(*fitasks.Keypair)
			masterKeypair.AlternateNameTasks = append(masterKeypair.AlternateNameTasks, lbfipTask)
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
		c.AddTask(listenerTask)

		for _, mastersg := range masters {
			associateTask := &openstacktasks.PoolAssociation{
				Name:          mastersg.Name,
				Pool:          poolTask,
				ServerGroup:   mastersg,
				InterfaceName: fi.String(clusterName),
				ProtocolPort:  fi.Int(443),
				Lifecycle:     b.Lifecycle,
			}

			c.AddTask(associateTask)
		}

	}

	return nil
}
