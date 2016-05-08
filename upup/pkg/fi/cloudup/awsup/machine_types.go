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
}

type EphemeralDevice struct {
	DeviceName  string
	VirtualName string
	SizeGB      int
}

func (m *AWSMachineTypeInfo) EphemeralDevices() ([]*EphemeralDevice, error) {
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
	return disks, nil
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
}
