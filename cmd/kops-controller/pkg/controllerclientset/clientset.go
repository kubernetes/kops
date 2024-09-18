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

package controllerclientset

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// New constructs a new client for querying cluster and instancegroup information
// We split it out because we expect it to support CRDs etc in future.
func New(vfsContext *vfs.VFSContext, clusterBasePath vfs.Path, clusterName string, clusterKeystore fi.CAStore, clusterSecretStore fi.SecretStore) (simple.Clientset, error) {
	return &client{
		vfsContext: vfsContext,

		clusterBasePath: clusterBasePath,
		clusterName:     clusterName,

		clusterKeystore:    clusterKeystore,
		clusterSecretStore: clusterSecretStore,
	}, nil
}

type client struct {
	vfsContext         *vfs.VFSContext
	clusterBasePath    vfs.Path
	clusterName        string
	clusterKeystore    fi.CAStore
	clusterSecretStore fi.SecretStore
}

// GetCluster reads a cluster by name
func (c *client) GetCluster(ctx context.Context, name string) (*kops.Cluster, error) {
	if name != c.clusterName {
		return nil, fmt.Errorf("clientset bound to cluster %q, got cluster %q", c.clusterName, name)
	}

	p := c.clusterBasePath.Join("config")
	b, err := p.ReadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading file %v: %w", p, err)
	}

	gvk := kops.SchemeGroupVersion.WithKind("Cluster")
	object, _, err := kopscodecs.Decode(b, &gvk)
	if err != nil {
		return nil, fmt.Errorf("error parsing %v: %w", p, err)
	}

	cluster, ok := object.(*kops.Cluster)
	if !ok {
		return nil, fmt.Errorf("unexpected kind for cluster, got %T, want kops.Cluster", object)
	}
	return cluster, nil
}

// VFSContext returns a VFSContext.
func (c *client) VFSContext() *vfs.VFSContext {
	return c.vfsContext
}

// CreateCluster creates a cluster
func (c *client) CreateCluster(ctx context.Context, cluster *kops.Cluster) (*kops.Cluster, error) {
	return nil, fmt.Errorf("method CreateCluster not supported in server-side client")
}

// UpdateCluster updates a cluster
func (c *client) UpdateCluster(ctx context.Context, cluster *kops.Cluster, status *kops.ClusterStatus) (*kops.Cluster, error) {
	return nil, fmt.Errorf("method UpdateCluster not supported in server-side client")
}

// ListClusters returns all clusters
func (c *client) ListClusters(ctx context.Context, options metav1.ListOptions) (*kops.ClusterList, error) {
	return nil, fmt.Errorf("method ListClusters not supported in server-side client")
}

// ConfigBaseFor returns the vfs path where we will read configuration information from
func (c *client) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	return nil, fmt.Errorf("method ConfigBaseFor not supported in server-side client")
}

// InstanceGroupsFor returns the InstanceGroupInterface bound to the namespace for a particular Cluster
func (c *client) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	return newInstanceGroups(c.vfsContext, cluster, c.clusterBasePath)
}

// AddonsFor returns the client for addon objects for a particular Cluster
func (c *client) AddonsFor(cluster *kops.Cluster) simple.AddonsClient {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	basePath := c.clusterBasePath.Join("clusteraddons")

	return newAddonsClient(basePath, cluster)
}

// SecretStore builds the secret store for the specified cluster
func (c *client) SecretStore(cluster *kops.Cluster) (fi.SecretStore, error) {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	return c.clusterSecretStore, nil
}

// KeystoreReader builds the read-only key store for the specified cluster
func (c *client) KeystoreReader(cluster *kops.Cluster) (fi.KeystoreReader, error) {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	return c.clusterKeystore, nil
}

// KeyStore builds the Keystore Writer for the specified cluster
func (c *client) KeyStore(cluster *kops.Cluster) (fi.CAStore, error) {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	return c.clusterKeystore, nil
}

// SSHCredentialStore builds the SSHCredential store for the specified cluster
func (c *client) SSHCredentialStore(cluster *kops.Cluster) (fi.SSHCredentialStore, error) {
	clusterName := cluster.Name
	if clusterName != c.clusterName {
		klog.Fatalf("clientset bound to cluster %q, got cluster %q", c.clusterName, clusterName)
	}

	return newSSHCredentialStore(c.clusterBasePath, cluster), nil
}

// DeleteCluster deletes all the state for the specified cluster
func (c *client) DeleteCluster(ctx context.Context, cluster *kops.Cluster) error {
	return fmt.Errorf("method DeleteCluster not supported in server-side client")
}
