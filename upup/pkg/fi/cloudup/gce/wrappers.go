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
	"fmt"

	context "golang.org/x/net/context"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
)

// DeleteInstanceGroupManager deletes the specified InstanceGroupManager in GCE
func DeleteInstanceGroupManager(c GCECloud, t *compute.InstanceGroupManager) error {
	klog.V(2).Infof("Deleting GCE InstanceGroupManager %s", t.SelfLink)
	u, err := ParseGoogleCloudURL(t.SelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().InstanceGroupManagers.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if IsNotFound(err) {
			klog.Infof("InstanceGroupManager not found, assuming deleted: %q", t.SelfLink)
			return nil
		}
		return fmt.Errorf("error deleting InstanceGroupManager %s: %v", t.SelfLink, err)
	}

	return c.WaitForOp(op)
}

// DeleteInstanceTemplate deletes the specified InstanceTemplate (by URL) in GCE
func DeleteInstanceTemplate(c GCECloud, selfLink string) error {
	klog.V(2).Infof("Deleting GCE InstanceTemplate %s", selfLink)
	u, err := ParseGoogleCloudURL(selfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().InstanceTemplates.Delete(u.Project, u.Name).Do()
	if err != nil {
		if IsNotFound(err) {
			klog.Infof("instancetemplate not found, assuming deleted: %q", selfLink)
			return nil
		}
		return fmt.Errorf("error deleting InstanceTemplate %s: %v", selfLink, err)
	}

	return c.WaitForOp(op)
}

// DeleteInstance deletes the specified instance (by URL) in GCE
func DeleteInstance(c GCECloud, instanceSelfLink string) error {
	klog.V(2).Infof("Deleting GCE Instance %s", instanceSelfLink)
	u, err := ParseGoogleCloudURL(instanceSelfLink)
	if err != nil {
		return err
	}

	op, err := c.Compute().Instances.Delete(u.Project, u.Zone, u.Name).Do()
	if err != nil {
		if IsNotFound(err) {
			klog.Infof("Instance not found, assuming deleted: %q", instanceSelfLink)
			return nil
		}
		return fmt.Errorf("error deleting Instance %s: %v", instanceSelfLink, err)
	}

	return c.WaitForOp(op)
}

// ListManagedInstances lists the specified InstanceGroupManagers in GCE
func ListManagedInstances(c GCECloud, igm *compute.InstanceGroupManager) ([]*compute.ManagedInstance, error) {
	ctx := context.Background()
	project := c.Project()

	zoneName := LastComponent(igm.Zone)

	// TODO: Only select a subset of fields
	//	req.Fields(
	//		googleapi.Field("items/selfLink"),
	//		googleapi.Field("items/metadata/items[key='cluster-name']"),
	//		googleapi.Field("items/metadata/items[key='instance-template']"),
	//	)

	var instances []*compute.ManagedInstance
	err := c.Compute().InstanceGroupManagers.ListManagedInstances(project, zoneName, igm.Name).Pages(ctx,
		func(page *compute.InstanceGroupManagersListManagedInstancesResponse) error {
			instances = append(instances, page.ManagedInstances...)
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("error listing ManagedInstances in %s: %v", igm.Name, err)
	}

	return instances, nil
}
