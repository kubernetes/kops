/*
Copyright 2016 The Kubernetes Authors.

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

package resources

import (
	"context"
	"github.com/golang/glog"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

const (
	typeVM = "VM"
)

type clusterDiscoveryVSphere struct {
	cloud        fi.Cloud
	vsphereCloud *vsphere.VSphereCloud
	clusterName  string
}

type vsphereListFn func() ([]*ResourceTracker, error)

func (c *ClusterResources) listResourcesVSphere() (map[string]*ResourceTracker, error) {
	vsphereCloud := c.Cloud.(*vsphere.VSphereCloud)

	resources := make(map[string]*ResourceTracker)

	d := &clusterDiscoveryVSphere{
		cloud:        c.Cloud,
		vsphereCloud: vsphereCloud,
		clusterName:  c.ClusterName,
	}

	listFunctions := []vsphereListFn{
		d.listVMs,
	}

	for _, fn := range listFunctions {
		trackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range trackers {
			resources[GetResourceTrackerKey(t)] = t
		}
	}

	return resources, nil
}

func (d *clusterDiscoveryVSphere) listVMs() ([]*ResourceTracker, error) {
	c := d.vsphereCloud

	regexForMasterVMs := "*" + "." + "masters" + "." + d.clusterName + "*"
	regexForNodeVMs := "nodes" + "." + d.clusterName + "*"

	vms, err := c.GetVirtualMachines([]string{regexForMasterVMs, regexForNodeVMs})
	if err != nil {
		if _, ok := err.(*find.NotFoundError); !ok {
			return nil, err
		}
		glog.Warning(err)
	}

	var trackers []*ResourceTracker
	for _, vm := range vms {
		tracker := &ResourceTracker{
			Name:    vm.Name(),
			ID:      vm.Name(),
			Type:    typeVM,
			deleter: deleteVM,
			Dumper:  DumpVMInfo,
			obj:     vm,
		}
		trackers = append(trackers, tracker)
	}
	return trackers, nil
}

func deleteVM(cloud fi.Cloud, r *ResourceTracker) error {
	vsphereCloud := cloud.(*vsphere.VSphereCloud)

	vm := r.obj.(*object.VirtualMachine)

	task, err := vm.PowerOff(context.TODO())
	if err != nil {
		return err
	}
	task.Wait(context.TODO())

	vsphereCloud.DeleteCloudInitISO(fi.String(vm.Name()))

	task, err = vm.Destroy(context.TODO())
	if err != nil {
		return err
	}

	err = task.Wait(context.TODO())
	if err != nil {
		glog.Fatalf("Destroy VM failed: %q", err)
	}

	return nil
}

func DumpVMInfo(r *ResourceTracker) (interface{}, error) {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.obj
	return data, nil
}

func GetResourceTrackerKey(t *ResourceTracker) string {
	return t.Type + ":" + t.ID
}
