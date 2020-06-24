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
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/pkg/selector/outputs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	cpuArchitecture        = "cpuArchitecture"
	usageClass             = "usageClass"
	rootDeviceType         = "rootDeviceType"
	hibernationSupported   = "hibernationSupported"
	vcpusRange             = "vcpusRange"
	memoryRange            = "memoryRange"
	gpuMemoryRange         = "gpuMemoryRange"
	gpusRange              = "gpusRange"
	placementGroupStrategy = "placementGroupStrategy"
	hypervisor             = "hypervisor"
	baremetal              = "baremetal"
	burstable              = "burstable"
	fpga                   = "fpga"
	enaSupport             = "enaSupport"
	vcpusToMemoryRatio     = "vcpusToMemoryRatio"
	currentGeneration      = "currentGeneration"
	networkInterfaces      = "networkInterfaces"
	networkPerformance     = "networkPerformance"
	allowList              = "allowList"
	denyList               = "denyList"
)

// New creates an instance of Selector provided an aws session
func New(sess *session.Session) *Selector {
	userAgentTag := fmt.Sprintf("%s-v%s", sdkName, versionID)
	userAgentHandler := request.MakeAddToUserAgentFreeFormHandler(userAgentTag)
	sess.Handlers.Build.PushBack(userAgentHandler)
	return &Selector{
		EC2: ec2.New(sess),
	}
}

// Filter accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a simple list of instance type strings
func (itf Selector) Filter(filters Filters) ([]string, error) {
	outputFn := InstanceTypesOutputFn(outputs.SimpleInstanceTypeOutput)
	return itf.FilterWithOutput(filters, outputFn)
}

// FilterVerbose accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list instanceTypeInfo
func (itf Selector) FilterVerbose(filters Filters) ([]*ec2.InstanceTypeInfo, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(filters)
	if err != nil {
		return nil, err
	}
	instanceTypeInfoSlice = itf.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	return instanceTypeInfoSlice, nil
}

// FilterWithOutput accepts a Filters struct which is used to select the available instance types
// matching the criteria within Filters and returns a list of strings based on the custom outputFn
func (itf Selector) FilterWithOutput(filters Filters, outputFn InstanceTypesOutput) ([]string, error) {
	instanceTypeInfoSlice, err := itf.rawFilter(filters)
	if err != nil {
		return nil, err
	}
	instanceTypeInfoSlice = itf.truncateResults(filters.MaxResults, instanceTypeInfoSlice)
	output := outputFn.Output(instanceTypeInfoSlice)
	return output, nil
}

func (itf Selector) truncateResults(maxResults *int, instanceTypeInfoSlice []*ec2.InstanceTypeInfo) []*ec2.InstanceTypeInfo {
	if maxResults == nil {
		return instanceTypeInfoSlice
	}
	upperIndex := *maxResults
	if *maxResults > len(instanceTypeInfoSlice) {
		upperIndex = len(instanceTypeInfoSlice)
	}
	return instanceTypeInfoSlice[0:upperIndex]
}

// AggregateFilterTransform takes higher level filters which are used to affect multiple raw filters in an opinionated way.
func (itf Selector) AggregateFilterTransform(filters Filters) (Filters, error) {
	transforms := []FiltersTransform{
		TransformFn(itf.TransformBaseInstanceType),
		TransformFn(itf.TransformFlexible),
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
func (itf Selector) rawFilter(filters Filters) ([]*ec2.InstanceTypeInfo, error) {
	filters, err := itf.AggregateFilterTransform(filters)
	if err != nil {
		return nil, err
	}
	var locations []string

	// Support the deprecated singular availabilityZone filter in favor of the plural
	if filters.AvailabilityZone != nil {
		if filters.AvailabilityZones != nil {
			*filters.AvailabilityZones = append(*filters.AvailabilityZones, *filters.AvailabilityZone)
		} else {
			filters.AvailabilityZones = &[]string{*filters.AvailabilityZone}
		}
	}

	if filters.AvailabilityZones != nil {
		locations = *filters.AvailabilityZones
	} else if filters.Region != nil {
		locations = []string{*filters.Region}
	}
	locationInstanceOfferings, err := itf.RetrieveInstanceTypesSupportedInLocations(locations)
	if err != nil {
		return nil, err
	}

	instanceTypesInput := &ec2.DescribeInstanceTypesInput{}
	instanceTypeCandidates := map[string]*ec2.InstanceTypeInfo{}
	// innerErr will hold any error while processing DescribeInstanceTypes pages
	var innerErr error

	err = itf.EC2.DescribeInstanceTypesPages(instanceTypesInput, func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
		for _, instanceTypeInfo := range page.InstanceTypes {
			instanceTypeName := *instanceTypeInfo.InstanceType
			instanceTypeCandidates[instanceTypeName] = instanceTypeInfo
			isFpga := instanceTypeInfo.FpgaInfo != nil

			// filterToInstanceSpecMappingPairs is a map of filter name [key] to filter pair [value].
			// A filter pair includes user input filter value and instance spec value retrieved from DescribeInstanceTypes
			filterToInstanceSpecMappingPairs := map[string]filterPair{
				cpuArchitecture:        {filters.CPUArchitecture, instanceTypeInfo.ProcessorInfo.SupportedArchitectures},
				usageClass:             {filters.UsageClass, instanceTypeInfo.SupportedUsageClasses},
				rootDeviceType:         {filters.RootDeviceType, instanceTypeInfo.SupportedRootDeviceTypes},
				hibernationSupported:   {filters.HibernationSupported, instanceTypeInfo.HibernationSupported},
				vcpusRange:             {filters.VCpusRange, instanceTypeInfo.VCpuInfo.DefaultVCpus},
				memoryRange:            {filters.MemoryRange, instanceTypeInfo.MemoryInfo.SizeInMiB},
				gpuMemoryRange:         {filters.GpuMemoryRange, getTotalGpuMemory(instanceTypeInfo.GpuInfo)},
				gpusRange:              {filters.GpusRange, getTotalGpusCount(instanceTypeInfo.GpuInfo)},
				placementGroupStrategy: {filters.PlacementGroupStrategy, instanceTypeInfo.PlacementGroupInfo.SupportedStrategies},
				hypervisor:             {filters.Hypervisor, instanceTypeInfo.Hypervisor},
				baremetal:              {filters.BareMetal, instanceTypeInfo.BareMetal},
				burstable:              {filters.Burstable, instanceTypeInfo.BurstablePerformanceSupported},
				fpga:                   {filters.Fpga, &isFpga},
				enaSupport:             {filters.EnaSupport, supportSyntaxToBool(instanceTypeInfo.NetworkInfo.EnaSupport)},
				vcpusToMemoryRatio:     {filters.VCpusToMemoryRatio, calculateVCpusToMemoryRatio(instanceTypeInfo.VCpuInfo.DefaultVCpus, instanceTypeInfo.MemoryInfo.SizeInMiB)},
				currentGeneration:      {filters.CurrentGeneration, instanceTypeInfo.CurrentGeneration},
				networkInterfaces:      {filters.NetworkInterfaces, instanceTypeInfo.NetworkInfo.MaximumNetworkInterfaces},
				networkPerformance:     {filters.NetworkPerformance, getNetworkPerformance(instanceTypeInfo.NetworkInfo.NetworkPerformance)},
			}

			if isInDenyList(filters.DenyList, instanceTypeName) || !isInAllowList(filters.AllowList, instanceTypeName) {
				delete(instanceTypeCandidates, instanceTypeName)
			}

			if !isSupportedInLocation(locationInstanceOfferings, instanceTypeName) {
				delete(instanceTypeCandidates, instanceTypeName)
			}

			var isInstanceSupported bool
			isInstanceSupported, innerErr = itf.executeFilters(filterToInstanceSpecMappingPairs, instanceTypeName)
			if innerErr != nil {
				// stops paging through instance types
				return false
			}
			if !isInstanceSupported {
				delete(instanceTypeCandidates, instanceTypeName)
			}
		}
		// continue paging through instance types
		return true
	})
	if err != nil {
		return nil, err
	}
	if innerErr != nil {
		return nil, innerErr
	}

	instanceTypeInfoSlice := []*ec2.InstanceTypeInfo{}
	for _, instanceTypeInfo := range instanceTypeCandidates {
		instanceTypeInfoSlice = append(instanceTypeInfoSlice, instanceTypeInfo)
	}
	return sortInstanceTypeInfo(instanceTypeInfoSlice), nil
}

// sortInstanceTypeInfo will sort based on instance type info alpha-numerically
func sortInstanceTypeInfo(instanceTypeInfoSlice []*ec2.InstanceTypeInfo) []*ec2.InstanceTypeInfo {
	sort.Slice(instanceTypeInfoSlice, func(i, j int) bool {
		iInstanceInfo := instanceTypeInfoSlice[i]
		jInstanceInfo := instanceTypeInfoSlice[j]
		return strings.Compare(*iInstanceInfo.InstanceType, *jInstanceInfo.InstanceType) <= 0
	})
	return instanceTypeInfoSlice
}

// executeFilters accepts a mapping of filter name to filter pairs which are iterated through
// to determine if the instance type matches the filter values.
func (itf Selector) executeFilters(filterToInstanceSpecMapping map[string]filterPair, instanceType string) (bool, error) {
	for filterName, filterPair := range filterToInstanceSpecMapping {
		filterVal := filterPair.filterValue
		instanceSpec := filterPair.instanceSpec
		// if filter is nil, user did not specify a filter, so skip evaluation
		if reflect.ValueOf(filterVal).IsNil() {
			continue
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
		case *float64:
			switch iSpec := instanceSpec.(type) {
			case *float64:
				if !isSupportedWithFloat64(iSpec, filter) {
					return false, nil
				}
			default:
				return false, fmt.Errorf(invalidInstanceSpecTypeMsg)
			}
		default:
			return false, fmt.Errorf("No filter handler found for %s", filterDetailsMsg)
		}
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
