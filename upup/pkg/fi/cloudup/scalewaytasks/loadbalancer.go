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

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// +kops:fitask
type LoadBalancer struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Zone                  *string
	LBID                  *string
	LBAddresses           []string
	Tags                  []string
	Description           string
	SslCompatibilityLevel string
	ForAPIServer          bool

	VPCId *string // set if Cluster.Spec.NetworkID is
	//VPCName     *string // set if Cluster.Spec.NetworkCIDR is
	//NetworkCIDR *string // set if Cluster.Spec.NetworkCIDR is
}

var _ fi.CompareWithID = &LoadBalancer{}
var _ fi.HasAddress = &LoadBalancer{}

func (l *LoadBalancer) CompareWithID() *string {
	return l.LBID
}

func (l *LoadBalancer) IsForAPIServer() bool {
	return l.ForAPIServer
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
		Name:         fi.PtrTo(loadBalancer.Name),
		LBID:         fi.PtrTo(loadBalancer.ID),
		Zone:         fi.PtrTo(string(loadBalancer.Zone)),
		LBAddresses:  lbIPs,
		Tags:         loadBalancer.Tags,
		Lifecycle:    l.Lifecycle,
		ForAPIServer: l.ForAPIServer,
	}, nil
}

func (l *LoadBalancer) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	if l.LBID == nil {
		return nil, nil
	}

	loadBalancer, err := lbService.GetLB(&lb.ZonedAPIGetLBRequest{
		Zone: scw.Zone(cloud.Zone()),
		LBID: fi.ValueOf(l.LBID),
	})
	if err != nil {
		return nil, err
	}

	addresses := []string(nil)
	for _, address := range loadBalancer.IP {
		addresses = append(addresses, address.IPAddress)
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

		klog.Infof("Updating existing load-balancer with name %q", expected.Name)

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

		klog.Infof("Creating new load-balancer with name %q", expected.Name)

		lbCreated, err := lbService.CreateLB(&lb.ZonedAPICreateLBRequest{
			Zone: scw.Zone(fi.ValueOf(expected.Zone)),
			Name: fi.ValueOf(expected.Name),
			Tags: expected.Tags,
		})
		if err != nil {
			return fmt.Errorf("creating load-balancer: %w", err)
		}

		_, err = lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
			LBID: lbCreated.ID,
			Zone: scw.Zone(fi.ValueOf(expected.Zone)),
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
