/*
Copyright 2017 The Kubernetes Authors.

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

	"google.golang.org/api/storage/v1"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// StorageObjectAcl represents an ACL rule on a google cloud storage object
//go:generate fitask -type=StorageObjectAcl
type StorageObjectAcl struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Bucket *string
	Object *string
	Entity *string

	Role *string
}

var _ fi.CompareWithID = &StorageObjectAcl{}

func (e *StorageObjectAcl) CompareWithID() *string {
	return e.Name
}

func (e *StorageObjectAcl) Find(c *fi.Context) (*StorageObjectAcl, error) {
	cloud := c.Cloud.(gce.GCECloud)

	bucket := fi.StringValue(e.Bucket)
	object := fi.StringValue(e.Object)
	entity := fi.StringValue(e.Entity)

	klog.V(2).Infof("Checking GCS object ACL for gs://%s/%s for %s", bucket, object, entity)
	r, err := cloud.Storage().ObjectAccessControls.Get(bucket, object, entity).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying GCS object ACL for gs://%s/%s for %s: %v", bucket, object, entity, err)
	}

	actual := &StorageObjectAcl{}
	actual.Name = e.Name
	actual.Bucket = &r.Bucket
	actual.Object = &r.Object
	actual.Entity = &r.Entity

	actual.Role = &r.Role

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *StorageObjectAcl) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *StorageObjectAcl) CheckChanges(a, e, changes *StorageObjectAcl) error {
	if fi.StringValue(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.StringValue(e.Object) == "" {
		return fi.RequiredField("Object")
	}
	if fi.StringValue(e.Entity) == "" {
		return fi.RequiredField("Entity")
	}
	return nil
}

func (_ *StorageObjectAcl) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageObjectAcl) error {
	bucket := fi.StringValue(e.Bucket)
	object := fi.StringValue(e.Object)
	entity := fi.StringValue(e.Entity)
	role := fi.StringValue(e.Role)

	acl := &storage.ObjectAccessControl{
		Entity: entity,
		Role:   role,
	}

	if a == nil {
		klog.V(2).Infof("Creating GCS object ACL for gs://%s/%s for %s as %s", bucket, object, entity, role)

		_, err := t.Cloud.Storage().ObjectAccessControls.Insert(bucket, object, acl).Do()
		if err != nil {
			return fmt.Errorf("error creating GCS object ACL for gs://%s/%s for %s as %s: %v", bucket, object, entity, role, err)
		}
	} else {
		klog.V(2).Infof("Updating GCS object ACL for gs://%s/%s for %s as %s", bucket, object, entity, role)

		_, err := t.Cloud.Storage().ObjectAccessControls.Update(bucket, object, entity, acl).Do()
		if err != nil {
			return fmt.Errorf("error updating GCS object ACL for gs://%s/%s for %s as %s: %v", bucket, object, entity, role, err)
		}
	}

	return nil
}

// terraformStorageObjectAcl is the model for a terraform google_storage_object_acl rule
type terraformStorageObjectAcl struct {
	Bucket     string   `json:"bucket,omitempty"`
	Object     string   `json:"object,omitempty"`
	RoleEntity []string `json:"role_entity,omitempty"`
}

func (_ *StorageObjectAcl) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageObjectAcl) error {
	var roleEntities []string
	roleEntities = append(roleEntities, fi.StringValue(e.Role)+":"+fi.StringValue(e.Name))
	tf := &terraformStorageObjectAcl{
		Bucket:     fi.StringValue(e.Bucket),
		Object:     fi.StringValue(e.Object),
		RoleEntity: roleEntities,
	}

	return t.RenderResource("google_storage_object_acl", *e.Name, tf)
}
