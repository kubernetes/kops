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
	"log"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	supported = "supported"
	required  = "required"
)

var amdRegex = regexp.MustCompile("[a-zA-Z0-9]+a\\.[a-zA-Z0-9]")

func isSupportedFromString(instanceTypeValue *string, target *string) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	return *instanceTypeValue == *target
}

func isSupportedFromStrings(instanceTypeValues []*string, target *string) bool {
	if target == nil {
		return true
	}
	return contains(instanceTypeValues, *target)
}

func isSupportedWithRangeInt(instanceTypeValue *int, target *IntRangeFilter) bool {
	var instanceTypeValueInt64 *int64
	if instanceTypeValue != nil {
		nonPtr := int64(*instanceTypeValue)
		instanceTypeValueInt64 = &nonPtr
	}
	return isSupportedWithRangeInt64(instanceTypeValueInt64, target)
}

func isSupportedWithFloat64(instanceTypeValue *float64, target *float64) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	// compare up to values' two decimal floor
	return math.Floor(*instanceTypeValue*100)/100 == math.Floor(*target*100)/100
}

func isSupportedUsageClassType(instanceTypeValue []ec2types.UsageClassType, target *ec2types.UsageClassType) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}

	for _, potentialType := range instanceTypeValue {
		if potentialType == *target {
			return true
		}
	}
	return false
}

func isSupportedArchitectureType(instanceTypeValue []ec2types.ArchitectureType, target *ec2types.ArchitectureType) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}

	for _, potentialType := range instanceTypeValue {
		if potentialType == *target {
			return true
		}
	}
	return false
}

func isSupportedVirtualizationType(instanceTypeValue []ec2types.VirtualizationType, target *ec2types.VirtualizationType) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}
	for _, potentialType := range instanceTypeValue {
		if potentialType == *target {
			return true
		}
	}
	return false
}

func isSupportedInstanceTypeHypervisorType(instanceTypeValue ec2types.InstanceTypeHypervisor, target *ec2types.InstanceTypeHypervisor) bool {
	if target == nil {
		return true
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}
	if instanceTypeValue == *target {
		return true
	}
	return false
}

func isSupportedRootDeviceType(instanceTypeValue []ec2types.RootDeviceType, target *ec2types.RootDeviceType) bool {
	if target == nil {
		return true
	}
	if instanceTypeValue == nil {
		return false
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}
	for _, potentialType := range instanceTypeValue {
		if potentialType == *target {
			return true
		}
	}
	return false
}

func isMatchingCpuArchitecture(instanceTypeValue CPUManufacturer, target *CPUManufacturer) bool {
	if target == nil {
		return true
	}
	if reflect.ValueOf(*target).IsZero() {
		return true
	}
	if instanceTypeValue == *target {
		return true
	}
	return false
}

func isSupportedWithRangeInt64(instanceTypeValue *int64, target *IntRangeFilter) bool {
	if target == nil {
		return true
	} else if instanceTypeValue == nil && target.LowerBound == 0 && target.UpperBound == 0 {
		return true
	} else if instanceTypeValue == nil {
		return false
	}
	return int(*instanceTypeValue) >= target.LowerBound && int(*instanceTypeValue) <= target.UpperBound
}

func isSupportedWithRangeInt32(instanceTypeValue *int32, target *Int32RangeFilter) bool {
	if target == nil {
		return true
	} else if instanceTypeValue == nil && target.LowerBound == 0 && target.UpperBound == 0 {
		return true
	} else if instanceTypeValue == nil {
		return false
	}
	return *instanceTypeValue >= target.LowerBound && *instanceTypeValue <= target.UpperBound
}

func isSupportedWithRangeUint64(instanceTypeValue *int64, target *Uint64RangeFilter) bool {
	if target == nil {
		return true
	} else if instanceTypeValue == nil && target.LowerBound == 0 && target.UpperBound == 0 {
		return true
	} else if instanceTypeValue == nil {
		return false
	}
	if target.UpperBound > math.MaxInt64 {
		target.UpperBound = math.MaxInt64
	}
	return uint64(*instanceTypeValue) >= target.LowerBound && uint64(*instanceTypeValue) <= target.UpperBound
}

func isSupportedWithRangeFloat64(instanceTypeValue *float64, target *Float64RangeFilter) bool {
	if target == nil {
		return true
	} else if instanceTypeValue == nil && target.LowerBound == 0.0 && target.UpperBound == 0.0 {
		return true
	} else if instanceTypeValue == nil {
		return false
	}
	return float64(*instanceTypeValue) >= target.LowerBound && float64(*instanceTypeValue) <= target.UpperBound
}

func isSupportedWithBool(instanceTypeValue *bool, target *bool) bool {
	if target == nil {
		return true
	}
	return *target == *instanceTypeValue
}

// Helper functions for aggregating data parsed from AWS API calls

func getTotalAcceleratorsCount(acceleratorInfo *ec2types.InferenceAcceleratorInfo) *int32 {
	if acceleratorInfo == nil {
		return nil
	}
	total := int32(0)
	for _, accel := range acceleratorInfo.Accelerators {
		total = total + *accel.Count
	}
	return &total
}

func getTotalGpusCount(gpusInfo *ec2types.GpuInfo) *int32 {
	if gpusInfo == nil {
		return nil
	}
	total := int32(0)
	for _, gpu := range gpusInfo.Gpus {
		total = total + *gpu.Count
	}
	return &total
}

func getTotalGpuMemory(gpusInfo *ec2types.GpuInfo) *int64 {
	if gpusInfo == nil {
		return nil
	}
	return aws.Int64(int64(*gpusInfo.TotalGpuMemoryInMiB))
}

func getGPUManufacturers(gpusInfo *ec2types.GpuInfo) []*string {
	if gpusInfo == nil {
		return nil
	}
	var manufacturers []*string
	for _, info := range gpusInfo.Gpus {
		manufacturers = append(manufacturers, info.Manufacturer)
	}
	return manufacturers
}

func getGPUModels(gpusInfo *ec2types.GpuInfo) []*string {
	if gpusInfo == nil {
		return nil
	}
	var models []*string
	for _, info := range gpusInfo.Gpus {
		models = append(models, info.Name)
	}
	return models
}

func getInferenceAcceleratorManufacturers(acceleratorInfo *ec2types.InferenceAcceleratorInfo) []*string {
	if acceleratorInfo == nil {
		return nil
	}
	var manufacturers []*string
	for _, info := range acceleratorInfo.Accelerators {
		manufacturers = append(manufacturers, info.Manufacturer)
	}
	return manufacturers
}

func getInferenceAcceleratorModels(acceleratorInfo *ec2types.InferenceAcceleratorInfo) []*string {
	if acceleratorInfo == nil {
		return nil
	}
	var models []*string
	for _, info := range acceleratorInfo.Accelerators {
		models = append(models, info.Name)
	}
	return models
}

func getNetworkPerformance(networkPerformance *string) *int {
	if networkPerformance == nil {
		return aws.Int(-1)
	}
	re, err := regexp.Compile(`[0-9]+ Gigabit`)
	if err != nil {
		log.Printf("Unable to compile regexp to parse network performance: %s\n", *networkPerformance)
		return nil
	}
	networkBandwidth := re.FindString(*networkPerformance)
	if networkBandwidth == "" {
		return aws.Int(-1)
	}
	bandwidthAndUnit := strings.Split(networkBandwidth, " ")
	if len(bandwidthAndUnit) != 2 {
		return aws.Int(-1)
	}
	bandwidthNumber, err := strconv.Atoi(bandwidthAndUnit[0])
	if err != nil {
		return aws.Int(-1)
	}
	return aws.Int(bandwidthNumber)
}

func getInstanceStorage(instanceStorageInfo *ec2types.InstanceStorageInfo) *int64 {
	if instanceStorageInfo == nil {
		return aws.Int64(0)
	}
	return aws.Int64(*instanceStorageInfo.TotalSizeInGB * 1024)
}

func getDiskType(instanceStorageInfo *ec2types.InstanceStorageInfo) *string {
	if instanceStorageInfo == nil || len(instanceStorageInfo.Disks) == 0 {
		return nil
	}
	return aws.String(string(instanceStorageInfo.Disks[0].Type))
}

func getNVMESupport(instanceStorageInfo *ec2types.InstanceStorageInfo, ebsInfo *ec2types.EbsInfo) *bool {
	if instanceStorageInfo != nil {
		return supportSyntaxToBool(aws.String(string(instanceStorageInfo.NvmeSupport)))
	}
	if ebsInfo != nil {
		return supportSyntaxToBool(aws.String(string(ebsInfo.EbsOptimizedSupport)))
	}
	return aws.Bool(false)
}

func getDiskEncryptionSupport(instanceStorageInfo *ec2types.InstanceStorageInfo, ebsInfo *ec2types.EbsInfo) *bool {
	if instanceStorageInfo != nil {
		encryptionSupport := string(instanceStorageInfo.EncryptionSupport)
		return supportSyntaxToBool(&encryptionSupport)
	}
	if ebsInfo != nil {
		ebsEncryptionSupport := string(ebsInfo.EncryptionSupport)
		return supportSyntaxToBool(&ebsEncryptionSupport)
	}
	return aws.Bool(false)
}

func getEBSOptimizedBaselineBandwidth(ebsInfo *ec2types.EbsInfo) *int32 {
	if ebsInfo == nil || ebsInfo.EbsOptimizedInfo == nil {
		return nil
	}
	return ebsInfo.EbsOptimizedInfo.BaselineBandwidthInMbps
}

func getEBSOptimizedBaselineThroughput(ebsInfo *ec2types.EbsInfo) *float64 {
	if ebsInfo == nil || ebsInfo.EbsOptimizedInfo == nil {
		return nil
	}
	return ebsInfo.EbsOptimizedInfo.BaselineThroughputInMBps
}

func getEBSOptimizedBaselineIOPS(ebsInfo *ec2types.EbsInfo) *int32 {
	if ebsInfo == nil || ebsInfo.EbsOptimizedInfo == nil {
		return nil
	}
	return ebsInfo.EbsOptimizedInfo.BaselineIops
}

func getCPUManufacturer(instanceTypeInfo *ec2types.InstanceTypeInfo) CPUManufacturer {
	for _, it := range instanceTypeInfo.ProcessorInfo.SupportedArchitectures {
		if it == ec2types.ArchitectureTypeArm64 {
			return CPUManufacturerAWS
		}
	}

	if amdRegex.Match([]byte(instanceTypeInfo.InstanceType)) {
		return CPUManufacturerAMD
	}
	return CPUManufacturerIntel
}

// supportSyntaxToBool takes an instance spec field that uses ["unsupported", "supported", "required", or "default"]
// and transforms it to a *bool to use in filter execution
func supportSyntaxToBool(instanceTypeSupport *string) *bool {
	if instanceTypeSupport == nil {
		return nil
	}
	if strings.ToLower(*instanceTypeSupport) == required || strings.ToLower(*instanceTypeSupport) == supported || strings.ToLower(*instanceTypeSupport) == "default" {
		return aws.Bool(true)
	}
	return aws.Bool(false)
}

func calculateVCpusToMemoryRatio(vcpusVal *int32, memoryVal *int64) *float64 {
	if vcpusVal == nil || *vcpusVal == 0 || memoryVal == nil {
		return nil
	}
	// normalize vcpus to a mebivcpu value
	result := math.Ceil(float64(*memoryVal) / float64(*vcpusVal*1024))
	return &result
}

// Slice helper function

func contains(slice []*string, target string) bool {
	for _, it := range slice {
		if it != nil && strings.EqualFold(*it, target) {
			return true
		}
	}
	return false
}
