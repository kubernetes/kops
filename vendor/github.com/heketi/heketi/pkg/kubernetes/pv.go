//
// Copyright (c) 2016 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package kubernetes

import (
	"fmt"

	"github.com/heketi/heketi/pkg/glusterfs/api"

	"k8s.io/kubernetes/pkg/api/resource"
	kubeapi "k8s.io/kubernetes/pkg/api/v1"
)

func VolumeToPv(volume *api.VolumeInfoResponse,
	name, endpoint string) *kubeapi.PersistentVolume {
	// Initialize object
	pv := &kubeapi.PersistentVolume{}
	pv.Kind = "PersistentVolume"
	pv.APIVersion = "v1"
	pv.Spec.PersistentVolumeReclaimPolicy = kubeapi.PersistentVolumeReclaimRetain
	pv.Spec.AccessModes = []kubeapi.PersistentVolumeAccessMode{
		kubeapi.ReadWriteMany,
	}
	pv.Spec.Capacity = make(kubeapi.ResourceList)
	pv.Spec.Glusterfs = &kubeapi.GlusterfsVolumeSource{}

	// Set path
	pv.Spec.Capacity[kubeapi.ResourceStorage] =
		resource.MustParse(fmt.Sprintf("%vGi", volume.Size))
	pv.Spec.Glusterfs.Path = volume.Name

	// Set name
	if name == "" {
		pv.ObjectMeta.Name = "glusterfs-" + volume.Id[:8]
	} else {
		pv.ObjectMeta.Name = name

	}

	// Set endpoint
	if endpoint == "" {
		pv.Spec.Glusterfs.EndpointsName = "TYPE ENDPOINT HERE"
	} else {
		pv.Spec.Glusterfs.EndpointsName = endpoint
	}

	return pv
}
