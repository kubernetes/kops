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

package instancegroups

import (
	"context"
	"errors"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

// SignalCh is a signalling channel
type SignalCh chan struct{}

// ResultCh is a signalling channel to get the result
type ResultCh chan error

// Signal is just an empty signal
var Signal = struct{}{}

var (
	// ErrRolloutCancelled indicates the rollout has been cancelled
	ErrRolloutCancelled = errors.New("rollout cancelled")
)

// RollingUpdateCluster is a struct containing cluster information for a rolling update.
type RollingUpdateCluster struct {
	// BastionInterval is the interval between with bastion node terminations
	BastionInterval time.Duration
	// Batch is the instances per instancegroup we should rollout concurrently
	Batch int
	// Client is the kubernetes the client
	Client kubernetes.Interface
	// CliendConfig is kubernetes configuration
	ClientConfig clientcmd.ClientConfig
	// Clientset is a client interface to kops
	Clientset simple.Clientset
	// Cloud is the cloud provider interface
	Cloud fi.Cloud
	// CloudOnly indicated this a cloud only operation
	CloudOnly bool
	// Cluster is a references to the cluster
	Cluster *api.Cluster
	// ClusterName is the name of the cluster
	ClusterName string
	// Count is the instances per instance group we should rollout
	Count int
	// Drain indicates we should drain the node before terminating
	Drain *bool
	// DrainTimeout is the amount of time will are willing we are willing to wait to drain the node
	DrainTimeout time.Duration
	// FailOnDrainError indicates we should stop on rollout on a drain failure
	FailOnDrainError bool
	// FailOnValidate indicates we should fail on any cluster validation errors
	FailOnValidate bool
	// FailOnValidateTimeout is the time were willing to wait to validate the cluster
	FailOnValidateTimeout time.Duration
	// Force all the instance groups to rollout
	Force bool
	// InstanceGroups is a list of groups to rollout on
	InstanceGroups []string
	// MasterInterval is the delay between master node terminations
	MasterInterval time.Duration
	// NodeBatch is the number of nodes groups to run concurrently
	NodeBatch int
	// NodeInterval is the delay between node terminations
	NodeInterval time.Duration
	// PostDrainDelay is the duration we should wait post draining node
	PostDrainDelay time.Duration
	// Strategy is the default strategy to employ
	Strategy api.RolloutStrategy
}

// RollingUpdateOptions are the options for a rolling update
type RollingUpdateOptions struct {
	// InstanceGroups is the groups we are rolling out to
	InstanceGroups map[string]*cloudinstances.CloudInstanceGroup
	// List of list of all the instancegroups within the cluster
	List *api.InstanceGroupList
}

// Rollout is the interface to a rollout provider
type Rollout interface {
	// RollingUpdate performs a rollout of the instance groups
	RollingUpdate(context.Context, *api.InstanceGroupList) error
}

// DrainOptions defined the options to drain a instance group
type DrainOptions struct {
	// Batch is the number of instances to do concurrently
	Batch int
	// Count is the number of instances to do
	Count int
	// CloudOnly indicates we only perform cloud api operations
	CloudOnly bool
	// Delete indicates we should delete the instance afterwards
	Delete bool
	// DrainPods indicates we should drain pods
	DrainPods bool
	// FailOnValidation indicates we should fail if the cluster does not validate
	FailOnValidation bool
	// Interval is the time between nodes to wait
	Interval time.Duration
	// PostDelay is the duration to watch after draining
	PostDelay time.Duration
	// Timeout is the time we will wait for the drain to finish
	Timeout time.Duration
	// ValidateCluster indicates we should valiate the cluster after each batch
	ValidateCluster bool
	// ValidationTimeout is the max time for a validation to occur
	ValidationTimeout time.Duration
}
