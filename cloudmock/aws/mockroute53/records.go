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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/klog/v2"
)

func (m *MockRoute53) ListResourceRecordSets(ctx context.Context, request *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ListResourceRecordSetsPages %v", request)

	if request.HostedZoneId == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("HostedZoneId required")
	}

	if request.StartRecordIdentifier != nil || request.StartRecordName != nil || len(request.StartRecordType) > 0 || request.MaxItems != nil {
		klog.Fatalf("Unsupported options: %v", request)
	}

	zone := m.findZone(*request.HostedZoneId)

	if zone == nil {
		// TODO: Use correct error
		return nil, fmt.Errorf("NOT FOUND")
	}

	page := &route53.ListResourceRecordSetsOutput{}
	for _, r := range zone.records {
		copy := *r
		page.ResourceRecordSets = append(page.ResourceRecordSets, copy)
	}

	return page, nil
}

func (m *MockRoute53) ChangeResourceRecordSets(ctx context.Context, request *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
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
		ChangeInfo: &route53types.ChangeInfo{},
	}
	for _, change := range request.ChangeBatch.Changes {
		changeType := change.ResourceRecordSet.Type
		changeName := aws.ToString(change.ResourceRecordSet.Name)

		foundIndex := -1
		for i, rr := range zone.records {
			if rr.Type != changeType {
				continue
			}
			if aws.ToString(rr.Name) != changeName {
				continue
			}
			foundIndex = i
			break
		}

		switch change.Action {
		case route53types.ChangeActionUpsert:
			if foundIndex == -1 {
				zone.records = append(zone.records, change.ResourceRecordSet)
			} else {
				zone.records[foundIndex] = change.ResourceRecordSet
			}

		case route53types.ChangeActionCreate:
			if foundIndex == -1 {
				zone.records = append(zone.records, change.ResourceRecordSet)
			} else {
				// TODO: Use correct error
				return nil, fmt.Errorf("duplicate record %s %q", changeType, changeName)
			}

		case route53types.ChangeActionDelete:
			if foundIndex == -1 {
				// TODO: Use correct error
				return nil, fmt.Errorf("record not found %s %q", changeType, changeName)
			}
			zone.records = append(zone.records[:foundIndex], zone.records[foundIndex+1:]...)

		default:
			// TODO: Use correct error
			return nil, fmt.Errorf("Unsupported action: %q", change.Action)
		}
	}

	return response, nil
}
