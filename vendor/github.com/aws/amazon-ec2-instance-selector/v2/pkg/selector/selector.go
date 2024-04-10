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

// Package selector provides filtering logic for Amazon EC2 Instance Types based on declarative resource specfications.
package selector

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/ec2pricing"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector/outputs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"
)

var (
	// Version is overridden at compilation with the version based on the git tag
	versionID = "dev"
)

const (
	locationFilterKey      = "location"
	zoneIDLocationType     = ec2types.LocationTypeAvailabilityZoneId
	zoneNameLocationType   = ec2types.LocationTypeAvailabilityZone
	regionNameLocationType = ec2types.LocationTypeRegion
	sdkName                = "instance-selector"

	// Filter Keys

	cpuArchitecture                  = "cpuArchitecture"
	cpuManufacturer                  = "cpuManufacturer"
	usageClass                       = "usageClass"
	rootDeviceType                   = "rootDeviceType"
	hibernationSupported             = "hibernationSupported"
	vcpusRange                       = "vcpusRange"
	memoryRange                      = "memoryRange"
	gpuMemoryRange                   = "gpuMemoryRange"
	gpusRange                        = "gpusRange"
	gpuManufacturer                  = "gpuManufacturer"
	gpuModel                         = "gpuModel"
	inferenceAcceleratorsRange       = "inferenceAcceleratorsRange"
	inferenceAcceleratorManufacturer = "inferenceAcceleartorManufacturer"
	inferenceAcceleratorModel        = "inferenceAcceleratorModel"
	placementGroupStrategy           = "placementGroupStrategy"
	hypervisor                       = "hypervisor"
	baremetal                        = "baremetal"
	burstable                        = "burstable"
	fpga                             = "fpga"
	enaSupport                       = "enaSupport"
	efaSupport                       = "efaSupport"
	vcpusToMemoryRatio               = "vcpusToMemoryRatio"
	currentGeneration                = "currentGeneration"
	networkInterfaces                = "networkInterfaces"
	networkPerformance               = "networkPerformance"
	networkEncryption                = "networkEncryption"
	ipv6                             = "ipv6"
	allowList                        = "allowList"
	denyList                         = "denyList"
	instanceTypes                    = "instanceTypes"
	virtualizationType               = "virtualizationType"
	instanceStorageRange             = "instanceStorageRange"
	diskEncryption                   = "diskEncryption"
	diskType                         = "diskType"
	nvme                             = "nvme"
	ebsOptimized                     = "ebsOptimized"
	ebsOptimizedBaselineBandwidth    = "ebsOptimizedBaselineBandwidth"
	ebsOptimizedBaselineIOPS         = "ebsOptimizedBaselineIOPS"
	ebsOptimizedBaselineThroughput   = "ebsOptimizedBaselineThroughput"
	freeTier                         = "freeTier"
	autoRecovery                     = "autoRecovery"
	dedicatedHosts                   = "dedicatedHosts"

	cpuArchitectureAMD64 = "amd64"

	virtualizationTypePV = "pv"

	pricePerHour = "pricePerHour"
)

// New creates an instance of Selector provided an aws session
func New(ctx context.Context, cfg aws.Config) (*Selector, error) {
	serviceRegistry := NewRegistry()
	serviceRegistry.RegisterAWSServices()
	ec2Client := ec2.NewFromConfig(cfg, func(options *ec2.Options) {
		options.APIOptions = append(options.APIOptions, middleware.AddUserAgentKeyValue(sdkName, versionID))
	})
	pricingClient, err := ec2pricing.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Selector{
		EC2:                   ec2Client,
		EC2Pricing:            pricingClient,
		InstanceTypesProvider: instancetypes.LoadFromOrNew("", cfg.Region, 0, ec2Client),
		ServiceRegistry:       serviceRegistry,
	}, nil
}

func NewWithCache(ctx context.Context, cfg aws.Config, ttl time.Duration, cacheDir string) (*Selector, error) {
	serviceRegistry := NewRegistry()
	serviceRegistry.RegisterAWSServices()
	ec2Client := ec2.NewFromConfig(cfg, func(options *ec2.Options) {
		options.APIOptions = append(options.APIOptions, middleware.AddUserAgentKeyValue(sdkName, versionID))
	})
	pricingClient, err := ec2pricing.NewWithCache(ctx, cfg, ttl, cacheDir)
	if err != nil {
		return nil, err
	}

	return &Selector{
		EC2:                   ec2Client,
		EC2Pricing:            pricingClient,
		InstanceTypesProvider: instancetypes.LoadFromOrNew(cacheDir, cfg.Region, ttl, ec2Client),
		ServiceRegistry:       serviceRegistry,
	}, nil
}

func (itf Selector) Save() error {
	return multierr.Append(itf.EC2Pricing.Save(), itf.InstanceTypesProvider.Save())
}

// Filter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a simple list of instance type strings
//
// Deprecated: This function will be replaced with GetFilteredInstanceTypes() and
// OutputInstanceTypes() in the next major version.
func (itf Selector) Filter(ctx context.Context, filters Filters) ([]string, error) {
	outputFn := InstanceTypesOutputFn(outputs.SimpleInstanceTypeOutput)
	output, _, err := itf.FilterWithOutput(ctx, filters, outputFn)
	return output, err
}

// FilterVerbose accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list instanceTypeInfo
//
// Deprecated: This function will be replaced with GetFilteredInstanceTypes() in the next
// major version.
func (itf Selector) FilterVerbose(ctx context.Context, filters Filters) ([]*instancetypes.Details, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(ctx, filters)
	if err != nil {
		return nil, err
	}
	instanceTypeInfoSlice, _ = itf.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	return instanceTypeInfoSlice, nil
}

// FilterWithOutput accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list of strings based on the custom outputFn
//
// Deprecated: This function will be replaced with GetFilteredInstanceTypes() and
// OutputInstanceTypes() in the next major version.
func (itf Selector) FilterWithOutput(ctx context.Context, filters Filters, outputFn InstanceTypesOutput) ([]string, int, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	instanceTypeInfoSlice, numOfItemsTruncated := itf.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	output := outputFn.Output(instanceTypeInfoSlice)
	return output, numOfItemsTruncated, nil
}

func (itf Selector) truncateResults(maxResults *int, instanceTypeInfoSlice []*instancetypes.Details) ([]*instancetypes.Details, int) {
	if maxResults == nil {
		return instanceTypeInfoSlice, 0
	}
	upperIndex := *maxResults
	if *maxResults > len(instanceTypeInfoSlice) {
		upperIndex = len(instanceTypeInfoSlice)
	}
	return instanceTypeInfoSlice[0:upperIndex], len(instanceTypeInfoSlice) - upperIndex
}

// AggregateFilterTransform takes higher level filters which are used to affect multiple raw filters in an opinionated way.
func (itf Selector) AggregateFilterTransform(ctx context.Context, filters Filters) (Filters, error) {
	transforms := []FiltersTransform{
		TransformFn(itf.TransformBaseInstanceType),
		TransformFn(itf.TransformFlexible),
		TransformFn(itf.TransformForService),
	}
	var err error
	for _, transform := range transforms {
		filters, err = transform.Transform(ctx, filters)
		if err != nil {
			return filters, err
		}
	}
	return filters, nil
}

// rawFilter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns the detailed specs of matching instance types
func (itf Selector) rawFilter(ctx context.Context, filters Filters) ([]*instancetypes.Details, error) {
	filters, err := itf.AggregateFilterTransform(ctx, filters)
	if err != nil {
		return nil, err
	}
	var locations, availabilityZones []string

	if filters.CPUArchitecture != nil && *filters.CPUArchitecture == cpuArchitectureAMD64 {
		*filters.CPUArchitecture = ec2types.ArchitectureTypeX8664
	}
	if filters.VirtualizationType != nil && *filters.VirtualizationType == virtualizationTypePV {
		*filters.VirtualizationType = ec2types.VirtualizationTypeParavirtual
	}
	if filters.AvailabilityZones != nil {
		availabilityZones = *filters.AvailabilityZones
		locations = *filters.AvailabilityZones
	} else if filters.Region != nil {
		locations = []string{*filters.Region}
	}
	locationInstanceOfferings, err := itf.RetrieveInstanceTypesSupportedInLocations(ctx, locations)
	if err != nil {
		return nil, err
	}

	instanceTypeDetails, err := itf.InstanceTypesProvider.Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	filteredInstanceTypes := []*instancetypes.Details{}
	var wg sync.WaitGroup
	instanceTypes := make(chan *instancetypes.Details, len(instanceTypeDetails))
	for _, instanceTypeInfo := range instanceTypeDetails {
		wg.Add(1)
		go func(instanceTypeInfo instancetypes.Details) {
			defer wg.Done()
			it, err := itf.prepareFilter(ctx, filters, instanceTypeInfo, availabilityZones, locationInstanceOfferings)
			if err != nil {
				log.Println(err)
			}
			if it != nil {
				instanceTypes <- it
			}
		}(*instanceTypeInfo)
	}
	wg.Wait()
	close(instanceTypes)
	for it := range instanceTypes {
		filteredInstanceTypes = append(filteredInstanceTypes, it)
	}
	return sortInstanceTypeInfo(filteredInstanceTypes), nil
}

func (itf Selector) prepareFilter(ctx context.Context, filters Filters, instanceTypeInfo instancetypes.Details, availabilityZones []string, locationInstanceOfferings map[ec2types.InstanceType]string) (*instancetypes.Details, error) {
	instanceTypeName := instanceTypeInfo.InstanceType
	isFpga := instanceTypeInfo.FpgaInfo != nil
	var instanceTypeHourlyPriceForFilter float64 // Price used to filter based on usage class
	var instanceTypeHourlyPriceOnDemand, instanceTypeHourlyPriceSpot *float64
	// If prices are fetched, populate the fields irrespective of the price filters
	if itf.EC2Pricing.OnDemandCacheCount() > 0 {
		price, err := itf.EC2Pricing.GetOnDemandInstanceTypeCost(ctx, instanceTypeName)
		if err != nil {
			log.Printf("Could not retrieve instantaneous hourly on-demand price for instance type %s - %s\n", instanceTypeName, err)
		} else {
			instanceTypeHourlyPriceOnDemand = &price
			instanceTypeInfo.OndemandPricePerHour = instanceTypeHourlyPriceOnDemand
		}
	}

	isSpotUsageClass := false
	for _, it := range instanceTypeInfo.SupportedUsageClasses {
		if it == ec2types.UsageClassTypeSpot {
			isSpotUsageClass = true
		}
	}

	if itf.EC2Pricing.SpotCacheCount() > 0 && isSpotUsageClass {
		price, err := itf.EC2Pricing.GetSpotInstanceTypeNDayAvgCost(ctx, instanceTypeName, availabilityZones, 30)
		if err != nil {
			log.Printf("Could not retrieve 30 day avg hourly spot price for instance type %s\n", instanceTypeName)
		} else {
			instanceTypeHourlyPriceSpot = &price
			instanceTypeInfo.SpotPrice = instanceTypeHourlyPriceSpot
		}
	}
	if filters.PricePerHour != nil {
		// If price filter is present, prices should be already fetched
		// If prices are not fetched, filter should fail and the corresponding error is already printed
		if filters.UsageClass != nil && *filters.UsageClass == ec2types.UsageClassTypeSpot && instanceTypeHourlyPriceSpot != nil {
			instanceTypeHourlyPriceForFilter = *instanceTypeHourlyPriceSpot
		} else if instanceTypeHourlyPriceOnDemand != nil {
			instanceTypeHourlyPriceForFilter = *instanceTypeHourlyPriceOnDemand
		}
	}
	eneaSupport := string(instanceTypeInfo.NetworkInfo.EnaSupport)
	ebsOptimizedSupport := string(instanceTypeInfo.EbsInfo.EbsOptimizedSupport)

	// filterToInstanceSpecMappingPairs is a map of filter name [key] to filter pair [value].
	// A filter pair includes user input filter value and instance spec value retrieved from DescribeInstanceTypes
	filterToInstanceSpecMappingPairs := map[string]filterPair{
		cpuArchitecture:                  {filters.CPUArchitecture, instanceTypeInfo.ProcessorInfo.SupportedArchitectures},
		cpuManufacturer:                  {filters.CPUManufacturer, getCPUManufacturer(&instanceTypeInfo.InstanceTypeInfo)},
		usageClass:                       {filters.UsageClass, instanceTypeInfo.SupportedUsageClasses},
		rootDeviceType:                   {filters.RootDeviceType, instanceTypeInfo.SupportedRootDeviceTypes},
		hibernationSupported:             {filters.HibernationSupported, instanceTypeInfo.HibernationSupported},
		vcpusRange:                       {filters.VCpusRange, instanceTypeInfo.VCpuInfo.DefaultVCpus},
		memoryRange:                      {filters.MemoryRange, instanceTypeInfo.MemoryInfo.SizeInMiB},
		gpuMemoryRange:                   {filters.GpuMemoryRange, getTotalGpuMemory(instanceTypeInfo.GpuInfo)},
		gpusRange:                        {filters.GpusRange, getTotalGpusCount(instanceTypeInfo.GpuInfo)},
		inferenceAcceleratorsRange:       {filters.InferenceAcceleratorsRange, getTotalAcceleratorsCount(instanceTypeInfo.InferenceAcceleratorInfo)},
		placementGroupStrategy:           {filters.PlacementGroupStrategy, instanceTypeInfo.PlacementGroupInfo.SupportedStrategies},
		hypervisor:                       {filters.Hypervisor, instanceTypeInfo.Hypervisor},
		baremetal:                        {filters.BareMetal, instanceTypeInfo.BareMetal},
		burstable:                        {filters.Burstable, instanceTypeInfo.BurstablePerformanceSupported},
		fpga:                             {filters.Fpga, &isFpga},
		enaSupport:                       {filters.EnaSupport, supportSyntaxToBool(&eneaSupport)},
		efaSupport:                       {filters.EfaSupport, instanceTypeInfo.NetworkInfo.EfaSupported},
		vcpusToMemoryRatio:               {filters.VCpusToMemoryRatio, calculateVCpusToMemoryRatio(instanceTypeInfo.VCpuInfo.DefaultVCpus, instanceTypeInfo.MemoryInfo.SizeInMiB)},
		currentGeneration:                {filters.CurrentGeneration, instanceTypeInfo.CurrentGeneration},
		networkInterfaces:                {filters.NetworkInterfaces, instanceTypeInfo.NetworkInfo.MaximumNetworkInterfaces},
		networkPerformance:               {filters.NetworkPerformance, getNetworkPerformance(instanceTypeInfo.NetworkInfo.NetworkPerformance)},
		networkEncryption:                {filters.NetworkEncryption, instanceTypeInfo.NetworkInfo.EncryptionInTransitSupported},
		ipv6:                             {filters.IPv6, instanceTypeInfo.NetworkInfo.Ipv6Supported},
		instanceTypes:                    {filters.InstanceTypes, instanceTypeInfo.InstanceType},
		virtualizationType:               {filters.VirtualizationType, instanceTypeInfo.SupportedVirtualizationTypes},
		pricePerHour:                     {filters.PricePerHour, &instanceTypeHourlyPriceForFilter},
		instanceStorageRange:             {filters.InstanceStorageRange, getInstanceStorage(instanceTypeInfo.InstanceStorageInfo)},
		diskType:                         {filters.DiskType, getDiskType(instanceTypeInfo.InstanceStorageInfo)},
		nvme:                             {filters.NVME, getNVMESupport(instanceTypeInfo.InstanceStorageInfo, instanceTypeInfo.EbsInfo)},
		ebsOptimized:                     {filters.EBSOptimized, supportSyntaxToBool(&ebsOptimizedSupport)},
		diskEncryption:                   {filters.DiskEncryption, getDiskEncryptionSupport(instanceTypeInfo.InstanceStorageInfo, instanceTypeInfo.EbsInfo)},
		ebsOptimizedBaselineBandwidth:    {filters.EBSOptimizedBaselineBandwidth, getEBSOptimizedBaselineBandwidth(instanceTypeInfo.EbsInfo)},
		ebsOptimizedBaselineThroughput:   {filters.EBSOptimizedBaselineThroughput, getEBSOptimizedBaselineThroughput(instanceTypeInfo.EbsInfo)},
		ebsOptimizedBaselineIOPS:         {filters.EBSOptimizedBaselineIOPS, getEBSOptimizedBaselineIOPS(instanceTypeInfo.EbsInfo)},
		freeTier:                         {filters.FreeTier, instanceTypeInfo.FreeTierEligible},
		autoRecovery:                     {filters.AutoRecovery, instanceTypeInfo.AutoRecoverySupported},
		gpuManufacturer:                  {filters.GPUManufacturer, getGPUManufacturers(instanceTypeInfo.GpuInfo)},
		gpuModel:                         {filters.GPUModel, getGPUModels(instanceTypeInfo.GpuInfo)},
		inferenceAcceleratorManufacturer: {filters.InferenceAcceleratorManufacturer, getInferenceAcceleratorManufacturers(instanceTypeInfo.InferenceAcceleratorInfo)},
		inferenceAcceleratorModel:        {filters.InferenceAcceleratorModel, getInferenceAcceleratorModels(instanceTypeInfo.InferenceAcceleratorInfo)},
		dedicatedHosts:                   {filters.DedicatedHosts, instanceTypeInfo.DedicatedHostsSupported},
	}

	if isInDenyList(filters.DenyList, instanceTypeName) || !isInAllowList(filters.AllowList, instanceTypeName) {
		return nil, nil
	}

	if !isSupportedInLocation(locationInstanceOfferings, instanceTypeName) {
		return nil, nil
	}

	var isInstanceSupported bool
	isInstanceSupported, err := itf.executeFilters(ctx, filterToInstanceSpecMappingPairs, instanceTypeName)
	if err != nil {
		return nil, err
	}
	if !isInstanceSupported {
		return nil, nil
	}
	return &instanceTypeInfo, nil
}

// sortInstanceTypeInfo will sort based on instance type info alpha-numerically
func sortInstanceTypeInfo(instanceTypeInfoSlice []*instancetypes.Details) []*instancetypes.Details {
	if len(instanceTypeInfoSlice) < 2 {
		return instanceTypeInfoSlice
	}
	sort.Slice(instanceTypeInfoSlice, func(i, j int) bool {
		iInstanceInfo := instanceTypeInfoSlice[i]
		jInstanceInfo := instanceTypeInfoSlice[j]
		return strings.Compare(string(iInstanceInfo.InstanceType), string(jInstanceInfo.InstanceType)) <= 0
	})
	return instanceTypeInfoSlice
}

// executeFilters accepts a mapping of filter name to filter pairs which are iterated through
// to determine if the instance type matches the filter values.
func (itf Selector) executeFilters(ctx context.Context, filterToInstanceSpecMapping map[string]filterPair, instanceType ec2types.InstanceType) (bool, error) {
	verdict := make(chan bool, len(filterToInstanceSpecMapping)+1)
	errs := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	for filterName, filter := range filterToInstanceSpecMapping {
		wg.Add(1)
		go func(ctx context.Context, filterName string, filter filterPair) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				ok, err := exec(instanceType, filterName, filter)
				if err != nil {
					errs <- err
				}
				if !ok {
					verdict <- false
				}
			}
		}(ctx, filterName, filter)
	}
	go func() {
		wg.Wait()
		verdict <- true
	}()

	if <-verdict {
		return true, nil
	}
	cancel()
	var err error
	for {
		select {
		case e := <-errs:
			err = multierr.Append(err, e)
		default:
			return false, err
		}
	}
}

func exec(instanceType ec2types.InstanceType, filterName string, filter filterPair) (bool, error) {
	filterVal := filter.filterValue
	instanceSpec := filter.instanceSpec
	filterValReflection := reflect.ValueOf(filterVal)
	// if filter is nil, user did not specify a filter, so skip evaluation
	if filterValReflection.IsNil() {
		return true, nil
	}
	instanceSpecType := reflect.ValueOf(instanceSpec).Type()
	filterType := filterValReflection.Type()
	filterDetailsMsg := fmt.Sprintf("filter (%s: %s => %s) corresponding to instance spec (%s => %s) for instance type %s", filterName, filterVal, filterType, instanceSpec, instanceSpecType, instanceType)
	invalidInstanceSpecTypeMsg := fmt.Sprintf("Unable to process for %s", filterDetailsMsg)

	// Determine appropriate filter comparator by switching on filter type
	switch filter := filterVal.(type) {
	case *string:
		switch iSpec := instanceSpec.(type) {
		case []*string:
			if !isSupportedFromStrings(iSpec, filter) {
				return false, nil
			}
		case *string:
			if !isSupportedFromString(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *bool:
		switch iSpec := instanceSpec.(type) {
		case *bool:
			if !isSupportedWithBool(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *IntRangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *int64:
			if !isSupportedWithRangeInt64(iSpec, filter) {
				return false, nil
			}
		case *int:
			if !isSupportedWithRangeInt(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *Int32RangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *int32:
			if !isSupportedWithRangeInt32(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *Float64RangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *float64:
			if !isSupportedWithRangeFloat64(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ByteQuantityRangeFilter:
		mibRange := Uint64RangeFilter{
			LowerBound: filter.LowerBound.Quantity,
			UpperBound: filter.UpperBound.Quantity,
		}
		switch iSpec := instanceSpec.(type) {
		case *int:
			var iSpec64 *int64
			if iSpec != nil {
				iSpecVal := int64(*iSpec)
				iSpec64 = &iSpecVal
			}
			if !isSupportedWithRangeUint64(iSpec64, &mibRange) {
				return false, nil
			}
		case *int64:
			if !isSupportedWithRangeUint64(iSpec, &mibRange) {
				return false, nil
			}
		case *float64:
			floatMiBRange := Float64RangeFilter{
				LowerBound: float64(filter.LowerBound.Quantity),
				UpperBound: float64(filter.UpperBound.Quantity),
			}
			if !isSupportedWithRangeFloat64(iSpec, &floatMiBRange) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *float64:
		switch iSpec := instanceSpec.(type) {
		case *float64:
			if !isSupportedWithFloat64(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ec2types.ArchitectureType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.ArchitectureType:
			if !isSupportedArchitectureType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ec2types.UsageClassType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.UsageClassType:
			if !isSupportedUsageClassType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *CPUManufacturer:
		switch iSpec := instanceSpec.(type) {
		case CPUManufacturer:
			if !isMatchingCpuArchitecture(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ec2types.VirtualizationType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.VirtualizationType:
			if !isSupportedVirtualizationType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ec2types.InstanceTypeHypervisor:
		switch iSpec := instanceSpec.(type) {
		case ec2types.InstanceTypeHypervisor:
			if !isSupportedInstanceTypeHypervisorType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *ec2types.RootDeviceType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.RootDeviceType:
			if !isSupportedRootDeviceType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	case *[]string:
		switch iSpec := instanceSpec.(type) {
		case *string:
			filterOfPtrs := []*string{}
			for _, f := range *filter {
				// this allows us to copy a static pointer to f into filterOfPtrs
				// since the pointer to f is updated on each loop iteration
				temp := f
				filterOfPtrs = append(filterOfPtrs, &temp)
			}
			if !isSupportedFromStrings(filterOfPtrs, iSpec) {
				return false, nil
			}
		default:
			return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
		}
	default:
		return false, fmt.Errorf("No filter handler found for %s", filterDetailsMsg)
	}
	return true, nil
}

// RetrieveInstanceTypesSupportedInLocations returns a map of instance type -> AZ or Region for all instance types supported in the intersected locations passed in
// The location can be a zone-id (ie. use1-az1), a zone-name (us-east-1a), or a region name (us-east-1).
// Note that zone names are not necessarily the same across accounts
func (itf Selector) RetrieveInstanceTypesSupportedInLocations(ctx context.Context, locations []string) (map[ec2types.InstanceType]string, error) {
	if len(locations) == 0 {
		return nil, nil
	}
	availableInstanceTypes := map[ec2types.InstanceType]int{}
	for _, location := range locations {
		locationType, err := itf.getLocationType(ctx, location)
		if err != nil {
			return nil, err
		}

		instanceTypeOfferingsInput := &ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: locationType,
			Filters: []ec2types.Filter{
				{
					Name:   aws.String(locationFilterKey),
					Values: []string{location},
				},
			},
		}

		p := ec2.NewDescribeInstanceTypeOfferingsPaginator(itf.EC2, instanceTypeOfferingsInput)

		for p.HasMorePages() {
			instanceTypeOfferings, err := p.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("Encountered an error when describing instance type offerings: %w", err)
			}

			for _, instanceType := range instanceTypeOfferings.InstanceTypeOfferings {
				if i, ok := availableInstanceTypes[instanceType.InstanceType]; !ok {
					availableInstanceTypes[instanceType.InstanceType] = 1
				} else {
					availableInstanceTypes[instanceType.InstanceType] = i + 1
				}
			}
		}
	}
	availableInstanceTypesAllLocations := map[ec2types.InstanceType]string{}
	for instanceType, locationsSupported := range availableInstanceTypes {
		if locationsSupported == len(locations) {
			availableInstanceTypesAllLocations[instanceType] = ""
		}
	}

	return availableInstanceTypesAllLocations, nil
}

func (itf Selector) getLocationType(ctx context.Context, location string) (ec2types.LocationType, error) {
	azs, err := itf.EC2.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return "", err
	}
	for _, zone := range azs.AvailabilityZones {
		if location == *zone.RegionName {
			return regionNameLocationType, nil
		} else if location == *zone.ZoneName {
			return zoneNameLocationType, nil
		} else if location == *zone.ZoneId {
			return zoneIDLocationType, nil
		}
	}
	return "", fmt.Errorf("The location passed in (%s) is not a valid zone-id, zone-name, or region name", location)
}

func isSupportedInLocation(instanceOfferings map[ec2types.InstanceType]string, instanceType ec2types.InstanceType) bool {
	if instanceOfferings == nil {
		return true
	}
	_, ok := instanceOfferings[instanceType]
	return ok
}

func isInDenyList(denyRegex *regexp.Regexp, instanceTypeName ec2types.InstanceType) bool {
	if denyRegex == nil {
		return false
	}
	return denyRegex.MatchString(string(instanceTypeName))
}

func isInAllowList(allowRegex *regexp.Regexp, instanceTypeName ec2types.InstanceType) bool {
	if allowRegex == nil {
		return true
	}
	return allowRegex.MatchString(string(instanceTypeName))
}
