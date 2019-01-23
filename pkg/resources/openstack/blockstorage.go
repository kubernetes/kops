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

	"github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeVolume = "Volume"
)

var listBlockStorageFunctions = []listFn{
	listVolumes,
}

func listVolumes(cloud openstack.OpenstackCloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	opts := volumes.ListOpts{
		Metadata: cloud.GetCloudTags(),
	}
	vs, err := cloud.ListVolumes(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %s", err)
	}

	for _, v := range vs {
		resourceTracker := &resources.Resource{
			Name: v.Name,
			ID:   v.ID,
			Type: typeVolume,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return volumes.Delete(cloud.(openstack.OpenstackCloud).BlockStorageClient(), r.ID).ExtractErr()
			},
			Obj: v,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
