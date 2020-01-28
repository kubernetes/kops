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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources/digitalocean/dns"
	"k8s.io/kops/upup/pkg/fi"
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

// Cloud exposes all the interfaces required to operate on DigitalOcean resources
type Cloud struct {
	Client *godo.Client

	dns dnsprovider.Interface

	// RegionName holds the region, renamed to avoid conflict with Region()
	RegionName string
}

var _ fi.Cloud = &Cloud{}

// NewCloud returns a Cloud, expecting the env var DIGITALOCEAN_ACCESS_TOKEN
// NewCloud will return an err if DIGITALOCEAN_ACCESS_TOKEN is not defined
func NewCloud(region string) (*Cloud, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := godo.NewClient(oauthClient)

	return &Cloud{
		Client:     client,
		dns:        dns.NewProvider(client),
		RegionName: region,
	}, nil
}

// GetCloudGroups is not implemented yet, that needs to return the instances and groups that back a kops cluster.
func (c *Cloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Info("digitalocean cloud provider GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("digital ocean cloud provider does not support getting cloud groups at this time")
}

// DeleteGroup is not implemented yet, is a func that needs to delete a DO instance group.
func (c *Cloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	klog.V(8).Info("digitalocean cloud provider DeleteGroup not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud groups at this time")
}

// DeleteInstance is not implemented yet, is func needs to delete a DO instance.
func (c *Cloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Info("digitalocean cloud provider DeleteInstance not implemented yet")
	return fmt.Errorf("digital ocean cloud provider does not support deleting cloud instances at this time")
}

// ProviderID returns the kops api identifier for DigitalOcean cloud provider
func (c *Cloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderDO
}

// Region returns the DO region we will target
func (c *Cloud) Region() string {
	return c.RegionName
}

// DNS returns a DO implementation for dnsprovider.Interface
func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	return c.dns, nil
}

// Volumes returns an implementation of godo.StorageService
func (c *Cloud) Volumes() godo.StorageService {
	return c.Client.Storage
}

// VolumeActions returns an implementation of godo.StorageActionsService
func (c *Cloud) VolumeActions() godo.StorageActionsService {
	return c.Client.StorageActions
}

func (c *Cloud) Droplets() godo.DropletsService {
	return c.Client.Droplets
}

func (c *Cloud) LoadBalancers() godo.LoadBalancersService {
	return c.Client.LoadBalancers
}

// FindVPCInfo is not implemented, it's only here to satisfy the fi.Cloud interface
func (c *Cloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("not implemented")
}

func (c *Cloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	var ingresses []kops.ApiIngressStatus
	if cluster.Spec.MasterPublicName != "" {
		// Note that this must match Digital Ocean's lb name
		klog.V(2).Infof("Querying DO to find Loadbalancers for API (%q)", cluster.Name)

		loadBalancers, err := getAllLoadBalancers(c)
		if err != nil {
			return nil, fmt.Errorf("LoadBalancers.List returned error: %v", err)
		}

		lbName := "api-" + strings.Replace(cluster.Name, ".", "-", -1)

		for _, lb := range loadBalancers {
			if lb.Name == lbName {
				klog.V(10).Infof("Matching LB name found for API (%q)", cluster.Name)

				if lb.Status != "active" {
					return nil, fmt.Errorf("load-balancer is not yet active (current status: %s)", lb.Status)
				}

				address := lb.IP
				ingresses = append(ingresses, kops.ApiIngressStatus{IP: address})

				return ingresses, nil
			}
		}
	}

	return nil, nil
}
