package scaleway

import "k8s.io/kops/upup/pkg/fi"

type ScwAPITarget struct {
	Cloud ScwCloud
}

var _ fi.Target = &ScwAPITarget{}

func NewScwAPITarget(cloud ScwCloud) *ScwAPITarget {
	return &ScwAPITarget{
		Cloud: cloud,
	}
}

func (s ScwAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (s ScwAPITarget) ProcessDeletions() bool {
	return true
}
