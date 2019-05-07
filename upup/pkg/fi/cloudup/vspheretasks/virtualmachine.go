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

// virtualmachine houses task that creates VM on vSphere cloud.

import (
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// VirtualMachine represents a VMware VM
//go:generate fitask -type=VirtualMachine
type VirtualMachine struct {
	Name           *string
	VMTemplateName *string
}

var _ fi.CompareWithID = &VirtualMachine{}
var _ fi.HasName = &VirtualMachine{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *VirtualMachine) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *VirtualMachine) SetName(name string) {
	o.Name = &name
}

// String is the stringer function for the task, producing readable output using fi.TaskAsString
func (o *VirtualMachine) String() string {
	return fi.TaskAsString(o)
}

// CompareWithID is returning name of this VirtualMachine.
func (e *VirtualMachine) CompareWithID() *string {
	klog.V(4).Info("VirtualMachine.CompareWithID invoked!")
	return e.Name
}

// Find is a no-op for vSphere.
func (e *VirtualMachine) Find(c *fi.Context) (*VirtualMachine, error) {
	klog.V(4).Info("VirtualMachine.Find invoked!")
	return nil, nil
}

// Run executes DefaultDeltaRunMethod for this task.
func (e *VirtualMachine) Run(c *fi.Context) error {
	klog.V(4).Info("VirtualMachine.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

// CheckChanges is a no-op for vSphere, for now.
func (_ *VirtualMachine) CheckChanges(a, e, changes *VirtualMachine) error {
	klog.V(4).Info("VirtualMachine.CheckChanges invoked!")
	return nil
}

// RenderVSphere executes the actual VM clone creation for vSphere cloud.
func (_ *VirtualMachine) RenderVSphere(t *vsphere.VSphereAPITarget, a, e, changes *VirtualMachine) error {
	klog.V(4).Infof("VirtualMachine.RenderVSphere invoked with a(%+v) e(%+v) and changes(%+v)", a, e, changes)
	_, err := t.Cloud.CreateLinkClonedVm(changes.Name, changes.VMTemplateName)
	if err != nil {
		return err
	}
	return nil
}
