package kubemanifest

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type ImageRemapFunction func(image string) (string, error)

func (m *Manifest) RemapImages(mapper ImageRemapFunction) error {
	visitor := &imageRemapVisitor{
		mapper: mapper,
	}
	err := m.accept(visitor)
	if err != nil {
		return err
	}
	//if changed {
	//	// invalidate cached rendering
	//	m.bytes = nil
	//}
	return nil
}

type imageRemapVisitor struct {
	visitorBase
	mapper ImageRemapFunction
}

func (m *imageRemapVisitor) VisitString(path []string, v string, mutator func(string)) error {
	n := len(path)
	if n < 1 || path[n-1] != "image" {
		return nil
	}

	// Deployments look like spec.template.spec.containers.[2].image
	if n < 3 || path[n-3] != "containers" {
		glog.Warningf("Skipping likely image field: %s", strings.Join(path, "."))
		return nil
	}

	image := v
	glog.V(4).Infof("Consider image for re-mapping: %q", image)
	remapped, err := m.mapper(v)
	if err != nil {
		return fmt.Errorf("error remapping image %q: %v", image, err)
	}
	if remapped != image {
		mutator(remapped)
	}
	return nil
}
