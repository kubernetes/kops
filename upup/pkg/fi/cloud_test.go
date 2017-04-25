/*
Copyright 2017 The Kubernetes Authors.

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

package fi

import (
	"github.com/aws/aws-sdk-go/awstesting"
	"testing"
)

func TestFindBiggestFreeSubnet(t *testing.T) {
	subnet := &SubnetInfo{
		ID:   "testsubnet",
		CIDR: "192.168.100.0/24",
		Zone: "ru-spb-narva",
	}
	vpcInfo := &VPCInfo{
		CIDR: "192.168.0.0/16",
	}
	vpcInfo.Subnets = append(vpcInfo.Subnets, subnet)

	result := vpcInfo.FindBiggestFreeSubnet()
	shouldBe := "192.168.0.0/18"
	awstesting.Match(t, result, shouldBe)
}

func TestFindBiggestFreeSubnet4(t *testing.T) {
	subnet := &SubnetInfo{
		ID:   "testsubnet",
		CIDR: "192.168.0.0/18",
		Zone: "ru-spb-narva",
	}
	subnet2 := &SubnetInfo{
		ID:   "testsubnet2",
		CIDR: "192.168.64.0/18",
		Zone: "ru-spb-narva",
	}
	subnet3 := &SubnetInfo{
		ID:   "testsubnet3",
		CIDR: "192.168.192.0/18",
		Zone: "ru-spb-narva",
	}
	vpcInfo := &VPCInfo{
		CIDR: "192.168.0.0/16",
	}
	vpcInfo.Subnets = append(vpcInfo.Subnets, subnet, subnet2, subnet3)

	result := vpcInfo.FindBiggestFreeSubnet()
	shouldBe := "192.168.128.0/18"
	awstesting.Match(t, result, shouldBe)
}
