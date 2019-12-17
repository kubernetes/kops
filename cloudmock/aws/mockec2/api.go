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

package mockec2

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type MockEC2 struct {
	// Stub out interface
	ec2iface.EC2API

	mutex sync.Mutex

	addressNumber int
	Addresses     map[string]*ec2.Address

	RouteTables map[string]*ec2.RouteTable

	DhcpOptions map[string]*ec2.DhcpOptions

	Images []*ec2.Image

	securityGroupNumber int
	SecurityGroups      map[string]*ec2.SecurityGroup

	subnets map[string]*subnetInfo

	Volumes map[string]*ec2.Volume

	KeyPairs map[string]*ec2.KeyPairInfo

	Tags []*ec2.TagDescription

	Vpcs map[string]*vpcInfo

	InternetGateways map[string]*ec2.InternetGateway

	LaunchTemplates map[string]*ec2.ResponseLaunchTemplateData

	NatGateways map[string]*ec2.NatGateway

	idsMutex sync.Mutex
	ids      map[string]*idAllocator
}

var _ ec2iface.EC2API = &MockEC2{}

func (m *MockEC2) All() map[string]interface{} {
	all := make(map[string]interface{})

	for _, o := range m.Addresses {
		all[aws.StringValue(o.AllocationId)] = o
	}
	for id, o := range m.RouteTables {
		all[id] = o
	}
	for id, o := range m.DhcpOptions {
		all[id] = o
	}
	for _, o := range m.Images {
		all[aws.StringValue(o.ImageId)] = o
	}
	for id, o := range m.SecurityGroups {
		all[id] = o
	}
	for id, o := range m.subnets {
		all[id] = &o.main
	}
	for id, o := range m.Volumes {
		all[id] = o
	}
	for id, o := range m.KeyPairs {
		all["sshkey-"+id] = o
	}
	for id, o := range m.Vpcs {
		all[id] = o
	}
	for id, o := range m.InternetGateways {
		all[id] = o
	}
	for id, o := range m.NatGateways {
		all[id] = o
	}

	return all
}

type idAllocator struct {
	NextId int
}

func (m *MockEC2) allocateId(prefix string) string {
	m.idsMutex.Lock()
	defer m.idsMutex.Unlock()

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
