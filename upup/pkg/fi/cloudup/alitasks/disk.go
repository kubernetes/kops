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

package alitasks

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"k8s.io/klog"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// Disk represents a ALI Cloud Disk
//go:generate fitask -type=Disk
const (
	DiskResource = "disk"
	DiskType     = ecs.DiskTypeAllData
)

type Disk struct {
	Lifecycle    *fi.Lifecycle
	Name         *string
	DiskId       *string
	ZoneId       *string
	DiskCategory *string
	Encrypted    *bool
	SizeGB       *int
	Tags         map[string]string
}

var _ fi.CompareWithID = &Disk{}

func (d *Disk) CompareWithID() *string {
	return d.DiskId
}

func (d *Disk) Find(c *fi.Context) (*Disk, error) {
	cloud := c.Cloud.(aliup.ALICloud)
	clusterTags := cloud.GetClusterTags()

	request := &ecs.DescribeDisksArgs{
		DiskType: DiskType,
		RegionId: common.Region(cloud.Region()),
		ZoneId:   fi.StringValue(d.ZoneId),
		Tag:      clusterTags,
		DiskName: fi.StringValue(d.Name),
	}

	responseDisks, _, err := cloud.EcsClient().DescribeDisks(request)
	if err != nil {
		return nil, fmt.Errorf("error finding Disks: %v", err)
	}
	// Don't exist disk with specified ClusterTags or Name.
	if len(responseDisks) == 0 {
		return nil, nil
	}
	if len(responseDisks) > 1 {
		klog.V(4).Infof("The number of specified disk with the same name and ClusterTags exceeds 1, diskName:%q", *d.Name)
	}

	klog.V(2).Infof("found matching Disk with name: %q", *d.Name)

	disk := responseDisks[0]
	actual := &Disk{
		Name:         fi.String(disk.DiskName),
		DiskCategory: fi.String(string(disk.Category)),
		ZoneId:       fi.String(disk.ZoneId),
		SizeGB:       fi.Int(disk.Size),
		DiskId:       fi.String(disk.DiskId),
		Encrypted:    fi.Bool(disk.Encrypted),
	}

	tags, err := cloud.GetTags(fi.StringValue(actual.DiskId), DiskResource)

	if err != nil {
		klog.V(4).Infof("Error getting tags on resourceId:%q", *actual.DiskId)
	}
	actual.Tags = tags

	// Ignore "system" fields
	actual.Lifecycle = d.Lifecycle
	d.DiskId = actual.DiskId
	return actual, nil
}

func (d *Disk) Run(c *fi.Context) error {
	if d.Tags == nil {
		d.Tags = make(map[string]string)
	}
	c.Cloud.(aliup.ALICloud).AddClusterTags(d.Tags)
	return fi.DefaultDeltaRunMethod(d, c)
}

func (_ *Disk) CheckChanges(a, e, changes *Disk) error {
	if a == nil {
		if e.ZoneId == nil {
			return fi.RequiredField("ZoneId")
		}
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.DiskCategory != nil {
			return fi.CannotChangeField("DiskCategory")
		}
	}
	return nil
}

//Disk can only modify tags.
func (_ *Disk) RenderALI(t *aliup.ALIAPITarget, a, e, changes *Disk) error {
	if a == nil {
		klog.V(2).Infof("Creating Disk with Name:%q", fi.StringValue(e.Name))

		request := &ecs.CreateDiskArgs{
			DiskName:     fi.StringValue(e.Name),
			RegionId:     common.Region(t.Cloud.Region()),
			ZoneId:       fi.StringValue(e.ZoneId),
			Encrypted:    fi.BoolValue(e.Encrypted),
			DiskCategory: ecs.DiskCategory(fi.StringValue(e.DiskCategory)),
			Size:         fi.IntValue(e.SizeGB),
		}
		diskId, err := t.Cloud.EcsClient().CreateDisk(request)
		if err != nil {
			return fmt.Errorf("error creating disk: %v", err)
		}
		e.DiskId = fi.String(diskId)
	}

	if changes != nil && changes.Tags != nil {
		klog.V(2).Infof("Modifying tags of disk with Name:%q", fi.StringValue(e.Name))
		if err := t.Cloud.CreateTags(*e.DiskId, DiskResource, e.Tags); err != nil {
			return fmt.Errorf("error adding Tags to ALI YunPan: %v", err)
		}
	}

	if a != nil && (len(a.Tags) > 0) {

		tagsToDelete := e.getDiskTagsToDelete(a.Tags)
		if len(tagsToDelete) > 0 {
			klog.V(2).Infof("Deleting tags of disk with Name:%q", fi.StringValue(e.Name))
			if err := t.Cloud.RemoveTags(*e.DiskId, DiskResource, tagsToDelete); err != nil {
				return fmt.Errorf("error removing Tags from ALI YunPan: %v", err)
			}
		}
	}

	return nil
}

// getDiskTagsToDelete loops through the currently set tags and builds a list of tags to be deleted from the specified disk
func (d *Disk) getDiskTagsToDelete(currentTags map[string]string) map[string]string {
	tagsToDelete := map[string]string{}
	for k, v := range currentTags {
		if _, ok := d.Tags[k]; !ok {
			tagsToDelete[k] = v
		}
	}

	return tagsToDelete
}

type terraformDiskTag struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}

type terraformDisk struct {
	DiskName     *string             `json:"name,omitempty"`
	DiskCategory *string             `json:"category,omitempty"`
	SizeGB       *int                `json:"size,omitempty"`
	Zone         *string             `json:"availability_zone,omitempty"`
	Tags         []*terraformDiskTag `json:"tags,omitempty"`
}

func (_ *Disk) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Disk) error {
	tf := &terraformDisk{
		DiskName:     e.Name,
		DiskCategory: e.DiskCategory,
		SizeGB:       e.SizeGB,
		Zone:         e.ZoneId,
	}

	for key, value := range e.Tags {
		tf.Tags = append(tf.Tags, &terraformDiskTag{
			Key:   &key,
			Value: &value,
		})
	}
	return t.RenderResource("alicloud_disk", *e.Name, tf)
}
