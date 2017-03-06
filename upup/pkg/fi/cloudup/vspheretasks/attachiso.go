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

import (
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// AttachISO represents the cloud-init ISO file attached to a VMware VM
//go:generate fitask -type=AttachISO
type AttachISO struct {
	Name *string
	VM   *VirtualMachine
}

var _ fi.HasName = &AttachISO{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *AttachISO) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *AttachISO) SetName(name string) {
	o.Name = &name
}

func (e *AttachISO) Run(c *fi.Context) error {
	glog.Info("AttachISO.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *AttachISO) Find(c *fi.Context) (*AttachISO, error) {
	glog.Info("AttachISO.Find invoked!")
	return nil, nil
}

func (_ *AttachISO) CheckChanges(a, e, changes *AttachISO) error {
	glog.Info("AttachISO.CheckChanges invoked!")
	return nil
}

func (_ *AttachISO) RenderVC(t *vsphere.VSphereAPITarget, a, e, changes *AttachISO) error {
	glog.Info("AttachISO.RenderVC invoked!")
	return nil
}
