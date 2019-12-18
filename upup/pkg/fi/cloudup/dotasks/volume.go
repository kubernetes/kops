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

package dotasks

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"

	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Volume
type Volume struct {
	Name      *string
	ID        *string
	Lifecycle *fi.Lifecycle

	SizeGB *int64
	Region *string
	Tags   map[string]string
}

var _ fi.CompareWithID = &Volume{}

func (v *Volume) CompareWithID() *string {
	return v.ID
}

func (v *Volume) Find(c *fi.Context) (*Volume, error) {
	cloud := c.Cloud.(*digitalocean.Cloud)
	volService := cloud.Volumes()

	volumes, _, err := volService.ListVolumes(context.TODO(), &godo.ListVolumeParams{
		Region: cloud.Region(),
		Name:   fi.StringValue(v.Name),
	})
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		if volume.Name == fi.StringValue(v.Name) {
			return &Volume{
				Name:      fi.String(volume.Name),
				ID:        fi.String(volume.ID),
				Lifecycle: v.Lifecycle,
				SizeGB:    fi.Int64(volume.SizeGigaBytes),
				Region:    fi.String(volume.Region.Slug),
			}, nil
		}
	}

	// Volume = nil if not found
	return nil, nil
}

func (v *Volume) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Volume) CheckChanges(a, e, changes *Volume) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.SizeGB == nil {
			return fi.RequiredField("SizeGB")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
	}
	return nil
}

func (_ *Volume) RenderDO(t *do.DOAPITarget, a, e, changes *Volume) error {
	if a != nil {
		// in general, we shouldn't need to render changes to a volume
		// however there can be cases where we may want to resize or rename.
		// consider this in later stages of DO support on kops
		return nil
	}

	tagArray := []string{}

	for k, v := range e.Tags {
		// DO tags don't accept =. Separate the key and value with an ":"
		klog.V(10).Infof("DO - Join the volume tag - %s", fmt.Sprintf("%s:%s", k, v))
		tagArray = append(tagArray, fmt.Sprintf("%s:%s", k, v))
	}

	volService := t.Cloud.Volumes()
	_, _, err := volService.CreateVolume(context.TODO(), &godo.VolumeCreateRequest{
		Name:          fi.StringValue(e.Name),
		Region:        fi.StringValue(e.Region),
		SizeGigaBytes: fi.Int64Value(e.SizeGB),
		Tags:          tagArray,
	})

	return err
}

// terraformVolume represents the digitalocean_volume resource in terraform
// https://www.terraform.io/docs/providers/do/r/volume.html
type terraformVolume struct {
	Name   *string `json:"name"`
	SizeGB *int64  `json:"size"`
	Region *string `json:"region"`
}

func (_ *Volume) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Volume) error {
	tf := &terraformVolume{
		Name:   e.Name,
		SizeGB: e.SizeGB,
		Region: e.Region,
	}
	return t.RenderResource("digitalocean_volume", *e.Name, tf)
}
