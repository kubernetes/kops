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

package yandextasks

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
)

// Disk represents a Yandex.Cloud PD
// +kops:fitask
type Disk struct {
	Name      *string
	Lifecycle fi.Lifecycle

	FolderId *string
	DiskId   *string
	TypeId   *string
	Size     *int64
	ZoneId   *string
	Labels   map[string]string
}

var _ fi.CompareWithID = &Disk{}

func (e *Disk) CompareWithID() *string {
	return e.DiskId
}

func (e *Disk) Find(c *fi.Context) (*Disk, error) {
	sdk := c.Cloud.(yandex.YandexCloud).SDK()
	name := strings.ReplaceAll(*e.Name, ".", "--")
	filter := fmt.Sprintf("name=\"%s\"", name) // only filter by name supported atm 08.2022
	r, err := sdk.Compute().Disk().List(context.TODO(), &compute.ListDisksRequest{
		FolderId: *e.FolderId,
		Filter:   filter,
		PageSize: 100,
	})

	if err != nil {
		return nil, err
	}
	for _, disk := range r.Disks {
		if disk.Name != name {
			continue
		}
		// TODO(YuraBeznos): find a better naming for disks
		name := strings.ReplaceAll(disk.Name, "--", ".")
		actual := &Disk{
			FolderId:  &disk.FolderId,
			TypeId:    &disk.TypeId,
			Size:      &disk.Size,
			Name:      &name,
			Lifecycle: e.Lifecycle,
			Labels:    disk.Labels,
			ZoneId:    &disk.ZoneId,
		}
		return actual, nil
	}
	return nil, nil
}

func (e *Disk) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Disk) CheckChanges(a, e, changes *Disk) error {
	if a != nil {
		if changes.Size != nil {
			return fi.CannotChangeField("Size")
		}
		if changes.ZoneId != nil {
			return fi.CannotChangeField("Zone")
		}
		if changes.TypeId != nil {
			return fi.CannotChangeField("TypeId")
		}
	} else {
		if e.ZoneId == nil {
			return fi.RequiredField("ZoneId")
		}
	}
	return nil
}

func (_ *Disk) RenderYandex(t *yandex.YandexAPITarget, a, e, changes *Disk) error {
	sdk := t.Cloud.SDK()

	name := strings.ReplaceAll(*e.Name, ".", "--")
	disk := &compute.CreateDiskRequest{
		FolderId: *e.FolderId,
		Name:     name,
		Size:     *e.Size,
		TypeId:   *e.TypeId,
		Labels:   e.Labels,
		ZoneId:   *e.ZoneId,
	}

	if a == nil {
		if _, err := sdk.Compute().Disk().Create(context.TODO(), disk); err != nil {
			return fmt.Errorf("error creating Disk: %v", err)
		}
	}

	if changes.Labels != nil {
		// TODO(YuraBeznos): disk labels update case implementation if supported by Yandex
	}

	if a != nil && changes != nil {
		empty := &Disk{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to Disk: %v", changes)
		}
	}

	return nil
}
