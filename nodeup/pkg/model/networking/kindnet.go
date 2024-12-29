/*
Copyright 2024 The Kubernetes Authors.

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

// KindnetBuilder configures the etcd TLS support for Calico
type KindnetBuilder struct {
	*model.NodeupModelContext
}

var _ fi.NodeupModelBuilder = &KindnetBuilder{}

// Build is responsible for performing setup for Kindnet.
func (b *KindnetBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.NodeupConfig.Networking.Kindnet == nil {
		return nil
	}

	return nil
}
