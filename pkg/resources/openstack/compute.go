/*
Copyright 2018 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeKeypair     = "keypair"
	typeServer      = "server"
	typeServerGroup = "serverGroup"
)

func listInstances(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	clusterTag := "KubernetesCluster:" + strings.Replace(clusterName, ".", "-", -1)

	opts := servers.ListOpts{
		Name: clusterTag,
	}
	ss, err := cloud.ListInstances(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %v", err)
	}

	for _, s := range ss {
		resourceTracker := &resources.Resource{
			Name: s.Name,
			ID:   s.ID,
			Type: typeServer,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return servers.Delete(cloud.(openstack.OpenstackCloud).ComputeClient(), s.ID).ExtractErr()
			},
			Obj: s,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listServerGroups(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	sgsAll, err := cloud.ListServerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to extract server group pages: %s", err)
	}

	var sgs []servergroups.ServerGroup
	for _, sg := range sgsAll {
		if strings.HasPrefix(sg.Name, clusterName) {
			sgs = append(sgs, sg)
		}
	}

	for _, sg := range sgs {
		resourceTracker := &resources.Resource{
			Name: sg.ID,
			ID:   sg.Name,
			Type: typeServerGroup,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return servergroups.Delete(cloud.(openstack.OpenstackCloud).ComputeClient(), r.ID).ExtractErr()
			},
			Obj: sg,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listKeypairs(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	kp, err := cloud.ListKeypair(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get keypair: %s", err)
	}

	if kp == nil {
		return resourceTrackers, nil
	}

	resourceTracker := &resources.Resource{
		Name: kp.Name,
		ID:   kp.Name,
		Type: typeKeypair,
		Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
			return keypairs.Delete(cloud.(openstack.OpenstackCloud).ComputeClient(), r.ID).ExtractErr()
		},
		Obj: kp,
	}

	resourceTrackers = append(resourceTrackers, resourceTracker)

	return resourceTrackers, nil
}
