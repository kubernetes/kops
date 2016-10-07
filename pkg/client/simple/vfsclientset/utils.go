package vfsclientset

import (
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"fmt"
)

func listChildNames(vfsPath vfs.Path) ([]string, error) {
	children, err := vfsPath.ReadDir()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing children of %s: %v", vfsPath, err)
	}

	var names []string
	for _, child := range children {
		names = append(names, child.Base())
	}
	return names, nil
}
