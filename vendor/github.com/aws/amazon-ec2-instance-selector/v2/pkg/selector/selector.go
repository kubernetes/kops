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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"go.uber.org/multierr"
)

var (
	// Version is overridden at compilation with the version based on the git tag
	versionID = "dev"
)

const (
	locationFilterKey      = "location"
	zoneIDLocationType     = "availability-zone-id"
	zoneNameLocationType   = "availability-zone"
	regionNameLocationType = "region"
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
	cpuArchitectureX8664 = "x86_64"

	virtualizationTypeParaVirtual = "paravirtual"
	virtualizationTypePV          = "pv"

	pricePerHour = "pricePerHour"
)

// New creates an instance of Selector provided an aws session
func New(sess *session.Session) *Selector {
	serviceRegistry := NewRegistry()
	serviceRegistry.RegisterAWSServices()
	ec2Client := ec2.New(userAgentWith(sess))
	return &Selector{
		EC2:                   ec2Client,
		EC2Pricing:            ec2pricing.New(sess),
		InstanceTypesProvider: instancetypes.LoadFromOrNew("", *sess.Config.Region, 0, ec2Client),
		ServiceRegistry:       serviceRegistry,
	}
}

func NewWithCache(sess *session.Session, ttl time.Duration, cacheDir string) *Selector {
	serviceRegistry := NewRegistry()
	serviceRegistry.RegisterAWSServices()
	ec2Client := ec2.New(userAgentWith(sess))
	return &Selector{
		EC2:                   ec2Client,
		EC2Pricing:            ec2pricing.NewWithCache(sess, ttl, cacheDir),
		InstanceTypesProvider: instancetypes.LoadFromOrNew(cacheDir, *sess.Config.Region, ttl, ec2Client),
		ServiceRegistry:       serviceRegistry,
	}
}

func (itf Selector) Save() error {
	return multierr.Append(itf.EC2Pricing.Save(), itf.InstanceTypesProvider.Save())
}

// Filter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a simple list of instance type strings
//
// Deprecated: This function will be replaced with GetFilteredInstanceTypes() and
// OutputInstanceTypes() in the next major version.
func (itf Selector) Filter(filters Filters) ([]string, error) {
	outputFn := InstanceTypesOutputFn(outputs.SimpleInstanceTypeOutput)
	output, _, err := itf.FilterWithOutput(filters, outputFn)
	return output, err
}

// FilterVerbose accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list instanceTypeInfo
//
// Deprecated: This function will be replaced with GetFilteredInstanceTypes() in the next
// major version.
func (itf Selector) FilterVerbose(filters Filters) ([]*instancetypes.Details, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(filters)
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
func (itf Selector) FilterWithOutput(filters Filters, outputFn InstanceTypesOutput) ([]string, int, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(filters)
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
func (itf Selector) AggregateFilterTransform(filters Filters) (Filters, error) {
	transforms := []FiltersTransform{
		TransformFn(itf.TransformBaseInstanceType),
		TransformFn(itf.TransformFlexible),
		TransformFn(itf.TransformForService),
	}
	var err error
	for _, transform := range transforms {
		filters, err = transform.Transform(filters)
		if err != nil {
			return filters, err
		}
	}
	return filters, nil
}

// rawFilter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns the detailed specs of matching instance types
func (itf Selector) rawFilter(filters Filters) ([]*instancetypes.Details, error) {
	filters, err := itf.AggregateFilterTransform(filters)
	if err != nil {
		return nil, err
	}
	var locations, availabilityZones []string

	if filters.CPUArchitecture != nil && *filters.CPUArchitecture == cpuArchitectureAMD64 {
		*filters.CPUArchitecture = cpuArchitectureX8664
	}
	if filters.VirtualizationType != nil && *filters.VirtualizationType == virtualizationTypePV {
		*filters.VirtualizationType = virtualizationTypeParaVirtual
	}
	if filters.AvailabilityZones != nil {
		availabilityZones = *filters.AvailabilityZones
		locations = *filters.AvailabilityZones
	} else if filters.Region != nil {
		locations = []string{*filters.Region}
	}
	locationInstanceOfferings, err := itf.RetrieveInstanceTypesSupportedInLocations(locations)
	if err != nil {
		return nil, err
	}

	instanceTypeDetails, err := itf.InstanceTypesProvider.Get(nil)
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
			it, err := itf.prepareFilter(filters, instanceTypeInfo, availabilityZones, locationInstanceOfferings)
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

func (itf Selector) prepareFilter(filters Filters, instanceTypeInfo instancetypes.Details, availabilityZones []string, locationInstanceOfferings map[string]string) (*instancetypes.Details, error) {
	instanceTypeName := *instanceTypeInfo.InstanceType
	isFpga := instanceTypeInfo.FpgaInfo != nil
	var instanceTypeHourlyPriceForFilter float64 // Price used to filter based on usage class
	var instanceTypeHourlyPriceOnDemand, instanceTypeHourlyPriceSpot *float64
	// If prices are fetched, populate the fields irrespective of the price filters
	if itf.EC2Pricing.OnDemandCacheCount() > 0 {
		price, err := itf.EC2Pricing.GetOnDemandInstanceTypeCost(instanceTypeName)
		if err != nil {
			log.Printf("Could not retrieve instantaneous hourly on-demand price for instance type %s\n", instanceTypeName)
		} else {
			instanceTypeHourlyPriceOnDemand = &price
			instanceTypeInfo.OndemandPricePerHour = instanceTypeHourlyPriceOnDemand
		}
	}
	if itf.EC2Pricing.SpotCacheCount() > 0 && contains(instanceTypeInfo.SupportedUsageClasses, "spot") {
		price, err := itf.EC2Pricing.GetSpotInstanceTypeNDayAvgCost(instanceTypeName, availabilityZones, 30)
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
		if filters.UsageClass != nil && *filters.UsageClass == "spot" && instanceTypeHourlyPriceSpot != nil {
			instanceTypeHourlyPriceForFilter = *instanceTypeHourlyPriceSpot
		} else if instanceTypeHourlyPriceOnDemand != nil {
			instanceTypeHourlyPriceForFilter = *instanceTypeHourlyPriceOnDemand
		}
	}

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
		enaSupport:                       {filters.EnaSupport, supportSyntaxToBool(instanceTypeInfo.NetworkInfo.EnaSupport)},
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
		ebsOptimized:                     {filters.EBSOptimized, supportSyntaxToBool(instanceTypeInfo.EbsInfo.EbsOptimizedSupport)},
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
	isInstanceSupported, err := itf.executeFilters(filterToInstanceSpecMappingPairs, instanceTypeName)
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
		return strings.Compare(aws.StringValue(iInstanceInfo.InstanceType), aws.StringValue(jInstanceInfo.InstanceType)) <= 0
	})
	return instanceTypeInfoSlice
}

// executeFilters accepts a mapping of filter name to filter pairs which are iterated through
// to determine if the instance type matches the filter values.
func (itf Selector) executeFilters(filterToInstanceSpecMapping map[string]filterPair, instanceType string) (bool, error) {
	abort := make(chan bool, len(filterToInstanceSpecMapping))
	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
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
					abort <- true
				}
			}
		}(ctx, filterName, filter)
	}
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()
	select {
	case <-abort:
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
	case <-done:
		return true, nil
	}
}

func exec(instanceType string, filterName string, filter filterPair) (bool, error) {
	filterVal := filter.filterValue
	instanceSpec := filter.instanceSpec
	// if filter is nil, user did not specify a filter, so skip evaluation
	if reflect.ValueOf(filterVal).IsNil() {
		return true, nil
	}
	instanceSpecType := reflect.ValueOf(instanceSpec).Type()
	filterType := reflect.ValueOf(filterVal).Type()
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
func (itf Selector) RetrieveInstanceTypesSupportedInLocations(locations []string) (map[string]string, error) {
	if len(locations) == 0 {
		return nil, nil
	}
	availableInstanceTypes := map[string]int{}
	for _, location := range locations {
		instanceTypeOfferingsInput := &ec2.DescribeInstanceTypeOfferingsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String(locationFilterKey),
					Values: []*string{aws.String(location)},
				},
			},
		}
		locationType, err := itf.getLocationType(location)
		if err != nil {
			return nil, err
		}
		instanceTypeOfferingsInput.SetLocationType(locationType)

		err = itf.EC2.DescribeInstanceTypeOfferingsPages(instanceTypeOfferingsInput, func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool {
			for _, instanceType := range page.InstanceTypeOfferings {
				if i, ok := availableInstanceTypes[*instanceType.InstanceType]; !ok {
					availableInstanceTypes[*instanceType.InstanceType] = 1
				} else {
					availableInstanceTypes[*instanceType.InstanceType] = i + 1
				}
			}
			return true
		})
		if err != nil {
			return nil, fmt.Errorf("Encountered an error when describing instance type offerings: %w", err)
		}
	}
	availableInstanceTypesAllLocations := map[string]string{}
	for instanceType, locationsSupported := range availableInstanceTypes {
		if locationsSupported == len(locations) {
			availableInstanceTypesAllLocations[instanceType] = ""
		}
	}

	return availableInstanceTypesAllLocations, nil
}

func (itf Selector) getLocationType(location string) (string, error) {
	azs, err := itf.EC2.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
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

func isSupportedInLocation(instanceOfferings map[string]string, instanceType string) bool {
	if instanceOfferings == nil {
		return true
	}
	_, ok := instanceOfferings[instanceType]
	return ok
}

func isInDenyList(denyRegex *regexp.Regexp, instanceTypeName string) bool {
	if denyRegex == nil {
		return false
	}
	return denyRegex.MatchString(instanceTypeName)
}

func isInAllowList(allowRegex *regexp.Regexp, instanceTypeName string) bool {
	if allowRegex == nil {
		return true
	}
	return allowRegex.MatchString(instanceTypeName)
}

func userAgentWith(sess *session.Session) *session.Session {
	userAgentHandler := request.MakeAddToUserAgentFreeFormHandler(fmt.Sprintf("%s-%s", sdkName, versionID))
	sess.Handlers.Build.PushBack(userAgentHandler)
	return sess
}
