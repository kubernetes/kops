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
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
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
	GetOnDemandInstanceTypeCost(ctx context.Context, instanceType ec2types.InstanceType) (float64, error)
	GetSpotInstanceTypeNDayAvgCost(ctx context.Context, instanceType ec2types.InstanceType, availabilityZones []string, days int) (float64, error)
	RefreshOnDemandCache(ctx context.Context) error
	RefreshSpotCache(ctx context.Context, days int) error
	OnDemandCacheCount() int
	SpotCacheCount() int
	Save() error
}

// use us-east-1 since pricing only has endpoints in us-east-1 and ap-south-1
// TODO: In the future we may want to allow the client to select which endpoint is used through some mechanism
//       but that would likely happen through overriding this entire function as its signature is fixed
func modifyPricingRegion(opt *pricing.Options) {
	opt.Region = "us-east-1"
}

// New creates an instance of instance-selector EC2Pricing
func New(ctx context.Context, cfg aws.Config) (*EC2Pricing, error) {
	pricingClient := pricing.NewFromConfig(cfg, modifyPricingRegion)
	ec2Client := ec2.NewFromConfig(cfg)
	return &EC2Pricing{
		ODPricing:   LoadODCacheOrNew(ctx, pricingClient, cfg.Region, 0, ""),
		SpotPricing: LoadSpotCacheOrNew(ctx, ec2Client, cfg.Region, 0, "", DefaultSpotDaysBack),
	}, nil
}

func NewWithCache(ctx context.Context, cfg aws.Config, ttl time.Duration, cacheDir string) (*EC2Pricing, error) {
	pricingClient := pricing.NewFromConfig(cfg, modifyPricingRegion)
	ec2Client := ec2.NewFromConfig(cfg)
	return &EC2Pricing{
		ODPricing:   LoadODCacheOrNew(ctx, pricingClient, cfg.Region, ttl, cacheDir),
		SpotPricing: LoadSpotCacheOrNew(ctx, ec2Client, cfg.Region, ttl, cacheDir, DefaultSpotDaysBack),
	}, nil
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
func (p *EC2Pricing) GetSpotInstanceTypeNDayAvgCost(ctx context.Context, instanceType ec2types.InstanceType, availabilityZones []string, days int) (float64, error) {
	if len(availabilityZones) == 0 {
		return p.SpotPricing.Get(ctx, instanceType, "", days)
	}
	costs := []float64{}
	var errs error
	for _, zone := range availabilityZones {
		cost, err := p.SpotPricing.Get(ctx, instanceType, zone, days)
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
func (p *EC2Pricing) GetOnDemandInstanceTypeCost(ctx context.Context, instanceType ec2types.InstanceType) (float64, error) {
	return p.ODPricing.Get(ctx, instanceType)
}

// RefreshOnDemandCache makes a bulk request to the pricing api to retrieve all instance type pricing and stores them in a local cache
func (p *EC2Pricing) RefreshOnDemandCache(ctx context.Context) error {
	return p.ODPricing.Refresh(ctx)
}

// RefreshSpotCache makes a bulk request to the ec2 api to retrieve all spot instance type pricing and stores them in a local cache
func (p *EC2Pricing) RefreshSpotCache(ctx context.Context, days int) error {
	return p.SpotPricing.Refresh(ctx, days)
}

func (p *EC2Pricing) Save() error {
	return multierr.Append(p.ODPricing.Save(), p.SpotPricing.Save())
}
