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
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
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

	// Not currently in the API
	t2CreditsPerHour := map[string]float32{
		"t1.micro":   1,
		"t2.nano":    3,
		"t2.micro":   6,
		"t2.small":   12,
		"t2.medium":  24,
		"t2.large":   36,
		"t2.xlarge":  54,
		"t2.2xlarge": 81.6,
		"t3.nano":    6,
		"t3.micro":   12,
		"t3.small":   24,
		"t3.medium":  24,
		"t3.large":   36,
		"t3.xlarge":  96,
		"t3.2xlarge": 192,
	}

	machines := []awsup.AWSMachineTypeInfo{}
	families := make(map[string]struct{})

	prices := []aws.JSONValue{}

	config := aws.NewConfig()
	// Give verbose errors on auth problems
	config = config.WithCredentialsChainVerboseErrors(true)
	// Default to us-east-1
	config = config.WithRegion("us-east-1")

	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	svc := pricing.New(sess, config)
	typeTerm := pricing.FilterTypeTermMatch
	input := &pricing.GetProductsInput{
		Filters: []*pricing.Filter{
			{
				Field: aws.String("operatingSystem"),
				Type:  &typeTerm,
				Value: aws.String("Linux"),
			},
			{
				Field: aws.String("tenancy"),
				Type:  &typeTerm,
				Value: aws.String("shared"),
			},
			{
				Field: aws.String("location"),
				Type:  &typeTerm,
				Value: aws.String("US East (N. Virginia)"),
			},
			{
				Field: aws.String("preInstalledSw"),
				Type:  &typeTerm,
				Value: aws.String("NA"),
			},
		},
		FormatVersion: aws.String("aws_v1"),
		ServiceCode:   aws.String("AmazonEC2"),
	}

	for {
		result, err := svc.GetProducts(input)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case pricing.ErrCodeInternalErrorException:
					return fmt.Errorf("%s: %v", pricing.ErrCodeInternalErrorException, aerr)
				case pricing.ErrCodeInvalidParameterException:
					return fmt.Errorf("%s: %v", pricing.ErrCodeInvalidParameterException, aerr)
				case pricing.ErrCodeNotFoundException:
					return fmt.Errorf("%s: %v", pricing.ErrCodeNotFoundException, aerr)
				case pricing.ErrCodeInvalidNextTokenException:
					return fmt.Errorf("%s: %v", pricing.ErrCodeInvalidNextTokenException, aerr)
				case pricing.ErrCodeExpiredNextTokenException:
					return fmt.Errorf("%s: %v", pricing.ErrCodeExpiredNextTokenException, aerr)
				default:
					return aerr
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				return err
			}
		}

		prices = append(prices, result.PriceList...)

		if result.NextToken != nil {
			input.NextToken = result.NextToken
		} else {
			break
		}
	}

	var warnings []string

	seen := map[string]bool{}
	for _, item := range prices {
		for k, v := range item {
			if k == "product" {
				product := v.(map[string]interface{})
				attributes := map[string]string{}
				for k, v := range product["attributes"].(map[string]interface{}) {
					attributes[k] = v.(string)
				}

				instanceType := attributes["instanceType"]

				if _, ok := seen[instanceType]; ok {
					continue
				}
				seen[instanceType] = true

				machine := awsup.AWSMachineTypeInfo{
					Name:  instanceType,
					Cores: stringToInt(attributes["vcpu"]),
				}

				memory := strings.TrimSuffix(attributes["memory"], " GiB")
				machine.MemoryGB = stringToFloat32(memory)

				if attributes["storage"] != "EBS only" {
					storage := strings.Split(attributes["storage"], " ")
					var size int
					var count int
					if len(storage) > 1 {
						count = stringToInt(storage[0])
						if storage[2] == "NVMe" {
							count = 1
							size = stringToInt(storage[0])
						} else {
							size = stringToInt(storage[2])
						}
					} else {
						count = 0
					}

					ephemeralDisks := []int{}
					for i := 0; i < count; i++ {
						ephemeralDisks = append(ephemeralDisks, size)
					}

					machine.EphemeralDisks = ephemeralDisks
				}

				if attributes["instanceFamily"] == "GPU instance" {
					machine.GPU = true
				}

				if attributes["ecu"] == "Variable" {
					machine.Burstable = true
					machine.ECU = t2CreditsPerHour[machine.Name] // This is actually credits * ECUs, but we'll add that later
				} else if attributes["ecu"] == "NA" {
					machine.ECU = 0
				} else {
					machine.ECU = stringToFloat32(attributes["ecu"])
				}

				if enis, enisOK := InstanceENIsAvailable[instanceType]; enisOK {
					machine.InstanceENIs = enis
				} else {
					warnings = append(warnings, fmt.Sprintf("ENIs not known for %s", instanceType))
				}

				if ipsPerENI, ipsOK := InstanceIPsAvailable[instanceType]; ipsOK {
					machine.InstanceIPsPerENI = int(ipsPerENI)
				} else {
					warnings = append(warnings, fmt.Sprintf("IPs per ENI not known for %s", instanceType))
				}

				machines = append(machines, machine)

				family := strings.Split(instanceType, ".")[0]
				families[family] = struct{}{}

			}
		}
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
				var ecu string
				if m.Burstable {
					ecu = fmt.Sprintf("%v * BurstableCreditsToECUS", m.ECU)
				} else {
					ecu = fmt.Sprint(m.ECU)
				}

				body := fmt.Sprintf(`
	{
		Name: "%s",
		MemoryGB: %v,
		ECU: %v,
		Cores: %v,
		InstanceENIs: %v,
		InstanceIPsPerENI: %v,
	`, m.Name, m.MemoryGB, ecu, m.Cores, m.InstanceENIs, m.InstanceIPsPerENI)
				output = output + body

				// Avoid awkward []int(nil) syntax
				if len(m.EphemeralDisks) == 0 {
					output = output + "EphemeralDisks: nil,\n"
				} else {
					output = output + fmt.Sprintf("EphemeralDisks: %#v,\n", m.EphemeralDisks)
				}

				if m.Burstable {
					output = output + "Burstable: true,\n"
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

func stringToFloat32(s string) float32 {
	// For 1,000 case
	clean := strings.Replace(s, ",", "", -1)
	value, err := strconv.ParseFloat(clean, 32)
	if err != nil {
		klog.Errorf("error converting string to float32: %v", err)
	}
	return float32(value)
}

func stringToInt(s string) int {
	// For 1,000 case
	clean := strings.Replace(s, ",", "", -1)
	value, err := strconv.Atoi(clean)
	if err != nil {
		klog.Error(err)
	}
	return value
}
