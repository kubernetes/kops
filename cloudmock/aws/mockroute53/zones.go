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

package mockroute53

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog"
)

func (m *MockRoute53) GetHostedZoneRequest(*route53.GetHostedZoneInput) (*request.Request, *route53.GetHostedZoneOutput) {
	panic("MockRoute53 GetHostedZoneRequest not implemented")
}

func (m *MockRoute53) GetHostedZoneWithContext(aws.Context, *route53.GetHostedZoneInput, ...request.Option) (*route53.GetHostedZoneOutput, error) {
	panic("Not implemented")
}

func (m *MockRoute53) GetHostedZone(request *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("GetHostedZone %v", request)

	if request.Id == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("Id is required FOUND")
	}
	zone := m.findZone(*request.Id)
	if zone == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("NOT FOUND")
	}

	copy := *zone.hostedZone
	response := &route53.GetHostedZoneOutput{
		// DelegationSet ???
		HostedZone: &copy,
		VPCs:       zone.vpcs,
	}
	return response, nil
}

func (m *MockRoute53) GetHostedZoneCountRequest(*route53.GetHostedZoneCountInput) (*request.Request, *route53.GetHostedZoneCountOutput) {
	panic("MockRoute53 GetHostedZoneCountRequest not implemented")
}
func (m *MockRoute53) GetHostedZoneCountWithContext(aws.Context, *route53.GetHostedZoneCountInput, ...request.Option) (*route53.GetHostedZoneCountOutput, error) {
	panic("Not implemented")
}
func (m *MockRoute53) GetHostedZoneCount(*route53.GetHostedZoneCountInput) (*route53.GetHostedZoneCountOutput, error) {
	panic("MockRoute53 GetHostedZoneCount not implemented")
}

func (m *MockRoute53) ListHostedZonesRequest(*route53.ListHostedZonesInput) (*request.Request, *route53.ListHostedZonesOutput) {
	panic("MockRoute53 ListHostedZonesRequest not implemented")
}

func (m *MockRoute53) ListHostedZonesWithContext(aws.Context, *route53.ListHostedZonesInput, ...request.Option) (*route53.ListHostedZonesOutput, error) {
	panic("Not implemented")
}

func (m *MockRoute53) ListHostedZones(*route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	panic("MockRoute53 ListHostedZones not implemented")
}

func (m *MockRoute53) ListHostedZonesPagesWithContext(aws.Context, *route53.ListHostedZonesInput, func(*route53.ListHostedZonesOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockRoute53) ListHostedZonesPages(request *route53.ListHostedZonesInput, callback func(*route53.ListHostedZonesOutput, bool) bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListHostedZonesPages %v", request)

	page := &route53.ListHostedZonesOutput{}
	for _, zone := range m.Zones {
		copy := *zone.hostedZone
		page.HostedZones = append(page.HostedZones, &copy)
	}
	lastPage := true
	callback(page, lastPage)

	return nil
}

func (m *MockRoute53) ListHostedZonesByNameRequest(*route53.ListHostedZonesByNameInput) (*request.Request, *route53.ListHostedZonesByNameOutput) {
	panic("MockRoute53 ListHostedZonesByNameRequest not implemented")
}

func (m *MockRoute53) ListHostedZonesByNameWithContext(aws.Context, *route53.ListHostedZonesByNameInput, ...request.Option) (*route53.ListHostedZonesByNameOutput, error) {
	panic("Not implemented")
}

func (m *MockRoute53) ListHostedZonesByName(*route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var zones []*route53.HostedZone

	for _, z := range m.Zones {
		zones = append(zones, z.hostedZone)
	}

	return &route53.ListHostedZonesByNameOutput{
		HostedZones: zones,
	}, nil
}
