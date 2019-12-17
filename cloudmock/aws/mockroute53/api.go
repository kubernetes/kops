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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type zoneInfo struct {
	ID         string
	hostedZone *route53.HostedZone
	records    []*route53.ResourceRecordSet
	vpcs       []*route53.VPC
}

type MockRoute53 struct {
	// Mock out interface
	route53iface.Route53API

	mutex sync.Mutex
	Zones []*zoneInfo
}

var _ route53iface.Route53API = &MockRoute53{}

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

func (m *MockRoute53) MockCreateZone(z *route53.HostedZone, vpcs []*route53.VPC) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	zi := &zoneInfo{
		ID:         aws.StringValue(z.Id),
		hostedZone: z,
		vpcs:       vpcs,
	}
	m.Zones = append(m.Zones, zi)
}
