// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package outputs

const (
	capacityOptimized = "capacity-optimized"
	typeASG           = "AWS::AutoScaling::AutoScalingGroup"
)

// Resources is a struct to represent json for a cloudformation Resources definition block.
type Resources struct {
	Resources map[string]AutoScalingGroup `json:"Resources"`
}

// AutoScalingGroup is a struct to represent json for a cloudformation ASG definition
type AutoScalingGroup struct {
	Type       string                     `json:"Type"`
	Properties AutoScalingGroupProperties `json:"Properties"`
}

// AutoScalingGroupProperties is a struct to represent json for a cloudformation ASG Properties definition
type AutoScalingGroupProperties struct {
	AutoScalingGroupName string               `json:"AutoScalingGroupName"`
	MinSize              int                  `json:"MinSize,string"`
	MaxSize              int                  `json:"MaxSize,string"`
	DesiredCapacity      int                  `json:"DesiredCapacity,string"`
	VPCZoneIdentifier    []string             `json:"VPCZoneIdentifier"`
	MixedInstancesPolicy MixedInstancesPolicy `json:"MixedInstancesPolicy"`
}

// MixedInstancesPolicy is a struct to represent json for a cloudformation ASG MixedInstancesPolicy definition
type MixedInstancesPolicy struct {
	InstancesDistribution InstancesDistribution `json:"InstancesDistribution"`
	LaunchTemplate        LaunchTemplate        `json:"LaunchTemplate"`
}

// InstancesDistribution is a struct to represent json for a cloudformation ASG MixedInstancesPolicy InstancesDistribution definition
type InstancesDistribution struct {
	OnDemandAllocationStrategy          string `json:"OnDemandAllocationStrategy,omitempty"`
	OnDemandBaseCapacity                int    `json:"OnDemandBaseCapacity"`
	OnDemandPercentageAboveBaseCapacity int    `json:"OnDemandPercentageAboveBaseCapacity"`
	SpotAllocationStrategy              string `json:"SpotAllocationStrategy,omitempty"`
	SpotInstancePools                   int    `json:"SpotInstancePools,omitempty"`
	SpotMaxPrice                        string `json:"SpotMaxPrice,omitempty"`
}

// LaunchTemplate is a struct to represent json for a cloudformation LaunchTemplate definition
type LaunchTemplate struct {
	LaunchTemplateSpecification LaunchTemplateSpecification `json:"LaunchTemplateSpecification"`
	Overrides                   []InstanceTypeOverride      `json:"Overrides"`
}

// LaunchTemplateSpecification is a struct to represent json for a cloudformation LaunchTemplate LaunchTemplateSpecification definition
type LaunchTemplateSpecification struct {
	LaunchTemplateID   string `json:"LaunchTemplateId,omitempty"`
	LaunchTemplateName string `json:"LaunchTemplateName,omitempty"`
	Version            string `json:"Version"`
}

// InstanceTypeOverride is a struct to represent json for a cloudformation LaunchTemplate LaunchTemplateSpecification InstanceTypeOverrides definition
type InstanceTypeOverride struct {
	InstanceType     string `json:"InstanceType"`
	WeightedCapacity int    `json:"WeightedCapacity,omitempty"`
}
