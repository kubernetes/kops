package local

import "k8s.io/kops/upup/pkg/fi"

type LocalTarget struct {
	CacheDir string
}

var _ fi.Target = &LocalTarget{}

func (t *LocalTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}
