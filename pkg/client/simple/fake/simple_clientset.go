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

package fake

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	fakekopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion/fake"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/util/pkg/vfs"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(registry, scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	fakePtr := testing.Fake{}
	fakePtr.AddReactor("*", "*", testing.ObjectReaction(o, registry.RESTMapper()))

	fakePtr.AddWatchReactor("*", testing.DefaultWatchReactor(watch.NewFake(), nil))

	return &Clientset{fakePtr}
}

// Clientset implements simple.Clientset. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
}

var _ simple.Clientset = &Clientset{}

// ClustersFor returns the ClusterInterface bound to the namespace for a particular Cluster
func (c *Clientset) ClustersFor(cluster *kops.Cluster) kopsinternalversion.ClusterInterface {
	/*
		fk := &fakekopsinternalversion.FakeKops{&c.Fake}
		return fk.Clusters("")
	*/
	return nil
}

// GetCluster reads a cluster by name
func (c *Clientset) GetCluster(name string) (*kops.Cluster, error) {
	fk := &fakekopsinternalversion.FakeKops{&c.Fake}
	return fk.Clusters("").Get(name, metav1.GetOptions{})
}

// ListClusters returns all clusters
func (c *Clientset) ListClusters(options metav1.ListOptions) (*kops.ClusterList, error) {
	fk := &fakekopsinternalversion.FakeKops{&c.Fake}
	return fk.Clusters("").List(options)
}

// ConfigBaseFor returns the vfs path where we will read configuration information from
func (c *Clientset) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	return nil, fmt.Errorf("unimplemented")
}

// InstanceGroupsFor returns the InstanceGroupInterface bounds to the namespace for a particular Cluster
func (c *Clientset) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	clusterName := cluster.ObjectMeta.Name
	fk := &fakekopsinternalversion.FakeKops{&c.Fake}
	return fk.InstanceGroups(clusterName)
}

// FederationsFor returns the FederationInterface bounds to the namespace for a particular Federation
func (c *Clientset) FederationsFor(federation *kops.Federation) kopsinternalversion.FederationInterface {
	return nil
}

// ListFederations returns all federations
func (c *Clientset) ListFederations(options metav1.ListOptions) (*kops.FederationList, error) {
	fk := &fakekopsinternalversion.FakeKops{&c.Fake}
	return fk.Federations("").List(options)
}

// GetFederation reads a federation by name
func (c *Clientset) GetFederation(name string) (*kops.Federation, error) {
	fk := &fakekopsinternalversion.FakeKops{&c.Fake}
	return fk.Federations("").Get(name, metav1.GetOptions{})
}
