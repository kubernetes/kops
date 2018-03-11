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

package mockec2

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type MockEC2 struct {
	mutex sync.Mutex

	addressNumber int
	Addresses     []*ec2.Address

	RouteTables map[string]*ec2.RouteTable

	DhcpOptions map[string]*ec2.DhcpOptions

	Images []*ec2.Image

	securityGroupNumber int
	SecurityGroups      map[string]*ec2.SecurityGroup

	subnetNumber int
	subnets      map[string]*subnetInfo

	Volumes map[string]*ec2.Volume

	KeyPairs []*ec2.KeyPairInfo

	Tags []*ec2.TagDescription

	vpcNumber int
	Vpcs      map[string]*vpcInfo

	internetGatewayNumber int
	InternetGateways      map[string]*internetGatewayInfo

	NatGateways map[string]*ec2.NatGateway

	ids map[string]*idAllocator
}

var _ ec2iface.EC2API = &MockEC2{}

type idAllocator struct {
	NextId int
}

func (m *MockEC2) allocateId(prefix string) string {
	ids := m.ids[prefix]
	if ids == nil {
		if m.ids == nil {
			m.ids = make(map[string]*idAllocator)
		}
		ids = &idAllocator{NextId: 1}
		m.ids[prefix] = ids
	}
	id := ids.NextId
	ids.NextId++
	return fmt.Sprintf("%s-%d", prefix, id)
}
