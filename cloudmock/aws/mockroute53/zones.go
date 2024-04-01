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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/klog/v2"
)

func (m *MockRoute53) GetHostedZone(ctx context.Context, request *route53.GetHostedZoneInput, optFns ...func(*route53.Options)) (*route53.GetHostedZoneOutput, error) {
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
	vpcs := make([]route53types.VPC, len(zone.vpcs))
	for i := range zone.vpcs {
		vpcs[i] = *zone.vpcs[i]
	}

	copy := *zone.hostedZone
	response := &route53.GetHostedZoneOutput{
		// DelegationSet ???
		HostedZone: &copy,
		VPCs:       vpcs,
	}
	return response, nil
}

func (m *MockRoute53) ListHostedZones(ctx context.Context, request *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListHostedZonesPages %v", request)

	page := &route53.ListHostedZonesOutput{}
	for _, zone := range m.Zones {
		copy := *zone.hostedZone
		page.HostedZones = append(page.HostedZones, copy)
	}

	return page, nil
}

func (m *MockRoute53) ListHostedZonesByName(ctx context.Context, request *route53.ListHostedZonesByNameInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var zones []route53types.HostedZone

	for _, z := range m.Zones {
		zones = append(zones, *z.hostedZone)
	}

	return &route53.ListHostedZonesByNameOutput{
		HostedZones: zones,
	}, nil
}
