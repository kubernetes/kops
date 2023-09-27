package scalewaymodel

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/scw"
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
	clusterNameTag := scaleway.TagClusterName + "=" + b.ClusterName()
	resourceName := b.ClusterName()
	zone := scw.Zone(b.Cluster.Spec.Networking.Subnets[0].Zone)
	region, err := zone.Region()
	if err != nil {
		return fmt.Errorf("building network task: %w", err)
	}

	vpc := &scalewaytasks.VPC{
		Name:      fi.PtrTo(resourceName),
		Region:    fi.PtrTo(region.String()),
		Tags:      []string{clusterNameTag},
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(vpc)

	gateway := &scalewaytasks.Gateway{
		Name:      fi.PtrTo(resourceName),
		Zone:      fi.PtrTo(zone.String()),
		Tags:      []string{clusterNameTag},
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(gateway)

	ipRange := b.Cluster.Spec.Networking.NetworkCIDR
	if ipRange == "" {
		ipRange = "192.168.1.0/24"
	}

	privateNetwork := &scalewaytasks.PrivateNetwork{
		Name:      fi.PtrTo(resourceName),
		Region:    fi.PtrTo(region.String()),
		Tags:      []string{clusterNameTag},
		Lifecycle: b.Lifecycle,
		IPRange:   fi.PtrTo(ipRange),
		VPC:       vpc,
	}
	c.AddTask(privateNetwork)

	dhcpConfig := &scalewaytasks.DHCPConfig{
		Name:           fi.PtrTo(resourceName),
		Zone:           fi.PtrTo(zone.String()),
		Subnet:         fi.PtrTo(ipRange),
		Lifecycle:      b.Lifecycle,
		Gateway:        gateway,
		PrivateNetwork: privateNetwork,
	}
	c.AddTask(dhcpConfig)

	gatewayNetwork := &scalewaytasks.GatewayNetwork{
		Name:           fi.PtrTo(resourceName),
		Zone:           fi.PtrTo(zone.String()),
		Lifecycle:      b.Lifecycle,
		DHCPConfig:     dhcpConfig,
		Gateway:        gateway,
		PrivateNetwork: privateNetwork,
	}
	c.AddTask(gatewayNetwork)

	return nil
}
