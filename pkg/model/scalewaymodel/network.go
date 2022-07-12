package scalewaymodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/scalewaytasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*ScwModelContext
	Lifecycle fi.Lifecycle
}

func (b *NetworkModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {

	ipRange := b.Cluster.Spec.Networking.NetworkCIDR
	if ipRange == "" {
		ipRange = "192.168.1.0/24"
	}

	network := &scalewaytasks.Network{
		Name:      fi.PtrTo(b.ClusterName()),
		Zone:      fi.PtrTo(b.Cluster.Spec.Networking.Subnets[0].Zone),
		Lifecycle: b.Lifecycle,
		IPRange:   fi.PtrTo(ipRange),
		Tags:      []string{scaleway.TagClusterName + "=" + b.ClusterName()},
	}
	c.AddTask(network)

	return nil
}
