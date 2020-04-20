/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

var outputPath = ""

func main() {
	flag.StringVar(&outputPath, "out", outputPath, "file to write")

	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if outputPath == "" {
		return fmt.Errorf("must specify output file with --out")
	}

	klog.Info("Beginning AWS Machine Refresh")

	// These are instance types not available in every account
	// If they're not available, they wont be in the ec2.DescribeInstanceTypes response
	// so they are hardcoded here for reference.
	// Note that the m6g instances do not have ENI or IP information
	machines := []awsup.AWSMachineTypeInfo{
		{
			Name:              "cr1.8xlarge",
			MemoryGB:          244,
			Cores:             32,
			InstanceENIs:      8,
			InstanceIPsPerENI: 30,
			EphemeralDisks:    []int{120, 120},
		},
		{
			Name:              "hs1.8xlarge",
			MemoryGB:          117,
			Cores:             16,
			InstanceENIs:      8,
			InstanceIPsPerENI: 30,
			EphemeralDisks:    []int{2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000, 2000},
		},
		{
			Name:              "m6g.medium",
			MemoryGB:          4,
			Cores:             1,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.large",
			MemoryGB:          8,
			Cores:             2,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.xlarge",
			MemoryGB:          16,
			Cores:             4,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.2xlarge",
			MemoryGB:          32,
			Cores:             8,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.4xlarge",
			MemoryGB:          64,
			Cores:             16,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.8xlarge",
			MemoryGB:          128,
			Cores:             32,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.12xlarge",
			MemoryGB:          192,
			Cores:             48,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},

		{
			Name:              "m6g.16xlarge",
			MemoryGB:          256,
			Cores:             64,
			InstanceENIs:      0,
			InstanceIPsPerENI: 0,
			EphemeralDisks:    nil,
		},
	}
	families := map[string]struct{}{
		"cr1": {},
		"hs1": {},
		"m6g": {},
	}

	config := aws.NewConfig()
	// Give verbose errors on auth problems
	config = config.WithCredentialsChainVerboseErrors(true)
	// Default to us-east-1
	config = config.WithRegion("us-east-1")

	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	client := ec2.New(sess, config)
	instanceTypes := make([]*ec2.InstanceTypeInfo, 0)
	err = client.DescribeInstanceTypesPages(&ec2.DescribeInstanceTypesInput{},
		func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
			instanceTypes = append(instanceTypes, page.InstanceTypes...)
			return true
		})
	if err != nil {
		return err
	}

	var warnings []string

	seen := map[string]bool{}
	for _, typeInfo := range instanceTypes {
		instanceType := *typeInfo.InstanceType

		if _, ok := seen[instanceType]; ok {
			continue
		}
		seen[instanceType] = true
		machine := awsup.AWSMachineTypeInfo{
			Name:              instanceType,
			GPU:               typeInfo.GpuInfo != nil,
			InstanceENIs:      intValue(typeInfo.NetworkInfo.MaximumNetworkInterfaces),
			InstanceIPsPerENI: intValue(typeInfo.NetworkInfo.Ipv4AddressesPerInterface),
		}
		memoryGB := float64(intValue(typeInfo.MemoryInfo.SizeInMiB)) / 1024
		machine.MemoryGB = float32(math.Round(memoryGB*100) / 100)

		if typeInfo.VCpuInfo != nil && typeInfo.VCpuInfo.DefaultVCpus != nil {
			machine.Cores = intValue(typeInfo.VCpuInfo.DefaultVCpus)
		}
		if typeInfo.InstanceStorageInfo != nil && len(typeInfo.InstanceStorageInfo.Disks) > 0 {
			disks := make([]int, 0)
			for _, disk := range typeInfo.InstanceStorageInfo.Disks {
				for i := 0; i < intValue(disk.Count); i++ {
					disks = append(disks, intValue(disk.SizeInGB))
				}
			}
			machine.EphemeralDisks = disks
		}

		machines = append(machines, machine)

		family := strings.Split(instanceType, ".")[0]
		families[family] = struct{}{}

	}

	sortedFamilies := []string{}
	for f := range families {
		sortedFamilies = append(sortedFamilies, f)
	}
	sort.Strings(sortedFamilies)

	sort.Slice(machines, func(i, j int) bool {
		// Sort first by family
		tokensI := strings.Split(machines[i].Name, ".")
		tokensJ := strings.Split(machines[j].Name, ".")

		if tokensI[0] != tokensJ[0] {
			return tokensI[0] < tokensJ[0]
		}

		// Then sort by size within the family
		if machines[i].MemoryGB != machines[j].MemoryGB {
			return machines[i].MemoryGB < machines[j].MemoryGB
		}

		// Fallback: sort by name
		return machines[i].Name < machines[j].Name
	})

	var output string

	if len(warnings) != 0 {
		output = output + "\n"
		for _, warning := range warnings {
			output = output + "// WARNING: " + warning + "\n"
		}
		output = output + "\n"
	}

	for _, f := range sortedFamilies {
		output = output + fmt.Sprintf("\n// %s family", f)
		for _, m := range machines {
			if family := strings.Split(m.Name, ".")[0]; family == f {

				body := fmt.Sprintf(`
	{
		Name: "%s",
		MemoryGB: %v,
		Cores: %v,
		InstanceENIs: %v,
		InstanceIPsPerENI: %v,
	`, m.Name, m.MemoryGB, m.Cores, m.InstanceENIs, m.InstanceIPsPerENI)
				output = output + body

				// Avoid awkward []int(nil) syntax
				if len(m.EphemeralDisks) == 0 {
					output = output + "EphemeralDisks: nil,\n"
				} else {
					output = output + fmt.Sprintf("EphemeralDisks: %#v,\n", m.EphemeralDisks)
				}

				if m.GPU {
					output = output + "GPU: true,\n"
				}

				output = output + "},\n"
			}
		}
		output = output + "\n"
	}

	klog.Infof("Writing changes to %v", outputPath)

	fileInput, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", outputPath, err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(fileInput))

	scanner.Split(bufio.ScanLines)

	var newfile string
	flag := false
	done := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "END GENERATED CONTENT") {
			flag = false
			done = true
		}
		if !flag {
			newfile = newfile + line + "\n"
		}
		if strings.Contains(line, "BEGIN GENERATED CONTENT") {
			flag = true
			newfile = newfile + output
		}
	}

	if !done {
		return fmt.Errorf("BEGIN GENERATED CONTENT / END GENERATED CONTENT markers not found")
	}

	err = ioutil.WriteFile(outputPath, []byte(newfile), 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", outputPath, err)
	}

	klog.Info("Done.")
	klog.Flush()

	return nil
}

func intValue(v *int64) int {
	return int(aws.Int64Value(v))
}
