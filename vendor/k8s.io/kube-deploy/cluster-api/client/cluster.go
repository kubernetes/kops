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

package client

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

type ClustersGetter interface {
	Clusters() ClustersInterface
}

// ClustersInterface has methods to work with Cluster resources.
type ClustersInterface interface {
	Create(*clusterv1.Cluster) (*clusterv1.Cluster, error)
	Update(*clusterv1.Cluster) (*clusterv1.Cluster, error)
	Delete(string, *metav1.DeleteOptions) error
	List(metav1.ListOptions) (*clusterv1.ClusterList, error)
	Get(string, metav1.GetOptions) (*clusterv1.Cluster, error)
}

// clusters implements ClustersInterface
type clusters struct {
	client rest.Interface
}

// newClusters returns a clusters
func newClusters(c *ClusterAPIV1Alpha1Client) *clusters {
	return &clusters{
		client: c.RESTClient(),
	}
}

func (c *clusters) Create(cluster *clusterv1.Cluster) (result *clusterv1.Cluster, err error) {
	result = &clusterv1.Cluster{}
	err = c.client.Post().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.ClustersCRDPlural).
		Body(cluster).
		Do().
		Into(result)
	return
}

func (c *clusters) Update(cluster *clusterv1.Cluster) (result *clusterv1.Cluster, err error) {
	result = &clusterv1.Cluster{}
	err = c.client.Put().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.ClustersCRDPlural).
		Name(cluster.Name).
		Body(cluster).
		Do().
		Into(result)
	return
}

func (c *clusters) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.ClustersCRDPlural).
		Name(name).
		Body(options).
		Do().
		Error()
}

// List takes label and field selectors, and returns the list of machines that match those selectors.
func (c *clusters) List(opts metav1.ListOptions) (result *clusterv1.ClusterList, err error) {
	result = &clusterv1.ClusterList{}
	err = c.client.Get().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.ClustersCRDPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func (c *clusters) Get(name string, options metav1.GetOptions) (result *clusterv1.Cluster, err error) {
	result = &clusterv1.Cluster{}
	err = c.client.Get().
		Namespace(apiv1.NamespaceDefault).
		Resource(clusterv1.ClustersCRDPlural).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}
