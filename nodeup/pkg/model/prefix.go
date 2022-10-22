/*
Copyright 2021 The Kubernetes Authors.

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

package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

type PrefixBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &PrefixBuilder{}

func (b *PrefixBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.Cluster.Spec.IsKopsControllerIPAM() {
		return nil
	}
	c.AddTask(&nodetasks.Prefix{
		Name: "prefix",
	})
	return nil
}
