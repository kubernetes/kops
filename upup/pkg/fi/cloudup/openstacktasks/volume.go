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

package openstacktasks

import (
	"fmt"

	cinderv3 "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type Volume struct {
	ID               *string
	Name             *string
	AvailabilityZone *string
	VolumeType       *string
	SizeGB           *int64
	Tags             map[string]string
	Lifecycle        fi.Lifecycle
}

var _ fi.CompareWithID = &Volume{}
var _ fi.CloudupTaskNormalize = &Volume{}

func (c *Volume) CompareWithID() *string {
	return c.ID
}

func (c *Volume) Find(context *fi.CloudupContext) (*Volume, error) {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	opt := cinderv3.ListOpts{
		Name:     fi.ValueOf(c.Name),
		Metadata: c.Tags,
	}
	volumes, err := cloud.ListVolumes(opt)
	if err != nil {
		return nil, err
	}
	n := len(volumes)
	if n == 0 {
		return nil, nil
	} else if n != 1 {
		return nil, fmt.Errorf("found multiple Volumes with name: %s", fi.ValueOf(c.Name))
	}
	v := volumes[0]
	actual := &Volume{
		ID:               fi.PtrTo(v.ID),
		Name:             fi.PtrTo(v.Name),
		AvailabilityZone: fi.PtrTo(v.AvailabilityZone),
		VolumeType:       fi.PtrTo(v.VolumeType),
		SizeGB:           fi.PtrTo(int64(v.Size)),
		Tags:             v.Metadata,
		Lifecycle:        c.Lifecycle,
	}
	// remove tags "readonly" and "attached_mode", openstack are adding these and if not removed
	// kops will always try to update volumes
	delete(actual.Tags, "readonly")
	delete(actual.Tags, "attached_mode")
	c.ID = actual.ID
	c.AvailabilityZone = actual.AvailabilityZone
	return actual, nil
}

func (c *Volume) Normalize(context *fi.CloudupContext) error {
	cloud := context.T.Cloud.(openstack.OpenstackCloud)
	for k, v := range cloud.GetCloudTags() {
		c.Tags[k] = v
	}
	return nil
}

func (c *Volume) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(c, context)
}

func (_ *Volume) CheckChanges(a, e, changes *Volume) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.AvailabilityZone == nil {
			return fi.RequiredField("AvailabilityZone")
		}
		if e.VolumeType == nil {
			return fi.RequiredField("VolumeType")
		}
		if e.SizeGB == nil {
			return fi.RequiredField("SizeGB")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.AvailabilityZone != nil {
			return fi.CannotChangeField("AvailabilityZone")
		}
		if changes.VolumeType != nil {
			return fi.CannotChangeField("VolumeType")
		}
		if changes.SizeGB != nil {
			return fi.CannotChangeField("SizeGB")
		}
	}
	return nil
}

func (_ *Volume) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *Volume) error {
	if a == nil {
		klog.V(2).Infof("Creating PersistentVolume with Name:%q", fi.ValueOf(e.Name))

		storageAZ, err := t.Cloud.GetStorageAZFromCompute(fi.ValueOf(e.AvailabilityZone))
		if err != nil {
			return fmt.Errorf("Failed to get storage availability zone: %s", err)
		}

		opt := cinderv3.CreateOpts{
			Size:             int(*e.SizeGB),
			AvailabilityZone: storageAZ.ZoneName,
			Metadata:         e.Tags,
			Name:             fi.ValueOf(e.Name),
			VolumeType:       fi.ValueOf(e.VolumeType),
		}

		v, err := t.Cloud.CreateVolume(opt)
		if err != nil {
			return fmt.Errorf("error creating PersistentVolume: %v", err)
		}

		e.ID = fi.PtrTo(v.ID)
		e.AvailabilityZone = fi.PtrTo(v.AvailabilityZone)
		return nil
	}

	if changes != nil && changes.Tags != nil {
		klog.V(2).Infof("Update the tags on volume %q: %v, the differences are %v", fi.ValueOf(e.ID), e.Tags, changes.Tags)

		err := t.Cloud.SetVolumeTags(fi.ValueOf(e.ID), e.Tags)
		if err != nil {
			return fmt.Errorf("error updating the tags on volume %q: %v", fi.ValueOf(e.ID), err)
		}
	}

	klog.V(2).Infof("Openstack task Volume::RenderOpenstack did nothing")
	return nil
}
