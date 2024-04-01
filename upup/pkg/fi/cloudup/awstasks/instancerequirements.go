/*
Copyright 2021 The Kubernetes Authors.

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
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"

	"k8s.io/kops/upup/pkg/fi"
)

type InstanceRequirements struct {
	Architecture *string
	CPUMin       *int32
	CPUMax       *int32
	MemoryMin    *int32
	MemoryMax    *int32
}

var _ fi.CloudupHasDependencies = &InstanceRequirements{}

func (e *InstanceRequirements) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	return nil
}

func findInstanceRequirements(asg *autoscalingtypes.AutoScalingGroup) (*InstanceRequirements, error) {
	actual := &InstanceRequirements{}
	if asg.MixedInstancesPolicy != nil {
		for _, override := range asg.MixedInstancesPolicy.LaunchTemplate.Overrides {
			if override.InstanceRequirements != nil {
				if override.InstanceRequirements.VCpuCount != nil {
					actual.CPUMax = override.InstanceRequirements.VCpuCount.Max
					actual.CPUMin = override.InstanceRequirements.VCpuCount.Min
				}
				if override.InstanceRequirements.MemoryMiB != nil {
					actual.MemoryMax = override.InstanceRequirements.MemoryMiB.Max
					actual.MemoryMax = override.InstanceRequirements.MemoryMiB.Min
				}
				return actual, nil
			}
		}
	}
	return nil, nil
}

func overridesFromInstanceRequirements(ir *InstanceRequirements) autoscalingtypes.LaunchTemplateOverrides {
	return autoscalingtypes.LaunchTemplateOverrides{
		InstanceRequirements: &autoscalingtypes.InstanceRequirements{
			VCpuCount: &autoscalingtypes.VCpuCountRequest{
				Max: ir.CPUMax,
				Min: ir.CPUMin,
			},
			MemoryMiB: &autoscalingtypes.MemoryMiBRequest{
				Max: ir.MemoryMax,
				Min: ir.MemoryMin,
			},
			BurstablePerformance: autoscalingtypes.BurstablePerformanceIncluded,
		},
	}
}
