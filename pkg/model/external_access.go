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
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"strconv"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, KubernetesAPIAccess)
type ExternalAccessModelBuilder struct {
	*KopsModelContext
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
	if b.Cluster.IsTopologyPublic() {
		for i, sshAccess := range b.Cluster.Spec.SSHAccess {
			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          s("ssh-external-to-master-" + strconv.Itoa(i)),
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				Protocol:      s("tcp"),
				FromPort:      i64(22),
				ToPort:        i64(22),
				CIDR:          s(sshAccess),
			})

			c.AddTask(&awstasks.SecurityGroupRule{
				Name:          s("ssh-external-to-node-" + strconv.Itoa(i)),
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleNode),
				Protocol:      s("tcp"),
				FromPort:      i64(22),
				ToPort:        i64(22),
				CIDR:          s(sshAccess),
			})
		}

		// Configuration for the master, when not using a Loadbalancer (ELB)
		// We expect that either the IP address is published, or DNS is set up to point to the IPs
		// We need to open security groups directly to the master nodes (instead of via the ELB)

		// HTTPS to the master is allowed (for API access)
		for i, apiAccess := range b.Cluster.Spec.KubernetesAPIAccess {
			t := &awstasks.SecurityGroupRule{
				Name:          s("https-external-to-master-" + strconv.Itoa(i)),
				SecurityGroup: b.LinkToSecurityGroup(kops.InstanceGroupRoleMaster),
				Protocol:      s("tcp"),
				FromPort:      i64(443),
				ToPort:        i64(443),
				CIDR:          s(apiAccess),
			}
			c.AddTask(t)
		}
	}

	return nil
}
