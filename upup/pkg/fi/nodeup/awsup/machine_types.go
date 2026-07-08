/*
Copyright 2026 The Kubernetes Authors.

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

package awsup

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MachineTypeInfo holds the subset of instance type information used by nodeup.
type MachineTypeInfo struct {
	GPU               bool
	InstanceENIs      int32
	InstanceIPsPerENI int32
}

// GetMachineTypeInfo calls ec2.DescribeInstanceTypes to get information for a particular instance type, caching the result.
func (c *Cloud) GetMachineTypeInfo(ctx context.Context, machineType ec2types.InstanceType) (*MachineTypeInfo, error) {
	clients.mutex.Lock()
	defer clients.mutex.Unlock()

	if info, ok := clients.machineTypes[machineType]; ok {
		return info, nil
	}

	resp, err := clients.ec2.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []ec2types.InstanceType{machineType},
	})
	if err != nil {
		return nil, fmt.Errorf("describing instance type %q in region %q: %w", machineType, c.region, err)
	}
	if len(resp.InstanceTypes) != 1 {
		return nil, fmt.Errorf("instance type %q not found in region %q", machineType, c.region)
	}
	info := resp.InstanceTypes[0]

	machine := &MachineTypeInfo{
		GPU:               info.GpuInfo != nil,
		InstanceENIs:      aws.ToInt32(info.NetworkInfo.MaximumNetworkInterfaces),
		InstanceIPsPerENI: aws.ToInt32(info.NetworkInfo.Ipv4AddressesPerInterface),
	}

	if clients.machineTypes == nil {
		clients.machineTypes = make(map[ec2types.InstanceType]*MachineTypeInfo)
	}
	clients.machineTypes[machineType] = machine

	return machine, nil
}
