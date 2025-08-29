// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ec2pricing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
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
	pricingClient  pricing.GetProductsAPIClient
	logger         *log.Logger
	sync.RWMutex
}

type PricingList struct {
	Product         PricingListProduct `json:"product"`
	ServiceCode     string             `json:"serviceCode"`
	Terms           ProductTerms       `json:"terms"`
	Version         string             `json:"version"`
	PublicationDate string             `json:"publicationDate"`
}

type PricingListProduct struct {
	ProductFamily     string            `json:"productFamily"`
	ProductAttributes map[string]string `json:"attributes"`
	SKU               string            `json:"sku"`
}

type ProductTerms struct {
	OnDemand map[string]ProductPricingInfo `json:"OnDemand"`
	Reserved map[string]ProductPricingInfo `json:"Reserved"`
}

type ProductPricingInfo struct {
	PriceDimensions map[string]PriceDimensionInfo `json:"priceDimensions"`
	SKU             string                        `json:"sku"`
	EffectiveDate   string                        `json:"effectiveDate"`
	OfferTermCode   string                        `json:"offerTermCode"`
	TermAttributes  map[string]string             `json:"termAttributes"`
}

type PriceDimensionInfo struct {
	Unit         string            `json:"unit"`
	EndRange     string            `json:"endRange"`
	Description  string            `json:"description"`
	AppliesTo    []string          `json:"appliesTo"`
	RateCode     string            `json:"rateCode"`
	BeginRange   string            `json:"beginRange"`
	PricePerUnit map[string]string `json:"pricePerUnit"`
}

func LoadODCacheOrNew(ctx context.Context, pricingClient pricing.GetProductsAPIClient, region string, fullRefreshTTL time.Duration, directoryPath string) (*OnDemandPricing, error) {
	expandedDirPath, err := homedir.Expand(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load on-demand pricing cache directory %s: %w", expandedDirPath, err)
	}
	odPricing := &OnDemandPricing{
		Region:         region,
		FullRefreshTTL: fullRefreshTTL,
		DirectoryPath:  expandedDirPath,
		pricingClient:  pricingClient,
		cache:          cache.New(fullRefreshTTL, fullRefreshTTL),
		logger:         log.New(io.Discard, "", 0),
	}
	if fullRefreshTTL <= 0 {
		if err := odPricing.Clear(); err != nil {
			return nil, fmt.Errorf("unable to clear od pricing cache due to ttl <= 0 %w", err)
		}
		return odPricing, nil
	}
	// Start the cache refresh job
	go odPricing.odCacheRefreshJob(ctx)
	odCache, err := loadODCacheFrom(fullRefreshTTL, region, expandedDirPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("an on-demand pricing cache file could not be loaded: %v", err)
	}
	if err != nil {
		odCache = cache.New(0, 0)
	}
	odPricing.cache = odCache
	return odPricing, nil
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
	c := cache.NewFrom(itemTTL, itemTTL, *odCache)
	c.DeleteExpired()
	return c, nil
}

func getODCacheFilePath(region string, directoryPath string) string {
	return filepath.Join(directoryPath, fmt.Sprintf("%s-%s", region, ODCacheFileName))
}

func (c *OnDemandPricing) odCacheRefreshJob(ctx context.Context) {
	if c.FullRefreshTTL <= 0 {
		return
	}
	refreshTicker := time.NewTicker(c.FullRefreshTTL)
	for range refreshTicker.C {
		if err := c.Refresh(ctx); err != nil {
			c.logger.Printf("Periodic OD Cache Refresh Error: %v", err)
		}
	}
}

func (c *OnDemandPricing) SetLogger(logger *log.Logger) {
	c.logger = logger
}

func (c *OnDemandPricing) Refresh(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()
	odInstanceTypeCosts, err := c.fetchOnDemandPricing(ctx, "")
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

func (c *OnDemandPricing) Get(ctx context.Context, instanceType ec2types.InstanceType) (float64, error) {
	if cost, ok := c.cache.Get(string(instanceType)); ok {
		return cost.(float64), nil
	}
	c.RLock()
	defer c.RUnlock()
	costs, err := c.fetchOnDemandPricing(ctx, instanceType)
	if err != nil {
		return 0, fmt.Errorf("there was a problem fetching on-demand instance type pricing for %s: %v", instanceType, err)
	}
	c.cache.SetDefault(string(instanceType), costs[string(instanceType)])
	return costs[string(instanceType)], nil
}

// Count of items in the cache.
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
	if err := os.Mkdir(c.DirectoryPath, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return os.WriteFile(getODCacheFilePath(c.Region, c.DirectoryPath), cacheBytes, 0600)
}

func (c *OnDemandPricing) Clear() error {
	c.Lock()
	defer c.Unlock()
	c.cache.Flush()
	if err := os.Remove(getODCacheFilePath(c.Region, c.DirectoryPath)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// fetchOnDemandPricing makes a bulk request to the pricing api to retrieve all instance type pricing if the instanceType is the empty string
//
//	or, if instanceType is specified, it can request a specific instance type pricing
func (c *OnDemandPricing) fetchOnDemandPricing(ctx context.Context, instanceType ec2types.InstanceType) (map[string]float64, error) {
	start := time.Now()
	calls := 0
	defer func() {
		c.logger.Printf("Took %s and %d calls to collect OD pricing", time.Since(start), calls)
	}()
	odPricing := map[string]float64{}
	productInput := pricing.GetProductsInput{
		ServiceCode: c.StringMe(serviceCode),
		Filters:     c.getProductsInputFilters(instanceType),
	}
	var processingErr error

	p := pricing.NewGetProductsPaginator(c.pricingClient, &productInput)

	for p.HasMorePages() {
		calls++
		pricingOutput, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next OD pricing page, %w", err)
		}

		for _, priceDoc := range pricingOutput.PriceList {
			instanceTypeName, price, errParse := c.parseOndemandUnitPrice(priceDoc)
			if errParse != nil {
				processingErr = multierr.Append(processingErr, errParse)
				continue
			}
			odPricing[instanceTypeName] = price
		}
	}
	return odPricing, processingErr
}

// StringMe takes an interface and returns a pointer to a string value
// If the underlying interface kind is not string or *string then nil is returned.
func (c *OnDemandPricing) StringMe(i interface{}) *string {
	if i == nil {
		return nil
	}
	switch v := i.(type) {
	case *string:
		return v
	case string:
		return &v
	default:
		c.logger.Printf("%s cannot be converted to a string", i)
		return nil
	}
}

func (c *OnDemandPricing) getProductsInputFilters(instanceType ec2types.InstanceType) []pricingtypes.Filter {
	filters := []pricingtypes.Filter{
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("ServiceCode"), Value: c.StringMe(serviceCode)},
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("operatingSystem"), Value: c.StringMe("linux")},
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("regionCode"), Value: c.StringMe(c.Region)},
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("capacitystatus"), Value: c.StringMe("used")},
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("preInstalledSw"), Value: c.StringMe("NA")},
		{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("tenancy"), Value: c.StringMe("shared")},
	}
	if instanceType != "" {
		filters = append(filters, pricingtypes.Filter{Type: pricingtypes.FilterTypeTermMatch, Field: c.StringMe("instanceType"), Value: c.StringMe(string(instanceType))})
	}
	return filters
}

// parseOndemandUnitPrice takes a priceList from the pricing API and parses its weirdness.
func (c *OnDemandPricing) parseOndemandUnitPrice(priceList string) (string, float64, error) {
	var productPriceList PricingList
	err := json.Unmarshal([]byte(priceList), &productPriceList)
	if err != nil {
		return "", float64(-1.0), fmt.Errorf("unable to parse pricing doc: %w", err)
	}
	attributes := productPriceList.Product.ProductAttributes
	instanceTypeName := attributes["instanceType"]

	for _, priceDimensions := range productPriceList.Terms.OnDemand {
		dim := priceDimensions.PriceDimensions
		for _, dimension := range dim {
			pricePerUnit := dimension.PricePerUnit
			pricePerUnitInUSDStr, ok := pricePerUnit["USD"]
			if !ok {
				return instanceTypeName, float64(-1.0), fmt.Errorf("unable to find on-demand price per unit in USD")
			}
			var err error
			pricePerUnitInUSD, err := strconv.ParseFloat(pricePerUnitInUSDStr, 64)
			if err != nil {
				return instanceTypeName, float64(-1.0), fmt.Errorf("could not convert price per unit in USD to a float64")
			}
			return instanceTypeName, pricePerUnitInUSD, nil
		}
	}
	return instanceTypeName, float64(-1.0), fmt.Errorf("unable to parse pricing doc")
}
