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
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type zoneInfo struct {
	ID         string
	hostedZone *route53types.HostedZone
	records    []*route53types.ResourceRecordSet
	vpcs       []*route53types.VPC
}

type MockRoute53 struct {
	// Mock out interface
	awsinterfaces.Route53API

	mutex sync.Mutex
	Zones []*zoneInfo
}

var _ awsinterfaces.Route53API = &MockRoute53{}

func (m *MockRoute53) findZone(hostedZoneId string) *zoneInfo {
	if !strings.Contains(hostedZoneId, "/") {
		hostedZoneId = "/hostedzone/" + hostedZoneId
	}

	for _, z := range m.Zones {
		if z.ID == hostedZoneId {
			return z
		}
	}
	return nil
}

func (m *MockRoute53) MockCreateZone(z *route53types.HostedZone, vpcs []*route53types.VPC) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	zi := &zoneInfo{
		ID:         aws.ToString(z.Id),
		hostedZone: z,
		vpcs:       vpcs,
	}
	m.Zones = append(m.Zones, zi)
}
