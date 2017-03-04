package vsphere

import "k8s.io/kops/upup/pkg/fi"

type VSphereAPITarget struct {
	Cloud *VSphereCloud
}

var _ fi.Target = &VSphereAPITarget{}

func NewVSphereAPITarget(cloud *VSphereCloud) *VSphereAPITarget {
	return &VSphereAPITarget{
		Cloud: cloud,
	}
}

func (t *VSphereAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *VSphereAPITarget) ProcessDeletions() bool {
	return true
}
