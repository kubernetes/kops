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

// StorageBucketAcl represents an ACL rule on a google cloud storage bucket
//go:generate fitask -type=StorageBucketAcl
type StorageBucketAcl struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Bucket *string
	Entity *string

	Role *string
}

var _ fi.CompareWithID = &StorageBucketAcl{}

func (e *StorageBucketAcl) CompareWithID() *string {
	return e.Name
}

func (e *StorageBucketAcl) Find(c *fi.Context) (*StorageBucketAcl, error) {
	cloud := c.Cloud.(gce.GCECloud)

	bucket := fi.StringValue(e.Bucket)
	entity := fi.StringValue(e.Entity)

	klog.V(2).Infof("Checking GCS bucket ACL for gs://%s for %s", bucket, entity)
	r, err := cloud.Storage().BucketAccessControls.Get(bucket, entity).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking GCS bucket ACL for gs://%s for %s: %v", bucket, entity, err)
	}

	actual := &StorageBucketAcl{}
	actual.Name = e.Name
	actual.Bucket = &r.Bucket
	actual.Entity = &r.Entity

	actual.Role = &r.Role

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *StorageBucketAcl) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *StorageBucketAcl) CheckChanges(a, e, changes *StorageBucketAcl) error {
	if fi.StringValue(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.StringValue(e.Entity) == "" {
		return fi.RequiredField("Entity")
	}
	return nil
}

func (_ *StorageBucketAcl) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageBucketAcl) error {
	bucket := fi.StringValue(e.Bucket)
	entity := fi.StringValue(e.Entity)
	role := fi.StringValue(e.Role)

	acl := &storage.BucketAccessControl{
		Entity: entity,
		Role:   role,
	}

	if a == nil {
		klog.V(2).Infof("Creating GCS bucket ACL for gs://%s for %s as %s", bucket, entity, role)

		_, err := t.Cloud.Storage().BucketAccessControls.Insert(bucket, acl).Do()
		if err != nil {
			return fmt.Errorf("error creating GCS bucket ACL for gs://%s for %s as %s: %v", bucket, entity, role, err)
		}
	} else {
		klog.V(2).Infof("Updating GCS bucket ACL for gs://%s for %s as %s", bucket, entity, role)

		_, err := t.Cloud.Storage().BucketAccessControls.Update(bucket, entity, acl).Do()
		if err != nil {
			return fmt.Errorf("error updating GCS bucket ACL for gs://%s for %s as %s: %v", bucket, entity, role, err)
		}
	}

	return nil
}

// terraformStorageBucketAcl is the model for a terraform google_storage_bucket_acl rule
type terraformStorageBucketAcl struct {
	Bucket     string   `json:"bucket,omitempty" cty:"bucket"`
	RoleEntity []string `json:"role_entity,omitempty" cty:"role_entity"`
}

func (_ *StorageBucketAcl) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageBucketAcl) error {
	var roleEntities []string
	roleEntities = append(roleEntities, fi.StringValue(e.Role)+":"+fi.StringValue(e.Entity))
	tf := &terraformStorageBucketAcl{
		Bucket:     fi.StringValue(e.Bucket),
		RoleEntity: roleEntities,
	}

	return t.RenderResource("google_storage_bucket_acl", *e.Name, tf)
}
