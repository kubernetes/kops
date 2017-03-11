package gce

import (
	"fmt"
	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/protokube/pkg/gossip"
	"strings"
)

type SeedProvider struct {
	compute   *compute.Service
	projectID string
	region    string
}

var _ gossip.SeedProvider = &SeedProvider{}

// Each page can have 500 results, but we cap how many pages
// are iterated through to prevent infinite loops if the API
// were to continuously return a nextPageToken.
const maxPages = 100

func (p *SeedProvider) GetSeeds() ([]string, error) {
	zones, err := p.compute.Zones.List(p.projectID).Do()
	if err != nil {
		return nil, fmt.Errorf("error querying for GCE zones: %v", err)
	}

	var zoneNames []string
	for _, zone := range zones.Items {
		regionName := lastComponent(zone.Region)
		if regionName != p.region {
			continue
		}
		zoneNames = append(zoneNames, zone.Name)
	}

	var seeds []string
	// TODO: Does it suffice to just query one zone (as long as we sort so it is always the first)?
	// Or does that introduce edges cases where we have partitions / cliques

	for _, zoneName := range zoneNames {
		pageToken := ""
		page := 0
		for ; page == 0 || (pageToken != "" && page < maxPages); page++ {
			listCall := p.compute.Instances.List(p.projectID, zoneName)

			// TODO: Filter by fields (but ask about google issue 29524655)

			// TODO: Match clusterid?

			if pageToken != "" {
				listCall.PageToken(pageToken)
			}

			res, err := listCall.Do()
			if err != nil {
				return nil, err
			}
			pageToken = res.NextPageToken
			for _, i := range res.Items {
				// TODO: Expose multiple IPs topologies?

				for _, ni := range i.NetworkInterfaces {
					// TODO: Check e.g. Network

					if ni.NetworkIP != "" {
						seeds = append(seeds, ni.NetworkIP)
					}
				}
			}
		}
		if page >= maxPages {
			glog.Errorf("GetSeeds exceeded maxPages=%d for Instances.List: truncating.", maxPages)
		}
	}

	return seeds, nil
}

func NewSeedProvider(compute *compute.Service, region string, projectID string) (*SeedProvider, error) {
	return &SeedProvider{
		compute:   compute,
		region:    region,
		projectID: projectID,
	}, nil
}

// Returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func lastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
}
