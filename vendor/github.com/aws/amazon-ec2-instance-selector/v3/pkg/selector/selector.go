// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package selector provides filtering logic for Amazon EC2 Instance Types based on declarative resource specfications.
package selector

import (
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/multierr"

	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/ec2pricing"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/instancetypes"
	"github.com/aws/amazon-ec2-instance-selector/v3/pkg/selector/outputs"
)

// Version is overridden at compilation with the version based on the git tag
var versionID = "dev"

const (
	locationFilterKey      = "location"
	zoneIDLocationType     = ec2types.LocationTypeAvailabilityZoneId
	zoneNameLocationType   = ec2types.LocationTypeAvailabilityZone
	regionNameLocationType = ec2types.LocationTypeRegion
	sdkName                = "instance-selector"

	// Filter Keys.

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
	generation                       = "generation"

	cpuArchitectureAMD64 = "amd64"

	virtualizationTypePV = "pv"

	pricePerHour = "pricePerHour"
)

// New creates an instance of Selector provided an aws session.
func New(ctx context.Context, cfg aws.Config) (*Selector, error) {
	return NewWithCache(ctx, cfg, 0, "")
}

// NewWithCache creates an instance of Selector backed by an on-disk cache provided an aws session and cache configuration parameters.
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

	instanceTypeProvider, err := instancetypes.LoadFromOrNew(cacheDir, cfg.Region, ttl, ec2Client)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize instance type provider: %w", err)
	}

	return &Selector{
		EC2:                   ec2Client,
		EC2Pricing:            pricingClient,
		InstanceTypesProvider: instanceTypeProvider,
		ServiceRegistry:       serviceRegistry,
		Logger:                log.New(io.Discard, "", 0),
	}, nil
}

// SetLogger can be called to log more detailed logs about what selector is doing
// including things like API timings
// If SetLogger is not called, no logs will be displayed.
func (s *Selector) SetLogger(logger *log.Logger) {
	s.Logger = logger
	s.InstanceTypesProvider.SetLogger(logger)
	s.EC2Pricing.SetLogger(logger)
}

// Save persists the selector cache data to disk if caching is configured.
func (s Selector) Save() error {
	return multierr.Append(s.EC2Pricing.Save(), s.InstanceTypesProvider.Save())
}

// Filter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a simple list of instance type strings.
func (s Selector) Filter(ctx context.Context, filters Filters) ([]string, error) {
	outputFn := InstanceTypesOutputFn(outputs.SimpleInstanceTypeOutput)
	output, _, err := s.FilterWithOutput(ctx, filters, outputFn)
	return output, err
}

// FilterVerbose accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list instanceTypeInfo.
func (s Selector) FilterVerbose(ctx context.Context, filters Filters) ([]*instancetypes.Details, error) {
	instanceTypeInfoSlice, err := s.rawFilter(ctx, filters)
	if err != nil {
		return nil, err
	}
	instanceTypeInfoSlice, _ = s.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	return instanceTypeInfoSlice, nil
}

// FilterWithOutput accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list of strings based on the custom outputFn.
func (s Selector) FilterWithOutput(ctx context.Context, filters Filters, outputFn InstanceTypesOutput) ([]string, int, error) {
	instanceTypeInfoSlice, err := s.rawFilter(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	instanceTypeInfoSlice, numOfItemsTruncated := s.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	output := outputFn.Output(instanceTypeInfoSlice)
	return output, numOfItemsTruncated, nil
}

func (s Selector) truncateResults(maxResults *int, instanceTypeInfoSlice []*instancetypes.Details) ([]*instancetypes.Details, int) {
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
func (s Selector) AggregateFilterTransform(ctx context.Context, filters Filters) (Filters, error) {
	transforms := []FiltersTransform{
		TransformFn(s.TransformBaseInstanceType),
		TransformFn(s.TransformFlexible),
		TransformFn(s.TransformForService),
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
// matching the criteria within Filters and returns the detailed specs of matching instance types.
func (s Selector) rawFilter(ctx context.Context, filters Filters) ([]*instancetypes.Details, error) {
	filters, err := s.AggregateFilterTransform(ctx, filters)
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
	locationInstanceOfferings, err := s.RetrieveInstanceTypesSupportedInLocations(ctx, locations)
	if err != nil {
		return nil, err
	}

	instanceTypeDetails, err := s.InstanceTypesProvider.Get(ctx, nil)
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
			it, err := s.prepareFilter(ctx, filters, instanceTypeInfo, availabilityZones, locationInstanceOfferings)
			if err != nil {
				s.Logger.Printf("Unable to prepare filter for %s, %v", instanceTypeInfo.InstanceType, err)
			}
			if it != nil {
				instanceTypes <- it
			}
		}(*instanceTypeInfo)
	}
	go func() {
		wg.Wait()
		close(instanceTypes)
	}()
	for it := range instanceTypes {
		filteredInstanceTypes = append(filteredInstanceTypes, it)
	}
	return sortInstanceTypeInfo(filteredInstanceTypes), nil
}

func (s Selector) prepareFilter(ctx context.Context, filters Filters, instanceTypeInfo instancetypes.Details, availabilityZones []string, locationInstanceOfferings map[ec2types.InstanceType]string) (*instancetypes.Details, error) {
	instanceTypeName := instanceTypeInfo.InstanceType
	isFpga := instanceTypeInfo.FpgaInfo != nil
	var instanceTypeHourlyPriceForFilter float64 // Price used to filter based on usage class
	var instanceTypeHourlyPriceOnDemand, instanceTypeHourlyPriceSpot *float64
	// If prices are fetched, populate the fields irrespective of the price filters
	if s.EC2Pricing.OnDemandCacheCount() > 0 {
		price, err := s.EC2Pricing.GetOnDemandInstanceTypeCost(ctx, instanceTypeName)
		if err != nil {
			s.Logger.Printf("Could not retrieve instantaneous hourly on-demand price for instance type %s - %s\n", instanceTypeName, err)
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

	if s.EC2Pricing.SpotCacheCount() > 0 && isSpotUsageClass {
		price, err := s.EC2Pricing.GetSpotInstanceTypeNDayAvgCost(ctx, instanceTypeName, availabilityZones, 30)
		if err != nil {
			s.Logger.Printf("Could not retrieve 30 day avg hourly spot price for instance type %s\n", instanceTypeName)
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

	// If an empty slice is passed, treat the filter as nil
	filterInstanceTypes := filters.InstanceTypes
	if filterInstanceTypes != nil && len(*filterInstanceTypes) == 0 {
		filterInstanceTypes = nil
	}

	var cpuManufacturerFilter *string
	if filters.CPUManufacturer != nil {
		cpuManufacturerFilter = aws.String(string(*filters.CPUManufacturer))
	}

	// filterToInstanceSpecMappingPairs is a map of filter name [key] to filter pair [value].
	// A filter pair includes user input filter value and instance spec value retrieved from DescribeInstanceTypes
	filterToInstanceSpecMappingPairs := map[string]filterPair{
		cpuArchitecture:                  {filters.CPUArchitecture, instanceTypeInfo.ProcessorInfo.SupportedArchitectures},
		cpuManufacturer:                  {cpuManufacturerFilter, instanceTypeInfo.ProcessorInfo.Manufacturer},
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
		instanceTypes:                    {filterInstanceTypes, aws.String(string(instanceTypeInfo.InstanceType))},
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
		generation:                       {filters.Generation, getInstanceTypeGeneration(string(instanceTypeInfo.InstanceType))},
	}

	if isInDenyList(filters.DenyList, instanceTypeName) || !isInAllowList(filters.AllowList, instanceTypeName) {
		return nil, nil
	}

	if !isSupportedInLocation(locationInstanceOfferings, instanceTypeName) {
		return nil, nil
	}

	var isInstanceSupported bool
	isInstanceSupported, err := s.executeFilters(ctx, filterToInstanceSpecMappingPairs, instanceTypeName)
	if err != nil {
		return nil, err
	}
	if !isInstanceSupported {
		return nil, nil
	}
	return &instanceTypeInfo, nil
}

// sortInstanceTypeInfo will sort based on instance type info alpha-numerically.
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
func (s Selector) executeFilters(ctx context.Context, filterToInstanceSpecMapping map[string]filterPair, instanceType ec2types.InstanceType) (bool, error) {
	verdict := make(chan bool, len(filterToInstanceSpecMapping)+1)
	errs := make(chan error, len(filterToInstanceSpecMapping))
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

// exec executes a specific filterPair (user value & instance spec) with a specific instance type
// If the filterPair matches, true is returned.
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
	errInvalidInstanceSpec := fmt.Errorf("unable to process for %s", filterDetailsMsg)

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
			return false, errInvalidInstanceSpec
		}
	case *bool:
		switch iSpec := instanceSpec.(type) {
		case *bool:
			if !isSupportedWithBool(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *IntRangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *int64:
			if !isSupportedWithRangeInt64(iSpec, filter) {
				return false, nil
			}
		case *int32:
			var iSpec64 *int64
			if iSpec != nil {
				iSpec64 = aws.Int64(int64(*iSpec))
			}
			if !isSupportedWithRangeInt64(iSpec64, filter) {
				return false, nil
			}
		case *int:
			if !isSupportedWithRangeInt(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *Int32RangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *int32:
			if !isSupportedWithRangeInt32(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *Float64RangeFilter:
		switch iSpec := instanceSpec.(type) {
		case *float64:
			if !isSupportedWithRangeFloat64(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
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
			return false, errInvalidInstanceSpec
		}
	case *float64:
		switch iSpec := instanceSpec.(type) {
		case *float64:
			if !isSupportedWithFloat64(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *ec2types.ArchitectureType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.ArchitectureType:
			if !isSupportedArchitectureType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *ec2types.UsageClassType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.UsageClassType:
			if !isSupportedUsageClassType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *CPUManufacturer:
		switch iSpec := instanceSpec.(type) {
		case CPUManufacturer:
			if !isMatchingCpuArchitecture(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *ec2types.VirtualizationType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.VirtualizationType:
			if !isSupportedVirtualizationType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *ec2types.InstanceTypeHypervisor:
		switch iSpec := instanceSpec.(type) {
		case ec2types.InstanceTypeHypervisor:
			if !isSupportedInstanceTypeHypervisorType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
		}
	case *ec2types.RootDeviceType:
		switch iSpec := instanceSpec.(type) {
		case []ec2types.RootDeviceType:
			if !isSupportedRootDeviceType(iSpec, filter) {
				return false, nil
			}
		default:
			return false, errInvalidInstanceSpec
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
			return false, errInvalidInstanceSpec
		}
	default:
		return false, fmt.Errorf("no filter handler found for %s", filterDetailsMsg)
	}
	return true, nil
}

// RetrieveInstanceTypesSupportedInLocations returns a map of instance type -> AZ or Region for all instance types supported in the intersected locations passed in
// The location can be a zone-id (ie. use1-az1), a zone-name (us-east-1a), or a region name (us-east-1).
// Note that zone names are not necessarily the same across accounts.
func (s Selector) RetrieveInstanceTypesSupportedInLocations(ctx context.Context, locations []string) (map[ec2types.InstanceType]string, error) {
	if len(locations) == 0 {
		return nil, nil
	}
	availableInstanceTypes := map[ec2types.InstanceType]int{}
	for _, location := range locations {
		locationType, err := s.getLocationType(ctx, location)
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

		p := ec2.NewDescribeInstanceTypeOfferingsPaginator(s.EC2, instanceTypeOfferingsInput)

		for p.HasMorePages() {
			instanceTypeOfferings, err := p.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("encountered an error when describing instance type offerings: %w", err)
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

func (s Selector) getLocationType(ctx context.Context, location string) (ec2types.LocationType, error) {
	azs, err := s.EC2.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{})
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
	return "", fmt.Errorf("the location passed in (%s) is not a valid zone-id, zone-name, or region name", location)
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
