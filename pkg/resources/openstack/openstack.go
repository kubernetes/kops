/*
Copyright 2019 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type openstackListFn func() ([]*resources.Resource, error)

type clusterDiscoveryOS struct {
	cloud       fi.Cloud
	osCloud     openstack.OpenstackCloud
	clusterName string

	zones []string
}

// ListResources lists the OpenStack resources kops manages
func ListResources(cloud openstack.OpenstackCloud, clusterName string) (map[string]*resources.Resource, error) {
	resources := make(map[string]*resources.Resource)

	os := &clusterDiscoveryOS{
		cloud:       cloud,
		osCloud:     cloud,
		clusterName: clusterName,
	}

	listFunctions := []openstackListFn{
		os.ListKeypairs,
		os.ListInstances,
		os.ListServerGroups,
		os.ListVolumes,
		os.ListSecurityGroups,
		os.ListNetwork,
		os.ListDNSRecordsets,
	}
	for _, fn := range listFunctions {
		resourceTrackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}
	return resources, nil
}
