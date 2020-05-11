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
	cinder "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeVolume = "Volume"
)

func (os *clusterDiscoveryOS) ListVolumes() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	opt := cinder.ListOpts{
		Metadata: map[string]string{"KubernetesCluster": os.clusterName},
	}
	volumes, err := os.osCloud.ListVolumes(opt)
	if err != nil {
		return resourceTrackers, err
	}

	for _, volume := range volumes {
		resourceTracker := &resources.Resource{
			Name: volume.Name,
			ID:   volume.ID,
			Type: typeVolume,
			Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
				return cloud.(openstack.OpenstackCloud).DeleteVolume(r.ID)
			},
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}
