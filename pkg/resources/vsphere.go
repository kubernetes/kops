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

package resources

import (
	"context"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"k8s.io/klog"
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

type vsphereListFn func() ([]*Resource, error)

func ListResourcesVSphere(cloud *vsphere.VSphereCloud, clusterName string) (map[string]*Resource, error) {
	resources := make(map[string]*Resource)

	d := &clusterDiscoveryVSphere{
		cloud:        cloud,
		vsphereCloud: cloud,
		clusterName:  clusterName,
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

func (d *clusterDiscoveryVSphere) listVMs() ([]*Resource, error) {
	c := d.vsphereCloud

	regexForMasterVMs := "*" + "." + "masters" + "." + d.clusterName + "*"
	regexForNodeVMs := "nodes" + "." + d.clusterName + "*"

	vms, err := c.GetVirtualMachines([]string{regexForMasterVMs, regexForNodeVMs})
	if err != nil {
		if _, ok := err.(*find.NotFoundError); !ok {
			return nil, err
		}
		klog.Warning(err)
	}

	var trackers []*Resource
	for _, vm := range vms {
		tracker := &Resource{
			Name:    vm.Name(),
			ID:      vm.Name(),
			Type:    typeVM,
			Deleter: deleteVM,
			Dumper:  DumpVMInfo,
			Obj:     vm,
		}
		trackers = append(trackers, tracker)
	}
	return trackers, nil
}

func deleteVM(cloud fi.Cloud, r *Resource) error {
	vsphereCloud := cloud.(*vsphere.VSphereCloud)

	vm := r.Obj.(*object.VirtualMachine)

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
		klog.Fatalf("Destroy VM failed: %q", err)
	}

	return nil
}

func DumpVMInfo(op *DumpOperation, r *Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)
	return nil
}

func GetResourceTrackerKey(t *Resource) string {
	return t.Type + ":" + t.ID
}
