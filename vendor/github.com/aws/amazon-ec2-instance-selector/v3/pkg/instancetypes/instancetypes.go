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

package instancetypes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/mitchellh/go-homedir"
	"github.com/patrickmn/go-cache"
)

var CacheFileName = "ec2-instance-types.json"

// Details hold all the information on an ec2 instance type.
type Details struct {
	ec2types.InstanceTypeInfo
	OndemandPricePerHour *float64
	SpotPrice            *float64
}

type Provider struct {
	Region          string
	DirectoryPath   string
	FullRefreshTTL  time.Duration
	lastFullRefresh *time.Time
	ec2Client       ec2.DescribeInstanceTypesAPIClient
	cache           *cache.Cache
	logger          *log.Logger
}

// NewProvider creates a new Instance Types provider used to fetch Instance Type information from EC2.
func NewProvider(region string, ec2Client ec2.DescribeInstanceTypesAPIClient) *Provider {
	return &Provider{
		Region:         region,
		DirectoryPath:  "",
		FullRefreshTTL: 0,
		ec2Client:      ec2Client,
		cache:          cache.New(0, 0),
		logger:         log.New(io.Discard, "", 0),
	}
}

// NewProvider creates a new Instance Types provider used to fetch Instance Type information from EC2 and optionally cache.
func LoadFromOrNew(directoryPath string, region string, ttl time.Duration, ec2Client ec2.DescribeInstanceTypesAPIClient) (*Provider, error) {
	expandedDirPath, err := homedir.Expand(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load instance-type cache directory %s: %w", expandedDirPath, err)
	}
	if ttl <= 0 {
		provider := NewProvider(region, ec2Client)
		if err := provider.Clear(); err != nil {
			return nil, err
		}
		return provider, nil
	}
	itCache, err := loadFrom(ttl, region, expandedDirPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to load instance-type cache from %s: %w", expandedDirPath, err)
	}
	if err != nil {
		itCache = cache.New(0, 0)
	}
	return &Provider{
		Region:         region,
		DirectoryPath:  expandedDirPath,
		ec2Client:      ec2Client,
		cache:          itCache,
		FullRefreshTTL: ttl,
		logger:         log.New(io.Discard, "", 0),
	}, nil
}

func loadFrom(ttl time.Duration, region string, expandedDirPath string) (*cache.Cache, error) {
	itemTTL := ttl + time.Second
	cacheBytes, err := os.ReadFile(getCacheFilePath(region, expandedDirPath))
	if err != nil {
		return nil, err
	}
	itCache := &map[string]cache.Item{}
	if err := json.Unmarshal(cacheBytes, itCache); err != nil {
		return nil, err
	}
	return cache.NewFrom(itemTTL, itemTTL, *itCache), nil
}

func getCacheFilePath(region string, expandedDirPath string) string {
	return filepath.Join(expandedDirPath, fmt.Sprintf("%s-%s", region, CacheFileName))
}

func (p *Provider) SetLogger(logger *log.Logger) {
	p.logger = logger
}

func (p *Provider) Get(ctx context.Context, instanceTypes []ec2types.InstanceType) ([]*Details, error) {
	p.logger.Printf("Getting instance types %v", instanceTypes)
	start := time.Now()
	calls := 0
	defer func() {
		p.logger.Printf("Took %s and %d calls to collect Instance Types", time.Since(start), calls)
	}()
	instanceTypeDetails := []*Details{}
	describeInstanceTypeOpts := &ec2.DescribeInstanceTypesInput{}
	if len(instanceTypes) != 0 {
		for _, it := range instanceTypes {
			if cachedIT, ok := p.cache.Get(string(it)); ok {
				instanceTypeDetails = append(instanceTypeDetails, cachedIT.(*Details))
			} else {
				// need to reassign, so we're not sharing the loop iterators memory space
				instanceType := it
				describeInstanceTypeOpts.InstanceTypes = append(describeInstanceTypeOpts.InstanceTypes, instanceType)
			}
		}
		// if we were able to retrieve all from cache, return here, else continue to do a remote lookup
		if len(describeInstanceTypeOpts.InstanceTypes) == 0 {
			return instanceTypeDetails, nil
		}
	} else if p.lastFullRefresh != nil && !p.isFullRefreshNeeded() {
		for _, item := range p.cache.Items() {
			instanceTypeDetails = append(instanceTypeDetails, item.Object.(*Details))
		}
		return instanceTypeDetails, nil
	}

	s := ec2.NewDescribeInstanceTypesPaginator(p.ec2Client, describeInstanceTypeOpts)

	for s.HasMorePages() {
		calls++
		instanceTypeOutput, err := s.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next instance types page, %w", err)
		}
		for _, instanceTypeInfo := range instanceTypeOutput.InstanceTypes {
			itDetails := &Details{InstanceTypeInfo: instanceTypeInfo}
			instanceTypeDetails = append(instanceTypeDetails, itDetails)
			p.cache.SetDefault(string(instanceTypeInfo.InstanceType), itDetails)
		}
	}

	if len(instanceTypes) == 0 {
		now := time.Now().UTC()
		p.lastFullRefresh = &now
		if err := p.Save(); err != nil {
			return instanceTypeDetails, err
		}
	}
	return instanceTypeDetails, nil
}

func (p *Provider) isFullRefreshNeeded() bool {
	return time.Since(*p.lastFullRefresh) > p.FullRefreshTTL
}

func (p *Provider) Save() error {
	if p.FullRefreshTTL <= 0 || p.cache.ItemCount() == 0 {
		return nil
	}
	cacheBytes, err := json.Marshal(p.cache.Items())
	if err != nil {
		return err
	}
	if err := os.Mkdir(p.DirectoryPath, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return os.WriteFile(getCacheFilePath(p.Region, p.DirectoryPath), cacheBytes, 0600)
}

func (p *Provider) Clear() error {
	p.cache.Flush()
	if err := os.Remove(getCacheFilePath(p.Region, p.DirectoryPath)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *Provider) CacheCount() int {
	return p.cache.ItemCount()
}
