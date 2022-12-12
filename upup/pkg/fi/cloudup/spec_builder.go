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

package cloudup

import (
	"context"

	"k8s.io/klog/v2"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/util/pkg/reflectutils"
)

type SpecBuilder struct {
	OptionsLoader *loader.OptionsLoader
}

func (l *SpecBuilder) BuildCompleteSpec(ctx context.Context, clusterSpec *kopsapi.ClusterSpec) (*kopsapi.ClusterSpec, error) {
	loaded, err := l.OptionsLoader.Build(ctx, clusterSpec)
	if err != nil {
		return nil, err
	}
	completed := &kopsapi.ClusterSpec{}
	*completed = *(loaded.(*kopsapi.ClusterSpec))

	// Control-plane kubelet config = (base kubelet config + control-plane kubelet config)
	controlPlaneKubelet := &kopsapi.KubeletConfigSpec{}
	reflectutils.JSONMergeStruct(controlPlaneKubelet, completed.Kubelet)
	reflectutils.JSONMergeStruct(controlPlaneKubelet, completed.ControlPlaneKubelet)
	completed.ControlPlaneKubelet = controlPlaneKubelet

	klog.V(1).Infof("options: %s", fi.DebugAsJsonStringIndent(completed))
	return completed, nil
}
