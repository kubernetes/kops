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

package vsphere

// vsphere_utils houses various utility methods related to vSphere cloud.

import (
	"context"
	"path"
	"sync"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/klog"
)

var snapshotLock sync.Mutex

func createSnapshot(ctx context.Context, vm *object.VirtualMachine, snapshotName string, snapshotDesc string) (object.Reference, error) {
	snapshotLock.Lock()
	defer snapshotLock.Unlock()

	snapshotRef, err := findSnapshot(vm, ctx, snapshotName)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Template VM is %s and snapshot is %s", vm, snapshotRef)
	if snapshotRef != nil {
		return snapshotRef, nil
	}

	task, err := vm.CreateSnapshot(ctx, snapshotName, snapshotDesc, false, false)
	if err != nil {
		return nil, err
	}

	taskInfo, err := task.WaitForResult(ctx, nil)
	if err != nil {
		return nil, err
	}
	klog.Infof("taskInfo.Result is %s", taskInfo.Result)
	return taskInfo.Result.(object.Reference), nil
}

type snapshotMap map[string][]object.Reference

func (m snapshotMap) add(parent string, tree []types.VirtualMachineSnapshotTree) {
	for i, st := range tree {
		sname := st.Name
		names := []string{sname, st.Snapshot.Value}

		if parent != "" {
			sname = path.Join(parent, sname)
			// Add full path as an option to resolve duplicate names
			names = append(names, sname)
		}

		for _, name := range names {
			m[name] = append(m[name], &tree[i].Snapshot)
		}

		m.add(sname, st.ChildSnapshotList)
	}
}

func findSnapshot(v *object.VirtualMachine, ctx context.Context, name string) (object.Reference, error) {
	var o mo.VirtualMachine

	err := v.Properties(ctx, v.Reference(), []string{"snapshot"}, &o)
	if err != nil {
		return nil, err
	}

	if o.Snapshot == nil || len(o.Snapshot.RootSnapshotList) == 0 {
		return nil, nil
	}

	m := make(snapshotMap)
	m.add("", o.Snapshot.RootSnapshotList)

	s := m[name]
	switch len(s) {
	case 0:
		return nil, nil
	case 1:
		return s[0], nil
	default:
		klog.Warningf("VM %s seems to have more than one snapshots with name %s. Using a random snapshot.", v, name)
		return s[0], nil
	}
}
