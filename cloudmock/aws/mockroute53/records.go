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

func (m *MockRoute53) ListResourceRecordSetsRequest(*route53.ListResourceRecordSetsInput) (*request.Request, *route53.ListResourceRecordSetsOutput) {
	panic("MockRoute53 ListResourceRecordSetsRequest not implemented")
}

func (m *MockRoute53) ListResourceRecordSetsWithContext(aws.Context, *route53.ListResourceRecordSetsInput, ...request.Option) (*route53.ListResourceRecordSetsOutput, error) {
	panic("Not implemented")
}

func (m *MockRoute53) ListResourceRecordSets(*route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	panic("MockRoute53 ListResourceRecordSets not implemented")
}

func (m *MockRoute53) ListResourceRecordSetsPagesWithContext(aws.Context, *route53.ListResourceRecordSetsInput, func(*route53.ListResourceRecordSetsOutput, bool) bool, ...request.Option) error {
	panic("Not implemented")
}

func (m *MockRoute53) ListResourceRecordSetsPages(request *route53.ListResourceRecordSetsInput, callback func(*route53.ListResourceRecordSetsOutput, bool) bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListResourceRecordSetsPages %v", request)

	if request.HostedZoneId == nil {
		// TODO: Use correct error
		return fmt.Errorf("HostedZoneId required")
	}

	if request.StartRecordIdentifier != nil || request.StartRecordName != nil || request.StartRecordType != nil || request.MaxItems != nil {
		klog.Fatalf("Unsupported options: %v", request)
	}

	zone := m.findZone(*request.HostedZoneId)

	if zone == nil {
		// TODO: Use correct error
		return fmt.Errorf("NOT FOUND")
	}

	page := &route53.ListResourceRecordSetsOutput{}
	for _, r := range zone.records {
		copy := *r
		page.ResourceRecordSets = append(page.ResourceRecordSets, &copy)
	}
	lastPage := true
	callback(page, lastPage)

	return nil
}

func (m *MockRoute53) ChangeResourceRecordSetsRequest(*route53.ChangeResourceRecordSetsInput) (*request.Request, *route53.ChangeResourceRecordSetsOutput) {
	panic("MockRoute53 ChangeResourceRecordSetsRequest not implemented")
}

func (m *MockRoute53) ChangeResourceRecordSetsWithContext(aws.Context, *route53.ChangeResourceRecordSetsInput, ...request.Option) (*route53.ChangeResourceRecordSetsOutput, error) {
	panic("Not implemented")
}

func (m *MockRoute53) ChangeResourceRecordSets(request *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ChangeResourceRecordSets %v", request)

	if request.HostedZoneId == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("HostedZoneId required")
	}
	zone := m.findZone(*request.HostedZoneId)
	if zone == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("NOT FOUND")
	}

	response := &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &route53.ChangeInfo{},
	}
	for _, change := range request.ChangeBatch.Changes {
		changeType := aws.StringValue(change.ResourceRecordSet.Type)
		changeName := aws.StringValue(change.ResourceRecordSet.Name)

		foundIndex := -1
		for i, rr := range zone.records {
			if aws.StringValue(rr.Type) != changeType {
				continue
			}
			if aws.StringValue(rr.Name) != changeName {
				continue
			}
			foundIndex = i
			break
		}

		switch aws.StringValue(change.Action) {
		case "UPSERT":
			if foundIndex == -1 {
				zone.records = append(zone.records, change.ResourceRecordSet)
			} else {
				zone.records[foundIndex] = change.ResourceRecordSet
			}

		case "CREATE":
			if foundIndex == -1 {
				zone.records = append(zone.records, change.ResourceRecordSet)
			} else {
				// TODO: Use correct error
				return nil, fmt.Errorf("duplicate record %s %q", changeType, changeName)
			}

		case "DELETE":
			if foundIndex == -1 {
				// TODO: Use correct error
				return nil, fmt.Errorf("record not found %s %q", changeType, changeName)
			}
			zone.records = append(zone.records[:foundIndex], zone.records[foundIndex+1:]...)

		default:
			// TODO: Use correct error
			return nil, fmt.Errorf("Unsupported action: %q", aws.StringValue(change.Action))
		}
	}

	return response, nil
}
