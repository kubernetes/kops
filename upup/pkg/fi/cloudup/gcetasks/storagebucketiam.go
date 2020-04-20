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

// StorageBucketIam represents an IAM policy on a google cloud storage bucket
//go:generate fitask -type=StorageBucketIam
type StorageBucketIam struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Bucket *string
	Entity *string

	Role *string
}

var _ fi.CompareWithID = &StorageBucketIam{}

func (e *StorageBucketIam) CompareWithID() *string {
	return e.Name
}

func (e *StorageBucketIam) Find(c *fi.Context) (*StorageBucketIam, error) {
	cloud := c.Cloud.(gce.GCECloud)

	bucket := fi.StringValue(e.Bucket)
	entity := fi.StringValue(e.Entity)
	role := fi.StringValue(e.Role)

	klog.V(2).Infof("Checking GCS IAM policy for gs://%s for %s", bucket, entity)
	policy, err := cloud.Storage().Buckets.GetIamPolicy(bucket).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			policy = &storage.Policy{}
		} else {
			return nil, fmt.Errorf("error querying GCS IAM policy for gs://%s for %s: %v", bucket, entity, err)
		}
	}

	for _, binding := range policy.Bindings {
		if binding.Role != role {
			continue
		}

		for _, member := range binding.Members {
			if member == entity {
				return &StorageBucketIam{
					Name:      e.Name,
					Bucket:    e.Bucket,
					Entity:    e.Entity,
					Role:      e.Role,
					Lifecycle: e.Lifecycle,
				}, nil
			}
		}
	}

	return nil, nil
}

func (e *StorageBucketIam) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *StorageBucketIam) CheckChanges(a, e, changes *StorageBucketIam) error {
	if fi.StringValue(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.StringValue(e.Entity) == "" {
		return fi.RequiredField("Entity")
	}
	return nil
}

func (_ *StorageBucketIam) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageBucketIam) error {
	bucket := fi.StringValue(e.Bucket)
	entity := fi.StringValue(e.Entity)
	role := fi.StringValue(e.Role)

	policy, err := t.Cloud.Storage().Buckets.GetIamPolicy(bucket).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			policy = &storage.Policy{}
		} else {
			return fmt.Errorf("error querying GCS IAM policy for gs://%s: %v", bucket, err)
		}
	}

	foundRole := false
	for _, binding := range policy.Bindings {
		if binding.Role != role {
			continue
		}

		foundRole = true

		foundMember := false
		for _, member := range binding.Members {
			if member == entity {
				foundMember = true
			}
		}
		if foundMember {
			return nil
		}
		binding.Members = append(binding.Members, entity)
	}

	if !foundRole {
		binding := &storage.PolicyBindings{
			Role:    role,
			Members: []string{entity},
		}
		policy.Bindings = append(policy.Bindings, binding)
	}

	klog.V(2).Infof("Setting GCS IAM policy for gs://%s for %s: %s", bucket, entity, role)

	if _, err := t.Cloud.Storage().Buckets.SetIamPolicy(bucket, policy).Do(); err != nil {
		return fmt.Errorf("error setting GCS IAM policy for gs://%s for %s: %v", bucket, entity, err)
	}

	return nil
}

// type terraformStorageBucketIam struct {
// 	Bucket     string   `json:"bucket,omitempty" cty:"bucket"`
// 	RoleEntity []string `json:"role_entity,omitempty" cty:"role_entity"`
// }

func (_ *StorageBucketIam) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageBucketIam) error {
	//var roleEntities []string
	//roleEntities = append(roleEntities, fi.StringValue(e.Role)+":"+fi.StringValue(e.Name))
	//tf := &terraformStorageBucketIam{
	//	Bucket:     fi.StringValue(e.Bucket),
	//	RoleEntity: roleEntities,
	//}
	//
	//return t.RenderResource("google_storage_bucket_IAM policy", *e.Name, tf)

	klog.Warningf("terraform does not support GCE IAM policies on GCS buckets; please ensure your service account has access")
	return nil
}
