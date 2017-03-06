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

// VMPowerOn powers on a VMware VM
//go:generate fitask -type=VMPowerOn
type VMPowerOn struct {
	Name      *string
	AttachISO *AttachISO
}

var _ fi.HasName = &VMPowerOn{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *VMPowerOn) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *VMPowerOn) SetName(name string) {
	o.Name = &name
}

func (e *VMPowerOn) Run(c *fi.Context) error {
	glog.Info("VMPowerOn.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

func (e *VMPowerOn) Find(c *fi.Context) (*VMPowerOn, error) {
	glog.Info("VMPowerOn.Find invoked!")
	return nil, nil
}

func (_ *VMPowerOn) CheckChanges(a, e, changes *VMPowerOn) error {
	glog.Info("VMPowerOn.CheckChanges invoked!")
	return nil
}

func (_ *VMPowerOn) RenderVC(t *vsphere.VSphereAPITarget, a, e, changes *VMPowerOn) error {
	glog.Info("VMPowerOn.RenderVC invoked!")
	return nil
}
