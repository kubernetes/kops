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

package mockroute53

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"strings"
)

type zoneInfo struct {
}

func (m *MockRoute53) GetHostedZoneRequest(*route53.GetHostedZoneInput) (*request.Request, *route53.GetHostedZoneOutput) {
	panic("MockRoute53 GetHostedZoneRequest not implemented")
	return nil, nil
}
func (m *MockRoute53) GetHostedZone(request *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	glog.Infof("GetHostedZone %v", request)

	findID := aws.StringValue(request.Id)
	if !strings.Contains(findID, "/") {
		findID = "/hostedzone/" + findID
	}

	for _, z := range m.Zones {
		if *z.Id != findID {
			continue
		}

		copy := *z
		response := &route53.GetHostedZoneOutput{
			// DelegationSet ???
			HostedZone: &copy,
			// VPCs
		}
		return response, nil
	}

	// TODO: Correct error
	return nil, fmt.Errorf("NOT FOUND")
}

func (m *MockRoute53) GetHostedZoneCountRequest(*route53.GetHostedZoneCountInput) (*request.Request, *route53.GetHostedZoneCountOutput) {
	panic("MockRoute53 GetHostedZoneCountRequest not implemented")
	return nil, nil
}
func (m *MockRoute53) GetHostedZoneCount(*route53.GetHostedZoneCountInput) (*route53.GetHostedZoneCountOutput, error) {
	panic("MockRoute53 GetHostedZoneCount not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZonesRequest(*route53.ListHostedZonesInput) (*request.Request, *route53.ListHostedZonesOutput) {
	panic("MockRoute53 ListHostedZonesRequest not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZones(*route53.ListHostedZonesInput) (*route53.ListHostedZonesOutput, error) {
	panic("MockRoute53 ListHostedZones not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZonesPages(request *route53.ListHostedZonesInput, callback func(*route53.ListHostedZonesOutput, bool) bool) error {
	glog.Infof("ListHostedZonesPages %v", request)

	page := &route53.ListHostedZonesOutput{}
	for _, zone := range m.Zones {
		copy := *zone
		page.HostedZones = append(page.HostedZones, &copy)
	}
	lastPage := true
	callback(page, lastPage)

	return nil
}

func (m *MockRoute53) ListHostedZonesByNameRequest(*route53.ListHostedZonesByNameInput) (*request.Request, *route53.ListHostedZonesByNameOutput) {
	panic("MockRoute53 ListHostedZonesByNameRequest not implemented")
	return nil, nil
}

func (m *MockRoute53) ListHostedZonesByName(*route53.ListHostedZonesByNameInput) (*route53.ListHostedZonesByNameOutput, error) {
	panic("MockRoute53 ListHostedZonesByName not implemented")
	return nil, nil
}
