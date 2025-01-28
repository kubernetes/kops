package hget

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"k8s.io/klog/v2"
)

type Index struct {
	assets []*Asset
}

type Asset struct {
	URL    string
	SHA256 string
}

func NewIndex() *Index {
	return &Index{}
}

func (i *Index) AddToIndex(ctx context.Context, indexURL string) error {
	log := klog.FromContext(ctx)

	baseURL, err := url.Parse(indexURL)
	if err != nil {
		return fmt.Errorf("parsing base URL: %w", err)
	}

	lastToken := path.Base(baseURL.Path)
	baseURL.Path = strings.TrimSuffix(baseURL.Path, lastToken)

	results, err := readLines(ctx, indexURL, func(line string) error {
		tokens := strings.Fields(line)
		if len(tokens) != 2 {
			log.Info("ignoring unknown line", "line", line)
			return nil
		}
		assetURL := baseURL.JoinPath(tokens[1])
		asset := &Asset{
			URL:    assetURL.String(),
			SHA256: tokens[0],
		}
		i.assets = append(i.assets, asset)
		return nil
	})
	if err != nil {
		return err
	}
	log.Info("downloaded index", "results", results)
	return nil
}

// Lookup tries to find assets in the index that match the given SHA256 hash
func (i *Index) Lookup(ctx context.Context, sha256 string) (*Asset, error) {
	// log := klog.FromContext(ctx)

	for _, asset := range i.assets {
		if asset.SHA256 == sha256 {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("asset not found")
}
