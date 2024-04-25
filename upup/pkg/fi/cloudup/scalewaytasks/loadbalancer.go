/*
Copyright 2022 The Kubernetes Authors.

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
	"os"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const LbDefaultType = "LB-S"

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	Type      string
	Lifecycle fi.Lifecycle

	Zone                  *string
	LBID                  *string
	LBAddresses           []string
	Tags                  []string
	Description           string
	SslCompatibilityLevel string

	// WellKnownServices indicates which services are supported by this resource.
	// This field is internal and is not rendered to the cloud.
	WellKnownServices []wellknownservices.WellKnownService

	PrivateNetwork *PrivateNetwork
}

var _ fi.CloudupTask = &LoadBalancer{}
var _ fi.CompareWithID = &LoadBalancer{}
var _ fi.CloudupHasDependencies = &LoadBalancer{}
var _ fi.HasAddress = &LoadBalancer{}

func (l *LoadBalancer) CompareWithID() *string {
	return l.LBID
}

func (l *LoadBalancer) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

// GetWellKnownServices implements fi.HasAddress::GetWellKnownServices.
// It indicates which services we support with this load balancer.
func (l *LoadBalancer) GetWellKnownServices() []wellknownservices.WellKnownService {
	return l.WellKnownServices
}

func (l *LoadBalancer) Find(context *fi.CloudupContext) (*LoadBalancer, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	lbResponse, err := lbService.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: scw.Zone(cloud.Zone()),
		Name: l.Name,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("getting load-balancer %s: %w", fi.ValueOf(l.LBID), err)
	}
	if lbResponse.TotalCount != 1 {
		return nil, nil
	}
	loadBalancer := lbResponse.LBs[0]

	lbIPs := []string(nil)
	for _, IP := range loadBalancer.IP {
		lbIPs = append(lbIPs, IP.IPAddress)
	}

	return &LoadBalancer{
		Name:              fi.PtrTo(loadBalancer.Name),
		LBID:              fi.PtrTo(loadBalancer.ID),
		Zone:              fi.PtrTo(string(loadBalancer.Zone)),
		LBAddresses:       lbIPs,
		Tags:              loadBalancer.Tags,
		Lifecycle:         l.Lifecycle,
		WellKnownServices: l.WellKnownServices,
	}, nil
}

func (l *LoadBalancer) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	// Skip if we're running integration tests
	if profileName := os.Getenv("SCW_PROFILE"); profileName == "REDACTED" {
		return nil, nil
	}

	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	loadBalancers, err := lbService.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone: scw.Zone(cloud.Zone()),
		Name: l.Name,
	})
	if err != nil {
		return nil, err
	}

	addresses := []string(nil)
	for _, loadBalancer := range loadBalancers.LBs {
		for _, address := range loadBalancer.IP {
			addresses = append(addresses, address.IPAddress)
		}
	}

	return addresses, nil
}

func (l *LoadBalancer) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, context)
}

func (_ *LoadBalancer) CheckChanges(actual, expected, changes *LoadBalancer) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.LBID != nil {
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

func (l *LoadBalancer) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LoadBalancer) error {
	lbService := t.Cloud.LBService()

	if actual != nil {

		klog.Infof("Updating existing load-balancer with name %q", fi.ValueOf(expected.Name))

		// We update the tags
		if changes != nil || len(actual.Tags) != len(expected.Tags) {
			_, err := lbService.UpdateLB(&lb.ZonedAPIUpdateLBRequest{
				Zone:                  scw.Zone(fi.ValueOf(actual.Zone)),
				LBID:                  fi.ValueOf(actual.LBID),
				Name:                  fi.ValueOf(actual.Name),
				Description:           expected.Description,
				SslCompatibilityLevel: lb.SSLCompatibilityLevel(expected.SslCompatibilityLevel),
				Tags:                  expected.Tags,
			})
			if err != nil {
				return fmt.Errorf("updatings tags for load-balancer %q: %w", fi.ValueOf(expected.Name), err)
			}
		}

		expected.LBID = actual.LBID
		expected.LBAddresses = actual.LBAddresses

	} else {

		klog.Infof("Creating new load-balancer with name %q", fi.ValueOf(expected.Name))
		zone := scw.Zone(fi.ValueOf(expected.Zone))

		lbCreated, err := lbService.CreateLB(&lb.ZonedAPICreateLBRequest{
			Zone: zone,
			Name: fi.ValueOf(expected.Name),
			Type: LbDefaultType,
			Tags: expected.Tags,
			//AssignFlexibleIP: fi.PtrTo(true),
		})
		if err != nil {
			return fmt.Errorf("creating load-balancer: %w", err)
		}

		_, err = lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			LBID: lbCreated.ID,
			Zone: zone,
		})
		if err != nil {
			return fmt.Errorf("waiting for load-balancer %s: %w", lbCreated.ID, err)
		}

		_, err = lbService.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
			Zone:             zone,
			LBID:             lbCreated.ID,
			PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
		})
		if err != nil {
			return fmt.Errorf("attaching load-balancer to private network: %w", err)
		}

		_, err = lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			LBID: lbCreated.ID,
			Zone: zone,
		})
		if err != nil {
			return fmt.Errorf("waiting for load-balancer %s: %w", lbCreated.ID, err)
		}

		lbIPs := []string(nil)
		for _, ip := range lbCreated.IP {
			lbIPs = append(lbIPs, ip.IPAddress)
		}
		expected.LBID = &lbCreated.ID
		expected.LBAddresses = lbIPs
	}

	return nil
}

type terraformLBIP struct{}

type terraformLoadBalancer struct {
	Type        string                   `cty:"type"`
	Name        *string                  `cty:"name"`
	Description string                   `cty:"description"`
	Tags        []string                 `cty:"tags"`
	IPID        *terraformWriter.Literal `cty:"ip_id"`
}

func (_ *LoadBalancer) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *LoadBalancer) error {
	tfName := strings.ReplaceAll(fi.ValueOf(expected.Name), ".", "-")

	tfLBIP := terraformLBIP{}
	err := t.RenderResource("scaleway_lb_ip", tfName, tfLBIP)
	if err != nil {
		return err
	}

	tfLB := terraformLoadBalancer{
		Type:        LbDefaultType,
		Name:        expected.Name,
		Description: expected.Description,
		Tags:        expected.Tags,
		IPID:        terraformWriter.LiteralProperty("scaleway_lb_ip", tfName, "id"),
	}
	return t.RenderResource("scaleway_lb", tfName, tfLB)
}

func (l *LoadBalancer) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("scaleway_lb", fi.ValueOf(l.Name), "id")
}
