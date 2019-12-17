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
	"errors"

	"github.com/digitalocean/godo"

	"k8s.io/klog"
	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	_ "k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=Droplet
// Droplet represents a group of droplets. In the future it
// will be managed by the Machines API
type Droplet struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Region   *string
	Size     *string
	Image    *string
	SSHKey   *string
	Tags     []string
	Count    int
	UserData *fi.ResourceHolder
}

var _ fi.CompareWithID = &Droplet{}

func (d *Droplet) CompareWithID() *string {
	return d.Name
}

func (d *Droplet) Find(c *fi.Context) (*Droplet, error) {
	cloud := c.Cloud.(*digitalocean.Cloud)

	droplets, err := listDroplets(cloud)
	if err != nil {
		return nil, err
	}

	found := false
	count := 0
	var foundDroplet godo.Droplet
	for _, droplet := range droplets {
		if droplet.Name == fi.StringValue(d.Name) {
			found = true
			count++
			foundDroplet = droplet
		}
	}

	if !found {
		return nil, nil
	}

	return &Droplet{
		Name:      fi.String(foundDroplet.Name),
		Count:     count,
		Region:    fi.String(foundDroplet.Region.Slug),
		Size:      fi.String(foundDroplet.Size.Slug),
		Image:     fi.String(foundDroplet.Image.Slug),
		Tags:      foundDroplet.Tags,
		SSHKey:    d.SSHKey,   // TODO: get from droplet or ignore change
		UserData:  d.UserData, // TODO: get from droplet or ignore change
		Lifecycle: d.Lifecycle,
	}, nil
}

func listDroplets(cloud *digitalocean.Cloud) ([]godo.Droplet, error) {
	allDroplets := []godo.Droplet{}

	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := cloud.Droplets().List(context.TODO(), opt)
		if err != nil {
			return nil, err
		}

		allDroplets = append(allDroplets, droplets...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allDroplets, nil
}

func (d *Droplet) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(d, c)
}

func (_ *Droplet) RenderDO(t *do.DOAPITarget, a, e, changes *Droplet) error {
	userData, err := e.UserData.AsString()
	if err != nil {
		return err
	}

	var newDropletCount int
	if a == nil {
		newDropletCount = e.Count
	} else {

		expectedCount := e.Count
		actualCount := a.Count

		if expectedCount == actualCount {
			return nil
		}

		if actualCount > expectedCount {
			return errors.New("deleting droplets is not supported yet")
		}

		newDropletCount = expectedCount - actualCount
	}

	for i := 0; i < newDropletCount; i++ {
		_, _, err = t.Cloud.Droplets().Create(context.TODO(), &godo.DropletCreateRequest{
			Name:              fi.StringValue(e.Name),
			Region:            fi.StringValue(e.Region),
			Size:              fi.StringValue(e.Size),
			Image:             godo.DropletCreateImage{Slug: fi.StringValue(e.Image)},
			PrivateNetworking: true,
			Tags:              e.Tags,
			UserData:          userData,
			SSHKeys:           []godo.DropletCreateSSHKey{{Fingerprint: fi.StringValue(e.SSHKey)}},
		})

		if err != nil {
			klog.Errorf("Error creating droplet with Name=%s", fi.StringValue(e.Name))
			return err
		}
	}

	return err
}

func (_ *Droplet) CheckChanges(a, e, changes *Droplet) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
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
