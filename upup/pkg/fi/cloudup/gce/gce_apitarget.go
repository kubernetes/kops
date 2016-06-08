package gce

import (
	"k8s.io/kube-deploy/upup/pkg/fi"
)

type GCEAPITarget struct {
	Cloud *GCECloud
}

var _ fi.Target = &GCEAPITarget{}

func NewGCEAPITarget(cloud *GCECloud) *GCEAPITarget {
	return &GCEAPITarget{
		Cloud: cloud,
	}
}

func (t *GCEAPITarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}
