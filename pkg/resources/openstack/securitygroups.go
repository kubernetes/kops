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
	"fmt"
	"strings"

	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeSG = "SecurityGroup"
)

func (os *clusterDiscoveryOS) ListSecurityGroups() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	opt := sg.ListOpts{}
	sgs, err := os.osCloud.ListSecurityGroups(opt)
	if err != nil {
		return nil, err
	}

	for _, sg := range sgs {
		if strings.HasSuffix(sg.Name, fmt.Sprintf(".%s", os.clusterName)) {
			resourceTracker := &resources.Resource{
				Name: sg.Name,
				ID:   sg.ID,
				Type: typeSG,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteSecurityGroup(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}
