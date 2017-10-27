/*
Copyright 2016 The Kubernetes Authors.

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
	"strconv"

	"github.com/digitalocean/godo"

	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	_ "k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Droplet
type Droplet struct {
	Name      *string
	ID        *string
	Lifecycle *fi.Lifecycle

	Region   *string
	Size     *string
	Image    *string
	SSHKey   *string
	Tags     []string
	UserData *fi.ResourceHolder
}

var _ fi.CompareWithID = &Droplet{}

func (d *Droplet) CompareWithID() *string {
	return d.ID
}

func (d *Droplet) Find(c *fi.Context) (*Droplet, error) {
	cloud := c.Cloud.(*digitalocean.Cloud)
	dropletService := cloud.Droplets()

	droplets, _, err := dropletService.List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, droplet := range droplets {
		if droplet.Name == fi.StringValue(d.Name) {
			return &Droplet{
				Name:      fi.String(droplet.Name),
				ID:        fi.String(strconv.Itoa(droplet.ID)),
				Region:    fi.String(droplet.Region.Slug),
				Size:      fi.String(droplet.Size.Slug),
				Image:     fi.String(droplet.Image.Slug),
				Lifecycle: d.Lifecycle,
			}, nil
		}
	}

	return nil, nil
}

func (d *Droplet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(d, c)
}

func (_ *Droplet) RenderDO(t *do.DOAPITarget, a, e, changes *Droplet) error {
	if a != nil {
		return nil
	}

	userData, err := e.UserData.AsString()
	if err != nil {
		return err
	}

	dropletService := t.Cloud.Droplets()
	_, _, err = dropletService.Create(context.TODO(), &godo.DropletCreateRequest{
		Name:              fi.StringValue(e.Name),
		Region:            fi.StringValue(e.Region),
		Size:              fi.StringValue(e.Size),
		Image:             godo.DropletCreateImage{Slug: fi.StringValue(e.Image)},
		PrivateNetworking: true,
		Tags:              e.Tags,
		UserData:          userData,
		SSHKeys:           []godo.DropletCreateSSHKey{{Fingerprint: fi.StringValue(e.SSHKey)}},
	})
	return err
}

func (_ *Droplet) CheckChanges(a, e, changes *Droplet) error {
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
		if changes.Size != nil {
			return fi.CannotChangeField("Size")
		}
		if changes.Image != nil {
			return fi.CannotChangeField("Image")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
		if e.Size == nil {
			return fi.RequiredField("Size")
		}
		if e.Image == nil {
			return fi.RequiredField("Image")
		}
	}
	return nil
}
