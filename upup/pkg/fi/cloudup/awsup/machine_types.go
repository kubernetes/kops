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

package awsup

import (
	"fmt"

	"github.com/golang/glog"
)

// I believe one vCPU ~ 3 ECUS, and 60 CPU credits would be needed to use one vCPU for an hour
const BurstableCreditsToECUS float32 = 3.0 / 60.0

type AWSMachineTypeInfo struct {
	Name           string
	MemoryGB       float32
	ECU            float32
	Cores          int
	EphemeralDisks []int
	Burstable      bool
	GPU            bool
}

type EphemeralDevice struct {
	DeviceName  string
	VirtualName string
	SizeGB      int
}

func (m *AWSMachineTypeInfo) EphemeralDevices() []*EphemeralDevice {
	var disks []*EphemeralDevice
	for i, sizeGB := range m.EphemeralDisks {
		d := &EphemeralDevice{
			SizeGB: sizeGB,
		}

		if i >= 20 {
			// TODO: What drive letters do we use?
			glog.Fatalf("ephemeral devices for > 20 not yet implemented")
		}
		d.DeviceName = "/dev/sd" + string('c'+i)
		d.VirtualName = fmt.Sprintf("ephemeral%d", i)

		disks = append(disks, d)
	}
	return disks
}

func GetMachineTypeInfo(machineType string) (*AWSMachineTypeInfo, error) {
	for i := range MachineTypes {
		m := &MachineTypes[i]
		if m.Name == machineType {
			return m, nil
		}
	}

	return nil, fmt.Errorf("instance type not handled: %q", machineType)
}

var MachineTypes []AWSMachineTypeInfo = []AWSMachineTypeInfo{
	// This is tedious, but seems simpler than trying to have some logic and then a lot of exceptions

	// t2 family
	{
		Name:           "t2.nano",
		MemoryGB:       0.5,
		ECU:            3 * BurstableCreditsToECUS,
		Cores:          1,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.micro",
		MemoryGB:       1,
		ECU:            6 * BurstableCreditsToECUS,
		Cores:          1,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.small",
		MemoryGB:       2,
		ECU:            12 * BurstableCreditsToECUS,
		Cores:          1,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.medium",
		MemoryGB:       4,
		ECU:            24 * BurstableCreditsToECUS,
		Cores:          2,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.large",
		MemoryGB:       8,
		ECU:            36 * BurstableCreditsToECUS,
		Cores:          2,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.xlarge",
		MemoryGB:       16,
		ECU:            54 * BurstableCreditsToECUS,
		Cores:          4,
		EphemeralDisks: nil,
		Burstable:      true,
	},
	{
		Name:           "t2.2xlarge",
		MemoryGB:       32,
		ECU:            81 * BurstableCreditsToECUS,
		Cores:          8,
		EphemeralDisks: nil,
		Burstable:      true,
	},

	// m3 family
	{
		Name:           "m3.medium",
		MemoryGB:       3.75,
		ECU:            3,
		Cores:          1,
		EphemeralDisks: []int{4},
	},
	{
		Name:           "m3.large",
		MemoryGB:       7.5,
		ECU:            6.5,
		Cores:          2,
		EphemeralDisks: []int{32},
	},
	{
		Name:           "m3.xlarge",
		MemoryGB:       15,
		ECU:            13,
		Cores:          4,
		EphemeralDisks: []int{40, 40},
	},
	{
		Name:           "m3.2xlarge",
		MemoryGB:       30,
		ECU:            26,
		Cores:          8,
		EphemeralDisks: []int{80, 80},
	},

	// m4 family
	{
		Name:           "m4.large",
		MemoryGB:       8,
		ECU:            6.5,
		Cores:          2,
		EphemeralDisks: nil,
	},
	{
		Name:           "m4.xlarge",
		MemoryGB:       16,
		ECU:            13,
		Cores:          4,
		EphemeralDisks: nil,
	},
	{
		Name:           "m4.2xlarge",
		MemoryGB:       32,
		ECU:            26,
		Cores:          8,
		EphemeralDisks: nil,
	},
	{
		Name:           "m4.4xlarge",
		MemoryGB:       64,
		ECU:            53.5,
		Cores:          16,
		EphemeralDisks: nil,
	},
	{
		Name:           "m4.10xlarge",
		MemoryGB:       160,
		ECU:            124.5,
		Cores:          40,
		EphemeralDisks: nil,
	},
	{
		Name:           "m4.16xlarge",
		MemoryGB:       256,
		ECU:            188,
		Cores:          64,
		EphemeralDisks: nil,
	},

	// c3 family
	{
		Name:           "c3.large",
		MemoryGB:       3.75,
		ECU:            7,
		Cores:          2,
		EphemeralDisks: []int{16, 16},
	},
	{
		Name:           "c3.xlarge",
		MemoryGB:       7.5,
		ECU:            14,
		Cores:          4,
		EphemeralDisks: []int{40, 40},
	},
	{
		Name:           "c3.2xlarge",
		MemoryGB:       15,
		ECU:            28,
		Cores:          8,
		EphemeralDisks: []int{80, 80},
	},
	{
		Name:           "c3.4xlarge",
		MemoryGB:       30,
		ECU:            55,
		Cores:          16,
		EphemeralDisks: []int{160, 160},
	},
	{
		Name:           "c3.8xlarge",
		MemoryGB:       60,
		ECU:            108,
		Cores:          32,
		EphemeralDisks: []int{320, 320},
	},

	// c4 family
	{
		Name:           "c4.large",
		MemoryGB:       3.75,
		ECU:            8,
		Cores:          2,
		EphemeralDisks: nil,
	},
	{
		Name:           "c4.xlarge",
		MemoryGB:       7.5,
		ECU:            16,
		Cores:          4,
		EphemeralDisks: nil,
	},
	{
		Name:           "c4.2xlarge",
		MemoryGB:       15,
		ECU:            31,
		Cores:          8,
		EphemeralDisks: nil,
	},
	{
		Name:           "c4.4xlarge",
		MemoryGB:       30,
		ECU:            62,
		Cores:          16,
		EphemeralDisks: nil,
	},
	{
		Name:           "c4.8xlarge",
		MemoryGB:       60,
		ECU:            132,
		Cores:          32,
		EphemeralDisks: nil,
	},

	// cc2 family
	{
		Name:           "cc2.8xlarge",
		MemoryGB:       60.5,
		ECU:            88,
		Cores:          32,
		EphemeralDisks: []int{840, 840, 840, 840},
	},

	// cg1 family
	{
		Name:           "cg1.4xlarge",
		MemoryGB:       22.5,
		ECU:            33.5,
		Cores:          16,
		EphemeralDisks: []int{840, 840},
		GPU:            true,
	},

	// cr1 family
	{
		Name:           "cr1.8xlarge",
		MemoryGB:       244.0,
		ECU:            88,
		Cores:          32,
		EphemeralDisks: []int{120, 120},
	},

	// d2 family
	{
		Name:           "d2.xlarge",
		MemoryGB:       30.5,
		ECU:            14,
		Cores:          4,
		EphemeralDisks: []int{2000, 2000, 2000},
	},
	{
		Name:           "d2.2xlarge",
		MemoryGB:       61.0,
		ECU:            28,
		Cores:          8,
		EphemeralDisks: []int{2000, 2000, 2000, 2000, 2000, 2000},
	},
	{
		Name:     "d2.4xlarge",
		MemoryGB: 122.0,
		ECU:      56,
		Cores:    16,
		EphemeralDisks: []int{
			2000, 2000, 2000, 2000, 2000, 2000,
			2000, 2000, 2000, 2000, 2000, 2000,
		},
	},
	{
		Name:     "d2.8xlarge",
		MemoryGB: 244.0,
		ECU:      116,
		Cores:    36,
		EphemeralDisks: []int{
			2000, 2000, 2000, 2000, 2000, 2000,
			2000, 2000, 2000, 2000, 2000, 2000,
			2000, 2000, 2000, 2000, 2000, 2000,
			2000, 2000, 2000, 2000, 2000, 2000,
		},
	},

	// g2 family
	{
		Name:           "g2.2xlarge",
		MemoryGB:       15,
		ECU:            26,
		Cores:          8,
		EphemeralDisks: []int{60},
		GPU:            true,
	},
	{
		Name:           "g2.8xlarge",
		MemoryGB:       60,
		ECU:            104,
		Cores:          32,
		EphemeralDisks: []int{120, 120},
		GPU:            true,
	},

	// hi1 family
	{
		Name:           "hi1.4xlarge",
		MemoryGB:       60.5,
		ECU:            35,
		Cores:          16,
		EphemeralDisks: []int{1024, 1024},
	},

	// i2 family
	{
		Name:           "i2.xlarge",
		MemoryGB:       30.5,
		ECU:            14,
		Cores:          4,
		EphemeralDisks: []int{800},
	},
	{
		Name:           "i2.2xlarge",
		MemoryGB:       61,
		ECU:            27,
		Cores:          8,
		EphemeralDisks: []int{800, 800},
	},
	{
		Name:           "i2.4xlarge",
		MemoryGB:       122,
		ECU:            53,
		Cores:          16,
		EphemeralDisks: []int{800, 800, 800, 800},
	},
	{
		Name:           "i2.8xlarge",
		MemoryGB:       244,
		ECU:            104,
		Cores:          32,
		EphemeralDisks: []int{800, 800, 800, 800, 800, 800, 800, 800},
	},

	// i3 family
	{
		Name:           "i3.large",
		MemoryGB:       15.25,
		ECU:            9,
		Cores:          2,
		EphemeralDisks: []int{475},
	},
	{
		Name:           "i3.xlarge",
		MemoryGB:       30.5,
		ECU:            14,
		Cores:          4,
		EphemeralDisks: []int{950},
	},
	{
		Name:           "i3.2xlarge",
		MemoryGB:       61,
		ECU:            27,
		Cores:          8,
		EphemeralDisks: []int{1900},
	},
	{
		Name:           "i3.4xlarge",
		MemoryGB:       122,
		ECU:            53,
		Cores:          16,
		EphemeralDisks: []int{1900, 1900},
	},
	{
		Name:           "i3.8xlarge",
		MemoryGB:       244,
		ECU:            104,
		Cores:          32,
		EphemeralDisks: []int{1900, 1900, 1900, 1900},
	},
	{
		Name:           "i3.16xlarge",
		MemoryGB:       488,
		ECU:            208,
		Cores:          64,
		EphemeralDisks: []int{1900, 1900, 1900, 1900, 1900, 1900, 1900, 1900},
	},

	// r3 family
	{
		Name:           "r3.large",
		MemoryGB:       15.25,
		ECU:            6.5,
		Cores:          2,
		EphemeralDisks: []int{32},
	},
	{
		Name:           "r3.xlarge",
		MemoryGB:       30.5,
		ECU:            13,
		Cores:          4,
		EphemeralDisks: []int{80},
	},
	{
		Name:           "r3.2xlarge",
		MemoryGB:       61,
		ECU:            26,
		Cores:          8,
		EphemeralDisks: []int{160},
	},
	{
		Name:           "r3.4xlarge",
		MemoryGB:       122,
		ECU:            52,
		Cores:          16,
		EphemeralDisks: []int{320},
	},
	{
		Name:           "r3.8xlarge",
		MemoryGB:       244,
		ECU:            104,
		Cores:          32,
		EphemeralDisks: []int{320, 320},
	},

	// x1 family
	{
		Name:           "x1.32xlarge",
		MemoryGB:       1952,
		ECU:            349,
		Cores:          128,
		EphemeralDisks: []int{1920, 1920},
	},

	// r4 family
	{
		Name:           "r4.large",
		MemoryGB:       15.25,
		ECU:            7,
		Cores:          2,
		EphemeralDisks: nil,
	},
	{
		Name:           "r4.xlarge",
		MemoryGB:       30.5,
		ECU:            13.5,
		Cores:          4,
		EphemeralDisks: nil,
	},
	{
		Name:           "r4.2xlarge",
		MemoryGB:       61,
		ECU:            27,
		Cores:          8,
		EphemeralDisks: nil,
	},
	{
		Name:           "r4.4xlarge",
		MemoryGB:       122,
		ECU:            53,
		Cores:          16,
		EphemeralDisks: nil,
	},
	{
		Name:           "r4.8xlarge",
		MemoryGB:       244,
		ECU:            99,
		Cores:          32,
		EphemeralDisks: nil,
	},
	{
		Name:           "r4.16xlarge",
		MemoryGB:       488,
		ECU:            195,
		Cores:          64,
		EphemeralDisks: nil,
	},

	// p2 family
	{
		Name:           "p2.xlarge",
		MemoryGB:       61,
		ECU:            12,
		Cores:          4,
		EphemeralDisks: nil,
		GPU:            true,
	},
	{
		Name:           "p2.8xlarge",
		MemoryGB:       488,
		ECU:            94,
		Cores:          32,
		EphemeralDisks: nil,
		GPU:            true,
	},
	{
		Name:           "p2.16xlarge",
		MemoryGB:       732,
		ECU:            188,
		Cores:          64,
		EphemeralDisks: nil,
		GPU:            true,
	},

	// g3 family
	{
		Name:           "g3.4xlarge",
		MemoryGB:       122,
		Cores:          16,
		ECU:            47,
		EphemeralDisks: nil,
		GPU:            true,
	},
	{
		Name:           "g3.8xlarge",
		MemoryGB:       244,
		ECU:            94,
		Cores:          32,
		EphemeralDisks: nil,
		GPU:            true,
	},
	{
		Name:           "g3.16xlarge",
		MemoryGB:       488,
		ECU:            188,
		Cores:          64,
		EphemeralDisks: nil,
		GPU:            true,
	},
}
