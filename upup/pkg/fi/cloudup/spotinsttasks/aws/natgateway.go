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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinst"
)

//go:generate fitask -type=NatGateway
type NatGateway struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	ElasticIP *ElasticIP
	Subnet    *Subnet
	ID        *string

	EgressId *string

	// Shared is set if this is a shared NatGateway
	Shared *bool

	// We can't tag NatGateways, so we have to find through a surrogate
	AssociatedRouteTable *RouteTable
}

var _ fi.CompareWithID = &NatGateway{}

func (e *NatGateway) CompareWithID() *string {
	// Match by ID (NAT Gateways don't have tags, so they don't have a name in EC2)
	return e.ID
}

func (e *NatGateway) Find(c *fi.Context) (*NatGateway, error) {

	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)
	var ngw *ec2.NatGateway
	actual := &NatGateway{}

	if fi.StringValue(e.ID) != "" {
		// We have an existing NGW, lets look up the EIP
		var ngwIds []*string
		ngwIds = append(ngwIds, e.ID)

		request := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: ngwIds,
		}

		response, err := cloud.EC2().DescribeNatGateways(request)

		if err != nil {
			return nil, fmt.Errorf("error listing Nat Gateways %v", err)
		}

		if len(response.NatGateways) != 1 {
			return nil, fmt.Errorf("found %d Nat Gateways, expected 1", len(response.NatGateways))
		}
		ngw = response.NatGateways[0]

		if len(ngw.NatGatewayAddresses) != 1 {
			return nil, fmt.Errorf("found %d EIP Addresses for 1 NATGateway, expected 1", len(ngw.NatGatewayAddresses))
		}
		actual.ElasticIP = &ElasticIP{ID: ngw.NatGatewayAddresses[0].AllocationId}
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
		return nil, fmt.Errorf("found multiple elastic IPs attached to NatGateway %q", aws.StringValue(ngw.NatGatewayId))
	}

	// NATGateways don't have a Name (no tags), so we set the name to avoid spurious changes
	actual.Name = e.Name

	actual.AssociatedRouteTable = e.AssociatedRouteTable

	e.ID = actual.ID
	return actual, nil
}

func (e *NatGateway) findNatGateway(c *fi.Context) (*ec2.NatGateway, error) {
	cloud := c.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud)

	id := e.ID

	// Find via route on private route table
	if id == nil && e.AssociatedRouteTable != nil {
		ngw, err := findNatGatewayFromRouteTable(cloud, e.AssociatedRouteTable)
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
		var filters []*ec2.Filter
		filters = append(filters, awsup.NewEC2Filter("key", "AssociatedNatgateway"))
		if e.Subnet.ID == nil {
			glog.V(2).Infof("Unable to find subnet, bypassing Find() for NatGateway")
			return nil, nil
		}
		filters = append(filters, awsup.NewEC2Filter("resource-id", *e.Subnet.ID))

		request := &ec2.DescribeTagsInput{
			Filters: filters,
		}

		response, err := cloud.EC2().DescribeTags(request)
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
		glog.V(2).Infof("Found NatGateway via subnet tag: %v", *id)
	}

	if id != nil {
		return findNatGatewayById(cloud, id)
	}

	return nil, nil
}

func findNatGatewayById(cloud awsup.AWSCloud, id *string) (*ec2.NatGateway, error) {
	request := &ec2.DescribeNatGatewaysInput{}
	request.NatGatewayIds = []*string{id}
	response, err := cloud.EC2().DescribeNatGateways(request)
	if err != nil {
		return nil, fmt.Errorf("error listing NatGateway %q: %v", id, err)
	}

	if response == nil || len(response.NatGateways) == 0 {
		glog.V(2).Infof("Unable to find NatGateway %q", id)
		return nil, nil
	}
	if len(response.NatGateways) != 1 {
		return nil, fmt.Errorf("found multiple NatGateways with id %q", id)
	}
	return response.NatGateways[0], nil
}

func findNatGatewayFromRouteTable(cloud awsup.AWSCloud, routeTable *RouteTable) (*ec2.NatGateway, error) {
	// Find via route on private route table
	if routeTable.ID != nil {
		glog.V(2).Infof("trying to match NatGateway via RouteTable %s", *routeTable.ID)
		rt, err := routeTable.findEc2RouteTable(cloud)
		if err != nil {
			return nil, fmt.Errorf("error finding associated RouteTable to NatGateway: %v", err)
		}

		if rt != nil {
			var natGatewayIDs []*string
			for _, route := range rt.Routes {
				if route.NatGatewayId != nil {
					natGatewayIDs = append(natGatewayIDs, route.NatGatewayId)
				}
			}

			if len(natGatewayIDs) == 0 {
				glog.V(2).Infof("no NatGateway found in route table %s", *rt.RouteTableId)
			} else if len(natGatewayIDs) > 1 {
				return nil, fmt.Errorf("found multiple NatGateways in route table %s", *rt.RouteTableId)
			} else {
				return findNatGatewayById(cloud, natGatewayIDs[0])
			}
		}
	}

	return nil, nil
}

func (s *NatGateway) CheckChanges(a, e, changes *NatGateway) error {
	// New
	if a == nil {
		if !fi.BoolValue(e.Shared) {
			if e.ElasticIP == nil {
				return fi.RequiredField("ElasticIp")
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
			return fi.CannotChangeField("ElasticIp")
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

func (e *NatGateway) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *NatGateway) waitAvailable(cloud awsup.AWSCloud) error {
	// It takes 'forever' (up to 5 min...) for a NatGateway to become available after it has been created
	// We have to wait until it is actually up

	// TODO: Cache availability status

	id := aws.StringValue(e.ID)
	if id == "" {
		return fmt.Errorf("NAT Gateway %q did not have ID", e.Name)
	}

	glog.Infof("Waiting for NAT Gateway %q to be available (this often takes about 5 minutes)", id)
	params := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{e.ID},
	}
	err := cloud.EC2().WaitUntilNatGatewayAvailable(params)
	if err != nil {
		return fmt.Errorf("error waiting for NAT Gateway %q to be available: %v", id, err)
	}

	return nil
}

func (_ *NatGateway) Render(t *spotinst.Target, a, e, changes *NatGateway) error {
	// New NGW

	var id *string
	if a == nil {

		if fi.BoolValue(e.Shared) {
			return fmt.Errorf("NAT gateway %q not found", fi.StringValue(e.ID))
		}

		glog.V(2).Infof("Creating Nat Gateway")

		request := &ec2.CreateNatGatewayInput{}
		request.AllocationId = e.ElasticIP.ID
		request.SubnetId = e.Subnet.ID
		response, err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).EC2().CreateNatGateway(request)
		if err != nil {
			return fmt.Errorf("Error creating Nat Gateway: %v", err)
		}
		e.ID = response.NatGateway.NatGatewayId
		id = e.ID
	} else {
		id = a.ID
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
	err := t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(*e.Subnet.ID, tags)
	if err != nil {
		return fmt.Errorf("unable to tag subnet %v", err)
	}

	// If this is a shared NGW, we need to tag it
	// The tag that implies "shared" is `AssociatedNatgateway`=> NGW-ID
	// This is better than just a tag that's shared because this lets us create a whitelist of these NGWs
	// without doing a bunch more work in `kutil/delete_cluster.go`

	if fi.BoolValue(e.Shared) {
		if e.AssociatedRouteTable == nil {
			return fmt.Errorf("AssociatedRouteTable not provided")
		}
		glog.V(2).Infof("tagging route table %s to track shared NGW", fi.StringValue(e.AssociatedRouteTable.ID))
		err = t.Cloud.(spotinst.Cloud).Cloud().(awsup.AWSCloud).AddAWSTags(fi.StringValue(e.AssociatedRouteTable.ID), tags)
		if err != nil {
			return fmt.Errorf("unable to tag route table %v", err)
		}
	}

	return nil
}
