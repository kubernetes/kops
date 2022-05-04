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

package selector

import (
	"encoding/json"
	"regexp"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/ec2pricing"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// InstanceTypesOutput can be implemented to provide custom output to instance type results
type InstanceTypesOutput interface {
	Output([]*instancetypes.Details) []string
}

// InstanceTypesOutputFn is the func type definition for InstanceTypesOuput
type InstanceTypesOutputFn func([]*instancetypes.Details) []string

// Output implements InstanceTypesOutput interface on InstanceTypesOutputFn
// This allows any InstanceTypesOutputFn to be passed into funcs accepting InstanceTypesOutput interface
func (fn InstanceTypesOutputFn) Output(instanceTypes []*instancetypes.Details) []string {
	return fn(instanceTypes)
}

// Selector is used to filter instance type resource specs
type Selector struct {
	EC2                   ec2iface.EC2API
	EC2Pricing            ec2pricing.EC2PricingIface
	InstanceTypesProvider *instancetypes.Provider
	ServiceRegistry       ServiceRegistry
}

// IntRangeFilter holds an upper and lower bound int
// The lower and upper bound are used to range filter resource specs
type IntRangeFilter struct {
	UpperBound int
	LowerBound int
}

// Uint64RangeFilter holds an upper and lower bound uint64
// The lower and upper bound are used to range filter resource specs
type Uint64RangeFilter struct {
	UpperBound uint64
	LowerBound uint64
}

// ByteQuantityRangeFilter holds an upper and lower bound byte quantity
// The lower and upper bound are used to range filter resource specs
type ByteQuantityRangeFilter struct {
	UpperBound bytequantity.ByteQuantity
	LowerBound bytequantity.ByteQuantity
}

// Float64RangeFilter holds an upper and lower bound float64
// The lower and upper bound are used to range filter resource specs
type Float64RangeFilter struct {
	UpperBound float64
	LowerBound float64
}

// filterPair holds a tuple of the passed in filter value and the instance resource spec value
type filterPair struct {
	filterValue  interface{}
	instanceSpec interface{}
}

func getRegexpString(r *regexp.Regexp) *string {
	if r == nil {
		return nil
	}
	rStr := r.String()
	return &rStr
}

// MarshalIndent is used to return a pretty-print json representation of a Filters struct
func (f *Filters) MarshalIndent(prefix, indent string) ([]byte, error) {
	type Alias Filters
	return json.MarshalIndent(&struct {
		AllowList *string
		DenyList  *string
		*Alias
	}{
		AllowList: getRegexpString(f.AllowList),
		DenyList:  getRegexpString(f.DenyList),
		Alias:     (*Alias)(f),
	}, prefix, indent)
}

// Filters is used to group instance type resource attributes for filtering
type Filters struct {
	// AvailabilityZones is the AWS Availability Zones where instances will be provisioned.
	// Instance type capacity can vary between availability zones.
	// Will accept zone names or ids
	// Example: us-east-1a, us-east-1b, us-east-2a, etc. OR use1-az1, use2-az2, etc.
	AvailabilityZones *[]string

	// BareMetal is used to only return bare metal instance type results
	BareMetal *bool

	// Burstable is used to only return burstable instance type results like the t* series
	Burstable *bool

	// AutoRecovery is used to filter by instance types that support auto recovery
	AutoRecovery *bool

	// FreeTier is used to filter by instance types that can be used as part of the EC2 free tier
	FreeTier *bool

	// CPUArchitecture of the EC2 instance type
	// Possible values are: x86_64/amd64 or arm64
	CPUArchitecture *string

	// CPUManufacturer is used to filter instance types with a specific CPU manufacturer
	CPUManufacturer *string

	// CurrentGeneration returns the latest generation of instance types
	CurrentGeneration *bool

	// EnaSupport returns instances that can support an Elastic Network Adapter.
	EnaSupport *bool

	// EfaSupport returns instances that can support an Elastic Fabric Adapter.
	EfaSupport *bool

	// FPGA is used to only return FPGA instance type results
	Fpga *bool

	// GpusRange filter is a range of acceptable GPU count available to an EC2 instance type
	GpusRange *IntRangeFilter

	// GpuMemoryRange filter is a range of acceptable GPU memory in Gibibytes (GiB) available to an EC2 instance type in aggreagte across all GPUs.
	GpuMemoryRange *ByteQuantityRangeFilter

	// GPUManufacturer filters by GPU manufacturer
	GPUManufacturer *string

	// GPUModel filter by the GPU model name
	GPUModel *string

	// InferenceAcceleratorsRange filters inference accelerators available to the instance type
	InferenceAcceleratorsRange *IntRangeFilter

	// InferenceAcceleratorManufacturer filters by inference acceleartor manufacturer
	InferenceAcceleratorManufacturer *string

	// InferenceAcceleratorModel filters by inference accelerator model name
	InferenceAcceleratorModel *string

	// HibernationSupported denotes whether EC2 hibernate is supported
	// Possible values are: true or false
	HibernationSupported *bool

	// Hypervisor is used to return only a specific hypervisor backed instance type
	// Possibly values are: xen or nitro
	Hypervisor *string

	// MaxResults is the maximum number of instance types to return that match the filter criteria
	MaxResults *int

	// MemoryRange filter is a range of acceptable DRAM memory in Gibibytes (GiB) for the instance type
	MemoryRange *ByteQuantityRangeFilter

	// NetworkInterfaces filter is a range of the number of ENI attachments an instance type can support
	NetworkInterfaces *IntRangeFilter

	// NetworkPerformance filter is a range of network bandwidth an instance type can support
	NetworkPerformance *IntRangeFilter

	// NetworkEncryption filters for instance types that automatically encrypt network traffic in-transit
	NetworkEncryption *bool

	// IPv6 filters for instance types that support IPv6
	IPv6 *bool

	// PlacementGroupStrategy is used to return instance types based on its support
	// for a specific placement group strategy
	// Possible values are: cluster, spread, or partition
	PlacementGroupStrategy *string

	// Region is the AWS Region where instances will be provisioned.
	// Instance type availability can vary between AWS Regions.
	// Example: us-east-1, us-east-2, eu-west-1, etc.
	Region *string

	// RootDeviceType is the backing device of the root storage volume
	// Possible values are: instance-store or ebs
	RootDeviceType *string

	// UsageClass of the instance EC2 instance type
	// Possible values are: spot or on-demand
	UsageClass *string

	// VCpusRange filter is a range of acceptable VCpus for the instance type
	VCpusRange *IntRangeFilter

	// VcpusToMemoryRatio is a ratio of vcpus to memory expressed as a floating point
	VCpusToMemoryRatio *float64

	// AllowList is a regex of allowed instance types
	AllowList *regexp.Regexp

	// DenyList is a regex of excluded instance types
	DenyList *regexp.Regexp

	// InstanceTypeBase is a base instance type which is used to retrieve similarly spec'd instance types
	InstanceTypeBase *string

	// Flexible finds an opinionated set of general (c, m, r, t, a, etc.) instance types that match a criteria specified
	// or defaults to 4 vcpus
	Flexible *bool

	// Service filters instance types based on a service's supported list of instance types
	// Example: eks or emr
	Service *string

	// InstanceTypes filters instance types and only allows instance types in this slice
	InstanceTypes *[]string

	// VirtualizationType is used to return instance types that match either hvm or pv virtualization types
	VirtualizationType *string

	// PricePerHour is used to return instance types that are equal to or cheaper than the specified price
	PricePerHour *Float64RangeFilter

	// InstanceStorageRange filters on a range of storage available as local disk
	InstanceStorageRange *ByteQuantityRangeFilter

	// DiskType is the backing storage medium
	// Possible values are: hdd or ssd
	DiskType *string

	// NVME filters for NVME disks, including both EBS and local instance storage
	NVME *bool

	// EBSOptimized filters for instance types that support EBS Optimized
	EBSOptimized *bool

	// DiskEncryption filters for instance types that support EBS Encryption or local storage encryption
	DiskEncryption *bool

	// EBSOptimizedBaselineBandwidth filters on a range of bandwidth that an EBS Optimized volume supports
	EBSOptimizedBaselineBandwidth *ByteQuantityRangeFilter

	// EBSOptimizedBaselineThroughput filters on a range of throughput that an EBS Optimized volume supports
	EBSOptimizedBaselineThroughput *ByteQuantityRangeFilter

	// EBSOptimizedBaselineIOPS filters on a range of IOPS that an EBS Optimized volume supports
	EBSOptimizedBaselineIOPS *IntRangeFilter

	// DedicatedHosts filters on instance types that support dedicated hosts tenancy
	DedicatedHosts *bool
}
