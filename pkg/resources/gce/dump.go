/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	"context"
	"fmt"
	"strings"
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources"
	gce "k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// dumpState holds state for use during a GCE dump operation
type dumpState struct {
	// cloud is the a reference to the GCE cloud we are dumping
	cloud gce.GCECloud

	// mutex protects the follow resources
	mutex sync.Mutex

	// instances is a cache of instances by zone
	instances map[string]map[string]*compute.Instance

	// disks is a cache of disks by zone
	disks map[string]map[string]*compute.Disk
}

// DumpManagedInstance is responsible for dumping a resource for a ManagedInstance
func DumpManagedInstance(op *resources.DumpOperation, r *resources.Resource) error {
	instance := r.Obj.(*compute.ManagedInstance)

	u, err := gce.ParseGoogleCloudURL(instance.Instance)
	if err != nil {
		return fmt.Errorf("unable to parse instance url %q", instance.Instance)
	}

	// Fetch instance details
	instanceMap, err := getDumpState(op).getInstances(op.Context, u.Zone)
	if err != nil {
		return err
	}

	i := &resources.Instance{
		Name: u.Name,
	}

	instanceDetails := instanceMap[u.Name]
	if instanceDetails == nil {
		var sb strings.Builder
		fmt.Fprintf(&sb, "instance %q not found (currentAction=%q instanceStatus=%q)", instance.Instance, instance.CurrentAction, instance.InstanceStatus)
		if instance.LastAttempt != nil && instance.LastAttempt.Errors != nil {
			for _, e := range instance.LastAttempt.Errors.Errors {
				fmt.Fprintf(&sb, "; lastAttempt.error code=%q location=%q message=%q", e.Code, e.Location, e.Message)
			}
		}
		klog.Warning(sb.String())
		return nil
	}

	for _, ni := range instanceDetails.NetworkInterfaces {
		if ni == nil {
			continue
		}
		if ni.NetworkIP != "" {
			i.PrivateAddresses = append(i.PrivateAddresses, ni.NetworkIP)
		}
		if ni.Ipv6Address != "" {
			i.PrivateAddresses = append(i.PrivateAddresses, ni.Ipv6Address)
		}
		for _, ac := range ni.AccessConfigs {
			if ac == nil {
				continue
			}
			if ac.NatIP != "" {
				i.PublicAddresses = append(i.PublicAddresses, ac.NatIP)
			}
		}
	}

	isControlPlane := false
	for key, value := range instanceDetails.Labels {
		if !strings.HasPrefix(key, gce.GceLabelNameRolePrefix) {
			continue
		}
		if value == "control-plane" {
			isControlPlane = true
		} else {
			i.Roles = append(i.Roles, value)
		}
	}
	if isControlPlane {
		i.Roles = append(i.Roles, "control-plane")
	}

	if image, err := getDumpState(op).getBootDiskImage(op.Context, u.Zone, instanceDetails); err != nil {
		klog.Warningf("unable to determine boot disk image for instance %q: %v", u.Name, err)
	} else if image != "" {
		i.SSHUser = gce.SSHUsernameForImage(image)
	}

	op.Dump.Instances = append(op.Dump.Instances, i)

	op.Dump.Resources = append(op.Dump.Resources, instanceDetails)

	return nil
}

// getDumpState gets the dumpState from the dump context, or creates one if not yet initialized
func getDumpState(dumpContext *resources.DumpOperation) *dumpState {
	if dumpContext.CloudState == nil {
		dumpContext.CloudState = &dumpState{
			cloud: dumpContext.Cloud.(gce.GCECloud),
		}
	}
	return dumpContext.CloudState.(*dumpState)
}

// getInstances retrieves the list of instances from the cloud, using a cached copy if possible
func (s *dumpState) getInstances(ctx context.Context, zone string) (map[string]*compute.Instance, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.instances == nil {
		s.instances = make(map[string]map[string]*compute.Instance)
	}

	if s.instances[zone] != nil {
		return s.instances[zone], nil
	}

	l, err := s.cloud.Compute().Instances().List(ctx, s.cloud.Project(), zone)
	if err != nil {
		return nil, err
	}
	instances := make(map[string]*compute.Instance)
	for _, i := range l {
		instances[i.Name] = i
	}
	s.instances[zone] = instances
	return instances, nil
}

// getDisks retrieves the list of disks from the cloud, using a cached copy if possible
func (s *dumpState) getDisks(ctx context.Context, zone string) (map[string]*compute.Disk, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.disks == nil {
		s.disks = make(map[string]map[string]*compute.Disk)
	}

	if s.disks[zone] != nil {
		return s.disks[zone], nil
	}

	l, err := s.cloud.Compute().Disks().List(ctx, s.cloud.Project(), zone)
	if err != nil {
		return nil, err
	}
	disks := make(map[string]*compute.Disk)
	for _, d := range l {
		disks[d.Name] = d
	}
	s.disks[zone] = disks
	return disks, nil
}

// getBootDiskImage returns the source image of the instance's boot disk. The instance's attached
// disks don't carry the source image, so we look it up on the disk resource itself. Returns "" if
// the image cannot be determined (e.g. a disk created from a snapshot).
func (s *dumpState) getBootDiskImage(ctx context.Context, zone string, instance *compute.Instance) (string, error) {
	for _, d := range instance.Disks {
		if d == nil || !d.Boot {
			continue
		}
		disks, err := s.getDisks(ctx, zone)
		if err != nil {
			return "", err
		}
		disk := disks[gce.LastComponent(d.Source)]
		if disk == nil {
			return "", nil
		}
		return disk.SourceImage, nil
	}
	return "", nil
}

// DumpNetwork is responsible for dumping a resource for a Network
func DumpNetwork(op *resources.DumpOperation, r *resources.Resource) error {
	network := r.Obj.(*compute.Network)

	vpc := &resources.VPC{
		ID: gce.LastComponent(network.SelfLink),
	}
	op.Dump.VPC = vpc

	return nil
}

// DumpSubnetwork is responsible for dumping a resource for a Subnetwork
func DumpSubnetwork(op *resources.DumpOperation, r *resources.Resource) error {
	obj := r.Obj.(*compute.Subnetwork)

	subnet := &resources.Subnet{
		ID: gce.LastComponent(obj.SelfLink),
	}
	op.Dump.Subnets = append(op.Dump.Subnets, subnet)

	return nil
}
