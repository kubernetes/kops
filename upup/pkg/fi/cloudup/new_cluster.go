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
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
)

const (
	AuthorizationFlagAlwaysAllow = "AlwaysAllow"
	AuthorizationFlagRBAC        = "RBAC"
)

type NewClusterOptions struct {
	// ClusterName is the name of the cluster to initialize.
	ClusterName string

	// Authorization is the authorization mode to use. The options are "RBAC" (default) and "AlwaysAllow".
	Authorization string
	// Channel is a channel location for initializing the cluster. It defaults to "stable".
	Channel string
	// ConfigBase is the location where we will store the configuration. It defaults to the state store.
	ConfigBase string
}

func (o *NewClusterOptions) InitDefaults() {
	o.Channel = api.DefaultChannel
	o.Authorization = AuthorizationFlagRBAC
}

type NewClusterResult struct {
	// Cluster is the initialized Cluster resource.
	Cluster *api.Cluster

	// TODO remove after more create_cluster logic refactored in
	Channel *api.Channel
}

// NewCluster initializes cluster and instance groups specifications as
// intended for newly created clusters.
func NewCluster(opt *NewClusterOptions, clientset simple.Clientset) (*NewClusterResult, error) {
	if opt.ClusterName == "" {
		return nil, fmt.Errorf("name is required")
	}

	if opt.Channel == "" {
		opt.Channel = api.DefaultChannel
	}
	channel, err := api.LoadChannel(opt.Channel)
	if err != nil {
		return nil, err
	}

	cluster := api.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: opt.ClusterName,
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

	cluster.Spec.ConfigBase = opt.ConfigBase
	configBase, err := clientset.ConfigBaseFor(&cluster)
	if err != nil {
		return nil, fmt.Errorf("error building ConfigBase for cluster: %v", err)
	}
	cluster.Spec.ConfigBase = configBase.Path()

	cluster.Spec.Authorization = &api.AuthorizationSpec{}
	if strings.EqualFold(opt.Authorization, AuthorizationFlagAlwaysAllow) {
		cluster.Spec.Authorization.AlwaysAllow = &api.AlwaysAllowAuthorizationSpec{}
	} else if opt.Authorization == "" || strings.EqualFold(opt.Authorization, AuthorizationFlagRBAC) {
		cluster.Spec.Authorization.RBAC = &api.RBACAuthorizationSpec{}
	} else {
		return nil, fmt.Errorf("unknown authorization mode %q", opt.Authorization)
	}

	result := NewClusterResult{
		Cluster: &cluster,
		Channel: channel,
	}
	return &result, nil
}
