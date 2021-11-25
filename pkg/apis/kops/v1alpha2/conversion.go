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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/values"
)

func Convert_v1alpha2_ClusterSpec_To_kops_ClusterSpec(in *ClusterSpec, out *kops.ClusterSpec, s conversion.Scope) error {
	if err := autoConvert_v1alpha2_ClusterSpec_To_kops_ClusterSpec(in, out, s); err != nil {
		return err
	}
	if in.TagSubnets != nil {
		out.TagSubnets = values.Bool(!*in.TagSubnets)
	}
	for i, hook := range in.Hooks {
		if hook.Enabled != nil {
			out.Hooks[i].Enabled = values.Bool(!*hook.Enabled)
		}
	}
	return nil
}

func Convert_kops_ClusterSpec_To_v1alpha2_ClusterSpec(in *kops.ClusterSpec, out *ClusterSpec, s conversion.Scope) error {
	if err := autoConvert_kops_ClusterSpec_To_v1alpha2_ClusterSpec(in, out, s); err != nil {
		return err
	}
	if in.TagSubnets != nil {
		out.TagSubnets = values.Bool(!*in.TagSubnets)
	}
	for i, hook := range in.Hooks {
		if hook.Enabled != nil {
			out.Hooks[i].Enabled = values.Bool(!*hook.Enabled)
		}
	}
	return nil
}
