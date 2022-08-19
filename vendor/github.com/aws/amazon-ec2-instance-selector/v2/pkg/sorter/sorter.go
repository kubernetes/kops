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

package sorter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/oliveagle/jsonpath"
)

const (
	// Sort direction

	SortAscending  = "ascending"
	SortAsc        = "asc"
	SortDescending = "descending"
	SortDesc       = "desc"

	// Not all fields can be reached through a json path (Ex: gpu count)
	// so we have special flags for such cases.

	GPUCountField              = "gpus"
	InferenceAcceleratorsField = "inference-accelerators"

	// shorthand flags

	VCPUs                          = "vcpus"
	Memory                         = "memory"
	GPUMemoryTotal                 = "gpu-memory-total"
	NetworkInterfaces              = "network-interfaces"
	SpotPrice                      = "spot-price"
	ODPrice                        = "on-demand-price"
	InstanceStorage                = "instance-storage"
	EBSOptimizedBaselineBandwidth  = "ebs-optimized-baseline-bandwidth"
	EBSOptimizedBaselineThroughput = "ebs-optimized-baseline-throughput"
	EBSOptimizedBaselineIOPS       = "ebs-optimized-baseline-iops"

	// JSON field paths for shorthand flags

	instanceNamePath                   = ".InstanceType"
	vcpuPath                           = ".VCpuInfo.DefaultVCpus"
	memoryPath                         = ".MemoryInfo.SizeInMiB"
	gpuMemoryTotalPath                 = ".GpuInfo.TotalGpuMemoryInMiB"
	networkInterfacesPath              = ".NetworkInfo.MaximumNetworkInterfaces"
	spotPricePath                      = ".SpotPrice"
	odPricePath                        = ".OndemandPricePerHour"
	instanceStoragePath                = ".InstanceStorageInfo.TotalSizeInGB"
	ebsOptimizedBaselineBandwidthPath  = ".EbsInfo.EbsOptimizedInfo.BaselineBandwidthInMbps"
	ebsOptimizedBaselineThroughputPath = ".EbsInfo.EbsOptimizedInfo.BaselineThroughputInMBps"
	ebsOptimizedBaselineIOPSPath       = ".EbsInfo.EbsOptimizedInfo.BaselineIops"
)

// sorterNode represents a sortable instance type which holds the value
// to sort by instance sort
type sorterNode struct {
	instanceType *instancetypes.Details
	fieldValue   reflect.Value
}

// sorter is used to sort instance types based on a sorting field
// and direction
type sorter struct {
	sorters      []*sorterNode
	sortField    string
	isDescending bool
}

// Sort sorts the given instance types by the given field in the given direction
//
// sortField is a json path to a field in the instancetypes.Details struct which represents
// the field to sort instance types by (Ex: ".MemoryInfo.SizeInMiB"). Quantity flags present
// in the CLI (memory, gpus, etc.) are also accepted.
//
// sortDirection represents the direction to sort in. Valid options: "ascending", "asc", "descending", "desc".
func Sort(instanceTypes []*instancetypes.Details, sortField string, sortDirection string) ([]*instancetypes.Details, error) {
	sortingKeysMap := map[string]string{
		VCPUs:                          vcpuPath,
		Memory:                         memoryPath,
		GPUMemoryTotal:                 gpuMemoryTotalPath,
		NetworkInterfaces:              networkInterfacesPath,
		SpotPrice:                      spotPricePath,
		ODPrice:                        odPricePath,
		InstanceStorage:                instanceStoragePath,
		EBSOptimizedBaselineBandwidth:  ebsOptimizedBaselineBandwidthPath,
		EBSOptimizedBaselineThroughput: ebsOptimizedBaselineThroughputPath,
		EBSOptimizedBaselineIOPS:       ebsOptimizedBaselineIOPSPath,
	}

	// determine if user used a shorthand for sorting flag
	if sortFieldShorthandPath, ok := sortingKeysMap[sortField]; ok {
		sortField = sortFieldShorthandPath
	}

	sorter, err := newSorter(instanceTypes, sortField, sortDirection)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when preparing to sort instance types: %v", err)
	}

	if err := sorter.sort(); err != nil {
		return nil, fmt.Errorf("an error occurred when sorting instance types: %v", err)
	}

	return sorter.instanceTypes(), nil
}

// newSorter creates a new Sorter object to be used to sort the given instance types
// based on the sorting field and direction
//
// sortField is a json path to a field in the instancetypes.Details struct which represents
// the field to sort instance types by (Ex: ".MemoryInfo.SizeInMiB").
//
// sortDirection represents the direction to sort in. Valid options: "ascending", "asc", "descending", "desc".
func newSorter(instanceTypes []*instancetypes.Details, sortField string, sortDirection string) (*sorter, error) {
	var isDescending bool
	switch sortDirection {
	case SortDescending, SortDesc:
		isDescending = true
	case SortAscending, SortAsc:
		isDescending = false
	default:
		return nil, fmt.Errorf("invalid sort direction: %s (valid options: %s, %s, %s, %s)", sortDirection, SortAscending, SortAsc, SortDescending, SortDesc)
	}

	sortField = formatSortField(sortField)

	// Create sorterNode objects for each instance type
	sorters := []*sorterNode{}
	for _, instanceType := range instanceTypes {
		newSorter, err := newSorterNode(instanceType, sortField)
		if err != nil {
			return nil, fmt.Errorf("error creating sorting node: %v", err)
		}

		sorters = append(sorters, newSorter)
	}

	return &sorter{
		sorters:      sorters,
		sortField:    sortField,
		isDescending: isDescending,
	}, nil
}

// formatSortField reformats sortField to match the expected json path format
// of the json lookup library. Format is unchanged if the sorting field
// matches one of the special flags.
func formatSortField(sortField string) string {
	// check to see if the sorting field matched one of the special exceptions
	if sortField == GPUCountField || sortField == InferenceAcceleratorsField {
		return sortField
	}

	return "$" + sortField
}

// newSorterNode creates a new sorterNode object which represents the given instance type
// and can be used in sorting of instance types based on the given sortField
func newSorterNode(instanceType *instancetypes.Details, sortField string) (*sorterNode, error) {
	// some important fields (such as gpu count) can not be accessed directly in the instancetypes.Details
	// struct, so we have special hard-coded flags to handle such cases
	switch sortField {
	case GPUCountField:
		gpuCount := getTotalGpusCount(instanceType)
		return &sorterNode{
			instanceType: instanceType,
			fieldValue:   reflect.ValueOf(gpuCount),
		}, nil
	case InferenceAcceleratorsField:
		acceleratorsCount := getTotalAcceleratorsCount(instanceType)
		return &sorterNode{
			instanceType: instanceType,
			fieldValue:   reflect.ValueOf(acceleratorsCount),
		}, nil
	}

	// convert instance type into json
	jsonInstanceType, err := json.Marshal(instanceType)
	if err != nil {
		return nil, err
	}

	// unmarshal json instance types in order to get proper format
	// for json path parsing
	var jsonData interface{}
	err = json.Unmarshal(jsonInstanceType, &jsonData)
	if err != nil {
		return nil, err
	}

	// get the desired field from the json data based on the passed in
	// json path
	result, err := jsonpath.JsonPathLookup(jsonData, sortField)
	if err != nil {
		// handle case where parent objects in path are null
		// by setting result to nil
		if err.Error() == "get attribute from null object" {
			result = nil
		} else {
			return nil, fmt.Errorf("error during json path lookup: %v", err)
		}
	}

	return &sorterNode{
		instanceType: instanceType,
		fieldValue:   reflect.ValueOf(result),
	}, nil
}

// sort the instance types in the Sorter based on the Sorter's sort field and
// direction
func (s *sorter) sort() error {
	if len(s.sorters) <= 1 {
		return nil
	}

	var sortErr error = nil

	sort.Slice(s.sorters, func(i int, j int) bool {
		valI := s.sorters[i].fieldValue
		valJ := s.sorters[j].fieldValue

		less, err := isLess(valI, valJ, s.isDescending)
		if err != nil {
			sortErr = err
		}

		return less
	})

	return sortErr
}

// isLess determines whether the first value (valI) is less than the
// second value (valJ) or not
func isLess(valI, valJ reflect.Value, isDescending bool) (bool, error) {
	switch valI.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// if valJ is not an int (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Int && vaJKind != reflect.Int8 && vaJKind != reflect.Int16 && vaJKind != reflect.Int32 && vaJKind != reflect.Int64 {
			return true, nil
		}

		if isDescending {
			return valI.Int() > valJ.Int(), nil
		} else {
			return valI.Int() <= valJ.Int(), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// if valJ is not a uint (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Uint && vaJKind != reflect.Uint8 && vaJKind != reflect.Uint16 && vaJKind != reflect.Uint32 && vaJKind != reflect.Uint64 {
			return true, nil
		}

		if isDescending {
			return valI.Uint() > valJ.Uint(), nil
		} else {
			return valI.Uint() <= valJ.Uint(), nil
		}
	case reflect.Float32, reflect.Float64:
		// if valJ is not a float (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		vaJKind := valJ.Kind()
		if vaJKind != reflect.Float32 && vaJKind != reflect.Float64 {
			return true, nil
		}

		if isDescending {
			return valI.Float() > valJ.Float(), nil
		} else {
			return valI.Float() <= valJ.Float(), nil
		}
	case reflect.String:
		// if valJ is not a string (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		if valJ.Kind() != reflect.String {
			return true, nil
		}

		if isDescending {
			return strings.Compare(valI.String(), valJ.String()) > 0, nil
		} else {
			return strings.Compare(valI.String(), valJ.String()) <= 0, nil
		}
	case reflect.Pointer:
		// Handle nil values by making non nil values always less than the nil values. That way the
		// nil values can be bubbled up to the end of the list.
		if valI.IsNil() {
			return false, nil
		} else if valJ.Kind() != reflect.Pointer || valJ.IsNil() {
			return true, nil
		}

		return isLess(valI.Elem(), valJ.Elem(), isDescending)
	case reflect.Bool:
		// if valJ is not a bool (can occur if the other value is nil)
		// then valI is less. This will bubble invalid values to the end
		if valJ.Kind() != reflect.Bool {
			return true, nil
		}

		if isDescending {
			return !valI.Bool(), nil
		} else {
			return valI.Bool(), nil
		}
	case reflect.Invalid:
		// handle invalid values (like nil values) by making valid values
		// always less than the invalid values. That way the invalid values
		// always bubble up to the end of the list
		return false, nil
	default:
		// unsortable value
		return false, fmt.Errorf("unsortable value")
	}
}

// instanceTypes returns the list of instance types held in the Sorter
func (s *sorter) instanceTypes() []*instancetypes.Details {
	instanceTypes := []*instancetypes.Details{}

	for _, node := range s.sorters {
		instanceTypes = append(instanceTypes, node.instanceType)
	}

	return instanceTypes
}

// helper functions for special sorting fields

// getTotalGpusCount calculates the number of gpus in the given instance type
func getTotalGpusCount(instanceType *instancetypes.Details) *int64 {
	gpusInfo := instanceType.GpuInfo

	if gpusInfo == nil {
		return nil
	}

	total := aws.Int64(0)
	for _, gpu := range gpusInfo.Gpus {
		total = aws.Int64(*total + *gpu.Count)
	}

	return total
}

// getTotalAcceleratorsCount calculates the total number of inference accelerators
// in the given instance type
func getTotalAcceleratorsCount(instanceType *instancetypes.Details) *int64 {
	acceleratorInfo := instanceType.InferenceAcceleratorInfo

	if acceleratorInfo == nil {
		return nil
	}

	total := aws.Int64(0)
	for _, accel := range acceleratorInfo.Accelerators {
		total = aws.Int64(*total + *accel.Count)
	}

	return total
}
