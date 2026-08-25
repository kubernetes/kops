/*
Copyright 2026 The Kubernetes Authors.

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

package linodetasks

import (
	"context"
	"fmt"

	"github.com/linode/linodego/v2"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type Volume struct {
	Name      *string
	ID        *int
	Lifecycle fi.Lifecycle
	Region    *string
	SizeGB    *int
	Tags      []string
}

var _ fi.CloudupTask = &Volume{}
var _ fi.CompareWithID = &Volume{}

func (v *Volume) CompareWithID() *string {
	return v.Name
}

func (v *Volume) Find(c *fi.CloudupContext) (*Volume, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)
	name := fi.ValueOf(v.Name)
	if name == "" {
		return nil, fmt.Errorf("Volume.Name is required")
	}
	label := truncate.TruncateString(linode.NormalizeLinodeLabel(name), truncate.TruncateStringOptions{MaxLength: 32})
	listOptions, err := linode.ListOptionsForLabel(label)
	if err != nil {
		return nil, err
	}

	volumes, err := cloud.Client().ListVolumes(c.Context(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) volumes: %w", err)
	}

	if len(volumes) == 0 {
		return nil, nil
	}

	// Name is unique, so we should only have one match
	matched := volumes[0]

	actual := &Volume{
		ID:        new(matched.ID),
		Name:      new(name),
		Lifecycle: v.Lifecycle,
		Region:    new(matched.Region),
		SizeGB:    new(matched.Size),
		Tags:      matched.Tags,
	}
	v.ID = actual.ID

	return actual, nil
}

func (v *Volume) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(v, c)
}

func (_ *Volume) CheckChanges(actual, expected, changes *Volume) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
		if changes.SizeGB != nil {
			if fi.ValueOf(expected.SizeGB) < fi.ValueOf(actual.SizeGB) {
				return fmt.Errorf("SizeGB cannot be decreased")
			}
		}
		if changes.Tags != nil {
			return fi.CannotChangeField("Tags")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
		if expected.SizeGB == nil {
			return fi.RequiredField("SizeGB")
		}
		if fi.ValueOf(expected.SizeGB) < 10 {
			return fmt.Errorf("SizeGB must be at least 10 GB")
		}
	}

	return nil
}

func (*Volume) RenderLinode(t *linode.APITarget, actual, expected, changes *Volume) error {
	if actual != nil {
		expected.ID = actual.ID
		if changes.SizeGB == nil {
			return nil
		}
		if err := t.Cloud.Client().ResizeVolume(context.Background(), fi.ValueOf(actual.ID), linodego.VolumeResizeOptions{Size: fi.ValueOf(expected.SizeGB)}); err != nil {
			return fmt.Errorf("error resizing Akamai (Linode) volume %q: %w", fi.ValueOf(actual.Name), err)
		}
		klog.V(2).Infof("Resized Akamai (Linode) volume %q (id=%d) to %d GB", fi.ValueOf(actual.Name), fi.ValueOf(actual.ID), fi.ValueOf(expected.SizeGB))
		return nil
	}

	name := fi.ValueOf(expected.Name)
	label := truncate.TruncateString(linode.NormalizeLinodeLabel(name), truncate.TruncateStringOptions{MaxLength: 32})
	created, err := t.Cloud.Client().CreateVolume(context.Background(), linodego.VolumeCreateOptions{
		Label:  label,
		Region: fi.ValueOf(expected.Region),
		Size:   fi.ValueOf(expected.SizeGB),
		Tags:   expected.Tags,
	})
	if err != nil {
		return fmt.Errorf("error creating Akamai (Linode) volume %q: %w", name, err)
	}

	expected.ID = new(created.ID)
	klog.V(2).Infof("Created Akamai (Linode) volume %q (id=%d)", created.Label, created.ID)

	return nil
}
