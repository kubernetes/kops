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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// CalicoBuilder configures the etcd TLS support for Calico
type CalicoBuilder struct {
	*model.NodeupModelContext
}

var _ fi.NodeupModelBuilder = &CalicoBuilder{}

// Build is responsible for performing setup for Calico.
func (b *CalicoBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.NodeupConfig.Networking.Calico == nil {
		return nil
	}

	if b.Distribution.IsUbuntu() {
		c.AddTask(&nodetasks.Package{Name: "wireguard"})
	}

	return nil
}
