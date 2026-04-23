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
	"regexp"
	"slices"
	"strconv"

	"github.com/linode/linodego"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const maxLinodeVolumeLabelLength = 32

// +kops:fitask
type Volume struct {
	Name      *string
	ID        *int
	Lifecycle fi.Lifecycle

	Region *string
	SizeGB *int64
	Tags   []string
}

var _ fi.CloudupTask = &Volume{}
var _ fi.CompareWithID = &Volume{}

var invalidVolumeLabelChars = regexp.MustCompile(`[^a-z0-9_-]+`)

func (v *Volume) CompareWithID() *string {
	if v.ID == nil {
		return nil
	}
	id := strconv.Itoa(fi.ValueOf(v.ID))
	return fi.PtrTo(id)
}

func (v *Volume) Find(c *fi.CloudupContext) (*Volume, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	volumes, err := cloud.Client().ListVolumes(c.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) volumes: %w", err)
	}

	label := normalizedVolumeLabel(fi.ValueOf(v.Name))
	var found *linodego.Volume
	for i := range volumes {
		volume := &volumes[i]
		if volume.Label != label {
			continue
		}
		if !hasAllTags(volume.Tags, v.Tags) {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("found multiple Linode (Akamai) volumes named %q with matching tags", label)
		}
		found = volume
	}

	if found == nil {
		return nil, nil
	}

	actual := &Volume{
		// Preserve desired task identity to avoid a synthetic Name change when
		// the cloud label is normalized from the desired dotted etcd name.
		Name:      v.Name,
		ID:        fi.PtrTo(found.ID),
		Lifecycle: v.Lifecycle,
		Region:    fi.PtrTo(found.Region),
		SizeGB:    fi.PtrTo(int64(found.Size)),
		Tags:      slices.Clone(found.Tags),
	}
	v.ID = actual.ID

	return actual, nil
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
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
		if changes.SizeGB != nil {
			return fi.CannotChangeField("SizeGB")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Region == nil {
			return fi.RequiredField("Region")
		}
		if e.SizeGB == nil {
			return fi.RequiredField("SizeGB")
		}
		if fi.ValueOf(e.SizeGB) < 10 {
			return fmt.Errorf("Linode (Akamai) volume size must be at least 10GiB")
		}
	}

	return nil
}

func (_ *Volume) RenderLinode(t *linode.APITarget, a, e, changes *Volume) error {
	if a != nil {
		// We currently only support create-once semantics for etcd volumes.
		return nil
	}

	_, err := t.Cloud.Client().CreateVolume(context.Background(), linodego.VolumeCreateOptions{
		Label:  normalizedVolumeLabel(fi.ValueOf(e.Name)),
		Region: fi.ValueOf(e.Region),
		Size:   int(fi.ValueOf(e.SizeGB)),
		Tags:   slices.Clone(e.Tags),
	})
	if err != nil {
		return fmt.Errorf("error creating Linode (Akamai) volume %q: %w", fi.ValueOf(e.Name), err)
	}

	return nil
}

func normalizedVolumeLabel(name string) string {
	clean := sanitizeVolumeLabelPart(name)
	if clean == "" {
		clean = "kops-etcd"
	}

	if len(clean) > maxLinodeVolumeLabelLength {
		clean = clean[:maxLinodeVolumeLabelLength]
	}
	clean = sanitizeVolumeLabelPart(clean)
	if clean == "" {
		clean = "kops-etcd"
	}

	return clean
}

func sanitizeVolumeLabelPart(s string) string {
	return sanitizeLabel(s, invalidVolumeLabelChars, "-_")
}
