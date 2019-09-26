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

package components

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeSchedulerOptionsBuilder adds options for kube-scheduler to the model
type KubeSchedulerOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &KubeSchedulerOptionsBuilder{}

func (b *KubeSchedulerOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.KubeScheduler == nil {
		clusterSpec.KubeScheduler = &kops.KubeSchedulerConfig{}
	}

	config := clusterSpec.KubeScheduler

	if config.UsePolicyConfigMap != nil && b.IsKubernetesLT("v1.7.0") {
		return fmt.Errorf("usePolicyConfigMap is only supported in Kubernetes 1.7.0 or later")
	}

	if config.LogLevel == 0 {
		// TODO: No way to set to 0?
		config.LogLevel = 2
	}

	if config.Image == "" {
		image, err := Image("kube-scheduler", b.Architecture(), clusterSpec, b.AssetBuilder)
		if err != nil {
			return err
		}
		config.Image = image
	}

	if config.LeaderElection == nil {
		//  Doesn't seem to be any real downside to always doing a leader election
		config.LeaderElection = &kops.LeaderElectionConfiguration{
			LeaderElect: fi.Bool(true),
		}
	}

	if config.Master == "" {
		if b.IsKubernetesLT("1.6") {
			// Backwards compatibility with pre-RBAC/pre-1.6 way of doing things
			config.Master = "http://127.0.0.1:8080"
		}
	}

	return nil
}
