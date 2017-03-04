package vspheretasks

import (
	"k8s.io/kops/upup/pkg/fi"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// AttachISO represents the cloud-init ISO file attached to a VMware VM
//go:generate fitask -type=AttachISO
type AttachISO struct {
	Name       *string
	VM *VirtualMachine
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

