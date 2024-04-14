package selector

import (
	"context"
	"fmt"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"regexp"
)

const (
	// AggregateLowPercentile is the default lower percentile for resource ranges on similar instance type comparisons
	AggregateLowPercentile = 0.9
	// AggregateHighPercentile is the default upper percentile for resource ranges on similar instance type comparisons
	AggregateHighPercentile = 1.2
)

// FiltersTransform can be implemented to provide custom transforms
type FiltersTransform interface {
	Transform(context.Context, Filters) (Filters, error)
}

// TransformFn is the func type definition for a FiltersTransform
type TransformFn func(context.Context, Filters) (Filters, error)

// Transform implements FiltersTransform interface on TransformFn
// This allows any TransformFn to be passed into funcs accepting FiltersTransform interface
func (fn TransformFn) Transform(ctx context.Context, filters Filters) (Filters, error) {
	return fn(ctx, filters)
}

// TransformBaseInstanceType transforms lower level filters based on the instanceTypeBase specs
func (itf Selector) TransformBaseInstanceType(ctx context.Context, filters Filters) (Filters, error) {
	if filters.InstanceTypeBase == nil {
		return filters, nil
	}
	instanceTypesOutput, err := itf.EC2.DescribeInstanceTypes(ctx, &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []ec2types.InstanceType{
			ec2types.InstanceType(*filters.InstanceTypeBase),
		},
	})
	if err != nil {
		return filters, err
	}
	if len(instanceTypesOutput.InstanceTypes) == 0 {
		return filters, fmt.Errorf("error instance type %s is not a valid instance type", *filters.InstanceTypeBase)
	}
	instanceTypeInfo := instanceTypesOutput.InstanceTypes[0]
	if filters.BareMetal == nil {
		filters.BareMetal = instanceTypeInfo.BareMetal
	}
	if filters.CPUArchitecture == nil && len(instanceTypeInfo.ProcessorInfo.SupportedArchitectures) == 1 {
		filters.CPUArchitecture = &instanceTypeInfo.ProcessorInfo.SupportedArchitectures[0]
	}
	if filters.Fpga == nil {
		isFpgaSupported := instanceTypeInfo.FpgaInfo != nil
		filters.Fpga = &isFpgaSupported
	}
	if filters.GpusRange == nil {
		gpuCount := int32(0)
		if instanceTypeInfo.GpuInfo != nil {
			gpuCount = *getTotalGpusCount(instanceTypeInfo.GpuInfo)
		}
		filters.GpusRange = &Int32RangeFilter{LowerBound: gpuCount, UpperBound: gpuCount}
	}
	if filters.MemoryRange == nil {
		lowerBound := bytequantity.ByteQuantity{Quantity: uint64(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB) * AggregateLowPercentile)}
		upperBound := bytequantity.ByteQuantity{Quantity: uint64(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB) * AggregateHighPercentile)}
		filters.MemoryRange = &ByteQuantityRangeFilter{LowerBound: lowerBound, UpperBound: upperBound}
	}
	if filters.VCpusRange == nil {
		lowerBound := int32(float32(*instanceTypeInfo.VCpuInfo.DefaultVCpus) * AggregateLowPercentile)
		upperBound := int32(float32(*instanceTypeInfo.VCpuInfo.DefaultVCpus) * AggregateHighPercentile)
		filters.VCpusRange = &Int32RangeFilter{LowerBound: lowerBound, UpperBound: upperBound}
	}
	if filters.VirtualizationType == nil && len(instanceTypeInfo.SupportedVirtualizationTypes) == 1 {
		filters.VirtualizationType = &instanceTypeInfo.SupportedVirtualizationTypes[0]
	}
	filters.InstanceTypeBase = nil

	return filters, nil
}

// TransformFlexible transforms lower level filters based on a set of opinions
func (itf Selector) TransformFlexible(ctx context.Context, filters Filters) (Filters, error) {
	if filters.Flexible == nil {
		return filters, nil
	}
	if filters.CPUArchitecture == nil {
		defaultArchitecture := ec2types.ArchitectureTypeX8664
		filters.CPUArchitecture = &defaultArchitecture
	}
	if filters.BareMetal == nil {
		bareMetalDefault := false
		filters.BareMetal = &bareMetalDefault
	}
	if filters.Fpga == nil {
		fpgaDefault := false
		filters.Fpga = &fpgaDefault
	}

	if filters.AllowList == nil {
		baseAllowedInstanceTypes, err := regexp.Compile("^[cmr][3-9][ag]?\\..*$|^a[1-9]\\..*$|^t[2-9]\\..*$")
		if err != nil {
			return filters, err
		}
		filters.AllowList = baseAllowedInstanceTypes
	}

	if filters.VCpusRange == nil && filters.MemoryRange == nil {
		defaultVcpus := int32(4)
		filters.VCpusRange = &Int32RangeFilter{LowerBound: defaultVcpus, UpperBound: defaultVcpus}
	}

	return filters, nil
}

// TransformForService transforms lower level filters based on the service
func (itf Selector) TransformForService(ctx context.Context, filters Filters) (Filters, error) {
	return itf.ServiceRegistry.ExecuteTransforms(filters)
}
