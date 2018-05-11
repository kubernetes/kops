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

package aws

import (
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, KubernetesAPIAccess)
type ExternalAccessModelBuilder struct {
	*ModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ExternalAccessModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if len(b.Cluster.Spec.KubernetesAPIAccess) == 0 {
		glog.Warningf("KubernetesAPIAccess is empty")
	}

	if len(b.Cluster.Spec.SSHAccess) == 0 {
		glog.Warningf("SSHAccess is empty")
	}

	// SSH is open to AdminCIDR set
	if b.UsesSSHBastion() {
		// If we are using a bastion, we only access through the bastion
		// This is admittedly a little odd... adding a bastion shuts down direct access to the masters/nodes
		// But I think we can always add more permissions in this case later, but we can't easily take them away
		glog.V(2).Infof("bastion is in use; won't configure SSH access to master / node instances")
	} else {
		for _, sshAccess := range b.Cluster.Spec.SSHAccess {
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          fi.String("ssh-external-to-master-" + sshAccess),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				Protocol:      fi.String("tcp"),
				FromPort:      fi.Int64(22),
				ToPort:        fi.Int64(22),
				CIDR:          fi.String(sshAccess),
			})

			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          fi.String("ssh-external-to-node-" + sshAccess),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
				Protocol:      fi.String("tcp"),
				FromPort:      fi.Int64(22),
				ToPort:        fi.Int64(22),
				CIDR:          fi.String(sshAccess),
			})
		}
	}

	if !b.UseLoadBalancerForAPI() {
		// Configuration for the master, when not using a Loadbalancer (ELB)
		// We expect that either the IP address is published, or DNS is set up to point to the IPs
		// We need to open security groups directly to the master nodes (instead of via the ELB)

		// HTTPS to the master is allowed (for API access)
		for _, apiAccess := range b.Cluster.Spec.KubernetesAPIAccess {
			t := &awstasks.SecurityGroupRule{
				Name:          fi.String("https-external-to-master-" + apiAccess),
				Lifecycle:     b.Lifecycle,
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				Protocol:      fi.String("tcp"),
				FromPort:      fi.Int64(443),
				ToPort:        fi.Int64(443),
				CIDR:          fi.String(apiAccess),
			}
			c.AddTask(t)
		}
	}

	return nil
}
