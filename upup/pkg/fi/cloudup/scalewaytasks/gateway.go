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

const GatewayDefaultType = "VPC-GW-S"

// +kops:fitask
type Gateway struct {
	ID   *string
	Name *string
	Zone *string
	Tags []string

	Lifecycle fi.Lifecycle
	//PrivateNetwork *PrivateNetwork
}

//func (g *Gateway) IsForAPIServer() bool {
//	return true
//}
//
//func (g *Gateway) FindAddresses(context *fi.CloudupContext) ([]string, error) {
//	if g.ID == nil {
//		return nil, nil
//	}
//
//	cloud := context.T.Cloud.(scaleway.ScwCloud)
//	//gwFound, err := cloud.GetClusterGateways(scaleway.ClusterNameFromTags(g.Tags))
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	var gatewayAddresses []string
//	region, err := scw.Zone(fi.ValueOf(g.Zone)).Region()
//	if err != nil {
//		return nil, fmt.Errorf("finding public gateway's region: %w", err)
//	}
//
//	//for _, gw := range gwFound {
//	ips, err := cloud.IPAMService().ListIPs(&ipam.ListIPsRequest{
//		Region: region,
//		//Zonal:  g.Zone,
//		//ResourceID:       &gw.ID,
//		PrivateNetworkID: g.PrivateNetwork.ID,
//		ResourceName:     g.Name,
//		ResourceType:     ipam.ResourceTypeVpcGateway,
//	}, scw.WithContext(context.Context()), scw.WithAllPages())
//	if err != nil {
//		return nil, fmt.Errorf("listing public gateway's IPs: %w", err)
//	}
//	for _, ip := range ips.IPs {
//		gatewayAddresses = append(gatewayAddresses, ip.Address.IP.String())
//	}
//	//}
//	return gatewayAddresses, nil
//}

var _ fi.CloudupTask = &Gateway{}
var _ fi.CompareWithID = &Gateway{}

//var _ fi.HasAddress = &Gateway{}

func (g *Gateway) CompareWithID() *string {
	return g.ID
}

func (g *Gateway) Find(context *fi.CloudupContext) (*Gateway, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	gateways, err := cloud.GatewayService().ListGateways(&vpcgw.ListGatewaysRequest{
		Zone: scw.Zone(cloud.Zone()),
		Name: g.Name,
		Tags: []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, scaleway.ClusterNameFromTags(g.Tags))},
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing gateways: %w", err)
	}

	if gateways.TotalCount == 0 {
		return nil, nil
	}
	if gateways.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 gateway, got %d", gateways.TotalCount)
	}
	gatewayFound := gateways.Gateways[0]

	return &Gateway{
		ID:        fi.PtrTo(gatewayFound.ID),
		Name:      fi.PtrTo(gatewayFound.Name),
		Zone:      fi.PtrTo(gatewayFound.Zone.String()),
		Tags:      gatewayFound.Tags,
		Lifecycle: g.Lifecycle,
	}, nil
}

func (g *Gateway) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(g, context)
}

func (_ *Gateway) CheckChanges(actual, expected, changes *Gateway) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *Gateway) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *Gateway) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))

	gatewayCreated, err := cloud.GatewayService().CreateGateway(&vpcgw.CreateGatewayRequest{
		Zone:          zone,
		Name:          fi.ValueOf(expected.Name),
		Tags:          expected.Tags,
		Type:          GatewayDefaultType,
		EnableBastion: true,
		BastionPort:   scw.Uint32Ptr(1042),
	})
	if err != nil {
		return fmt.Errorf("creating gateway: %w", err)
	}

	_, err = cloud.GatewayService().WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gatewayCreated.ID,
		Zone:      zone,
	})
	if err != nil {
		return fmt.Errorf("waiting for gateway: %w", err)
	}

	expected.ID = &gatewayCreated.ID

	return nil
}

type terraformGateway struct {
	Type string   `cty:"type"`
	Name *string  `cty:"name"`
	Tags []string `cty:"tags"`
}

func (_ *Gateway) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *Gateway) error {
	tfName := strings.ReplaceAll(fi.ValueOf(expected.Name), ".", "-")

	tfGW := terraformGateway{
		Type: GatewayDefaultType,
		Name: expected.Name,
		Tags: expected.Tags,
	}
	return t.RenderResource("scaleway_vpc_public_gateway", tfName, tfGW)
}

func (g *Gateway) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("scaleway_vpc_public_gateway", fi.ValueOf(g.Name), "id")
}
