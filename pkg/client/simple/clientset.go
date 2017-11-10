/*
Copyright 2016 The Kubernetes Authors.

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

package simple

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// Clientset provides an interface to kops resources
type Clientset interface {
	// ConfigBaseFor returns the vfs path where we will read configuration information from
	ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error)
	// CreateCluster creates a cluster
	CreateCluster(cluster *kops.Cluster) (*kops.Cluster, error)
	// DeleteCluster deletes all the state for the specified cluster
	DeleteCluster(cluster *kops.Cluster) error
	// GetCluster reads a cluster by name
	GetCluster(name string) (*kops.Cluster, error)
	// InstanceGroupsFor returns the InstanceGroupInterface bounds to the namespace for a particular Cluster
	InstanceGroupsFor(cluster *kops.Cluster) internalversion.InstanceGroupInterface
	// KeyStore builds the key store for the specified cluster
	KeyStore(cluster *kops.Cluster) (fi.CAStore, error)
	// ListClusters returns all clusters
	ListClusters(options metav1.ListOptions) (*kops.ClusterList, error)
	// SecretStore builds the secret store for the specified cluster
	SecretStore(cluster *kops.Cluster) (fi.SecretStore, error)
	// SSHCredentialStore builds the SSHCredential store for the specified cluster
	SSHCredentialStore(cluster *kops.Cluster) (fi.SSHCredentialStore, error)
	// UpdateCluster updates a cluster
	UpdateCluster(cluster *kops.Cluster, status *kops.ClusterStatus) (*kops.Cluster, error)
}
