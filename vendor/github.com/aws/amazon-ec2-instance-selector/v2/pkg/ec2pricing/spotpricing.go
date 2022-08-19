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
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/mitchellh/go-homedir"
	"github.com/patrickmn/go-cache"
	"go.uber.org/multierr"
)

const (
	SpotCacheFileName = "spot-pricing-cache.gob"
)

type SpotPricing struct {
	Region         string
	FullRefreshTTL time.Duration
	DirectoryPath  string
	cache          *cache.Cache
	ec2Client      ec2iface.EC2API
	sync.RWMutex
}

type spotPricingEntry struct {
	Timestamp time.Time
	SpotPrice float64
	Zone      string
}

func LoadSpotCacheOrNew(ec2Client ec2iface.EC2API, region string, fullRefreshTTL time.Duration, directoryPath string, days int) *SpotPricing {
	expandedDirPath, err := homedir.Expand(directoryPath)
	if err != nil {
		log.Printf("Unable to load spot pricing cache directory %s: %v", expandedDirPath, err)
		return &SpotPricing{
			Region:         region,
			FullRefreshTTL: 0,
			DirectoryPath:  directoryPath,
			cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
			ec2Client:      ec2Client,
		}
	}
	spotPricing := &SpotPricing{
		Region:         region,
		FullRefreshTTL: fullRefreshTTL,
		DirectoryPath:  expandedDirPath,
		ec2Client:      ec2Client,
		cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
	}
	if fullRefreshTTL <= 0 {
		spotPricing.Clear()
		return spotPricing
	}
	gob.Register([]*spotPricingEntry{})
	// Start the cache refresh job
	go spotCacheRefreshJob(spotPricing, days)
	spotCache, err := loadSpotCacheFrom(fullRefreshTTL, region, expandedDirPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("A spot pricing cache file could not be loaded: %v", err)
		}
		return spotPricing
	}
	spotPricing.cache = spotCache
	return spotPricing
}

func loadSpotCacheFrom(itemTTL time.Duration, region string, expandedDirPath string) (*cache.Cache, error) {
	file, err := os.Open(getSpotCacheFilePath(region, expandedDirPath))
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	spotTimeSeries := &map[string]cache.Item{}
	if err := decoder.Decode(spotTimeSeries); err != nil {
		return nil, err
	}
	c := cache.NewFrom(itemTTL, itemTTL, *spotTimeSeries)
	c.DeleteExpired()
	return c, nil
}

func getSpotCacheFilePath(region string, directoryPath string) string {
	return filepath.Join(directoryPath, fmt.Sprintf("%s-%s", region, SpotCacheFileName))
}

func spotCacheRefreshJob(spotPricing *SpotPricing, days int) {
	if spotPricing.FullRefreshTTL <= 0 {
		return
	}
	refreshTicker := time.NewTicker(spotPricing.FullRefreshTTL)
	for range refreshTicker.C {
		if err := spotPricing.Refresh(days); err != nil {
			log.Println(err)
		}
	}
}

func (c *SpotPricing) Refresh(days int) error {
	c.Lock()
	defer c.Unlock()
	spotInstanceTypeCosts, err := c.fetchSpotPricingTimeSeries("", days)
	if err != nil {
		return fmt.Errorf("there was a problem refreshing the spot instance type pricing cache: %v", err)
	}
	for instanceTypeAndZone, cost := range spotInstanceTypeCosts {
		c.cache.SetDefault(instanceTypeAndZone, cost)
	}
	if err := c.Save(); err != nil {
		return fmt.Errorf("unable to save the refreshed spot instance type pricing cache file: %v", err)
	}
	return nil
}

func (c *SpotPricing) Get(instanceType string, zone string, days int) (float64, error) {
	entries, ok := c.cache.Get(instanceType)
	if zone != "" && ok {
		if !c.contains(zone, entries.([]*spotPricingEntry)) {
			ok = false
		}
	}
	if !ok {
		c.RLock()
		defer c.RUnlock()
		zonalSpotPricing, err := c.fetchSpotPricingTimeSeries(instanceType, days)
		if err != nil {
			return -1, fmt.Errorf("there was a problem fetching spot instance type pricing for %s: %v", instanceType, err)
		}
		for instanceType, costs := range zonalSpotPricing {
			c.cache.SetDefault(instanceType, costs)
		}
	}

	entries, ok = c.cache.Get(instanceType)
	if !ok {
		return -1, fmt.Errorf("unable to get spot pricing for %s in zone %s for %d days back", instanceType, zone, days)
	}
	return c.calculateSpotAggregate(c.filterOn(zone, entries.([]*spotPricingEntry))), nil
}

func (c *SpotPricing) contains(zone string, entries []*spotPricingEntry) bool {
	for _, entry := range entries {
		if entry.Zone == zone {
			return true
		}
	}
	return false
}

func (c *SpotPricing) calculateSpotAggregate(spotPriceEntries []*spotPricingEntry) float64 {
	if len(spotPriceEntries) == 0 {
		return 0.0
	}
	if len(spotPriceEntries) == 1 {
		return spotPriceEntries[0].SpotPrice
	}
	// Sort slice by timestamp in descending order from the end time (most likely, now)
	sort.Slice(spotPriceEntries, func(i, j int) bool {
		return spotPriceEntries[i].Timestamp.After(spotPriceEntries[j].Timestamp)
	})

	endTime := spotPriceEntries[0].Timestamp
	startTime := spotPriceEntries[len(spotPriceEntries)-1].Timestamp
	totalDuration := endTime.Sub(startTime).Minutes()

	priceSum := float64(0)
	for i, entry := range spotPriceEntries {
		duration := spotPriceEntries[int(math.Max(float64(i-1), 0))].Timestamp.Sub(entry.Timestamp).Minutes()
		priceSum += duration * entry.SpotPrice
	}
	return priceSum / totalDuration
}

func (c *SpotPricing) filterOn(zone string, pricingEntries []*spotPricingEntry) []*spotPricingEntry {
	filtered := []*spotPricingEntry{}
	for _, entry := range pricingEntries {
		// this takes the first zone, might be better to do all zones instead...
		if zone == "" {
			zone = entry.Zone
		}
		if entry.Zone == zone {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Count of items in the cache
func (c *SpotPricing) Count() int {
	return c.cache.ItemCount()
}

func (c *SpotPricing) Save() error {
	if c.FullRefreshTTL <= 0 || c.Count() == 0 {
		return nil
	}
	if err := os.Mkdir(c.DirectoryPath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	file, err := os.Create(getSpotCacheFilePath(c.Region, c.DirectoryPath))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	return encoder.Encode(c.cache.Items())
}

func (c *SpotPricing) Clear() error {
	c.Lock()
	defer c.Unlock()
	c.cache.Flush()
	return os.Remove(getSpotCacheFilePath(c.Region, c.DirectoryPath))
}

// fetchSpotPricingTimeSeries makes a bulk request to the ec2 api to retrieve all spot instance type pricing for the past n days
// If instanceType is empty, it will fetch for all instance types
func (c *SpotPricing) fetchSpotPricingTimeSeries(instanceType string, days int) (map[string][]*spotPricingEntry, error) {
	spotTimeSeries := map[string][]*spotPricingEntry{}
	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Hour * time.Duration(24*-1*days))
	spotPriceHistInput := ec2.DescribeSpotPriceHistoryInput{
		ProductDescriptions: []*string{aws.String(productDescription)},
		StartTime:           &startTime,
		EndTime:             &endTime,
	}
	if instanceType != "" {
		spotPriceHistInput.InstanceTypes = append(spotPriceHistInput.InstanceTypes, &instanceType)
	}
	var processingErr error
	errAPI := c.ec2Client.DescribeSpotPriceHistoryPages(&spotPriceHistInput, func(dspho *ec2.DescribeSpotPriceHistoryOutput, b bool) bool {
		for _, history := range dspho.SpotPriceHistory {
			spotPrice, errFloat := strconv.ParseFloat(*history.SpotPrice, 64)
			if errFloat != nil {
				processingErr = multierr.Append(processingErr, errFloat)
				continue
			}
			instanceType := *history.InstanceType
			spotTimeSeries[instanceType] = append(spotTimeSeries[instanceType], &spotPricingEntry{
				Timestamp: *history.Timestamp,
				SpotPrice: spotPrice,
				Zone:      *history.AvailabilityZone,
			})
		}
		return true
	})
	if errAPI != nil {
		return spotTimeSeries, errAPI
	}
	return spotTimeSeries, processingErr
}
