package selector

import (
	"fmt"
	"regexp"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/bytequantity"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// AggregateLowPercentile is the default lower percentile for resource ranges on similar instance type comparisons
	AggregateLowPercentile = 0.9
	// AggregateHighPercentile is the default upper percentile for resource ranges on similar instance type comparisons
	AggregateHighPercentile = 1.2
)

// FiltersTransform can be implemented to provide custom transforms
type FiltersTransform interface {
	Transform(Filters) (Filters, error)
}

// TransformFn is the func type definition for a FiltersTransform
type TransformFn func(Filters) (Filters, error)

// Transform implements FiltersTransform interface on TransformFn
// This allows any TransformFn to be passed into funcs accepting FiltersTransform interface
func (fn TransformFn) Transform(filters Filters) (Filters, error) {
	return fn(filters)
}

// TransformBaseInstanceType transforms lower level filters based on the instanceTypeBase specs
func (itf Selector) TransformBaseInstanceType(filters Filters) (Filters, error) {
	if filters.InstanceTypeBase == nil {
		return filters, nil
	}
	instanceTypesOutput, err := itf.EC2.DescribeInstanceTypes(&ec2.DescribeInstanceTypesInput{
		InstanceTypes: []*string{filters.InstanceTypeBase},
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
	if filters.CPUArchitecture == nil {
		filters.CPUArchitecture = instanceTypeInfo.ProcessorInfo.SupportedArchitectures[0]
	}
	if filters.Fpga == nil {
		isFpgaSupported := instanceTypeInfo.FpgaInfo != nil
		filters.Fpga = &isFpgaSupported
	}
	if filters.GpusRange == nil {
		gpuCount := 0
		if instanceTypeInfo.GpuInfo != nil {
			gpuCount = int(*getTotalGpusCount(instanceTypeInfo.GpuInfo))
		}
		filters.GpusRange = &IntRangeFilter{LowerBound: gpuCount, UpperBound: gpuCount}
	}
	if filters.MemoryRange == nil {
		lowerBound := bytequantity.ByteQuantity{Quantity: uint64(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB) * AggregateLowPercentile)}
		upperBound := bytequantity.ByteQuantity{Quantity: uint64(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB) * AggregateHighPercentile)}
		filters.MemoryRange = &ByteQuantityRangeFilter{LowerBound: lowerBound, UpperBound: upperBound}
	}
	if filters.VCpusRange == nil {
		lowerBound := int(float64(*instanceTypeInfo.VCpuInfo.DefaultVCpus) * AggregateLowPercentile)
		upperBound := int(float64(*instanceTypeInfo.VCpuInfo.DefaultVCpus) * AggregateHighPercentile)
		filters.VCpusRange = &IntRangeFilter{LowerBound: lowerBound, UpperBound: upperBound}
	}
	filters.InstanceTypeBase = nil

	return filters, nil
}

// TransformFlexible transforms lower level filters based on a set of opinions
func (itf Selector) TransformFlexible(filters Filters) (Filters, error) {
	if filters.Flexible == nil {
		return filters, nil
	}
	if filters.CPUArchitecture == nil {
		filters.CPUArchitecture = aws.String("x86_64")
	}
	if filters.BareMetal == nil {
		filters.BareMetal = aws.Bool(false)
	}
	if filters.Fpga == nil {
		filters.Fpga = aws.Bool(false)
	}

	if filters.AllowList == nil {
		baseAllowedInstanceTypes, err := regexp.Compile("^[cmr][3-9][ag]?\\..*$|^a[1-9]\\..*$|^t[2-9]\\..*$")
		if err != nil {
			return filters, err
		}
		filters.AllowList = baseAllowedInstanceTypes
	}

	if filters.VCpusRange == nil && filters.MemoryRange == nil {
		defaultVcpus := 4
		filters.VCpusRange = &IntRangeFilter{LowerBound: defaultVcpus, UpperBound: defaultVcpus}
	}

	return filters, nil
}
