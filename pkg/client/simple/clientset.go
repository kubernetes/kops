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
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/util/pkg/vfs"
	"net/url"
	"strings"
)

type Clientset interface {
	// ClustersFor returns the ClusterInterface bound to the namespace for a particular Cluster
	ClustersFor(cluster *kops.Cluster) kopsinternalversion.ClusterInterface

	// GetCluster reads a cluster by name
	GetCluster(name string) (*kops.Cluster, error)

	// ListClusters returns all clusters
	ListClusters(options metav1.ListOptions) (*kops.ClusterList, error)

	// ConfigBaseFor returns the vfs path where we will read configuration information from
	ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error)

	// InstanceGroupsFor returns the InstanceGroupInterface bounds to the namespace for a particular Cluster
	InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface

	// FederationsFor returns the FederationInterface bounds to the namespace for a particular Federation
	FederationsFor(federation *kops.Federation) kopsinternalversion.FederationInterface

	// GetFederation reads a federation by name
	GetFederation(name string) (*kops.Federation, error)

	// ListFederations returns all federations
	ListFederations(options metav1.ListOptions) (*kops.FederationList, error)
}

// RESTClientset is an implementation of clientset that uses a "real" generated REST client
type RESTClientset struct {
	BaseURL    *url.URL
	KopsClient kopsinternalversion.KopsInterface
}

func (c *RESTClientset) ClustersFor(cluster *kops.Cluster) kopsinternalversion.ClusterInterface {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.Clusters(namespace)
}

func (c *RESTClientset) GetCluster(name string) (*kops.Cluster, error) {
	namespace := restNamespaceForClusterName(name)
	return c.KopsClient.Clusters(namespace).Get(name, metav1.GetOptions{})
}

func (c *RESTClientset) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	// URL for clusters looks like  https://<server>/apis/kops/v1alpha2/namespaces/<cluster>/clusters/<cluster>
	// We probably want to add a subresource for full resources
	return vfs.Context.BuildVfsPath(c.BaseURL.String())
}

func (c *RESTClientset) ListClusters(options metav1.ListOptions) (*kops.ClusterList, error) {
	return c.KopsClient.Clusters(metav1.NamespaceAll).List(options)
}

func (c *RESTClientset) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.InstanceGroups(namespace)
}

func (c *RESTClientset) FederationsFor(federation *kops.Federation) kopsinternalversion.FederationInterface {
	// Unsure if this should be namespaced or not - probably, so that we can RBAC it...
	panic("Federations are curently not supported by the server API")
	//namespace := restNamespaceForFederationName(federation.Name)
	//return c.KopsClient.Federations(namespace)
}

func (c *RESTClientset) ListFederations(options metav1.ListOptions) (*kops.FederationList, error) {
	return c.KopsClient.Federations(metav1.NamespaceAll).List(options)
}

func (c *RESTClientset) GetFederation(name string) (*kops.Federation, error) {
	namespace := restNamespaceForFederationName(name)
	return c.KopsClient.Federations(namespace).Get(name, metav1.GetOptions{})
}

func restNamespaceForClusterName(clusterName string) string {
	// We are not allowed dots, so we map them to dashes
	// This can conflict, but this will simply be a limitation that we pass on to the user
	// i.e. it will not be possible to create a.b.example.com and a-b.example.com
	namespace := strings.Replace(clusterName, ".", "-", -1)
	return namespace
}

func restNamespaceForFederationName(clusterName string) string {
	namespace := strings.Replace(clusterName, ".", "-", -1)
	return namespace
}
