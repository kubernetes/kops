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

// Package outputs provides types for implementing instance type output functions as well as prebuilt output functions.
package outputs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
)

// SimpleInstanceTypeOutput is an OutputFn which outputs a slice of instance type names
func SimpleInstanceTypeOutput(instanceTypeInfoSlice []*instancetypes.Details) []string {
	instanceTypeStrings := []string{}
	for _, instanceTypeInfo := range instanceTypeInfoSlice {
		instanceTypeStrings = append(instanceTypeStrings, *instanceTypeInfo.InstanceType)
	}
	return instanceTypeStrings
}

// VerboseInstanceTypeOutput is an OutputFn which outputs a slice of instance type names
func VerboseInstanceTypeOutput(instanceTypeInfoSlice []*instancetypes.Details) []string {
	output, err := json.MarshalIndent(instanceTypeInfoSlice, "", "    ")
	if err != nil {
		log.Println("Unable to convert instance type info to JSON")
		return []string{}
	}
	if string(output) == "[]" || string(output) == "null" {
		return []string{}
	}
	return []string{string(output)}
}

// TableOutputShort is an OutputFn which returns a CLI table for easy reading
func TableOutputShort(instanceTypeInfoSlice []*instancetypes.Details) []string {
	if len(instanceTypeInfoSlice) == 0 {
		return nil
	}
	w := new(tabwriter.Writer)
	buf := new(bytes.Buffer)
	w.Init(buf, 8, 8, 8, ' ', 0)
	defer w.Flush()

	headers := []interface{}{
		"Instance Type",
		"VCPUs",
		"Mem (GiB)",
	}
	separators := []interface{}{}

	headerFormat := ""
	for _, header := range headers {
		headerFormat = headerFormat + "%s\t"
		separators = append(separators, strings.Repeat("-", len(header.(string))))
	}
	fmt.Fprintf(w, headerFormat, headers...)
	fmt.Fprintf(w, "\n"+headerFormat, separators...)

	for _, instanceTypeInfo := range instanceTypeInfoSlice {
		fmt.Fprintf(w, "\n%s\t%d\t%s\t",
			*instanceTypeInfo.InstanceType,
			*instanceTypeInfo.VCpuInfo.DefaultVCpus,
			formatFloat(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB)/1024.0),
		)
	}
	w.Flush()
	return []string{buf.String()}
}

// TableOutputWide is an OutputFn which returns a detailed CLI table for easy reading
func TableOutputWide(instanceTypeInfoSlice []*instancetypes.Details) []string {
	if len(instanceTypeInfoSlice) == 0 {
		return nil
	}
	w := new(tabwriter.Writer)
	buf := new(bytes.Buffer)
	none := "none"
	w.Init(buf, 8, 8, 2, ' ', 0)
	defer w.Flush()

	onDemandPricePerHourHeader := "On-Demand Price/Hr"
	spotPricePerHourHeader := "Spot Price/Hr (30d avg)"

	headers := []interface{}{
		"Instance Type",
		"VCPUs",
		"Mem (GiB)",
		"Hypervisor",
		"Current Gen",
		"Hibernation Support",
		"CPU Arch",
		"Network Performance",
		"ENIs",
		"GPUs",
		"GPU Mem (GiB)",
		"GPU Info",
		onDemandPricePerHourHeader,
		spotPricePerHourHeader,
	}
	separators := make([]interface{}, 0)

	headerFormat := ""
	for _, header := range headers {
		headerFormat = headerFormat + "%s\t"
		separators = append(separators, strings.Repeat("-", len(header.(string))))
	}
	fmt.Fprintf(w, headerFormat, headers...)
	fmt.Fprintf(w, "\n"+headerFormat, separators...)

	for _, instanceTypeInfo := range instanceTypeInfoSlice {
		hypervisor := instanceTypeInfo.Hypervisor
		if hypervisor == nil {
			hypervisor = &none
		}
		cpuArchitectures := []string{}
		for _, cpuArch := range instanceTypeInfo.ProcessorInfo.SupportedArchitectures {
			cpuArchitectures = append(cpuArchitectures, *cpuArch)
		}
		gpus := int64(0)
		gpuMemory := int64(0)
		gpuType := []string{}
		if instanceTypeInfo.GpuInfo != nil {
			gpuMemory = *instanceTypeInfo.GpuInfo.TotalGpuMemoryInMiB
			for _, gpuInfo := range instanceTypeInfo.GpuInfo.Gpus {
				gpus = gpus + *gpuInfo.Count
				gpuType = append(gpuType, *gpuInfo.Manufacturer+" "+*gpuInfo.Name)
			}
		}

		onDemandPricePerHourStr := "-Not Fetched-"
		spotPricePerHourStr := "-Not Fetched-"
		if instanceTypeInfo.OndemandPricePerHour != nil {
			onDemandPricePerHourStr = fmt.Sprintf("$%s", formatFloat(*instanceTypeInfo.OndemandPricePerHour))
		}
		if instanceTypeInfo.SpotPrice != nil {
			spotPricePerHourStr = fmt.Sprintf("$%s", formatFloat(*instanceTypeInfo.SpotPrice))
		}

		fmt.Fprintf(w, "\n%s\t%d\t%s\t%s\t%t\t%t\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t",
			*instanceTypeInfo.InstanceType,
			*instanceTypeInfo.VCpuInfo.DefaultVCpus,
			formatFloat(float64(*instanceTypeInfo.MemoryInfo.SizeInMiB)/1024.0),
			*hypervisor,
			*instanceTypeInfo.CurrentGeneration,
			*instanceTypeInfo.HibernationSupported,
			strings.Join(cpuArchitectures, ", "),
			*instanceTypeInfo.NetworkInfo.NetworkPerformance,
			*instanceTypeInfo.NetworkInfo.MaximumNetworkInterfaces,
			gpus,
			formatFloat(float64(gpuMemory)/1024.0),
			strings.Join(gpuType, ", "),
			onDemandPricePerHourStr,
			spotPricePerHourStr,
		)
	}
	w.Flush()
	return []string{buf.String()}
}

// OneLineOutput is an output function which prints the instance type names on a single line separated by commas
func OneLineOutput(instanceTypeInfoSlice []*instancetypes.Details) []string {
	instanceTypeNames := []string{}
	for _, instanceType := range instanceTypeInfoSlice {
		instanceTypeNames = append(instanceTypeNames, *instanceType.InstanceType)
	}
	if len(instanceTypeNames) == 0 {
		return []string{}
	}
	return []string{strings.Join(instanceTypeNames, ",")}
}

func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 5, 64)
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return s
	}
	reversed := reverse(parts[0])
	withCommas := ""
	for i, p := range reversed {
		if i%3 == 0 && i != 0 {
			withCommas += ","
		}
		withCommas += string(p)
	}
	s = strings.Join([]string{reverse(withCommas), parts[1]}, ".")
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
