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

package openstacktasks

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/monitors"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type PoolMonitor struct {
	ID        *string
	Name      *string
	Lifecycle fi.Lifecycle
	Pool      *LBPool
}

// GetDependencies returns the dependencies of the Instance task
func (p *PoolMonitor) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*LBPool); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &PoolMonitor{}

func (p *PoolMonitor) CompareWithID() *string {
	return p.ID
}

func (p *PoolMonitor) Find(context *fi.CloudupContext) (*PoolMonitor, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)

	opt := monitors.ListOpts{
		Name:   fi.ValueOf(p.Name),
		PoolID: fi.ValueOf(p.Pool.ID),
	}

	rs, err := cloud.ListMonitors(opt)
	if err != nil {
		return nil, err
	}
	if rs == nil || len(rs) == 0 {
		return nil, nil
	} else if len(rs) != 1 {
		return nil, fmt.Errorf("found multiple monitors with name: %s", fi.ValueOf(p.Name))
	}
	found := rs[0]
	actual := &PoolMonitor{
		ID:        fi.PtrTo(found.ID),
		Name:      fi.PtrTo(found.Name),
		Pool:      p.Pool,
		Lifecycle: p.Lifecycle,
	}
	p.ID = actual.ID
	return actual, nil
}

func (p *PoolMonitor) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(p, context)
}

func (_ *PoolMonitor) CheckChanges(a, e, changes *PoolMonitor) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *PoolMonitor) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *PoolMonitor) error {
	if a == nil {
		klog.V(2).Infof("Creating PoolMonitor with Name: %q", fi.ValueOf(e.Name))

		poolMonitor, err := t.Cloud.CreatePoolMonitor(monitors.CreateOpts{
			Name:           fi.ValueOf(e.Name),
			PoolID:         fi.ValueOf(e.Pool.ID),
			Type:           monitors.TypeTCP,
			Delay:          10,
			Timeout:        5,
			MaxRetries:     3,
			MaxRetriesDown: 3,
		})
		if err != nil {
			return fmt.Errorf("error creating PoolMonitor: %v", err)
		}
		e.ID = fi.PtrTo(poolMonitor.ID)
	}
	return nil
}
