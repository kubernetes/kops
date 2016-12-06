package model

import (
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"strconv"
)

// ExternalAccessModelBuilder configures security group rules for external access
// (SSHAccess, APIAccess)
type ExternalAccessModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &ExternalAccessModelBuilder{}

func (b *ExternalAccessModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if len(b.Cluster.Spec.APIAccess) == 0 {
		glog.Warningf("APIAccess is empty")
	}

	if len(b.Cluster.Spec.SSHAccess) == 0 {
		glog.Warningf("SSHAccess is empty")

		//// AdminCIDR returns the CIDRs that are allowed to access the admin ports of the cluster
		//// (22, 443 on master and 22 on nodes)
		//func (tf *TemplateFunctions) AdminCIDR() []string {
		//	if len(tf.cluster.Spec.AdminAccess) == 0 {
		//		return []string{"0.0.0.0/0"}
		//	}
		//	return tf.cluster.Spec.AdminAccess
		//}
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
		for i, apiAccess := range b.Cluster.Spec.APIAccess {
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

	//upup/models/cloudup/_aws/master/_master_lb/master_lb.yaml:
	//# HTTPS to the master ELB is allowed (for API access)
	//# One security group rule is necessary per admin CIDR
	//{{ range $index, $cidr := AdminCIDR }}
	//securityGroupRule/https-external-to-api-{{ $index }}:
	//securityGroup: securityGroup/api.{{ ClusterName }}
	//cidr: {{ $cidr }}
	//protocol: tcp
	//fromPort: 443
	//toPort: 443
	//{{ end }}

	return nil
}
