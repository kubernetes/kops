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

package awstasks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	raws "k8s.io/kops/pkg/resources/aws"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type NatGateway struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ElasticIP *ElasticIP
	Subnet    *Subnet
	ID        *string

	EgressId *string

	// Shared is set if this is a shared NatGateway
	Shared *bool

	// Tags is a map of aws tags that are added to the NatGateway
	Tags map[string]string

	// We can't tag NatGateways, so we have to find through a surrogate
	AssociatedRouteTable *RouteTable
}

var _ fi.CompareWithID = &NatGateway{}

func (e *NatGateway) CompareWithID() *string {
	// Match by ID (NAT Gateways don't have tags, so they don't have a name in EC2)
	return e.ID
}

func (e *NatGateway) Find(c *fi.CloudupContext) (*NatGateway, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)
	var ngw *ec2types.NatGateway
	actual := &NatGateway{}

	if fi.ValueOf(e.ID) != "" {
		// We have an existing NGW, lets look up the EIP
		ngwIds := []string{fi.ValueOf(e.ID)}

		request := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: ngwIds,
		}

		response, err := cloud.EC2().DescribeNatGateways(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error listing Nat Gateways %v", err)
		}

		if len(response.NatGateways) != 1 {
			return nil, fmt.Errorf("found %d Nat Gateways with ID %q, expected 1", len(response.NatGateways), fi.ValueOf(e.ID))
		}
		ngw = &response.NatGateways[0]

		if len(ngw.NatGatewayAddresses) != 1 {
			return nil, fmt.Errorf("found %d EIP Addresses for 1 NATGateway, expected 1", len(ngw.NatGatewayAddresses))
		}
	} else {
		// This is the normal/default path
		var err error
		ngw, err = e.findNatGateway(c)
		if err != nil {
			return nil, err
		}
		if ngw == nil {
			return nil, nil
		}
	}

	actual.ID = ngw.NatGatewayId

	actual.Subnet = e.Subnet
	if len(ngw.NatGatewayAddresses) == 0 {
		// Not sure if this ever happens
		actual.ElasticIP = nil
	} else if len(ngw.NatGatewayAddresses) == 1 {
		actual.ElasticIP = &ElasticIP{ID: ngw.NatGatewayAddresses[0].AllocationId}
	} else {
		return nil, fmt.Errorf("found multiple elastic IPs attached to NatGateway %q", aws.ToString(ngw.NatGatewayId))
	}

	// NATGateways now have names and tags so lets pull from there instead.
	actual.Name = findNameTag(ngw.Tags)
	if e.Tags["Name"] == "" {
		// If we're not tagging by name, avoid spurious differences
		actual.Name = e.Name
	}
	actual.Tags = intersectTags(ngw.Tags, e.Tags)

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle
	actual.Shared = e.Shared
	actual.AssociatedRouteTable = e.AssociatedRouteTable

	e.ID = actual.ID
	return actual, nil
}

func (e *NatGateway) findNatGateway(c *fi.CloudupContext) (*ec2types.NatGateway, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)

	id := e.ID

	// Find via route on private route table
	if id == nil && e.AssociatedRouteTable != nil {
		ngw, err := findNatGatewayFromRouteTable(ctx, cloud, e.AssociatedRouteTable)
		if err != nil {
			return nil, err
		}
		if ngw != nil {
			return ngw, nil
		}
	}

	// Find via tag on subnet
	// TODO: Obsolete - we can get from the route table instead
	if id == nil && e.Subnet != nil {
		var filters []ec2types.Filter
		filters = append(filters, awsup.NewEC2Filter("key", "AssociatedNatgateway"))
		if e.Subnet.ID == nil {
			klog.V(2).Infof("Unable to find subnet, bypassing Find() for NatGateway")
			return nil, nil
		}
		filters = append(filters, awsup.NewEC2Filter("resource-id", *e.Subnet.ID))

		request := &ec2.DescribeTagsInput{
			Filters: filters,
		}

		response, err := cloud.EC2().DescribeTags(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("error listing tags: %v", err)
		}

		if response == nil || len(response.Tags) == 0 {
			return nil, nil
		}

		if len(response.Tags) != 1 {
			return nil, fmt.Errorf("found multiple tags for: %v", e)
		}
		t := response.Tags[0]
		id = t.Value
		klog.V(2).Infof("Found NatGateway via subnet tag: %v", *id)
	}

	if id != nil {
		return findNatGatewayById(ctx, cloud, fi.ValueOf(id))
	}

	return nil, nil
}

func findNatGatewayById(ctx context.Context, cloud awsup.AWSCloud, id string) (*ec2types.NatGateway, error) {
	request := &ec2.DescribeNatGatewaysInput{}
	request.NatGatewayIds = []string{id}
	response, err := cloud.EC2().DescribeNatGateways(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing NatGateway %q: %v", id, err)
	}

	if response == nil || len(response.NatGateways) == 0 {
		klog.V(2).Infof("Unable to find NatGateway %q", id)
		return nil, nil
	}
	if len(response.NatGateways) != 1 {
		return nil, fmt.Errorf("found multiple NatGateways with id %q", id)
	}
	return &response.NatGateways[0], nil
}

func findNatGatewayFromRouteTable(ctx context.Context, cloud awsup.AWSCloud, routeTable *RouteTable) (*ec2types.NatGateway, error) {
	// Find via route on private route table
	if routeTable.ID != nil {
		klog.V(2).Infof("trying to match NatGateway via RouteTable %s", *routeTable.ID)
		rt, err := findRouteTableByID(ctx, cloud, *routeTable.ID)
		if err != nil {
			return nil, fmt.Errorf("error finding associated RouteTable to NatGateway: %v", err)
		}

		if rt != nil {
			var natGatewayIDs []*string
			natGatewayIDsSeen := map[string]bool{}
			for _, route := range rt.Routes {
				if route.NatGatewayId != nil && route.State != ec2types.RouteStateBlackhole && !natGatewayIDsSeen[*route.NatGatewayId] {
					natGatewayIDs = append(natGatewayIDs, route.NatGatewayId)
					natGatewayIDsSeen[*route.NatGatewayId] = true
				}
			}

			if len(natGatewayIDs) == 0 {
				klog.V(2).Infof("no NatGateway found in route table %s", *rt.RouteTableId)
			} else if len(natGatewayIDs) > 1 {
				clusterName, ok := routeTable.Tags[awsup.TagClusterName]
				if !ok {
					return nil, fmt.Errorf("Could not find '%s' tag from route table", awsup.TagClusterName)
				}
				filteredNatGateways := []*ec2types.NatGateway{}
				for _, natGatewayID := range natGatewayIDs {
					gw, err := findNatGatewayById(ctx, cloud, fi.ValueOf(natGatewayID))
					if err != nil {
						return nil, err
					}

					if raws.HasOwnedTag(string(ec2types.ResourceTypeNatgateway)+":"+fi.ValueOf(natGatewayID), gw.Tags, clusterName) {
						filteredNatGateways = append(filteredNatGateways, gw)
					}
				}
				if len(filteredNatGateways) == 0 {
					klog.V(2).Infof("no kOps NatGateway found in route table %s", *rt.RouteTableId)
				} else if len(filteredNatGateways) > 1 {
					return nil, fmt.Errorf("found multiple kOps NatGateways in route table %s", *rt.RouteTableId)
				} else {
					return filteredNatGateways[0], nil
				}
			} else {
				return findNatGatewayById(ctx, cloud, fi.ValueOf(natGatewayIDs[0]))
			}
		}
	}

	return nil, nil
}

func (s *NatGateway) CheckChanges(a, e, changes *NatGateway) error {
	// New
	if a == nil {
		if !fi.ValueOf(e.Shared) {
			if e.ElasticIP == nil {
				return fi.RequiredField("ElasticIP")
			}
			if e.Subnet == nil {
				return fi.RequiredField("Subnet")
			}
		}
		if e.AssociatedRouteTable == nil {
			return fi.RequiredField("AssociatedRouteTable")
		}
	}

	// Delta
	if a != nil {
		if changes.ElasticIP != nil {
			eID := ""
			if e.ElasticIP != nil {
				eID = fi.ValueOf(e.ElasticIP.ID)
			}
			aID := ""
			if a.ElasticIP != nil {
				aID = fi.ValueOf(a.ElasticIP.ID)
			}
			return fi.FieldIsImmutable(eID, aID, field.NewPath("ElasticIP"))
		}
		if changes.Subnet != nil {
			return fi.CannotChangeField("Subnet")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
	}
	return nil
}

func (e *NatGateway) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *NatGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *NatGateway) error {
	// New NGW
	ctx := context.TODO()

	var id *string
	if a == nil {

		if fi.ValueOf(e.Shared) {
			return fmt.Errorf("NAT gateway %q not found", fi.ValueOf(e.ID))
		}

		klog.V(2).Infof("Creating Nat Gateway")

		request := &ec2.CreateNatGatewayInput{
			TagSpecifications: awsup.EC2TagSpecification(ec2types.ResourceTypeNatgateway, e.Tags),
		}
		request.AllocationId = e.ElasticIP.ID
		request.SubnetId = e.Subnet.ID
		response, err := t.Cloud.EC2().CreateNatGateway(ctx, request)
		if err != nil {
			return fmt.Errorf("Error creating Nat Gateway: %v", err)
		}
		e.ID = response.NatGateway.NatGatewayId
		id = e.ID
	} else {
		id = a.ID
	}

	err := t.AddAWSTags(*e.ID, e.Tags)
	if err != nil {
		return fmt.Errorf("unable to tag NatGateway")
	}

	// Tag the associated subnet
	if e.Subnet == nil {
		return fmt.Errorf("Subnet not set")
	} else if e.Subnet.ID == nil {
		return fmt.Errorf("Subnet ID not set")
	}

	// TODO: AssociatedNatgateway tag is obsolete - we can get from the route table instead
	tags := make(map[string]string)
	tags["AssociatedNatgateway"] = *id
	err = t.AddAWSTags(*e.Subnet.ID, tags)
	if err != nil {
		return fmt.Errorf("unable to tag subnet %v", err)
	}

	// If this is a shared NGW, we need to tag it
	// The tag that implies "shared" is `AssociatedNatgateway`=> NGW-ID
	// This is better than just a tag that's shared because this lets us create a whitelist of these NGWs
	// without doing a bunch more work in `kutil/delete_cluster.go`

	if fi.ValueOf(e.Shared) {
		if e.AssociatedRouteTable == nil {
			return fmt.Errorf("AssociatedRouteTable not provided")
		}
		klog.V(2).Infof("tagging route table %s to track shared NGW", fi.ValueOf(e.AssociatedRouteTable.ID))
		err = t.AddAWSTags(fi.ValueOf(e.AssociatedRouteTable.ID), tags)
		if err != nil {
			return fmt.Errorf("unable to tag route table %v", err)
		}
	}

	return nil
}

type terraformNATGateway struct {
	AllocationID *terraformWriter.Literal `cty:"allocation_id"`
	SubnetID     *terraformWriter.Literal `cty:"subnet_id"`
	Tag          map[string]string        `cty:"tags"`
}

func (_ *NatGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *NatGateway) error {
	if fi.ValueOf(e.Shared) {
		if e.ID == nil {
			return fmt.Errorf("ID must be set, if NatGateway is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing NatGateway with id %q", *e.ID)
		return nil
	}

	tf := &terraformNATGateway{
		AllocationID: e.ElasticIP.TerraformLink(),
		SubnetID:     e.Subnet.TerraformLink(),
		Tag:          e.Tags,
	}

	return t.RenderResource("aws_nat_gateway", *e.Name, tf)
}

func (e *NatGateway) TerraformLink() *terraformWriter.Literal {
	if fi.ValueOf(e.Shared) {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if NatGateway is shared: %s", e)
		}

		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_nat_gateway", *e.Name, "id")
}
