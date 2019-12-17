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
	"sync"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
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
		klog.Warningf("instance %q not found", instance.Instance)
	} else {
		for _, ni := range instanceDetails.NetworkInterfaces {
			for _, ac := range ni.AccessConfigs {
				if ac.NatIP != "" {
					i.PublicAddresses = append(i.PublicAddresses, ac.NatIP)
				}
			}
		}
	}

	op.Dump.Instances = append(op.Dump.Instances, i)

	// Unclear if we should include the instance details in the dump - assume YAGNI until someone needs it
	//dump.Resources = append(dump.Resources, instanceDetails)

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

	instances := make(map[string]*compute.Instance)
	err := s.cloud.Compute().Instances.List(s.cloud.Project(), zone).Pages(ctx, func(page *compute.InstanceList) error {
		for _, i := range page.Items {
			instances[i.Name] = i
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.instances[zone] = instances
	return instances, nil
}
