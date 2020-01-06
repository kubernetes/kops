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
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/mitchellh/mapstructure"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
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
)

// floatingBackoff is the backoff strategy for listing openstack floatingips
var floatingBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    20,
}

func (c *openstackCloud) CreateInstance(opt servers.CreateOptsBuilder) (*servers.Server, error) {
	var server *servers.Server

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := servers.Create(c.novaClient, opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating server %v: %v", opt, err)
		}
		server = v
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
				if c.floatingEnabled {
					if props.IPType == "floating" {
						result = append(result, fi.String(props.Addr))
					}
				} else {
					result = append(result, fi.String(props.Addr))
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

func (c *openstackCloud) DeleteInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.Warning("This does not work without running kops update cluster --yes in another terminal")
	return c.DeleteInstanceWithID(i.ID)
}

func (c *openstackCloud) DeleteInstanceWithID(instanceID string) error {
	return servers.Delete(c.novaClient, instanceID).ExtractErr()
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *openstackCloud) DetachInstance(i *cloudinstances.CloudInstanceGroupMember) error {
	klog.V(8).Info("openstack cloud provider DetachInstance not implemented yet")
	return fmt.Errorf("openstack cloud provider does not support surging")
}

func (c *openstackCloud) GetInstance(id string) (*servers.Server, error) {
	var server *servers.Server

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		instance, err := servers.Get(c.novaClient, id).Extract()
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
	var instances []servers.Server

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := servers.List(c.novaClient, opt).AllPages()
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
