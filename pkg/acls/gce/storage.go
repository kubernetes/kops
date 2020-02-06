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

package gce

import (
	"fmt"

	storage "google.golang.org/api/storage/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/util/pkg/vfs"
)

// gcsAclStrategy is the AclStrategy for objects written to google cloud storage
type gcsAclStrategy struct {
}

var _ acls.ACLStrategy = &gcsAclStrategy{}

// GetACL returns the ACL to use if this is a google cloud storage path
func (s *gcsAclStrategy) GetACL(p vfs.Path, cluster *kops.Cluster) (vfs.ACL, error) {
	if kops.CloudProviderID(cluster.Spec.CloudProvider) != kops.CloudProviderGCE {
		return nil, nil
	}
	gcsPath, ok := p.(*vfs.GSPath)
	if !ok {
		return nil, nil
	}

	bucketName := gcsPath.Bucket()
	client := gcsPath.Client()

	// TODO: Cache?
	bucket, err := client.Buckets.Get(bucketName).Do()
	if err != nil {
		return nil, fmt.Errorf("error querying bucket %q: %v", bucketName, err)
	}

	bucketPolicyOnly := false
	if bucket.IamConfiguration != nil && bucket.IamConfiguration.BucketPolicyOnly != nil {
		bucketPolicyOnly = bucket.IamConfiguration.BucketPolicyOnly.Enabled
	}

	if bucketPolicyOnly {
		klog.V(2).Infof("bucket gs://%s has bucket-policy only; won't try to set ACLs", bucketName)
		return nil, nil
	}

	// TODO: Cache?
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	serviceAccount, err := cloud.(gce.GCECloud).ServiceAccount()
	if err != nil {
		return nil, err
	}

	var acls []*storage.ObjectAccessControl
	for _, a := range bucket.DefaultObjectAcl {
		acls = append(acls, a)
	}

	acls = append(acls, &storage.ObjectAccessControl{
		Email:  serviceAccount,
		Entity: "user-" + serviceAccount,
		Role:   "READER",
	})

	return &vfs.GSAcl{
		Acl: acls,
	}, nil
}

func Register() {
	acls.RegisterPlugin("k8s.io/kops/acl/gce", &gcsAclStrategy{})
}
