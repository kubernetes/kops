/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaytasks

import (
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type Volume struct {
	Name      *string
	ID        *string
	Lifecycle fi.Lifecycle

	Size *int64
	Zone *string
	Tags []string
	Type *string
}

var _ fi.CompareWithID = &Volume{}

func (v *Volume) CompareWithID() *string {
	return v.ID
}

func (v *Volume) Find(c *fi.CloudupContext) (*Volume, error) {
	cloud := c.Cloud.(scaleway.ScwCloud)
	instanceService := cloud.InstanceService()
	zone := cloud.Zone()

	volumes, err := instanceService.ListVolumes(&instance.ListVolumesRequest{
		Name: v.Name,
		Zone: scw.Zone(zone),
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes.Volumes {
		if volume.Name == fi.ValueOf(v.Name) {
			return &Volume{
				Name:      fi.PtrTo(volume.Name),
				ID:        fi.PtrTo(volume.ID),
				Lifecycle: v.Lifecycle,
				Size:      fi.PtrTo(int64(volume.Size)),
				Zone:      fi.PtrTo(string(volume.Zone)),
				Type:      fi.PtrTo(string(volume.VolumeType)),
			}, nil
		}
	}

	return nil, nil
}

func (v *Volume) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (_ *Volume) CheckChanges(a, e, changes *Volume) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Size == nil {
			return fi.RequiredField("Size")
		}
		if e.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *Volume) RenderScw(t *scaleway.ScwAPITarget, a, e, changes *Volume) error {
	if a != nil {
		klog.Infof("Scaleway does not support changes to volumes for the moment")
		return nil
	}

	instanceService := t.Cloud.InstanceService()
	_, err := instanceService.CreateVolume(&instance.CreateVolumeRequest{
		Zone:       scw.Zone(fi.ValueOf(e.Zone)),
		Name:       fi.ValueOf(e.Name),
		VolumeType: instance.VolumeVolumeType(fi.ValueOf(e.Type)),
		Size:       scw.SizePtr(scw.Size(fi.ValueOf(e.Size))),
		Tags:       e.Tags,
	})

	return err
}
