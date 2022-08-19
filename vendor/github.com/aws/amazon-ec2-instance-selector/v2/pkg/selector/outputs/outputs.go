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
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/instancetypes"
)

const columnTag = "column"

// wideColumnsData stores the data that should be displayed on each column
// of a wide output row
type wideColumnsData struct {
	instanceName       string `column:"Instance Type"`
	vcpu               int64  `column:"VCPUs"`
	memory             string `column:"Mem (GiB)"`
	hypervisor         string `column:"Hypervisor"`
	currentGen         bool   `column:"Current Gen"`
	hibernationSupport bool   `column:"Hibernation Support"`
	cpuArch            string `column:"CPU Arch"`
	networkPerformance string `column:"Network Performance"`
	eni                int64  `column:"ENIs"`
	gpu                int64  `column:"GPUs"`
	gpuMemory          string `column:"GPU Mem (GiB)"`
	gpuInfo            string `column:"GPU Info"`
	odPrice            string `column:"On-Demand Price/Hr"`
	spotPrice          string `column:"Spot Price/Hr (30d avg)"`
}

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
	w.Init(buf, 8, 8, 2, ' ', 0)
	defer w.Flush()

	columnDataStruct := wideColumnsData{}
	headers := []interface{}{}
	structType := reflect.TypeOf(columnDataStruct)
	for i := 0; i < structType.NumField(); i++ {
		columnHeader := structType.Field(i).Tag.Get(columnTag)
		headers = append(headers, columnHeader)
	}
	separators := make([]interface{}, 0)

	headerFormat := ""
	for _, header := range headers {
		headerFormat = headerFormat + "%s\t"
		separators = append(separators, strings.Repeat("-", len(header.(string))))
	}
	fmt.Fprintf(w, headerFormat, headers...)
	fmt.Fprintf(w, "\n"+headerFormat, separators...)

	columnsData := getWideColumnsData(instanceTypeInfoSlice)

	for _, data := range columnsData {
		fmt.Fprintf(w, "\n%s\t%d\t%s\t%s\t%t\t%t\t%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\t",
			data.instanceName,
			data.vcpu,
			data.memory,
			data.hypervisor,
			data.currentGen,
			data.hibernationSupport,
			data.cpuArch,
			data.networkPerformance,
			data.eni,
			data.gpu,
			data.gpuMemory,
			data.gpuInfo,
			data.odPrice,
			data.spotPrice,
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

// getWideColumnsData returns the column data necessary for a wide output for each of
// the given instance types
func getWideColumnsData(instanceTypes []*instancetypes.Details) []*wideColumnsData {
	columnsData := []*wideColumnsData{}

	for _, instanceType := range instanceTypes {
		none := "none"
		hyperisor := instanceType.Hypervisor
		if hyperisor == nil {
			hyperisor = &none
		}

		cpuArchitectures := []string{}
		for _, cpuArch := range instanceType.ProcessorInfo.SupportedArchitectures {
			cpuArchitectures = append(cpuArchitectures, *cpuArch)
		}

		gpus := int64(0)
		gpuMemory := int64(0)
		gpuType := []string{}
		if instanceType.GpuInfo != nil {
			gpuMemory = *instanceType.GpuInfo.TotalGpuMemoryInMiB
			for _, gpuInfo := range instanceType.GpuInfo.Gpus {
				gpus = gpus + *gpuInfo.Count
				gpuType = append(gpuType, *gpuInfo.Manufacturer+" "+*gpuInfo.Name)
			}
		} else {
			gpuType = append(gpuType, none)
		}

		onDemandPricePerHourStr := "-Not Fetched-"
		spotPricePerHourStr := "-Not Fetched-"
		if instanceType.OndemandPricePerHour != nil {
			onDemandPricePerHourStr = "$" + formatFloat(*instanceType.OndemandPricePerHour)
		}
		if instanceType.SpotPrice != nil {
			spotPricePerHourStr = "$" + formatFloat(*instanceType.SpotPrice)
		}

		newColumn := wideColumnsData{
			instanceName:       *instanceType.InstanceType,
			vcpu:               *instanceType.VCpuInfo.DefaultVCpus,
			memory:             formatFloat(float64(*instanceType.MemoryInfo.SizeInMiB) / 1024.0),
			hypervisor:         *hyperisor,
			currentGen:         *instanceType.CurrentGeneration,
			hibernationSupport: *instanceType.HibernationSupported,
			cpuArch:            strings.Join(cpuArchitectures, ", "),
			networkPerformance: *instanceType.NetworkInfo.NetworkPerformance,
			eni:                *instanceType.NetworkInfo.MaximumNetworkInterfaces,
			gpu:                gpus,
			gpuMemory:          formatFloat(float64(gpuMemory) / 1024.0),
			gpuInfo:            strings.Join(gpuType, ", "),
			odPrice:            onDemandPricePerHourStr,
			spotPrice:          spotPricePerHourStr,
		}

		columnsData = append(columnsData, &newColumn)
	}

	return columnsData
}

// getUnderlyingValue returns the underlying value of the given
// reflect.Value type
func getUnderlyingValue(value reflect.Value) interface{} {
	var val interface{}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val = value.Uint()
	case reflect.Float32, reflect.Float64:
		val = value.Float()
	case reflect.String:
		val = value.String()
	case reflect.Pointer:
		val = value.Pointer()
	case reflect.Bool:
		val = value.Bool()
	case reflect.Complex128, reflect.Complex64:
		val = value.Complex()
	case reflect.Interface:
		val = value.Interface()
	case reflect.UnsafePointer:
		val = value.UnsafePointer()
	default:
		val = nil
	}

	return val
}
