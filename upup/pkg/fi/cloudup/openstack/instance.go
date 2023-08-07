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
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	v2pools "github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	INSTANCE_GROUP_GENERATION = "ig_generation"
	CLUSTER_GENERATION        = "cluster_generation"
	OS_ANNOTATION             = "openstack.kops.io/"
	BOOT_FROM_VOLUME          = "osVolumeBoot"
	BOOT_VOLUME_SIZE          = "osVolumeSize"
	SERVER_GROUP_AFFINITY     = "serverGroupAffinity"
	ALLOWED_ADDRESS_PAIR      = "allowedAddressPair"
	SERVER_GROUP_NAME         = "serverGroupName"

	defaultActiveTimeout = time.Second * 120
	activeStatus         = "ACTIVE"
	errorStatus          = "ERROR"
)

// floatingBackoff is the backoff strategy for listing openstack floatingips
var floatingBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    20,
}

func (c *openstackCloud) CreateInstance(opt servers.CreateOptsBuilder, portID string) (*servers.Server, error) {
	return createInstance(c, opt, portID)
}

func IsPortInUse(err error) bool {
	if _, ok := err.(gophercloud.ErrDefault409); ok {
		return true
	}
	return false
}

// waitForStatusActive uses gopherclouds WaitFor() func to determine when the server becomes "ACTIVE".
//
// The function will immediately fail if the server transistions into the status "ERROR"
// and will result in a timeout when not reaching status "ACTIVE" in time.
func waitForStatusActive(c OpenstackCloud, serverID string, timeout *time.Duration) error {
	if timeout == nil {
		timeout = fi.PtrTo(defaultActiveTimeout)
	}

	return gophercloud.WaitFor(int(timeout.Seconds()), func() (bool, error) {
		server, err := c.GetInstance(serverID)
		if err != nil {
			return false, err
		}

		if server.Status == errorStatus {
			return false, fmt.Errorf("unable to create server: %v", server.Fault)
		}

		if server.Status == activeStatus {
			return true, nil
		}

		return false, nil
	})
}

func createInstance(c OpenstackCloud, opt servers.CreateOptsBuilder, portID string) (*servers.Server, error) {
	var server *servers.Server

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {

		v, err := servers.Create(c.ComputeClient(), opt).Extract()
		if err != nil {
			if IsPortInUse(err) && portID != "" {
				port, err := c.GetPort(portID)
				if err != nil {
					return false, fmt.Errorf("error finding port %s: %v", portID, err)
				}
				// port is attached to deleted instance, we need reset the status of the DeviceID
				// this is bug in OpenStack APIs
				if port.DeviceID != "" && port.DeviceOwner == "" {
					klog.Warningf("Port %s is attached to Device that does not exist anymore, reseting the status of DeviceID", portID)
					_, err := c.UpdatePort(portID, ports.UpdateOpts{
						DeviceID: fi.PtrTo(""),
					})
					if err != nil {
						return false, fmt.Errorf("error updating port %s deviceid: %v", portID, err)
					}
				}
			}
			return false, fmt.Errorf("error creating server %v: %v", opt, err)
		}
		server = v

		err = waitForStatusActive(c, server.ID, nil)
		if err != nil {
			return true, err
		}

		return true, nil
	})
	if err != nil {
		return server, err
	} else if done {
		return server, nil
	} else {
		return server, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListServerFloatingIPs(instanceID string) ([]*string, error) {
	return listServerFloatingIPs(c, instanceID, c.floatingEnabled)
}

func listServerFloatingIPs(c OpenstackCloud, instanceID string, floatingEnabled bool) ([]*string, error) {
	var result []*string
	_, err := vfs.RetryWithBackoff(floatingBackoff, func() (bool, error) {
		server, err := c.GetInstance(instanceID)
		if err != nil {
			return true, fmt.Errorf("failed to find server with id (\"%s\"): %v", instanceID, err)
		}

		var addresses map[string][]Address
		err = mapstructure.Decode(server.Addresses, &addresses)
		if err != nil {
			return true, err
		}

		for _, addrList := range addresses {
			for _, props := range addrList {
				if floatingEnabled {
					if props.IPType == "floating" {
						result = append(result, fi.PtrTo(props.Addr))
					}
				} else {
					result = append(result, fi.PtrTo(props.Addr))
				}
			}
		}
		if len(result) > 0 {
			return true, nil
		}
		return false, nil
	})
	if len(result) == 0 || err != nil {
		return result, fmt.Errorf("could not find floating ip associated to server (\"%s\") %v", instanceID, err)
	}
	return result, nil
}

func (c *openstackCloud) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return deleteInstance(c, i)
}

func deleteInstance(c OpenstackCloud, i *cloudinstances.CloudInstance) error {
	return deleteInstanceWithID(c, i.ID)
}

func (c *openstackCloud) DeleteInstanceWithID(instanceID string) error {
	return deleteInstanceWithID(c, instanceID)
}

func deleteInstanceWithID(c OpenstackCloud, instanceID string) error {
	done, err := vfs.RetryWithBackoff(deleteBackoff, func() (bool, error) {
		err := servers.Delete(c.ComputeClient(), instanceID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting instance: %s", err)
		}
		if isNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

// DeregisterInstance drains a cloud instance and loadbalancers.
func (c *openstackCloud) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	return deregisterInstance(c, i.ID)
}

// deregisterInstance will drain all the loadbalancers attached to instance
func deregisterInstance(c OpenstackCloud, instanceID string) error {
	instance, err := c.GetInstance(instanceID)
	if err != nil {
		return err
	}

	// Kubernetes creates loadbalancers that member name matches to instance name
	// However, kOps uses different name format in API LB which is <cluster>-<ig>
	instanceName := instance.Name
	kopsName := ""
	ig, igok := instance.Metadata[TagKopsInstanceGroup]
	clusterName, clusterok := instance.Metadata[TagClusterName]
	if igok && clusterok {
		kopsName = fmt.Sprintf("%s-%s", clusterName, ig)
	}

	lbs, err := c.ListLBs(loadbalancers.ListOpts{})
	if err != nil {
		return err
	}
	ctx := context.Background()
	eg, _ := errgroup.WithContext(ctx)
	for i := range lbs {
		func(lb loadbalancers.LoadBalancer) {
			eg.Go(func() error {
				return drainSingleLB(c, lb, instanceName, kopsName)
			})
		}(lbs[i])
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to deregister instance from load balancers: %v", err)
	}

	return nil
}

// drainSingleLB will drain single loadbalancer that is attached to instance
func drainSingleLB(c OpenstackCloud, lb loadbalancers.LoadBalancer, instanceName string, kopsName string) error {
	oldStats, err := c.GetLBStats(lb.ID)
	if err != nil {
		return err
	}

	draining := false
	pools, err := c.ListPools(v2pools.ListOpts{
		LoadbalancerID: lb.ID,
	})
	if err != nil {
		return err
	}
	for _, pool := range pools {
		members, err := c.ListPoolMembers(pool.ID, v2pools.ListMembersOpts{})
		if err != nil {
			return err
		}
		for _, member := range members {
			if member.Name == instanceName || (member.Name == kopsName && len(kopsName) > 0) {
				// https://docs.openstack.org/api-ref/load-balancer/v2/?expanded=update-a-member-detail
				// Setting the member weight to 0 means that the member will not receive new requests but will finish any existing connections.
				// This “drains” the backend member of active connections.
				_, err := c.UpdateMemberInPool(pool.ID, member.ID, v2pools.UpdateMemberOpts{
					Weight: fi.PtrTo(0),
				})
				if err != nil {
					return err
				}
				draining = true
				break
			}
		}
	}

	if draining {
		time.Sleep(20 * time.Second)

		newStats, err := c.GetLBStats(lb.ID)
		if err != nil {
			return err
		}

		// NOTE! this is total loadbalancer connections NOT member connections
		klog.V(4).Infof("Loadbalancer %s connections before draining %d and after %d", lb.Name, oldStats.ActiveConnections, newStats.ActiveConnections)
	}
	return nil
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *openstackCloud) DetachInstance(i *cloudinstances.CloudInstance) error {
	return detachInstance(c, i)
}

func detachInstance(c OpenstackCloud, i *cloudinstances.CloudInstance) error {
	klog.V(8).Info("openstack cloud provider DetachInstance not implemented yet")
	return fmt.Errorf("openstack cloud provider does not support surging")
}

func (c *openstackCloud) GetInstance(id string) (*servers.Server, error) {
	return getInstance(c, id)
}

func getInstance(c OpenstackCloud, id string) (*servers.Server, error) {
	var server *servers.Server

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		instance, err := servers.Get(c.ComputeClient(), id).Extract()
		if err != nil {
			return false, err
		}
		server = instance
		return true, nil
	})
	if err != nil {
		return server, err
	} else if done {
		return server, nil
	} else {
		return server, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListInstances(opt servers.ListOptsBuilder) ([]servers.Server, error) {
	return listInstances(c, opt)
}

func listInstances(c OpenstackCloud, opt servers.ListOptsBuilder) ([]servers.Server, error) {
	var instances []servers.Server

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := servers.List(c.ComputeClient(), opt).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing servers %v: %v", opt, err)
		}

		ss, err := servers.ExtractServers(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting servers from pages: %v", err)
		}
		instances = ss
		return true, nil
	})
	if err != nil {
		return instances, err
	} else if done {
		return instances, nil
	} else {
		return instances, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) GetFlavor(name string) (*flavors.Flavor, error) {
	return getFlavor(c, name)
}

func getFlavor(c OpenstackCloud, name string) (*flavors.Flavor, error) {
	opts := flavors.ListOpts{}
	pager := flavors.ListDetail(c.ComputeClient(), opts)
	page, err := pager.AllPages()
	if err != nil {
		return nil, fmt.Errorf("failed to list flavors: %v", err)
	}

	fs, err := flavors.ExtractFlavors(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract flavors: %v", err)
	}
	for _, f := range fs {
		if f.Name == name {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("could not find flavor with name %v", name)
}
