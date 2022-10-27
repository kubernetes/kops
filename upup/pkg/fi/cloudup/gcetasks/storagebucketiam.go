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

	"google.golang.org/api/storage/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// StorageBucketIAM represents an IAM rule on a google cloud storage bucket
// +kops:fitask
type StorageBucketIAM struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Bucket *string
	Member *string
	Role   *string
}

var _ fi.CompareWithID = &StorageBucketIAM{}

func (e *StorageBucketIAM) CompareWithID() *string {
	return e.Name
}

func (e *StorageBucketIAM) Find(c *fi.CloudContext) (*StorageBucketIAM, error) {
	ctx := context.TODO()

	cloud := c.Cloud.(gce.GCECloud)

	bucket := fi.StringValue(e.Bucket)
	member := fi.StringValue(e.Member)
	role := fi.StringValue(e.Role)

	klog.V(2).Infof("Checking GCS bucket IAM for gs://%s for %s", bucket, member)
	policy, err := cloud.Storage().Buckets.GetIamPolicy(bucket).Context(ctx).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error checking GCS bucket IAM for gs://%s: %w", bucket, err)
	}

	changed := patchPolicy(policy, member, role)
	if changed {
		return nil, nil
	}

	actual := &StorageBucketIAM{}
	actual.Bucket = e.Bucket
	actual.Member = e.Member
	actual.Role = e.Role

	// Ignore "system" fields
	actual.Name = e.Name
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *StorageBucketIAM) Run(c *fi.CloudContext) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *StorageBucketIAM) CheckChanges(a, e, changes *StorageBucketIAM) error {
	if fi.StringValue(e.Bucket) == "" {
		return fi.RequiredField("Bucket")
	}
	if fi.StringValue(e.Member) == "" {
		return fi.RequiredField("Member")
	}
	if fi.StringValue(e.Role) == "" {
		return fi.RequiredField("Role")
	}
	return nil
}

func (_ *StorageBucketIAM) RenderGCE(t *gce.GCEAPITarget, a, e, changes *StorageBucketIAM) error {
	ctx := context.TODO()

	bucket := fi.StringValue(e.Bucket)
	member := fi.StringValue(e.Member)
	role := fi.StringValue(e.Role)

	klog.V(2).Infof("Creating GCS bucket IAM for gs://%s for %s as %s", bucket, member, role)

	policy, err := t.Cloud.Storage().Buckets.GetIamPolicy(bucket).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error creating IAM policy for bucket gs://%s: %w", bucket, err)
	}

	changed := patchPolicy(policy, member, role)

	if !changed {
		klog.Warningf("did not need to change policy (concurrent change?)")
		return nil
	}

	if _, err := t.Cloud.Storage().Buckets.SetIamPolicy(bucket, policy).Context(ctx).Do(); err != nil {
		return fmt.Errorf("error updating GCS bucket IAM for gs://%s: %v", bucket, err)
	}

	return nil
}

// terraformStorageBucketIAM is the model for a terraform google_storage_bucket_iam_member rule
type terraformStorageBucketIAM struct {
	Bucket string `cty:"bucket"`
	Role   string `cty:"role"`
	Member string `cty:"member"`
}

func (_ *StorageBucketIAM) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *StorageBucketIAM) error {
	tf := &terraformStorageBucketIAM{
		Bucket: fi.StringValue(e.Bucket),
		Role:   fi.StringValue(e.Role),
		Member: fi.StringValue(e.Member),
	}

	return t.RenderResource("google_storage_bucket_iam_member", *e.Name, tf)
}

func patchPolicy(policy *storage.Policy, wantMember string, wantRole string) bool {
	for _, binding := range policy.Bindings {
		if binding.Condition != nil {
			continue
		}
		if binding.Role != wantRole {
			continue
		}
		exists := false
		for _, member := range binding.Members {
			if member == wantMember {
				exists = true
			}
		}
		if exists {
			return false
		}

		if !exists {
			binding.Members = append(binding.Members, wantMember)
			return true
		}
	}

	policy.Bindings = append(policy.Bindings, &storage.PolicyBindings{
		Members: []string{wantMember},
		Role:    wantRole,
	})
	return true
}
