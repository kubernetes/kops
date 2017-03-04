package vspheretasks

import (
	"k8s.io/kops/upup/pkg/fi"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// VMPowerOn powers on a VMware VM
//go:generate fitask -type=VMPowerOn
type VMPowerOn struct {
	Name       *string
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

