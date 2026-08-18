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

// StorageBucketAcl represents an ACL rule on a google cloud storage bucket
// +kops:fitask
type StorageBucketAcl struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Bucket *string
	Entity *string

	Role *string
}

var _ fi.CompareWithID = (*StorageBucketAcl)(nil)

func (e *StorageBucketAcl) CompareWithID() *string {
	return e.Name
}

func (e *StorageBucketAcl) Find(c *fi.CloudupContext) (*StorageBucketAcl, error) {
	cloud := c.T.Cloud.(gce.GCECloud)

	bucket := fi.ValueOf(e.Bucket)
	entity := fi.ValueOf(e.Entity)

	klog.V(2).Infof("Checking GCS bucket ACL for gs://%s for %s", bucket, entity)
	rules, err := cloud.Storage().Bucket(bucket).ACL().List(context.TODO())
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking GCS bucket ACL for gs://%s for %s: %v", bucket, entity, err)
	}

	for _, r := range rules {
		if string(r.Entity) != entity {
			continue
		}

		foundEntity := string(r.Entity)
		foundRole := string(r.Role)

		actual := &StorageBucketAcl{}
		actual.Name = e.Name
		actual.Bucket = e.Bucket
		actual.Entity = &foundEntity

		actual.Role = &foundRole

		// Ignore "system" fields
		actual.Lifecycle = e.Lifecycle

		return actual, nil
	}

	return nil, nil
}

func (e *StorageBucketAcl) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *StorageBucketAcl) CheckChanges(a, e, changes *StorageBucketAcl) error {
	if fi.ValueOf(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.ValueOf(e.Entity) == "" {
		return fi.RequiredField("Entity")
	}
	return nil
}

func (_ *StorageBucketAcl) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageBucketAcl) error {
	bucket := fi.ValueOf(e.Bucket)
	entity := fi.ValueOf(e.Entity)
	role := fi.ValueOf(e.Role)

	if a == nil {
		klog.V(2).Infof("Creating GCS bucket ACL for gs://%s for %s as %s", bucket, entity, role)
	} else {
		klog.V(2).Infof("Updating GCS bucket ACL for gs://%s for %s as %s", bucket, entity, role)
	}

	err := t.Cloud.Storage().Bucket(bucket).ACL().Set(context.TODO(), storage.ACLEntity(entity), storage.ACLRole(role))
	if err != nil {
		return fmt.Errorf("error setting GCS bucket ACL for gs://%s for %s as %s: %v", bucket, entity, role, err)
	}

	return nil
}

// terraformStorageBucketAcl is the model for a terraform google_storage_bucket_acl rule
type terraformStorageBucketAcl struct {
	Bucket     string   `cty:"bucket"`
	RoleEntity []string `cty:"role_entity"`
}

func (_ *StorageBucketAcl) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageBucketAcl) error {
	var roleEntities []string
	roleEntities = append(roleEntities, fi.ValueOf(e.Role)+":"+fi.ValueOf(e.Entity))
	tf := &terraformStorageBucketAcl{
		Bucket:     fi.ValueOf(e.Bucket),
		RoleEntity: roleEntities,
	}

	return t.RenderResource("google_storage_bucket_acl", *e.Name, tf)
}
