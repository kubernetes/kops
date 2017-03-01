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

package gcemodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

type GCEModelContext struct {
	*model.KopsModelContext
}

func (b *GCEModelContext) LinkToNetwork() *gcetasks.Network {
	return &gcetasks.Network{Name: s("default")}
}

// SafeClusterName returns the cluster name escaped for the given cloud
func (c *GCEModelContext) SafeObjectName(name string) string {
	gceName := name + "-" + c.Cluster.ObjectMeta.Name

	// TODO: If the cluster name > some max size (32?) we should curtail it
	return gce.SafeClusterName(gceName)
}

func (c *GCEModelContext) GCETagForRole(role kops.InstanceGroupRole) string {
	return components.GCETagForRole(c.Cluster.ObjectMeta.Name, role)
}
