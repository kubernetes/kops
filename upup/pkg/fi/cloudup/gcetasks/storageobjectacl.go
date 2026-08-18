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
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// StorageObjectAcl represents an ACL rule on a google cloud storage object
// +kops:fitask
type StorageObjectAcl struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Bucket *string
	Object *string
	Entity *string

	Role *string
}

var _ fi.CompareWithID = (*StorageObjectAcl)(nil)

func (e *StorageObjectAcl) CompareWithID() *string {
	return e.Name
}

func (e *StorageObjectAcl) Find(c *fi.CloudupContext) (*StorageObjectAcl, error) {
	cloud := c.T.Cloud.(gce.GCECloud)

	bucket := fi.ValueOf(e.Bucket)
	object := fi.ValueOf(e.Object)
	entity := fi.ValueOf(e.Entity)

	klog.V(2).Infof("Checking GCS object ACL for gs://%s/%s for %s", bucket, object, entity)
	rules, err := cloud.Storage().Bucket(bucket).Object(object).ACL().List(context.TODO())
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying GCS object ACL for gs://%s/%s for %s: %v", bucket, object, entity, err)
	}

	for _, r := range rules {
		if string(r.Entity) != entity {
			continue
		}

		foundEntity := string(r.Entity)
		foundRole := string(r.Role)

		actual := &StorageObjectAcl{}
		actual.Name = e.Name
		actual.Bucket = e.Bucket
		actual.Object = e.Object
		actual.Entity = &foundEntity

		actual.Role = &foundRole

		// Ignore "system" fields
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *StorageObjectAcl) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *StorageObjectAcl) CheckChanges(a, e, changes *StorageObjectAcl) error {
	if fi.ValueOf(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.ValueOf(e.Object) == "" {
		return fi.RequiredField("Object")
	}
	if fi.ValueOf(e.Entity) == "" {
		return fi.RequiredField("Entity")
	}
	return nil
}

func (_ *StorageObjectAcl) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageObjectAcl) error {
	bucket := fi.ValueOf(e.Bucket)
	object := fi.ValueOf(e.Object)
	entity := fi.ValueOf(e.Entity)
	role := fi.ValueOf(e.Role)

	if a == nil {
		klog.V(2).Infof("Creating GCS object ACL for gs://%s/%s for %s as %s", bucket, object, entity, role)
	} else {
		klog.V(2).Infof("Updating GCS object ACL for gs://%s/%s for %s as %s", bucket, object, entity, role)
	}

	err := t.Cloud.Storage().Bucket(bucket).Object(object).ACL().Set(context.TODO(), storage.ACLEntity(entity), storage.ACLRole(role))
	if err != nil {
		return fmt.Errorf("error setting GCS object ACL for gs://%s/%s for %s as %s: %v", bucket, object, entity, role, err)
	}

	return nil
}

// terraformStorageObjectAcl is the model for a terraform google_storage_object_acl rule
type terraformStorageObjectAcl struct {
	Bucket     string   `cty:"bucket"`
	Object     string   `cty:"object"`
	RoleEntity []string `cty:"role_entity"`
}

func (_ *StorageObjectAcl) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageObjectAcl) error {
	var roleEntities []string
	roleEntities = append(roleEntities, fi.ValueOf(e.Role)+":"+fi.ValueOf(e.Name))
	tf := &terraformStorageObjectAcl{
		Bucket:     fi.ValueOf(e.Bucket),
		Object:     fi.ValueOf(e.Object),
		RoleEntity: roleEntities,
	}

	return t.RenderResource("google_storage_object_acl", *e.Name, tf)
}
