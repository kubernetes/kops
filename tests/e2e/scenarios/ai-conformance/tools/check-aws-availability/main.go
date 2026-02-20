/*
Copyright The Kubernetes Authors.

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
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	var region string
	var instanceType string

	flag.StringVar(&region, "region", "", "AWS Region")
	flag.StringVar(&instanceType, "instance-type", "", "EC2 Instance Type")
	flag.Parse()

	if region == "" || instanceType == "" {
		fmt.Println("Usage: check-aws-availability -region <region> -instance-type <type>")
		os.Exit(1)
	}

	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: types.LocationTypeAvailabilityZone,
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-type"),
				Values: []string{instanceType},
			},
		},
	}

	result, err := client.DescribeInstanceTypeOfferings(ctx, input)
	if err != nil {
		fmt.Printf("Error describing instance type offerings: %v\n", err)
		os.Exit(1)
	}

	if len(result.InstanceTypeOfferings) > 0 {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}
