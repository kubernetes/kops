/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vfsclientset

import (
	"context"
	"fmt"
	"os"

	"k8s.io/kops/util/pkg/vfs"
)

func listChildNames(ctx context.Context, vfsPath vfs.Path) ([]string, error) {
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
