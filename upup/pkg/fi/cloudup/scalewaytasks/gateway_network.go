/*
Copyright 2024 The Kubernetes Authors.

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

package scalewaytasks

import (
	"fmt"
	"strings"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type GatewayNetwork struct {
	ID   *string
	Name *string
	Zone *string

	//Address *string
	//IsForAPIServer bool

	Lifecycle      fi.Lifecycle
	Gateway        *Gateway
	PrivateNetwork *PrivateNetwork
}

//func (g *GatewayNetwork) IsForAPIServer() bool {
//	return g.
//}

//func (g *GatewayNetwork) FindAddresses(context *fi.CloudupContext) ([]string, error) {
//	//TODO implement me
//	panic("implement me")
//}

var _ fi.CloudupTask = &GatewayNetwork{}
var _ fi.CompareWithID = &GatewayNetwork{}
var _ fi.CloudupHasDependencies = &GatewayNetwork{}

//var _ fi.HasAddress = &GatewayNetwork{}

func (g *GatewayNetwork) CompareWithID() *string {
	return g.ID
}

func (g *GatewayNetwork) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*Gateway); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (g *GatewayNetwork) Find(context *fi.CloudupContext) (*GatewayNetwork, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	gwns, err := cloud.GatewayService().ListGatewayNetworks(&vpcgw.ListGatewayNetworksRequest{
		Zone:             scw.Zone(cloud.Zone()),
		GatewayID:        g.Gateway.ID,
		PrivateNetworkID: g.PrivateNetwork.ID,
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing gateway networks: %w", err)
	}

	if gwns.TotalCount == 0 {
		return nil, nil
	}
	if gwns.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 gateway network, got %d", gwns.TotalCount)
	}
	gwnFound := gwns.GatewayNetworks[0]

	return &GatewayNetwork{
		ID:   fi.PtrTo(gwnFound.ID),
		Zone: fi.PtrTo(gwnFound.Zone.String()),
		//Address:        fi.PtrTo(gwnFound.Address.IP.String()),
		Lifecycle:      g.Lifecycle,
		Gateway:        g.Gateway,
		PrivateNetwork: g.PrivateNetwork,
	}, nil
}

func (g *GatewayNetwork) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(g, context)
}

func (_ *GatewayNetwork) CheckChanges(actual, expected, changes *GatewayNetwork) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *GatewayNetwork) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *GatewayNetwork) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))

	gwnCreated, err := cloud.GatewayService().CreateGatewayNetwork(&vpcgw.CreateGatewayNetworkRequest{
		Zone:             zone,
		GatewayID:        fi.ValueOf(expected.Gateway.ID),
		PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
		IpamConfig: &vpcgw.CreateGatewayNetworkRequestIpamConfig{
			PushDefaultRoute: false,
		},
		EnableMasquerade: true,
	})
	if err != nil {
		return fmt.Errorf("creating gateway network: %w", err)
	}

	_, err = cloud.GatewayService().WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gwnCreated.ID,
		Zone:             zone,
	})
	if err != nil {
		return fmt.Errorf("waiting for gateway: %v", err)
	}

	expected.ID = &gwnCreated.ID

	return nil
}

type gwnIpamConfig struct {
	PushDefaultRoute bool `cty:"push_default_route"`
}

type terraformGatewayNetwork struct {
	GatewayID        *terraformWriter.Literal `cty:"gateway_id"`
	PrivateNetworkID *terraformWriter.Literal `cty:"private_network_id"`
	EnableMasquerade bool                     `cty:"enable_masquerade"`
	EnableDHCP       bool                     `cty:"enable_dhcp"`
	IpamConfig       *gwnIpamConfig           `cty:"ipam_config"`
}

func (_ *GatewayNetwork) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *GatewayNetwork) error {
	tfName := strings.ReplaceAll(fi.ValueOf(expected.Name), ".", "-")

	tfGWN := terraformGatewayNetwork{
		GatewayID:        expected.Gateway.TerraformLink(),
		PrivateNetworkID: expected.PrivateNetwork.TerraformLink(),
		IpamConfig: &gwnIpamConfig{
			PushDefaultRoute: false,
		},
	}

	return t.RenderResource("scaleway_vpc_gateway_network", tfName, tfGWN)
}
