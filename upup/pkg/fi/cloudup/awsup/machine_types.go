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

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/klog/v2"
)

type AWSMachineTypeInfo struct {
	Name              string
	MemoryGB          float32
	Cores             int
	EphemeralDisks    []int
	GPU               bool
	MaxPods           int
	InstanceENIs      int
	InstanceIPsPerENI int
}

type EphemeralDevice struct {
	DeviceName  string
	VirtualName string
	SizeGB      int
}

var (
	machineTypeInfo  map[string]*AWSMachineTypeInfo
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

func GetMachineTypeInfo(c AWSCloud, machineType string) (*AWSMachineTypeInfo, error) {

	machineTypeMutex.Lock()
	defer machineTypeMutex.Unlock()
	if machineTypeInfo == nil {
		machineTypeInfo = make(map[string]*AWSMachineTypeInfo)
	} else if i, ok := machineTypeInfo[machineType]; ok {
		return i, nil
	}

	info, err := c.DescribeInstanceType(machineType)
	if err != nil {
		return nil, err
	}
	machine := AWSMachineTypeInfo{
		Name:              machineType,
		GPU:               info.GpuInfo != nil,
		InstanceENIs:      intValue(info.NetworkInfo.MaximumNetworkInterfaces),
		InstanceIPsPerENI: intValue(info.NetworkInfo.Ipv4AddressesPerInterface),
	}
	memoryGB := float64(intValue(info.MemoryInfo.SizeInMiB)) / 1024
	machine.MemoryGB = float32(math.Round(memoryGB*100) / 100)

	if info.VCpuInfo != nil && info.VCpuInfo.DefaultVCpus != nil {
		machine.Cores = intValue(info.VCpuInfo.DefaultVCpus)
	}
	if info.InstanceStorageInfo != nil && len(info.InstanceStorageInfo.Disks) > 0 {
		disks := make([]int, 0)
		for _, disk := range info.InstanceStorageInfo.Disks {
			for i := 0; i < intValue(disk.Count); i++ {
				disks = append(disks, intValue(disk.SizeInGB))
			}
		}
		machine.EphemeralDisks = disks
	}
	machineTypeInfo[machineType] = &machine

	return &machine, nil
}

func intValue(v *int64) int {
	return int(aws.Int64Value(v))
}
