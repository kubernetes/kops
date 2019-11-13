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

package awstasks

import (
	"errors"
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// buildEphemeralDevices looks up the machine type and discovery any ephemeral device mappings
func buildEphemeralDevices(cloud awsup.AWSCloud, machineType string) (map[string]*BlockDeviceMapping, error) {
	mt, err := awsup.GetMachineTypeInfo(machineType)
	if err != nil {
		return nil, fmt.Errorf("failed to find instance type details on: %s, error: %s", machineType, err)
	}

	blockDeviceMappings := make(map[string]*BlockDeviceMapping)

	for _, ed := range mt.EphemeralDevices() {
		blockDeviceMappings[ed.DeviceName] = &BlockDeviceMapping{VirtualName: fi.String(ed.VirtualName)}
	}

	return blockDeviceMappings, nil
}

// buildAdditionalDevices is responsible for creating additional volumes in this lc
func buildAdditionalDevices(volumes []*BlockDeviceMapping) (map[string]*BlockDeviceMapping, error) {
	devices := make(map[string]*BlockDeviceMapping)

	// @step: iterate the volumes and create devices from them
	for _, x := range volumes {
		if x.DeviceName == nil {
			return nil, errors.New("DeviceName not set for volume")
		}
		devices[*x.DeviceName] = x
	}

	return devices, nil
}
