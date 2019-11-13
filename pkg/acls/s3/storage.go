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

package s3

import (
	"fmt"
	"net/url"

	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/acls"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/values"
	"k8s.io/kops/util/pkg/vfs"
)

// s3PublicAclStrategy is the AclStrategy for objects that are written with public read only ACL.
// This strategy is used by custom file assets.
type s3PublicAclStrategy struct {
}

var _ acls.ACLStrategy = &s3PublicAclStrategy{}

// GetACL creates a s3PublicAclStrategy object for writing public files with assets FileRepository.
// This strategy checks if the files are inside the state store, and if the files are located inside
// the state store, this returns nil and logs a message (level 8) that it will not run.
func (s *s3PublicAclStrategy) GetACL(p vfs.Path, cluster *kops.Cluster) (vfs.ACL, error) {
	if cluster.Spec.Assets == nil || cluster.Spec.Assets.FileRepository == nil {
		return nil, nil
	}

	s3Path, ok := p.(*vfs.S3Path)
	if !ok {
		return nil, nil
	}

	fileRepository := values.StringValue(cluster.Spec.Assets.FileRepository)

	u, err := url.Parse(fileRepository)
	if err != nil {
		return "", fmt.Errorf("unable to parse: %q", fileRepository)
	}

	// We are checking that the file repository url is in S3
	_, err = vfs.VFSPath(fileRepository)
	if err != nil {
		klog.V(8).Infof("path %q is not inside of a s3 bucket", u.String())
		return nil, nil
	}

	config, err := url.Parse(cluster.Spec.ConfigStore)
	if err != nil {
		return "", fmt.Errorf("unable to parse: %q", fileRepository)
	}

	// We are checking that the path defined is not the state store, if it is
	// we do NOT set the state store as public read.
	if strings.Contains(u.Path, config.Path) {
		klog.V(8).Infof("path %q is inside of config store %q, not setting public-read acl", u.Path, config.Path)
		return nil, nil
	}

	if strings.TrimPrefix(u.Path, "/") == s3Path.Bucket() {
		return &vfs.S3Acl{
			RequestACL: values.String("public-read"),
		}, nil
	}
	klog.V(8).Infof("path %q is not inside the file registry %q, not setting public-read acl", u.Path, config.Path)

	return nil, nil
}

func Register() {
	acls.RegisterPlugin("k8s.io/kops/acl/s3", &s3PublicAclStrategy{})
}
