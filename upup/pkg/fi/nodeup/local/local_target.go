package local

import "k8s.io/kops/upup/pkg/fi"

type LocalTarget struct {
	CacheDir string
	Tags     map[string]struct{}
}

var _ fi.Target = &LocalTarget{}

func (t *LocalTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *LocalTarget) HasTag(tag string) bool {
	_, found := t.Tags[tag]
	return found
}
