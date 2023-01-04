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

	Region       *string
	LBID         *string
	LBAddresses  []string
	Tags         []string
	ForAPIServer bool
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
	if fi.ValueOf(l.LBID) == "" {
		return nil, nil
	}

	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	loadBalancer, err := lbService.GetLB(&lb.GetLBRequest{
		Region: scw.Region(cloud.Region()),
		LBID:   fi.ValueOf(l.LBID),
	})
	if err != nil {
		return nil, fmt.Errorf("getting load-balancer %s: %s", fi.ValueOf(l.LBID), err)
	}

	lbIPs := []string(nil)
	for _, IP := range loadBalancer.IP {
		lbIPs = append(lbIPs, IP.IPAddress)
	}

	return &LoadBalancer{
		Name:         &loadBalancer.Name,
		LBID:         &loadBalancer.ID,
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

	loadBalancer, err := lbService.GetLB(&lb.GetLBRequest{
		Region: scw.Region(cloud.Region()),
		LBID:   fi.ValueOf(l.LBID),
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
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (l *LoadBalancer) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LoadBalancer) error {
	lbService := t.Cloud.LBService()
	region := scw.Region(fi.ValueOf(expected.Region))
	var loadBalancer *lb.LB
	backEndToCreate := true
	frontEndToCreate := true

	if actual != nil {
		klog.Infof("Updating existing load-balancer with name %q", expected.Name)

		lbToUpdate, err := lbService.GetLB(&lb.GetLBRequest{
			Region: region,
			LBID:   fi.ValueOf(actual.LBID),
		})
		if err != nil {
			return fmt.Errorf("getting load-balancer %q (%s): %w", fi.ValueOf(actual.Name), fi.ValueOf(actual.LBID), err)
		}

		// We update the tags
		if changes != nil || len(actual.Tags) != len(expected.Tags) {
			_, err = lbService.UpdateLB(&lb.UpdateLBRequest{
				Region:                region,
				LBID:                  lbToUpdate.ID,
				Name:                  lbToUpdate.Name,
				Description:           lbToUpdate.Description,
				SslCompatibilityLevel: lbToUpdate.SslCompatibilityLevel,
				Tags:                  expected.Tags,
			})
			if err != nil {
				return fmt.Errorf("updatings tags for load-balancer %q: %w", fi.ValueOf(expected.Name), err)
			}
		}

		// We check that the back-end exists
		backEnds, err := lbService.ListBackends(&lb.ListBackendsRequest{
			Region: region,
			LBID:   lbToUpdate.ID,
			Name:   scw.StringPtr("lb-backend"),
		})
		if err != nil {
			return fmt.Errorf("listing back-ends for load-balancer %q: %w", fi.ValueOf(expected.Name), err)
		}
		if backEnds.TotalCount > 0 {
			backEndToCreate = false
		}

		// We check that the front-end exists
		frontEnds, err := lbService.ListFrontends(&lb.ListFrontendsRequest{
			Region: region,
			LBID:   lbToUpdate.ID,
			Name:   scw.StringPtr("lb-frontend"),
		})
		if err != nil {
			return fmt.Errorf("listing front-ends for load-balancer %q: %w", fi.ValueOf(expected.Name), err)
		}
		if frontEnds.TotalCount > 0 {
			frontEndToCreate = false
		}

		lbIPs := []string(nil)
		for _, ip := range lbToUpdate.IP {
			lbIPs = append(lbIPs, ip.IPAddress)
		}
		expected.LBID = &lbToUpdate.ID
		expected.LBAddresses = lbIPs

		loadBalancer = lbToUpdate

	} else {
		klog.Infof("Creating new load-balancer with name %q", expected.Name)

		lbCreated, err := lbService.CreateLB(&lb.CreateLBRequest{
			Region: region,
			Name:   fi.ValueOf(expected.Name),
			Tags:   expected.Tags,
		})
		if err != nil {
			return fmt.Errorf("creating load-balancer: %w", err)
		}

		_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
			LBID:   lbCreated.ID,
			Region: region,
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

		loadBalancer = lbCreated
	}

	backEndID := ""

	// We create the load-balancer's backend if needed
	if backEndToCreate == true {
		backEnd, err := lbService.CreateBackend(&lb.CreateBackendRequest{
			Region:               region,
			LBID:                 loadBalancer.ID,
			Name:                 "lb-backend",
			ForwardProtocol:      "tcp",
			ForwardPort:          443,
			ForwardPortAlgorithm: "roundrobin",
			StickySessions:       "none",
			HealthCheck: &lb.HealthCheck{
				CheckMaxRetries: 5,
				TCPConfig:       &lb.HealthCheckTCPConfig{},
				Port:            443,
				CheckTimeout:    scw.TimeDurationPtr(3000),
				CheckDelay:      scw.TimeDurationPtr(1001),
			},
			ProxyProtocol: "proxy_protocol_none",
		})
		if err != nil {
			return fmt.Errorf("creating back-end for load-balancer %s: %w", loadBalancer.ID, err)
		}

		_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
			LBID:   loadBalancer.ID,
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("waiting for load-balancer %s: %w", loadBalancer.ID, err)
		}
		backEndID = backEnd.ID
	}

	// We create the load-balancer's front-end if needed
	if frontEndToCreate == true {
		_, err := lbService.CreateFrontend(&lb.CreateFrontendRequest{
			Region:      region,
			LBID:        loadBalancer.ID,
			Name:        "lb-frontend",
			InboundPort: 443,
			BackendID:   backEndID,
		})
		if err != nil {
			return fmt.Errorf("creating front-end for load-balancer %s: %w", loadBalancer.ID, err)
		}
		_, err = lbService.WaitForLb(&lb.WaitForLBRequest{
			LBID:   loadBalancer.ID,
			Region: region,
		})
		if err != nil {
			return fmt.Errorf("waiting for load-balancer %s: %w", loadBalancer.ID, err)
		}
	}

	return nil
}
