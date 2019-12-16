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

package api

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/validation"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
)

// RESTClientset is an implementation of clientset that uses a "real" generated REST client
type RESTClientset struct {
	BaseURL    *url.URL
	KopsClient kopsinternalversion.KopsInterface
}

// GetCluster implements the GetCluster method of Clientset for a kubernetes-API state store
func (c *RESTClientset) GetCluster(name string) (*kops.Cluster, error) {
	namespace := restNamespaceForClusterName(name)
	return c.KopsClient.Clusters(namespace).Get(name, metav1.GetOptions{})
}

// CreateCluster implements the CreateCluster method of Clientset for a kubernetes-API state store
func (c *RESTClientset) CreateCluster(cluster *kops.Cluster) (*kops.Cluster, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.Clusters(namespace).Create(cluster)
}

// UpdateCluster implements the UpdateCluster method of Clientset for a kubernetes-API state store
func (c *RESTClientset) UpdateCluster(cluster *kops.Cluster, status *kops.ClusterStatus) (*kops.Cluster, error) {
	klog.Warningf("validating cluster update client side; needs to move to server")
	old, err := c.GetCluster(cluster.Name)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidateClusterUpdate(cluster, status, old).ToAggregate(); err != nil {
		return nil, err
	}

	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.Clusters(namespace).Update(cluster)
}

// ConfigBaseFor implements the ConfigBaseFor method of Clientset for a kubernetes-API state store
func (c *RESTClientset) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	if cluster.Spec.ConfigBase != "" {
		return vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	}
	// URL for clusters looks like https://<server>/apis/kops/v1alpha2/namespaces/<cluster>/clusters/<cluster>
	// We probably want to add a subresource for full resources
	return vfs.Context.BuildVfsPath(c.BaseURL.String())
}

// ListClusters implements the ListClusters method of Clientset for a kubernetes-API state store
func (c *RESTClientset) ListClusters(options metav1.ListOptions) (*kops.ClusterList, error) {
	return c.KopsClient.Clusters(metav1.NamespaceAll).List(options)
}

// InstanceGroupsFor implements the InstanceGroupsFor method of Clientset for a kubernetes-API state store
func (c *RESTClientset) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.InstanceGroups(namespace)
}

func (c *RESTClientset) SecretStore(cluster *kops.Cluster) (fi.SecretStore, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return secrets.NewClientsetSecretStore(cluster, c.KopsClient, namespace), nil
}

func (c *RESTClientset) KeyStore(cluster *kops.Cluster) (fi.CAStore, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return fi.NewClientsetCAStore(cluster, c.KopsClient, namespace), nil
}

func (c *RESTClientset) SSHCredentialStore(cluster *kops.Cluster) (fi.SSHCredentialStore, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return fi.NewClientsetSSHCredentialStore(cluster, c.KopsClient, namespace), nil
}

func (c *RESTClientset) DeleteCluster(cluster *kops.Cluster) error {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return err
	}

	err = vfsclientset.DeleteAllClusterState(configBase)
	if err != nil {
		return err
	}

	name := cluster.Name
	namespace := restNamespaceForClusterName(name)

	{
		keysets, err := c.KopsClient.Keysets(namespace).List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing Keysets: %v", err)
		}

		for i := range keysets.Items {
			keyset := &keysets.Items[i]
			err = c.KopsClient.Keysets(namespace).Delete(keyset.Name, &metav1.DeleteOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// Unlikely...
					klog.Warningf("Keyset was concurrently deleted")
				} else {
					return fmt.Errorf("error deleting Keyset %q: %v", keyset.Name, err)
				}
			}
		}
	}

	{
		igs, err := c.KopsClient.InstanceGroups(namespace).List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing instance groups: %v", err)
		}

		for i := range igs.Items {
			ig := &igs.Items[i]
			err = c.KopsClient.InstanceGroups(namespace).Delete(ig.Name, &metav1.DeleteOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// Unlikely...
					klog.Warningf("instance group was concurrently deleted")
				} else {
					return fmt.Errorf("error deleting instance group %q: %v", ig.Name, err)
				}
			}
		}
	}

	err = c.KopsClient.Clusters(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Unlikely...
			klog.Warningf("cluster %q was concurrently deleted", name)
		} else {
			return fmt.Errorf("error deleting cluster%q: %v", name, err)
		}
	}

	return nil
}

func restNamespaceForClusterName(clusterName string) string {
	// We are not allowed dots, so we map them to dashes
	// This can conflict, but this will simply be a limitation that we pass on to the user
	// i.e. it will not be possible to create a.b.example.com and a-b.example.com
	namespace := strings.Replace(clusterName, ".", "-", -1)
	return namespace
}
