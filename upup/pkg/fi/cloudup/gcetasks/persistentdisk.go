package gcetasks

import (
	"fmt"

	"google.golang.org/api/compute/v1"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/gce"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/terraform"
	"strings"
)

//go:generate fitask -type=PersistentDisk
type PersistentDisk struct {
	Name       *string
	VolumeType *string
	SizeGB     *int64
	Zone       *string
}

var _ fi.CompareWithID = &PersistentDisk{}

func (e *PersistentDisk) CompareWithID() *string {
	return e.Name
}

// Returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func lastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
}

func (e *PersistentDisk) Find(c *fi.Context) (*PersistentDisk, error) {
	cloud := c.Cloud.(*gce.GCECloud)

	r, err := cloud.Compute.Disks.Get(cloud.Project, *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing PersistentDisks: %v", err)
	}

	actual := &PersistentDisk{}
	actual.Name = &r.Name
	actual.VolumeType = fi.String(lastComponent(r.Type))
	actual.Zone = fi.String(lastComponent(r.Zone))
	actual.SizeGB = &r.SizeGb

	return actual, nil
}

func (e *PersistentDisk) URL(project string) string {
	u := &gce.GoogleCloudURL{
		Project: project,
		Zone:    *e.Zone,
		Type:    "disks",
		Name:    *e.Name,
	}
	return u.BuildURL()
}

func (e *PersistentDisk) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *PersistentDisk) CheckChanges(a, e, changes *PersistentDisk) error {
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

func (_ *PersistentDisk) RenderGCE(t *gce.GCEAPITarget, a, e, changes *PersistentDisk) error {
	typeURL := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/diskTypes/%s",
		t.Cloud.Project,
		*e.Zone,
		*e.VolumeType)

	disk := &compute.Disk{
		Name:   *e.Name,
		SizeGb: *e.SizeGB,
		Type:   typeURL,
	}

	if a == nil {
		_, err := t.Cloud.Compute.Disks.Insert(t.Cloud.Project, *e.Zone, disk).Do()
		if err != nil {
			return fmt.Errorf("error creating PersistentDisk: %v", err)
		}
	} else {
		return fmt.Errorf("Cannot apply changes to PersistentDisk: %v", changes)
	}

	return nil
}

type terraformDisk struct {
	Name       *string `json:"name"`
	VolumeType *string `json:"type"`
	SizeGB     *int64  `json:"size"`
	Zone       *string `json:"zone"`
}

func (_ *PersistentDisk) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *PersistentDisk) error {
	tf := &terraformDisk{
		Name:       e.Name,
		VolumeType: e.VolumeType,
		SizeGB:     e.SizeGB,
		Zone:       e.Zone,
	}
	return t.RenderResource("google_compute_disk", *e.Name, tf)
}
