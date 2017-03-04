package vspheretasks

import (
	"k8s.io/kops/upup/pkg/fi"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

// VirtualMachine represents a VMware VM
//go:generate fitask -type=VirtualMachine
type VirtualMachine struct {
	Name       *string
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

func (e *VirtualMachine) CompareWithID() *string {
	glog.Info("VirtualMachine.CompareWithID invoked!")
	return e.Name
}

func (e *VirtualMachine) Find(c *fi.Context) (*VirtualMachine, error) {
	glog.Info("VirtualMachine.Find invoked!")
	return nil, nil
}

func (e *VirtualMachine) Run(c *fi.Context) error {
	glog.Info("VirtualMachine.Run invoked!")
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *VirtualMachine) CheckChanges(a, e, changes *VirtualMachine) error {
	glog.Info("VirtualMachine.CheckChanges invoked!")
	return nil
}

func (_ *VirtualMachine) RenderVSphere(t *vsphere.VSphereAPITarget, a, e, changes *VirtualMachine) error {
	glog.Info("VirtualMachine.RenderVSphere invoked!")
	return nil
}


