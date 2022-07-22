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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/pricing/pricingiface"
	"github.com/mitchellh/go-homedir"
	"github.com/patrickmn/go-cache"
	"go.uber.org/multierr"
)

const (
	ODCacheFileName = "on-demand-pricing-cache.json"
)

type OnDemandPricing struct {
	Region         string
	FullRefreshTTL time.Duration
	DirectoryPath  string
	cache          *cache.Cache
	pricingClient  pricingiface.PricingAPI
}

func LoadODCacheOrNew(pricingClient pricingiface.PricingAPI, region string, fullRefreshTTL time.Duration, directoryPath string) *OnDemandPricing {
	expandedDirPath, err := homedir.Expand(directoryPath)
	if err != nil {
		log.Printf("Unable to load on-demand pricing cache directory %s: %v", expandedDirPath, err)
		return &OnDemandPricing{
			Region:         region,
			FullRefreshTTL: 0,
			DirectoryPath:  directoryPath,
			cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
			pricingClient:  pricingClient,
		}
	}
	odPricing := &OnDemandPricing{
		Region:         region,
		FullRefreshTTL: fullRefreshTTL,
		DirectoryPath:  expandedDirPath,
		pricingClient:  pricingClient,
		cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
	}
	if fullRefreshTTL <= 0 {
		odPricing.Clear()
		return odPricing
	}
	// Start the cache refresh job
	go odCacheRefreshJob(odPricing)
	odCache, err := loadODCacheFrom(fullRefreshTTL, region, expandedDirPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("An on-demand pricing cache file could not be loaded: %v", err)
		}
		return odPricing
	}
	odPricing.cache = odCache
	return odPricing
}

func loadODCacheFrom(itemTTL time.Duration, region string, expandedDirPath string) (*cache.Cache, error) {
	cacheBytes, err := os.ReadFile(getODCacheFilePath(region, expandedDirPath))
	if err != nil {
		return nil, err
	}
	odCache := &map[string]cache.Item{}
	if err := json.Unmarshal(cacheBytes, odCache); err != nil {
		return nil, err
	}
	return cache.NewFrom(itemTTL, itemTTL, *odCache), nil
}

func getODCacheFilePath(region string, directoryPath string) string {
	return filepath.Join(directoryPath, fmt.Sprintf("%s-%s", region, ODCacheFileName))
}

func odCacheRefreshJob(odPricing *OnDemandPricing) {
	if odPricing.FullRefreshTTL <= 0 {
		return
	}
	refreshTicker := time.NewTicker(odPricing.FullRefreshTTL)
	for range refreshTicker.C {
		if err := odPricing.Refresh(); err != nil {
			log.Println(err)
		}
	}
}

func (c *OnDemandPricing) Refresh() error {
	odInstanceTypeCosts, err := c.fetchOnDemandPricing("")
	if err != nil {
		return fmt.Errorf("there was a problem refreshing the on-demand instance type pricing cache: %v", err)
	}
	for instanceType, cost := range odInstanceTypeCosts {
		c.cache.SetDefault(instanceType, cost)
	}
	if err := c.Save(); err != nil {
		return fmt.Errorf("unable to save the refreshed on-demand instance type pricing cache file: %v", err)
	}
	return nil
}

func (c *OnDemandPricing) Get(instanceType string) (float64, error) {
	if cost, ok := c.cache.Get(instanceType); ok {
		return cost.(float64), nil
	}
	costs, err := c.fetchOnDemandPricing(instanceType)
	if err != nil {
		return 0, fmt.Errorf("there was a problem fetching on-demand instance type pricing for %s: %v", instanceType, err)
	}
	c.cache.SetDefault(instanceType, costs[instanceType])
	return costs[instanceType], nil
}

// Count of items in the cache
func (c *OnDemandPricing) Count() int {
	return c.cache.ItemCount()
}

func (c *OnDemandPricing) Save() error {
	if c.FullRefreshTTL == 0 || c.Count() == 0 {
		return nil
	}
	cacheBytes, err := json.Marshal(c.cache.Items())
	if err != nil {
		return err
	}
	os.Mkdir(c.DirectoryPath, 0755)
	return ioutil.WriteFile(getODCacheFilePath(c.Region, c.DirectoryPath), cacheBytes, 0644)
}

func (c *OnDemandPricing) Clear() error {
	c.cache.Flush()
	return os.Remove(getODCacheFilePath(c.Region, c.DirectoryPath))
}

// fetchOnDemandPricing makes a bulk request to the pricing api to retrieve all instance type pricing if the instanceType is the empty string
//   or, if instanceType is specified, it can request a specific instance type pricing
func (c *OnDemandPricing) fetchOnDemandPricing(instanceType string) (map[string]float64, error) {
	odPricing := map[string]float64{}
	productInput := pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters:     c.getProductsInputFilters(instanceType),
	}
	var processingErr error
	errAPI := c.pricingClient.GetProductsPages(&productInput, func(pricingOutput *pricing.GetProductsOutput, nextPage bool) bool {
		for _, priceDoc := range pricingOutput.PriceList {
			instanceTypeName, price, errParse := c.parseOndemandUnitPrice(priceDoc)
			if errParse != nil {
				processingErr = multierr.Append(processingErr, errParse)
				continue
			}
			odPricing[instanceTypeName] = price
		}
		return true
	})
	if errAPI != nil {
		return odPricing, errAPI
	}
	return odPricing, processingErr
}

func (c *OnDemandPricing) getProductsInputFilters(instanceType string) []*pricing.Filter {
	regionDescription := c.getRegionForPricingAPI()
	filters := []*pricing.Filter{
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("ServiceCode"), Value: aws.String(serviceCode)},
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("operatingSystem"), Value: aws.String("linux")},
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("location"), Value: aws.String(regionDescription)},
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("capacitystatus"), Value: aws.String("used")},
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("preInstalledSw"), Value: aws.String("NA")},
		{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("tenancy"), Value: aws.String("shared")},
	}
	if instanceType != "" {
		filters = append(filters, &pricing.Filter{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("instanceType"), Value: aws.String(instanceType)})
	}
	return filters
}

// getRegionForPricingAPI attempts to retrieve the region description based on the AWS session used to create
// the ec2pricing struct. It then uses the endpoints package in the aws sdk to retrieve the region description
// This is necessary because the pricing API uses the region description rather than a region ID
func (c *OnDemandPricing) getRegionForPricingAPI() string {
	endpointResolver := endpoints.DefaultResolver()
	partitions := endpointResolver.(endpoints.EnumPartitions).Partitions()

	// use us-east-1 as the default
	regionDescription := "US East (N. Virginia)"
	for _, partition := range partitions {
		regions := partition.Regions()
		if region, ok := regions[c.Region]; ok {
			regionDescription = region.Description()
		}
	}

	// endpoints package returns European regions with the word "Europe," but the pricing API expects the word "EU."
	// This formatting mismatch is only present with European regions.
	// So replace "Europe" with "EU" if it exists in the regionDescription string.
	regionDescription = strings.ReplaceAll(regionDescription, "Europe", "EU")

	return regionDescription
}

// parseOndemandUnitPrice takes a priceList from the pricing API and parses its weirdness
func (c *OnDemandPricing) parseOndemandUnitPrice(priceList aws.JSONValue) (string, float64, error) {
	// TODO: this could probably be cleaned up a bit by adding a couple structs with json tags
	//       We still need to some weird for-loops to get at elements under json keys that are IDs...
	//       But it would probably be cleaner than this.
	attributes, ok := priceList["product"].(map[string]interface{})["attributes"]
	if !ok {
		return "", float64(-1.0), fmt.Errorf("unable to find product attributes")
	}
	instanceTypeName, ok := attributes.(map[string]interface{})["instanceType"].(string)
	if !ok {
		return "", float64(-1.0), fmt.Errorf("unable to find instance type name from product attributes")
	}
	terms, ok := priceList["terms"]
	if !ok {
		return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find pricing terms")
	}
	ondemandTerms, ok := terms.(map[string]interface{})["OnDemand"]
	if !ok {
		return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find on-demand pricing terms")
	}
	for _, priceDimensions := range ondemandTerms.(map[string]interface{}) {
		dim, ok := priceDimensions.(map[string]interface{})["priceDimensions"]
		if !ok {
			return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find on-demand pricing dimensions")
		}
		for _, dimension := range dim.(map[string]interface{}) {
			dims := dimension.(map[string]interface{})
			pricePerUnit, ok := dims["pricePerUnit"]
			if !ok {
				return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find on-demand price per unit in pricing dimensions")
			}
			pricePerUnitInUSDStr, ok := pricePerUnit.(map[string]interface{})["USD"]
			if !ok {
				return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find on-demand price per unit in USD")
			}
			var err error
			pricePerUnitInUSD, err := strconv.ParseFloat(pricePerUnitInUSDStr.(string), 64)
			if err != nil {
				return instanceTypeName, float64(-1.0), fmt.Errorf("could not convert price per unit in USD to a float64")
			}
			return instanceTypeName, pricePerUnitInUSD, nil
		}
	}
	return instanceTypeName, float64(-1.0), fmt.Errorf("unable to parse pricing doc")
}
