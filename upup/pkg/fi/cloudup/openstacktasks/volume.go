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

	cinderv2 "github.com/gophercloud/gophercloud/openstack/blockstorage/v2/volumes"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=Volume
type Volume struct {
	ID               *string
	Name             *string
	AvailabilityZone *string
	VolumeType       *string
	SizeGB           *int64
	Tags             map[string]string
	Lifecycle        *fi.Lifecycle
}

var _ fi.CompareWithID = &Volume{}

func (c *Volume) CompareWithID() *string {
	return c.ID
}

func (c *Volume) Find(context *fi.Context) (*Volume, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	opt := cinderv2.ListOpts{
		Name:     fi.StringValue(c.Name),
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
		return nil, fmt.Errorf("found multiple Volumes with name: %s", fi.StringValue(c.Name))
	}
	v := volumes[0]
	actual := &Volume{
		ID:               fi.String(v.ID),
		Name:             fi.String(v.Name),
		AvailabilityZone: fi.String(v.AvailabilityZone),
		VolumeType:       fi.String(v.VolumeType),
		SizeGB:           fi.Int64(int64(v.Size)),
		Tags:             v.Metadata,
		Lifecycle:        c.Lifecycle,
	}
	// remove tags "readonly" and "attached_mode", openstack are adding these and if not removed
	// kops will always try to update volumes
	if _, ok := actual.Tags["readonly"]; ok {
		delete(actual.Tags, "readonly")
	}
	if _, ok := actual.Tags["attached_mode"]; ok {
		delete(actual.Tags, "attached_mode")
	}
	c.ID = actual.ID
	c.AvailabilityZone = actual.AvailabilityZone
	return actual, nil
}

func (c *Volume) Run(context *fi.Context) error {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	for k, v := range cloud.GetCloudTags() {
		c.Tags[k] = v
	}

	return fi.DefaultDeltaRunMethod(c, context)
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
		klog.V(2).Infof("Creating PersistentVolume with Name:%q", fi.StringValue(e.Name))

		storageAZ, err := t.Cloud.GetStorageAZFromCompute(fi.StringValue(e.AvailabilityZone))
		if err != nil {
			return fmt.Errorf("Failed to get storage availability zone: %s", err)
		}

		opt := cinderv2.CreateOpts{
			Size:             int(*e.SizeGB),
			AvailabilityZone: storageAZ.ZoneName,
			Metadata:         e.Tags,
			Name:             fi.StringValue(e.Name),
			VolumeType:       fi.StringValue(e.VolumeType),
		}

		v, err := t.Cloud.CreateVolume(opt)
		if err != nil {
			return fmt.Errorf("error creating PersistentVolume: %v", err)
		}

		e.ID = fi.String(v.ID)
		e.AvailabilityZone = fi.String(v.AvailabilityZone)
		return nil
	}

	if changes != nil && changes.Tags != nil {
		klog.V(2).Infof("Update the tags on volume %q: %v, the differences are %v", fi.StringValue(e.ID), e.Tags, changes.Tags)

		err := t.Cloud.SetVolumeTags(fi.StringValue(e.ID), e.Tags)
		if err != nil {
			return fmt.Errorf("error updating the tags on volume %q: %v", fi.StringValue(e.ID), err)
		}
	}

	klog.V(2).Infof("Openstack task Volume::RenderOpenstack did nothing")
	return nil
}
