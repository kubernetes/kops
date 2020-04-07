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

package gcetasks

import (
	"fmt"
	"reflect"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// Disk represents a GCE PD
//go:generate fitask -type=Disk
type Disk struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	VolumeType *string
	SizeGB     *int64
	Zone       *string
	Labels     map[string]string
}

var _ fi.CompareWithID = &Disk{}

func (e *Disk) CompareWithID() *string {
	return e.Name
}

func (e *Disk) Find(c *fi.Context) (*Disk, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Disks.Get(cloud.Project(), *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing Disks: %v", err)
	}

	actual := &Disk{}
	actual.Name = &r.Name
	actual.VolumeType = fi.String(gce.LastComponent(r.Type))
	actual.Zone = fi.String(gce.LastComponent(r.Zone))
	actual.SizeGB = &r.SizeGb

	actual.Labels = r.Labels

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Disk) URL(project string) string {
	u := &gce.GoogleCloudURL{
		Project: project,
		Zone:    *e.Zone,
		Type:    "disks",
		Name:    *e.Name,
	}
	return u.BuildURL()
}

func (e *Disk) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Disk) CheckChanges(a, e, changes *Disk) error {
	if a != nil {
		if changes.SizeGB != nil {
			return fi.CannotChangeField("SizeGB")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
		if changes.VolumeType != nil {
			return fi.CannotChangeField("VolumeType")
		}
	} else {
		if e.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *Disk) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Disk) error {
	cloud := t.Cloud
	typeURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/diskTypes/%s",
		cloud.Project(),
		*e.Zone,
		*e.VolumeType)

	disk := &compute.Disk{
		Name:   *e.Name,
		SizeGb: *e.SizeGB,
		Type:   typeURL,
	}

	if a == nil {
		_, err := cloud.Compute().Disks.Insert(t.Cloud.Project(), *e.Zone, disk).Do()
		if err != nil {
			return fmt.Errorf("error creating Disk: %v", err)
		}
	}

	if changes.Labels != nil {
		d, err := cloud.Compute().Disks.Get(t.Cloud.Project(), *e.Zone, disk.Name).Do()
		if err != nil {
			return fmt.Errorf("error reading created Disk: %v", err)
		}

		labelsRequest := &compute.ZoneSetLabelsRequest{
			LabelFingerprint: d.LabelFingerprint,
			Labels:           make(map[string]string),
		}
		// Danger: labels replace tags on instances; but thankfully volumes don't have tags
		//for _, k := range d.Tags {
		//	labelsRequest.Labels[k] = ""
		//}
		for k, v := range d.Labels {
			labelsRequest.Labels[k] = v
		}
		for k, v := range t.Cloud.Labels() {
			labelsRequest.Labels[k] = v
		}
		for k, v := range e.Labels {
			labelsRequest.Labels[k] = v
		}
		klog.V(2).Infof("Setting labels on disk %q: %v", disk.Name, labelsRequest.Labels)
		_, err = t.Cloud.Compute().Disks.SetLabels(t.Cloud.Project(), *e.Zone, disk.Name, labelsRequest).Do()
		if err != nil {
			return fmt.Errorf("error setting labels on created Disk: %v", err)
		}
		changes.Labels = nil
	}

	if a != nil && changes != nil {
		empty := &Disk{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to Disk: %v", changes)
		}
	}

	return nil
}

type terraformDisk struct {
	Name       *string           `json:"name" cty:"name"`
	VolumeType *string           `json:"type" cty:"type"`
	SizeGB     *int64            `json:"size" cty:"size"`
	Zone       *string           `json:"zone" cty:"zone"`
	Labels     map[string]string `json:"labels,omitempty" cty:"labels"`
}

func (_ *Disk) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Disk) error {
	cloud := t.Cloud.(gce.GCECloud)

	labels := make(map[string]string)
	for k, v := range cloud.Labels() {
		labels[k] = v
	}
	for k, v := range e.Labels {
		labels[k] = v
	}

	tf := &terraformDisk{
		Name:       e.Name,
		VolumeType: e.VolumeType,
		SizeGB:     e.SizeGB,
		Zone:       e.Zone,
		Labels:     labels,
	}
	return t.RenderResource("google_compute_disk", *e.Name, tf)
}
