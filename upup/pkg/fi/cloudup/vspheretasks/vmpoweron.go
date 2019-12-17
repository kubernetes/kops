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

package vspheretasks

// vmpoweron houses task that powers on VM on vSphere cloud.

import (
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// VMPowerOn powers on a VMware VM
//go:generate fitask -type=VMPowerOn
type VMPowerOn struct {
	Name      *string
	AttachISO *AttachISO
}

var _ fi.HasName = &VMPowerOn{}
var _ fi.HasDependencies = &VMPowerOn{}

// GetDependencies returns map of tasks on which this task depends.
func (o *VMPowerOn) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	attachISOTask := tasks["AttachISO/"+*o.AttachISO.Name]
	if attachISOTask == nil {
		klog.Fatalf("Unable to find attachISO task %s dependency for VMPowerOn %s", *o.AttachISO.Name, *o.Name)
	}
	deps = append(deps, attachISOTask)
	return deps
}

// GetName returns the Name of the object, implementing fi.HasName
func (o *VMPowerOn) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *VMPowerOn) SetName(name string) {
	o.Name = &name
}

// Run executes DefaultDeltaRunMethod for this task.
func (e *VMPowerOn) Run(c *fi.Context) error {
	klog.Info("VMPowerOn.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

// Find is a no-op for vSphere cloud, for now.
func (e *VMPowerOn) Find(c *fi.Context) (*VMPowerOn, error) {
	klog.Info("VMPowerOn.Find invoked!")
	return nil, nil
}

// CheckChanges is a no-op for vSphere cloud, for now.
func (_ *VMPowerOn) CheckChanges(a, e, changes *VMPowerOn) error {
	klog.Info("VMPowerOn.CheckChanges invoked!")
	return nil
}

// RenderVSphere executes the actual power on operation for VM on vSphere cloud.
func (_ *VMPowerOn) RenderVSphere(t *vsphere.VSphereAPITarget, a, e, changes *VMPowerOn) error {
	klog.V(2).Infof("VMPowerOn.RenderVSphere invoked for vm %s", *changes.AttachISO.VM.Name)
	err := t.Cloud.PowerOn(*changes.AttachISO.VM.Name)
	return err
}
