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

package cloudup

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
)

type NewClusterOptions struct {
	// Name is the name of the cluster to initialize.
	Name string

	// Channel is a channel location for initializing the cluster.
	Channel string
}

type NewClusterResult struct {
	// Cluster is the initialized Cluster resource.
	Cluster *api.Cluster

	// TODO remove after more create_cluster logic refactored in
	Channel *api.Channel
}

// NewCluster initializes cluster and instance groups specifications as
// intended for newly created clusters.
func NewCluster(opt *NewClusterOptions) (*NewClusterResult, error) {
	if opt.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	channel, err := api.LoadChannel(opt.Channel)
	if err != nil {
		return nil, err
	}

	cluster := api.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: opt.Name,
		},
	}

	if channel.Spec.Cluster != nil {
		cluster.Spec = *channel.Spec.Cluster

		kubernetesVersion := api.RecommendedKubernetesVersion(channel, kops.Version)
		if kubernetesVersion != nil {
			cluster.Spec.KubernetesVersion = kubernetesVersion.String()
		}
	}
	cluster.Spec.Channel = opt.Channel

	result := NewClusterResult{
		Cluster: &cluster,
		Channel: channel,
	}
	return &result, nil
}
