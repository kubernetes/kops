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

package ec2pricing

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	"go.uber.org/multierr"
)

const (
	productDescription = "Linux/UNIX (Amazon VPC)"
	serviceCode        = "AmazonEC2"
)

var (
	DefaultSpotDaysBack = 30
)

// EC2Pricing is the public struct to interface with AWS pricing APIs
type EC2Pricing struct {
	ODPricing   *OnDemandPricing
	SpotPricing *SpotPricing
}

// EC2PricingIface is the EC2Pricing interface mainly used to mock out ec2pricing during testing
type EC2PricingIface interface {
	GetOnDemandInstanceTypeCost(instanceType string) (float64, error)
	GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error)
	RefreshOnDemandCache() error
	RefreshSpotCache(days int) error
	OnDemandCacheCount() int
	SpotCacheCount() int
	Save() error
}

// New creates an instance of instance-selector EC2Pricing
func New(sess *session.Session) *EC2Pricing {
	// use us-east-1 since pricing only has endpoints in us-east-1 and ap-south-1
	pricingClient := pricing.New(sess.Copy(aws.NewConfig().WithRegion("us-east-1")))
	return &EC2Pricing{
		ODPricing:   LoadODCacheOrNew(pricingClient, *sess.Config.Region, 0, ""),
		SpotPricing: LoadSpotCacheOrNew(ec2.New(sess), *sess.Config.Region, 0, "", DefaultSpotDaysBack),
	}
}

func NewWithCache(sess *session.Session, ttl time.Duration, cacheDir string) *EC2Pricing {
	// use us-east-1 since pricing only has endpoints in us-east-1 and ap-south-1
	pricingClient := pricing.New(sess.Copy(aws.NewConfig().WithRegion("us-east-1")))
	return &EC2Pricing{
		ODPricing:   LoadODCacheOrNew(pricingClient, *sess.Config.Region, ttl, cacheDir),
		SpotPricing: LoadSpotCacheOrNew(ec2.New(sess), *sess.Config.Region, ttl, cacheDir, DefaultSpotDaysBack),
	}
}

// OnDemandCacheCount returns the number of items in the OD cache
func (p *EC2Pricing) OnDemandCacheCount() int {
	return p.ODPricing.Count()
}

// SpotCacheCount returns the number of items in the spot cache
func (p *EC2Pricing) SpotCacheCount() int {
	return p.SpotPricing.Count()
}

// GetSpotInstanceTypeNDayAvgCost retrieves the spot price history for a given AZ from the past N days and averages the price
// Passing an empty list for availabilityZones will retrieve avg cost for all AZs in the current AWSSession's region
func (p *EC2Pricing) GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error) {
	if len(availabilityZones) == 0 {
		return p.SpotPricing.Get(instanceType, "", days)
	}
	costs := []float64{}
	var errs error
	for _, zone := range availabilityZones {
		cost, err := p.SpotPricing.Get(instanceType, zone, days)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		costs = append(costs, cost)
	}

	if len(multierr.Errors(errs)) == len(availabilityZones) {
		return -1, errs
	}
	return costs[0], nil
}

// GetOnDemandInstanceTypeCost retrieves the on-demand hourly cost for the specified instance type
func (p *EC2Pricing) GetOnDemandInstanceTypeCost(instanceType string) (float64, error) {
	return p.ODPricing.Get(instanceType)
}

// RefreshOnDemandCache makes a bulk request to the pricing api to retrieve all instance type pricing and stores them in a local cache
func (p *EC2Pricing) RefreshOnDemandCache() error {
	return p.ODPricing.Refresh()
}

// RefreshSpotCache makes a bulk request to the ec2 api to retrieve all spot instance type pricing and stores them in a local cache
func (p *EC2Pricing) RefreshSpotCache(days int) error {
	return p.SpotPricing.Refresh(days)
}

func (p *EC2Pricing) Save() error {
	return multierr.Append(p.ODPricing.Save(), p.SpotPricing.Save())
}
