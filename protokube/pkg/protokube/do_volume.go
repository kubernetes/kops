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

package protokube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"

	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdo "k8s.io/kops/protokube/pkg/gossip/do"
)

const (
	dropletRegionMetadataURL     = "http://169.254.169.254/metadata/v1/region"
	dropletNameMetadataURL       = "http://169.254.169.254/metadata/v1/hostname"
	dropletIDMetadataURL         = "http://169.254.169.254/metadata/v1/id"
	dropletIDMetadataTags        = "http://169.254.169.254/metadata/v1/tags"
	dropletInternalIPMetadataURL = "http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address"
)

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

type DOCloudProvider struct {
	ClusterID  string
	godoClient *godo.Client

	region      string
	dropletName string
	dropletID   int
	dropletTags []string
}

var _ CloudProvider = &DOCloudProvider{}

func GetClusterID() (string, error) {
	clusterID := ""

	dropletTags, err := getMetadataDropletTags()
	if err != nil {
		return clusterID, fmt.Errorf("GetClusterID failed - unable to retrieve droplet tags: %s", err)
	}

	for _, dropletTag := range dropletTags {
		if strings.Contains(dropletTag, "KubernetesCluster:") {
			clusterID = strings.Replace(dropletTag, ".", "-", -1)

			tokens := strings.Split(clusterID, ":")
			if len(tokens) != 2 {
				return clusterID, fmt.Errorf("invalid clusterID (expected two tokens): %q", clusterID)
			}

			clusterID := tokens[1]

			return clusterID, nil
		}
	}

	return clusterID, fmt.Errorf("failed to get droplet clusterID")
}

func NewDOCloudProvider() (*DOCloudProvider, error) {
	region, err := getMetadataRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet region: %s", err)
	}

	dropletIDStr, err := getMetadataDropletID()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet id: %s", err)
	}
	dropletID, err := strconv.Atoi(dropletIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert droplet ID to int: %s", err)
	}

	dropletName, err := getMetadataDropletName()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet name: %s", err)
	}

	godoClient, err := NewDOCloud()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize digitalocean cloud: %s", err)
	}

	dropletTags, err := getMetadataDropletTags()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet tags: %s", err)
	}

	clusterID, err := GetClusterID()
	if err != nil {
		return nil, fmt.Errorf("failed to get clusterID: %s", err)
	}

	return &DOCloudProvider{
		godoClient:  godoClient,
		ClusterID:   clusterID,
		dropletID:   dropletID,
		dropletName: dropletName,
		region:      region,
		dropletTags: dropletTags,
	}, nil
}

// Token() returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func NewDOCloud() (*godo.Client, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := godo.NewClient(oauthClient)

	return client, nil
}

// getEtcdClusterSpec returns etcd.EtcdClusterSpec which holds
// necessary information required for starting an etcd server.
// DigitalOcean support on kops only supports single master setup for now
// but in the future when it supports multiple masters this method be
// updated to handle that case.
// TODO: use tags once it's supported for volumes
func (d *DOCloudProvider) getEtcdClusterSpec(vol godo.Volume) (*etcd.EtcdClusterSpec, error) {
	nodeName := d.dropletName

	var clusterKey string
	if strings.Contains(vol.Name, "etcd-main") {
		clusterKey = "main"
	} else if strings.Contains(vol.Name, "etcd-events") {
		clusterKey = "events"
	} else {
		return nil, fmt.Errorf("could not determine etcd cluster type for volume: %s", vol.Name)
	}

	return &etcd.EtcdClusterSpec{
		ClusterKey: clusterKey,
		NodeName:   nodeName,
		NodeNames:  []string{nodeName},
	}, nil
}

func (d *DOCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	for _, dropletTag := range d.dropletTags {
		if strings.Contains(dropletTag, strings.Replace(d.ClusterID, ".", "-", -1)) {
			return gossipdo.NewSeedProvider(d.godoClient, dropletTag)
		}
	}

	return nil, fmt.Errorf("could not determine a matching droplet tag for gossip seeding")
}

func (d *DOCloudProvider) InstanceID() string {
	return d.dropletName
}

func getMetadataRegion() (string, error) {
	return getMetadata(dropletRegionMetadataURL)
}

func getMetadataDropletName() (string, error) {
	return getMetadata(dropletNameMetadataURL)
}

func getMetadataDropletID() (string, error) {
	return getMetadata(dropletIDMetadataURL)
}

func getMetadataDropletTags() ([]string, error) {
	tagString, err := getMetadata(dropletIDMetadataTags)
	return strings.Split(tagString, "\n"), err
}

func getMetadata(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("droplet metadata returned non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
