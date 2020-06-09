/*
Copyright 2017 The Kubernetes Authors.

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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

// CommonBuilder runs common tasks
type CommonBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &CommonBuilder{}

// Build is responsible for copying the common CNI binaries
func (b *CommonBuilder) Build(c *fi.ModelBuilderContext) error {
	assets := []string{
		"bandwidth",
		"bridge",
		"host-local",
		"loopback",
		"portmap",
		"ptp",
		"tuning",
	}

	if err := b.AddCNIBinAssets(c, assets); err != nil {
		return err
	}

	return nil
}
