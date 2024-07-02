/*
Copyright 2020 The Kubernetes Authors.

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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KarpenterOptionsBuilder adds options for karpenter to the model
type KarpenterOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KarpenterOptionsBuilder{}

func (b *KarpenterOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	c := clusterSpec.Karpenter
	if c == nil {
		return nil
	}

	if c.Image == "" {
		c.Image = "public.ecr.aws/karpenter/controller:v0.32.6"
	}

	if c.LogEncoding == "" {
		c.LogEncoding = "console"
	}

	if c.LogLevel == "" {
		c.LogLevel = "debug"
	}

	return nil
}
