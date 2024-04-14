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

package awsup

import (
	"fmt"
	"math"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

type AWSMachineTypeInfo struct {
	Name              ec2types.InstanceType
	MemoryGB          float32
	Cores             int32
	EphemeralDisks    []int64
	GPU               bool
	MaxPods           int32
	InstanceENIs      int32
	InstanceIPsPerENI int32
}

type EphemeralDevice struct {
	DeviceName  string
	VirtualName string
	SizeGB      int64
}

var (
	machineTypeInfo  map[ec2types.InstanceType]*AWSMachineTypeInfo
	machineTypeMutex sync.Mutex
)

func (m *AWSMachineTypeInfo) EphemeralDevices() []*EphemeralDevice {
	var disks []*EphemeralDevice
	for i, sizeGB := range m.EphemeralDisks {
		d := &EphemeralDevice{
			SizeGB: sizeGB,
		}

		if i >= 20 {
			// TODO: What drive letters do we use?
			klog.Fatalf("ephemeral devices for > 20 not yet implemented")
		}
		d.DeviceName = fmt.Sprintf("/dev/sd%c", 'c'+i)
		d.VirtualName = fmt.Sprintf("ephemeral%d", i)

		disks = append(disks, d)
	}
	return disks
}

func GetMachineTypeInfo(c AWSCloud, machineType ec2types.InstanceType) (*AWSMachineTypeInfo, error) {
	machineTypeMutex.Lock()
	defer machineTypeMutex.Unlock()
	if machineTypeInfo == nil {
		machineTypeInfo = make(map[ec2types.InstanceType]*AWSMachineTypeInfo)
	} else if i, ok := machineTypeInfo[machineType]; ok {
		return i, nil
	}

	info, err := c.DescribeInstanceType(string(machineType))
	if err != nil {
		return nil, err
	}
	machine := AWSMachineTypeInfo{
		Name:              machineType,
		GPU:               info.GpuInfo != nil,
		InstanceENIs:      aws.ToInt32(info.NetworkInfo.MaximumNetworkInterfaces),
		InstanceIPsPerENI: aws.ToInt32(info.NetworkInfo.Ipv4AddressesPerInterface),
	}
	memoryGB := float64(aws.ToInt64(info.MemoryInfo.SizeInMiB)) / 1024
	machine.MemoryGB = float32(math.Round(memoryGB*100) / 100)

	if info.VCpuInfo != nil && info.VCpuInfo.DefaultVCpus != nil {
		machine.Cores = aws.ToInt32(info.VCpuInfo.DefaultVCpus)
	}
	if info.InstanceStorageInfo != nil && len(info.InstanceStorageInfo.Disks) > 0 {
		disks := make([]int64, 0)
		for _, disk := range info.InstanceStorageInfo.Disks {
			for i := int32(0); i < aws.ToInt32(disk.Count); i++ {
				disks = append(disks, aws.ToInt64(disk.SizeInGB))
			}
		}
		machine.EphemeralDisks = disks
	}
	machineTypeInfo[machineType] = &machine

	return &machine, nil
}
