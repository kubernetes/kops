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

package hetznertasks

import (
	"context"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type Volume struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID       *int
	Location string
	Size     int

	Labels map[string]string
}

var _ fi.CompareWithID = &Volume{}

func (v *Volume) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

func (v *Volume) Find(c *fi.Context) (*Volume, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.VolumeClient()

	volumes, err := client.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		if volume.Name == fi.StringValue(v.Name) {
			matches := &Volume{
				Lifecycle: v.Lifecycle,
				Name:      fi.String(volume.Name),
				ID:        fi.Int(volume.ID),
				Size:      volume.Size,
				Labels:    volume.Labels,
			}

			if volume.Location != nil {
				matches.Location = volume.Location.Name
			}

			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *Volume) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Volume) CheckChanges(a, e, changes *Volume) error {
	if a != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Size != 0 {
			return fi.CannotChangeField("Size")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Size == 0 {
			return fi.RequiredField("Size")
		}
	}
	return nil
}

func (_ *Volume) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *Volume) error {
	client := t.Cloud.VolumeClient()

	if a == nil {
		opts := hcloud.VolumeCreateOpts{
			Name: fi.StringValue(e.Name),
			Location: &hcloud.Location{
				Name: e.Location,
			},
			Size:   e.Size,
			Labels: e.Labels,
		}
		_, _, err := client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}

	} else {
		volume, _, err := client.Get(context.TODO(), strconv.Itoa(fi.IntValue(a.ID)))
		if err != nil {
			return err
		}

		// Update the labels
		if changes.Name != nil || len(changes.Labels) != 0 {
			_, _, err := client.Update(context.TODO(), volume, hcloud.VolumeUpdateOpts{
				Name:   fi.StringValue(e.Name),
				Labels: e.Labels,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type terraformVolume struct {
	Name     *string           `cty:"name"`
	Size     *int              `cty:"size"`
	Location *string           `cty:"location"`
	Labels   map[string]string `cty:"labels"`
}

func (_ *Volume) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Volume) error {
	tf := &terraformVolume{
		Name:     e.Name,
		Size:     fi.Int(e.Size),
		Location: fi.String(e.Location),
		Labels:   e.Labels,
	}

	return t.RenderResource("hcloud_volume", *e.Name, tf)
}
