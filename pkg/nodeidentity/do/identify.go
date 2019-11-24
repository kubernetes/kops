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

package do

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/digitalocean/godo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/nodeidentity"
)

// nodeIdentifier identifies a node from DO
type nodeIdentifier struct {
	doClient *godo.Client
}

const (
	dropletRegionMetadataURL    = "http://169.254.169.254/metadata/v1/region"
	dropletTagInstanceGroupName = "kops-instancegroup"
)

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

// Token() returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// New creates and returns a nodeidentity.Identifier for Nodes running on DO
func New() (nodeidentity.Identifier, error) {
	region, err := getMetadataRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet region: %s", err)
	}

	godoClient, err := NewCloud(region)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize digitalocean cloud: %s", err)
	}

	return &nodeIdentifier{
		doClient: godoClient,
	}, nil
}

func getMetadataRegion() (string, error) {
	return getMetadata(dropletRegionMetadataURL)
}

// NewCloud returns a Cloud, expecting the env var DIGITALOCEAN_ACCESS_TOKEN
// NewCloud will return an err if DIGITALOCEAN_ACCESS_TOKEN is not defined
func NewCloud(region string) (*godo.Client, error) {
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

func getMetadata(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata URL %s: %v", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("droplet metadata returned non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata information %s: %v", url, err)
	}

	return string(bodyBytes), nil
}

// IdentifyNode queries DO for the node identity information
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID

	if providerID == "" {
		return nil, errors.New("provider ID cannot be empty")
	}

	const prefix = "digitalocean://"

	if !strings.HasPrefix(providerID, prefix) {
		return nil, fmt.Errorf("provider ID %q is missing prefix %q", providerID, prefix)
	}

	provIDNum := strings.TrimPrefix(providerID, prefix)
	if provIDNum == "" {
		return nil, errors.New("provider ID number cannot be empty")
	}

	dropletID, err := strconv.Atoi(provIDNum)
	if err != nil {
		return nil, fmt.Errorf("failed to convert provider ID number %q: %s", dropletID, err)
	}

	kopsGroup, err := i.getInstanceGroup(dropletID)
	if err != nil {
		return nil, err
	}

	info := &nodeidentity.Info{}
	info.InstanceGroup = kopsGroup

	return info, nil
}

func (i *nodeIdentifier) getInstanceGroup(instanceID int) (string, error) {

	ctx := context.TODO()
	droplet, _, err := i.doClient.Droplets.Get(ctx, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve droplet via api for dropletid = %d. Error = %v", instanceID, err)
	}

	for _, dropletTag := range droplet.Tags {
		if strings.Contains(dropletTag, dropletTagInstanceGroupName) {
			instancegrouptag := strings.SplitN(dropletTag, ":", 2)
			if len(instancegrouptag) < 2 {
				return "", fmt.Errorf("failed to parse droplet tag %q: expected colon-separated key/value pair", dropletTag)
			}
			instancegroupvalue := instancegrouptag[1]
			return instancegroupvalue, nil
		}
	}

	return "", errors.New("Could not find tag 'kops-instancegroup' from instance metadata")
}
