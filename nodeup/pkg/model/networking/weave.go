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

// WeaveBuilder installs weave
type WeaveBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &WeaveBuilder{}

// Build is responsible for configuring the weave
func (b *WeaveBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.Networking.Weave == nil {
		return nil
	}

	b.AddCNIBinAssets(c, []string{"portmap"})

	return nil
}
