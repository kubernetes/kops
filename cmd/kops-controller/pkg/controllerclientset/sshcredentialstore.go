/*
Copyright 2024 The Kubernetes Authors.

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

package controllerclientset

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type sshCredentialStore struct {
	clusterBasePath vfs.Path

	cluster *kops.Cluster
}

var _ fi.SSHCredentialStore = &sshCredentialStore{}

func newSSHCredentialStore(clusterBasePath vfs.Path, cluster *kops.Cluster) *sshCredentialStore {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	s := &sshCredentialStore{
		clusterBasePath: clusterBasePath,
		cluster:         cluster,
	}

	return s
}

// DeleteSSHCredential deletes the specified SSH credential.
func (s *sshCredentialStore) DeleteSSHCredential() error {
	return fmt.Errorf("method DeleteSSHCredential not supported in server-side client")
}

// AddSSHPublicKey adds an SSH public key.
func (s *sshCredentialStore) AddSSHPublicKey(ctx context.Context, data []byte) error {
	return fmt.Errorf("method AddSSHPublicKey not supported in server-side client")
}

// FindSSHPublicKeys retrieves the SSH public keys.
func (s *sshCredentialStore) FindSSHPublicKeys() ([]*kops.SSHCredential, error) {
	klog.Warningf("method FindSSHPublicKeys is stub-implemented supported in server-side client")
	return nil, nil
}
