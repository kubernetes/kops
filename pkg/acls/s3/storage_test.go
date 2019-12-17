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

package s3

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/values"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_Strategy(t *testing.T) {
	context := &vfs.VFSContext{}
	path, err := context.BuildVfsPath("s3://test/foo")
	if err != nil {
		t.Errorf("unable to create path: %v", err)
	}

	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			ConfigStore: "s3://my_state_store/cluster",
			Assets: &kops.Assets{
				FileRepository: values.String("https://s3.amazonaws.com/test"),
			},
		},
	}

	s := &s3PublicAclStrategy{}
	acl, err := s.GetACL(path, cluster)

	if err != nil {
		t.Errorf("error getting ACL: %v", err)
	}

	if acl == nil {
		t.Errorf("public ro ACL is nil and should not be, this test is a positive test case.")
	}
}

func Test_In_StateStore(t *testing.T) {
	context := &vfs.VFSContext{}
	stateStore, err := context.BuildVfsPath("s3://my_state_store/cluster")
	if err != nil {
		t.Errorf("unable to create path: %v", err)
	}

	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			ConfigStore: "s3://my_state_store/cluster",
			Assets: &kops.Assets{
				FileRepository: values.String("https://s3.amazonaws.com/my_state_store/opps"),
			},
		},
	}

	s := &s3PublicAclStrategy{}
	acl, err := s.GetACL(stateStore, cluster)

	if err != nil {
		t.Errorf("error getting ACL: %v", err)
	}

	if acl != nil {
		t.Errorf("public ro ACL is set but path is in the state store, this test is a negative test case.")
	}
}
