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
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/l7policies"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/pagination"
	version "github.com/hashicorp/go-version"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cloud-provider-openstack/pkg/cloudprovider/providers/openstack/metrics"
	klog "k8s.io/klog/v2"

	cpoerrors "k8s.io/cloud-provider-openstack/pkg/util/errors"
)

const (
	OctaviaFeatureTags              = 0
	OctaviaFeatureVIPACL            = 1
	OctaviaFeatureFlavors           = 2
	OctaviaFeatureTimeout           = 3
	OctaviaFeatureAvailabilityZones = 4

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
	mc := metrics.NewMetricContext("version", "list")
	allPages, err := apiversions.List(client).AllPages()
	if mc.ObserveRequest(err) != nil {
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
	case OctaviaFeatureFlavors:
		verFlavors, _ := version.NewVersion("v2.6")
		if currentVer.GreaterThanOrEqual(verFlavors) {
			return true
		}
	case OctaviaFeatureTimeout:
		verFlavors, _ := version.NewVersion("v2.1")
		if currentVer.GreaterThanOrEqual(verFlavors) {
			return true
		}
	case OctaviaFeatureAvailabilityZones:
		verAvailabilityZones, _ := version.NewVersion("v2.14")
		if currentVer.GreaterThanOrEqual(verAvailabilityZones) {
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
		mc := metrics.NewMetricContext("loadbalancer", "get")
		loadbalancer, err := loadbalancers.Get(client, loadbalancerID).Extract()
		if mc.ObserveRequest(err) != nil {
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

// GetLoadbalancerByName retrieves loadbalancer object
func GetLoadbalancerByName(client *gophercloud.ServiceClient, name string) (*loadbalancers.LoadBalancer, error) {
	opts := loadbalancers.ListOpts{
		Name: name,
	}
	mc := metrics.NewMetricContext("loadbalancer", "list")
	allPages, err := loadbalancers.List(client, opts).AllPages()
	if mc.ObserveRequest(err) != nil {
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

// DeleteLoadbalancer deletes a loadbalancer with all its child objects.
func DeleteLoadbalancer(client *gophercloud.ServiceClient, lbID string) error {
	mc := metrics.NewMetricContext("loadbalancer", "delete")
	err := loadbalancers.Delete(client, lbID, loadbalancers.DeleteOpts{Cascade: true}).ExtractErr()
	if err != nil && !cpoerrors.IsNotFound(err) {
		_ = mc.ObserveRequest(err)
		return fmt.Errorf("error deleting loadbalancer %s: %v", lbID, err)
	}

	return mc.ObserveRequest(nil)
}

// UpdateListener updates a listener and wait for the lb active
func UpdateListener(client *gophercloud.ServiceClient, lbID string, listenerID string, opts listeners.UpdateOpts) error {
	mc := metrics.NewMetricContext("loadbalancer_listener", "update")
	_, err := listeners.Update(client, listenerID, opts).Extract()
	if mc.ObserveRequest(err) != nil {
		return err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after updating listener: %v", err)
	}

	return nil
}

// CreateListener creates a new listener
func CreateListener(client *gophercloud.ServiceClient, lbID string, opts listeners.CreateOpts) (*listeners.Listener, error) {
	mc := metrics.NewMetricContext("loadbalancer_listener", "create")
	listener, err := listeners.Create(client, opts).Extract()
	if mc.ObserveRequest(err) != nil {
		return nil, err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return nil, fmt.Errorf("failed to wait for load balancer ACTIVE after creating listener: %v", err)
	}

	return listener, nil
}

// GetListenerByName gets a listener by its name, raise error if not found or get multiple ones.
func GetListenerByName(client *gophercloud.ServiceClient, name string, lbID string) (*listeners.Listener, error) {
	opts := listeners.ListOpts{
		Name:           name,
		LoadbalancerID: lbID,
	}
	mc := metrics.NewMetricContext("loadbalancer_listener", "list")
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
	if mc.ObserveRequest(err) != nil {
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

// CreatePool creates a new pool.
func CreatePool(client *gophercloud.ServiceClient, opts pools.CreateOptsBuilder, lbID string) (*pools.Pool, error) {
	mc := metrics.NewMetricContext("loadbalancer_pool", "create")
	pool, err := pools.Create(client, opts).Extract()
	if mc.ObserveRequest(err) != nil {
		return nil, err
	}

	if err = waitLoadbalancerActive(client, lbID); err != nil {
		return nil, fmt.Errorf("failed to wait for load balancer ACTIVE after creating pool: %v", err)
	}

	return pool, nil
}

// GetPoolByName gets a pool by its name, raise error if not found or get multiple ones.
func GetPoolByName(client *gophercloud.ServiceClient, name string, lbID string) (*pools.Pool, error) {
	var listenerPools []pools.Pool

	opts := pools.ListOpts{
		Name:           name,
		LoadbalancerID: lbID,
	}
	mc := metrics.NewMetricContext("loadbalancer_pool", "list")
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
	if mc.ObserveRequest(err) != nil {
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

// GetPoolsByListener finds pool for a listener. A listener always has exactly one pool.
func GetPoolByListener(client *gophercloud.ServiceClient, lbID, listenerID string) (*pools.Pool, error) {
	listenerPools := make([]pools.Pool, 0, 1)
	mc := metrics.NewMetricContext("loadbalancer_pool", "list")
	err := pools.List(client, pools.ListOpts{LoadbalancerID: lbID}).EachPage(func(page pagination.Page) (bool, error) {
		poolsList, err := pools.ExtractPools(page)
		if err != nil {
			return false, err
		}
		for _, p := range poolsList {
			for _, l := range p.Listeners {
				if l.ID == listenerID {
					listenerPools = append(listenerPools, p)
				}
			}
		}
		if len(listenerPools) > 1 {
			return false, ErrMultipleResults
		}
		return true, nil
	})
	if mc.ObserveRequest(err) != nil {
		if cpoerrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(listenerPools) == 0 {
		return nil, ErrNotFound
	}

	return &listenerPools[0], nil
}

// GetPools retrives the pools belong to the loadbalancer.
func GetPools(client *gophercloud.ServiceClient, lbID string) ([]pools.Pool, error) {
	var lbPools []pools.Pool

	opts := pools.ListOpts{
		LoadbalancerID: lbID,
	}
	allPages, err := pools.List(client, opts).AllPages()
	if err != nil {
		return nil, err
	}

	lbPools, err = pools.ExtractPools(allPages)
	if err != nil {
		return nil, err
	}

	return lbPools, nil
}

// GetMembersbyPool get all the members in the pool.
func GetMembersbyPool(client *gophercloud.ServiceClient, poolID string) ([]pools.Member, error) {
	var members []pools.Member

	mc := metrics.NewMetricContext("loadbalancer_member", "list")
	err := pools.ListMembers(client, poolID, pools.ListMembersOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		membersList, err := pools.ExtractMembers(page)
		if err != nil {
			return false, err
		}
		members = append(members, membersList...)

		return true, nil
	})
	if mc.ObserveRequest(err) != nil {
		return nil, err
	}

	return members, nil
}

// DeletePool deletes a pool.
func DeletePool(client *gophercloud.ServiceClient, poolID string, lbID string) error {
	mc := metrics.NewMetricContext("loadbalancer_pool", "delete")
	if err := pools.Delete(client, poolID).ExtractErr(); mc.ObserveRequest(err) != nil {
		return err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after deleting pool: %v", err)
	}

	return nil
}

// BatchUpdatePoolMembers updates pool members in batch.
func BatchUpdatePoolMembers(client *gophercloud.ServiceClient, lbID string, poolID string, opts []pools.BatchUpdateMemberOpts) error {
	mc := metrics.NewMetricContext("loadbalancer_members", "update")
	err := pools.BatchUpdateMembers(client, poolID, opts).ExtractErr()
	if mc.ObserveRequest(err) != nil {
		return err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after updating pool members for %s: %v", poolID, err)
	}

	return nil
}

// GetL7policies retrieves all l7 policies for the given listener.
func GetL7policies(client *gophercloud.ServiceClient, listenerID string) ([]l7policies.L7Policy, error) {
	var policies []l7policies.L7Policy
	opts := l7policies.ListOpts{
		ListenerID: listenerID,
	}
	err := l7policies.List(client, opts).EachPage(func(page pagination.Page) (bool, error) {
		v, err := l7policies.ExtractL7Policies(page)
		if err != nil {
			return false, err
		}
		policies = append(policies, v...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// CreateL7Policy creates a l7 policy.
func CreateL7Policy(client *gophercloud.ServiceClient, opts l7policies.CreateOpts, lbID string) (*l7policies.L7Policy, error) {
	mc := metrics.NewMetricContext("loadbalancer_l7policy", "create")
	policy, err := l7policies.Create(client, opts).Extract()
	if mc.ObserveRequest(err) != nil {
		return nil, err
	}

	if err = waitLoadbalancerActive(client, lbID); err != nil {
		return nil, fmt.Errorf("failed to wait for load balancer ACTIVE after creating l7policy: %v", err)
	}

	return policy, nil
}

// DeleteL7policy deletes a l7 policy.
func DeleteL7policy(client *gophercloud.ServiceClient, policyID string, lbID string) error {
	mc := metrics.NewMetricContext("loadbalancer_l7policy", "delete")
	if err := l7policies.Delete(client, policyID).ExtractErr(); mc.ObserveRequest(err) != nil {
		return err
	}

	if err := waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after deleting l7policy: %v", err)
	}

	return nil
}

// GetL7Rules gets all the rules for a l7 policy
func GetL7Rules(client *gophercloud.ServiceClient, policyID string) ([]l7policies.Rule, error) {
	listOpts := l7policies.ListRulesOpts{}
	allPages, err := l7policies.ListRules(client, policyID, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	allRules, err := l7policies.ExtractRules(allPages)
	if err != nil {
		return nil, err
	}

	return allRules, nil
}

// CreateL7Rule creates a l7 rule.
func CreateL7Rule(client *gophercloud.ServiceClient, policyID string, opts l7policies.CreateRuleOpts, lbID string) error {
	mc := metrics.NewMetricContext("loadbalancer_l7rule", "create")
	_, err := l7policies.CreateRule(client, policyID, opts).Extract()
	if mc.ObserveRequest(err) != nil {
		return err
	}

	if err = waitLoadbalancerActive(client, lbID); err != nil {
		return fmt.Errorf("failed to wait for load balancer ACTIVE after creating l7policy rule: %v", err)
	}

	return nil
}
