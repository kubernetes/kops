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

package gcemodel

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
	"k8s.io/kops/util/pkg/vfs"
)

// StorageAclBuilder configures storage acls
type StorageAclBuilder struct {
	*GCEModelContext
	Cloud     gce.GCECloud
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

// Build creates the tasks that set up storage acls

func (b *StorageAclBuilder) Build(c *fi.ModelBuilderContext) error {
<<<<<<< HEAD
	if featureflag.GoogleCloudBucketACL.Enabled() {
		if b.Cluster.Spec.CloudConfig.GCEServiceAccount == "" {
			return fmt.Errorf("featureflag GoogleCloudBucketACL not supported with per-instancegroup GCEServiceAccount")
		}

		klog.Warningf("featureflag GoogleCloudBucketACL is no longer recommended; use per-instancegroup GCEServiceAccounts instead")

		gceDefaultServiceAccount, err := b.Cloud.ServiceAccount()
		if err != nil {
			return fmt.Errorf("error fetching default ServiceAccount: %w", err)
		}

		clusterPath := b.Cluster.Spec.ConfigBase
		p, err := vfs.Context.BuildVfsPath(clusterPath)
		if err != nil {
			return fmt.Errorf("cannot parse cluster path %q: %w", clusterPath, err)
		}

		switch p := p.(type) {
		case *vfs.GSPath:
			// It's not ideal that we have to do this at the bucket level,
			// but GCS doesn't seem to have a way to do subtrees (like AWS IAM does)
			// Note this permission only lets us list objects, not read them
			c.AddTask(&gcetasks.StorageBucketAcl{
				Name:      s("serviceaccount-statestore-list"),
				Lifecycle: b.Lifecycle,
				Bucket:    s(p.Bucket()),
				Entity:    s("user-" + gceDefaultServiceAccount),
				Role:      s("READER"),
			})
=======
	done := make(map[string]bool)

	// clusterPath := b.Cluster.Spec.ConfigBase
	// p, err := vfs.Context.BuildVfsPath(clusterPath)
	// if err != nil {
	// 	return fmt.Errorf("cannot parse cluster path %q: %v", clusterPath, err)
	// }

	// switch p := p.(type) {
	// case *vfs.GSPath:
	// 	// It's not ideal that we have to do this at the bucket level,
	// 	// but GCS doesn't seem to have a way to do subtrees (like AWS IAM does)
	// 	// Note this permission only lets us list objects, not read them
	// 	c.AddTask(&gcetasks.StorageBucketAcl{
	// 		Name:      s("serviceaccount-statestore-list"),
	// 		Lifecycle: b.Lifecycle,
	// 		Bucket:    s(p.Bucket()),
	// 		Entity:    s("user-" + serviceAccount),
	// 		Role:      s("READER"),
	// 	})
	// }

	// klog.Warningf("we need to split master / node roles")
	// nodeRole, err := iam.BuildNodeRoleSubject(kops.InstanceGroupRoleMaster, false)
	// if err != nil {
	// 	return err
	// }

	for _, ig := range b.InstanceGroups {
		serviceAccount := b.LinkToServiceAccount(ig)

		email := *serviceAccount.Email
		if email == "default" {
			defaultServiceAccount, err := b.Cloud.DefaultServiceAccount()
			if err != nil {
				return err
			}
			email = defaultServiceAccount
		}

		if done[email] {
			continue
>>>>>>> e997ae1248 (gce: use serviceaccounts per role)
		}
		done[email] = true

		role := ig.Spec.Role

		nodeRole, err := iam.BuildNodeRoleSubject(role, false)
		if err != nil {
			return err
		}

		writeablePaths, err := iam.WriteableVFSPaths(b.Cluster, nodeRole)
		if err != nil {
			return err
		}

		buckets := sets.NewString()
		for _, p := range writeablePaths {
			gcsPath, ok := p.(*vfs.GSPath)
			if !ok {
				klog.Warningf("unknown path, can't apply IAM policy: %q", p)
				continue
			}

			bucket := gcsPath.Bucket()
			if buckets.Has(bucket) {
				continue
			}
			buckets.Insert(bucket)

<<<<<<< HEAD
			klog.Warningf("adding bucket level write ACL to gs://%s to support etcd backup", bucket)

			c.AddTask(&gcetasks.StorageBucketAcl{
				Name:      s("serviceaccount-backup-readwrite-" + bucket),
				Lifecycle: b.Lifecycle,
				Bucket:    s(bucket),
				Entity:    s("user-" + gceDefaultServiceAccount),
				Role:      s("WRITER"),
			})
		}

		return nil
	}

	if b.Cluster.Spec.CloudConfig.GCEServiceAccount != "" {
		// We can't set up per-instancegroup permissions if we're using a cluster-level account
		// Historically, we did not grant the serviceaccount permissions in this case.
		return nil
	}

	type serviceAccountRole struct {
		Email string
		Role  kops.InstanceGroupRole
	}
	serviceAccountRoles := make(map[serviceAccountRole]bool)

	for _, ig := range b.InstanceGroups {
		serviceAccount := b.LinkToServiceAccount(ig)

		email := *serviceAccount.Email
		serviceAccountRoles[serviceAccountRole{Email: email, Role: ig.Spec.Role}] = true
	}

	for serviceAccountRole := range serviceAccountRoles {
		role := serviceAccountRole.Role

		nodeRole, err := iam.BuildNodeRoleSubject(role, false)
		if err != nil {
			return err
		}

		buckets := sets.NewString()

		writeablePaths, err := iam.WriteableVFSPaths(b.Cluster, nodeRole)
		if err != nil {
			return err
		}
		for _, p := range writeablePaths {
			gcsPath, ok := p.(*vfs.GSPath)
			if !ok {
				klog.Warningf("unknown path, can't apply IAM policy: %q", p)
				continue
=======
			nameForTask := strings.ToLower(string(role))

			if featureflag.GoogleCloudBucketACL.Enabled() {
				if err := b.addBucketACL(c, nameForTask, email, bucket); err != nil {
					return err
				}
			} else if !fi.BoolValue(serviceAccount.Shared) {
				if err := b.addBucketIAM(c, nameForTask, email, bucket); err != nil {
					return err
				}
>>>>>>> e997ae1248 (gce: use serviceaccounts per role)
			}

			bucket := gcsPath.Bucket()
			if buckets.Has(bucket) {
				continue
			}
			buckets.Insert(bucket)

			nameForTask := strings.ToLower(string(role))

			klog.Warningf("adding bucket level write IAM for role %q to gs://%s to support etcd backup", bucket, role)

			c.AddTask(&gcetasks.StorageBucketIAM{
				Name:      s("objectadmin-" + bucket + "-serviceaccount-" + nameForTask),
				Lifecycle: b.Lifecycle,
				Bucket:    s(bucket),
				Member:    s("serviceAccount:" + serviceAccountRole.Email),
				Role:      s("roles/storage.objectAdmin"),
			})
		}

		// Add bucket read permissions if we need to read from the bucket
		readablePaths, err := iam.ReadableStatePaths(b.Cluster, nodeRole)
		if err != nil {
			return err
		}
		if len(readablePaths) != 0 {
			p, err := vfs.Context.BuildVfsPath(b.Cluster.Spec.ConfigStore)
			if err != nil {
				return fmt.Errorf("cannot parse VFS path %q: %v", b.Cluster.Spec.ConfigStore, err)
			}

			gcsPath, ok := p.(*vfs.GSPath)
			if !ok {
				klog.Warningf("unknown path, can't apply IAM policy: %q", p)
				continue
			}
			bucket := gcsPath.Bucket()
			if buckets.Has(bucket) {
				// Already marked as writeable; we can skip
				continue
			}
			buckets.Insert(bucket)

			nameForTask := strings.ToLower(string(role))

			klog.Warningf("adding bucket level read IAM to gs://%s for role %q", bucket, role)

			c.AddTask(&gcetasks.StorageBucketIAM{
				Name:      s("objectviewer-" + bucket + "-serviceaccount-" + nameForTask),
				Lifecycle: b.Lifecycle,
				Bucket:    s(bucket),
				Member:    s("serviceAccount:" + serviceAccountRole.Email),
				Role:      s("roles/storage.objectViewer"),
			})
		}
	}

<<<<<<< HEAD
=======
	return nil
}

func (b *StorageAclBuilder) addBucketACL(c *fi.ModelBuilderContext, nameForTask string, serviceAccountEmail string, bucket string) error {
	klog.Warningf("adding bucket level write ACL to gs://%s to support etcd backup", bucket)

	c.AddTask(&gcetasks.StorageBucketAcl{
		Name:      s("writer-" + bucket + "-serviceaccount-" + nameForTask),
		Lifecycle: b.Lifecycle,
		Bucket:    s(bucket),
		Entity:    s("user-" + serviceAccountEmail),
		Role:      s("WRITER"),
	})

	return nil
}

func (b *StorageAclBuilder) addBucketIAM(c *fi.ModelBuilderContext, nameForTask string, serviceAccountEmail string, bucket string) error {
	klog.Warningf("adding bucket level write IAM to gs://%s to support etcd backup", bucket)

	c.AddTask(&gcetasks.StorageBucketIAM{
		Name:      s("objectadmin-" + bucket + "-serviceaccount-" + nameForTask),
		Lifecycle: b.Lifecycle,
		Bucket:    s(bucket),
		Member:    s("serviceAccount:" + serviceAccountEmail),
		Role:      s("roles/storage.objectAdmin"),
	})

>>>>>>> e997ae1248 (gce: use serviceaccounts per role)
	return nil
}
