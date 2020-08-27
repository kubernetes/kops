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

package openstack

import (
	"errors"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/apiversions"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/pagination"
	version "github.com/hashicorp/go-version"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	cpoerrors "k8s.io/cloud-provider-openstack/pkg/util/errors"
)

const (
	OctaviaFeatureTags   = 0
	OctaviaFeatureVIPACL = 1

	loadbalancerActiveInitDelay = 1 * time.Second
	loadbalancerActiveFactor    = 1.2
	loadbalancerActiveSteps     = 19

	activeStatus = "ACTIVE"
	errorStatus  = "ERROR"
)

var (
	octaviaVersion string

	// ErrNotFound is used to inform that the object is missing
	ErrNotFound = errors.New("failed to find object")

	// ErrMultipleResults is used when we unexpectedly get back multiple results
	ErrMultipleResults = errors.New("multiple results where only one expected")
)

// getOctaviaVersion returns the current Octavia API version.
func getOctaviaVersion(client *gophercloud.ServiceClient) (string, error) {
	if octaviaVersion != "" {
		return octaviaVersion, nil
	}

	var defaultVer = "0.0"
	allPages, err := apiversions.List(client).AllPages()
	if err != nil {
		return defaultVer, err
	}
	versions, err := apiversions.ExtractAPIVersions(allPages)
	if err != nil {
		return defaultVer, err
	}
	if len(versions) == 0 {
		return defaultVer, fmt.Errorf("API versions for Octavia not found")
	}

	klog.V(4).Infof("Found Octavia API versions: %v", versions)

	// The current version is always the last one in the list
	octaviaVersion = versions[len(versions)-1].ID
	klog.V(4).Infof("The current Octavia API version: %v", octaviaVersion)

	return octaviaVersion, nil
}

// IsOctaviaFeatureSupported returns true if the given feature is supported in the deployed Octavia version.
func IsOctaviaFeatureSupported(client *gophercloud.ServiceClient, feature int) bool {
	octaviaVer, err := getOctaviaVersion(client)
	if err != nil {
		klog.Warningf("Failed to get current Octavia API version: %v", err)
		return false
	}

	currentVer, _ := version.NewVersion(octaviaVer)

	switch feature {
	case OctaviaFeatureVIPACL:
		verACL, _ := version.NewVersion("v2.12")
		if currentVer.GreaterThanOrEqual(verACL) {
			return true
		}
	case OctaviaFeatureTags:
		verTags, _ := version.NewVersion("v2.5")
		if currentVer.GreaterThanOrEqual(verTags) {
			return true
		}
	default:
		klog.Warningf("Feature %d not recognized", feature)
	}

	return false
}

func waitLoadbalancerActive(client *gophercloud.ServiceClient, loadbalancerID string) error {
	backoff := wait.Backoff{
		Duration: loadbalancerActiveInitDelay,
		Factor:   loadbalancerActiveFactor,
		Steps:    loadbalancerActiveSteps,
	}

	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		loadbalancer, err := loadbalancers.Get(client, loadbalancerID).Extract()
		if err != nil {
			return false, err
		}
		if loadbalancer.ProvisioningStatus == activeStatus {
			return true, nil
		} else if loadbalancer.ProvisioningStatus == errorStatus {
			return true, fmt.Errorf("loadbalancer has gone into ERROR state")
		} else {
			return false, nil
		}

	})

	return err
}

// UpdateListener updates a listener and wait for the lb active
func UpdateListener(client *gophercloud.ServiceClient, lbID string, listenerID string, opts listeners.UpdateOpts) error {
	if _, err := listeners.Update(client, listenerID, opts).Extract(); err != nil {
		return err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after updating listener: %v", err)
	}

	return nil
}

// CreateListener creates a new listener
func CreateListener(client *gophercloud.ServiceClient, lbID string, opts listeners.CreateOpts) (*listeners.Listener, error) {
	listener, err := listeners.Create(client, opts).Extract()
	if err != nil {
		return nil, err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return nil, fmt.Errorf("failed to wait for load balancer ACTIVE after creating listener: %v", err)
	}

	return listener, nil
}

// GetLoadbalancerByName retrieves loadbalancer object
func GetLoadbalancerByName(client *gophercloud.ServiceClient, name string) (*loadbalancers.LoadBalancer, error) {
	opts := loadbalancers.ListOpts{
		Name: name,
	}
	allPages, err := loadbalancers.List(client, opts).AllPages()
	if err != nil {
		return nil, err
	}
	loadbalancerList, err := loadbalancers.ExtractLoadBalancers(allPages)
	if err != nil {
		return nil, err
	}

	if len(loadbalancerList) > 1 {
		return nil, ErrMultipleResults
	}
	if len(loadbalancerList) == 0 {
		return nil, ErrNotFound
	}

	return &loadbalancerList[0], nil
}

// GetListenerByName gets a listener by its name, raise error if not found or get multiple ones.
func GetListenerByName(client *gophercloud.ServiceClient, name string, lbID string) (*listeners.Listener, error) {
	opts := listeners.ListOpts{
		Name:           name,
		LoadbalancerID: lbID,
	}
	pager := listeners.List(client, opts)
	var listenerList []listeners.Listener

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		v, err := listeners.ExtractListeners(page)
		if err != nil {
			return false, err
		}
		listenerList = append(listenerList, v...)
		if len(listenerList) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		if cpoerrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(listenerList) == 0 {
		return nil, ErrNotFound
	}

	return &listenerList[0], nil
}

// GetPoolByName gets a pool by its name, raise error if not found or get multiple ones.
func GetPoolByName(client *gophercloud.ServiceClient, name string, lbID string) (*pools.Pool, error) {
	var listenerPools []pools.Pool

	opts := pools.ListOpts{
		Name:           name,
		LoadbalancerID: lbID,
	}
	err := pools.List(client, opts).EachPage(func(page pagination.Page) (bool, error) {
		v, err := pools.ExtractPools(page)
		if err != nil {
			return false, err
		}
		listenerPools = append(listenerPools, v...)
		if len(listenerPools) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if err != nil {
		if cpoerrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(listenerPools) == 0 {
		return nil, ErrNotFound
	} else if len(listenerPools) > 1 {
		return nil, ErrMultipleResults
	}

	return &listenerPools[0], nil
}

// DeleteLoadbalancer deletes a loadbalancer with all its child objects.
func DeleteLoadbalancer(client *gophercloud.ServiceClient, lbID string) error {
	err := loadbalancers.Delete(client, lbID, loadbalancers.DeleteOpts{Cascade: true}).ExtractErr()
	if err != nil && !cpoerrors.IsNotFound(err) {
		return fmt.Errorf("error deleting loadbalancer %s: %v", lbID, err)
	}

	return nil
}
