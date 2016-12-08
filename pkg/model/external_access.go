package model

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
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
	}

	////// AdminCIDR returns the CIDRs that are allowed to access the admin ports of the cluster
	////// (22, 443 on master and 22 on nodes)
	////func (tf *TemplateFunctions) AdminCIDR() []string {
	////	if len(tf.cluster.Spec.AdminAccess) == 0 {
	////		return []string{"0.0.0.0/0"}
	////	}
	////	return tf.cluster.Spec.AdminAccess
	////}
	//
	//upup/models/cloudup/_aws/master/master.yaml:#
	//# SSH is open to AdminCIDR set
	//{{ if IsTopologyPublic }}
	//{{ range $index, $cidr := AdminCIDR }}
	//securityGroupRule/ssh-external-to-master-{{ $index }}:
	//securityGroup: securityGroup/masters.{{ ClusterName }}
	//cidr: {{ $cidr }}
	//protocol: tcp
	//fromPort: 22
	//toPort: 22
	//{{ end }}
	//{{ end }}
	//
	//upup/models/cloudup/_aws/master/_master_lb/master_lb.yaml:
	//
	//
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
	//
	//// upup/models/cloudup/_aws/master/_not_master_lb/not_master_lb.yaml:
	//# Configuration for the master, when not using a Loadbalancer (ELB)
	//# We expect that either the IP address is published, or DNS is set up to point to the IPs
	//# We need to open security groups directly to the master nodes (instead of via the ELB)
	//
	//# HTTPS to the master is allowed (for API access)
	//{{ range $index, $cidr := AdminCIDR }}
	//securityGroupRule/https-external-to-master-{{ $index }}:
	//securityGroup: securityGroup/masters.{{ ClusterName }}
	//cidr: {{ $cidr }}
	//protocol: tcp
	//fromPort: 443
	//toPort: 443
	//{{ end }}
	//
	//
	//// upup/models/cloudup/_aws/topologies/_topology_public/nodes.yaml:
	//
	//# SSH is open to CIDRs defined in the cluster configuration
	//{{ range $index, $cidr := AdminCIDR }}
	//securityGroupRule/ssh-external-to-node-{{ $index }}:
	//securityGroup: securityGroup/nodes.{{ ClusterName }}
	//cidr: {{ $cidr }}
	//protocol: tcp
	//fromPort: 22
	//toPort: 22
	//{{ end }}

	return fmt.Errorf("externalaccess.go NOT IMPLEMENTED")
}
