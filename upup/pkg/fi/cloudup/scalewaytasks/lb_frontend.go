/*
Copyright 2023 The Kubernetes Authors.

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

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type LBFrontend struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID          *string
	Zone        *string
	InboundPort *int32

	LoadBalancer *LoadBalancer
	LBBackend    *LBBackend
}

var _ fi.CloudupTask = &LBFrontend{}
var _ fi.CompareWithID = &LBFrontend{}

func (l *LBFrontend) CompareWithID() *string {
	return l.ID
}

func (l *LBFrontend) Find(context *fi.CloudupContext) (*LBFrontend, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	lbService := cloud.LBService()

	frontendResponse, err := lbService.ListFrontends(&lb.ZonedAPIListFrontendsRequest{
		Zone: scw.Zone(cloud.Zone()),
		LBID: fi.ValueOf(l.LoadBalancer.LBID),
		Name: l.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("listing front-ends for load-balancer %s: %w", fi.ValueOf(l.LoadBalancer.LBID), err)
	}
	if frontendResponse.TotalCount != 1 {
		return nil, nil
	}
	frontend := frontendResponse.Frontends[0]

	return &LBFrontend{
		Name:        fi.PtrTo(frontend.Name),
		Lifecycle:   l.Lifecycle,
		ID:          fi.PtrTo(frontend.ID),
		Zone:        fi.PtrTo(string(frontend.LB.Zone)),
		InboundPort: fi.PtrTo(frontend.InboundPort),
		LoadBalancer: &LoadBalancer{
			Name: fi.PtrTo(frontend.LB.Name),
		},
		LBBackend: &LBBackend{
			Name: fi.PtrTo(frontend.Backend.Name),
			ID:   fi.PtrTo(frontend.Backend.ID),
		},
	}, nil
}

func (l *LBFrontend) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, context)
}

func (_ *LBFrontend) CheckChanges(actual, expected, changes *LBFrontend) error {
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

func (l *LBFrontend) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *LBFrontend) error {
	lbService := t.Cloud.LBService()

	if actual != nil {

		_, err := lbService.UpdateFrontend(&lb.ZonedAPIUpdateFrontendRequest{
			Zone:        scw.Zone(fi.ValueOf(actual.Zone)),
			FrontendID:  fi.ValueOf(actual.ID),
			Name:        fi.ValueOf(actual.Name),
			InboundPort: fi.ValueOf(expected.InboundPort),
			BackendID:   fi.ValueOf(actual.LBBackend.ID),
		})
		if err != nil {
			return fmt.Errorf("updating front-end for load-balancer %s: %w", fi.ValueOf(actual.LoadBalancer.Name), err)
		}

	} else {

		frontendCreated, err := lbService.CreateFrontend(&lb.ZonedAPICreateFrontendRequest{
			Zone:        scw.Zone(fi.ValueOf(expected.Zone)),
			LBID:        fi.ValueOf(expected.LoadBalancer.LBID), // try expected instead of l
			Name:        fi.ValueOf(expected.Name),
			InboundPort: fi.ValueOf(expected.InboundPort),
			BackendID:   fi.ValueOf(expected.LBBackend.ID), // try expected instead of l
		})
		if err != nil {
			return fmt.Errorf("creating front-end for load-balancer %s: %w", fi.ValueOf(expected.LoadBalancer.Name), err)
		}

		expected.ID = &frontendCreated.ID

	}

	_, err := lbService.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID: fi.ValueOf(expected.LoadBalancer.LBID),
		Zone: scw.Zone(fi.ValueOf(expected.Zone)),
	})
	if err != nil {
		return fmt.Errorf("waiting for load-balancer %s: %w", fi.ValueOf(expected.LoadBalancer.Name), err)
	}

	return nil
}

type terraformLBFrontend struct {
	//BackendID   *terraformWriter.Literal `cty:"backend_id"`
	BackendID   *string                  `cty:"backend_id"`
	LBID        *terraformWriter.Literal `cty:"lb_id"`
	Name        *string                  `cty:"name"`
	InboundPort *int32                   `cty:"inbound_port"`
}

func (_ *LBFrontend) RenderTerraform(t *terraform.TerraformTarget, actual, expected, changes *LBFrontend) error {
	tfName := strings.Replace(fi.ValueOf(expected.LoadBalancer.Name), ".", "-", -1)
	tf := terraformLBFrontend{
		LBID:        expected.TerraformLinkLBID(tfName),
		BackendID:   expected.LBBackend.ID,
		Name:        expected.Name,
		InboundPort: expected.InboundPort,
	}
	return t.RenderResource("scaleway_lb_frontend", tfName, tf)
}

func (l *LBFrontend) TerraformLinkLBID(tfName string) *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("scaleway_lb", tfName, "id")
}
