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

package vfs

import (
	"fmt"
	"io"
	"path"
	"strings"

	"k8s.io/kops/util/pkg/hashing"
)

// KubernetesPath is a path for a VFS backed by the kubernetes API
// Currently all operations are no-ops
type KubernetesPath struct {
	k8sContext *KubernetesContext
	host       string
	key        string
}

var _ Path = &KubernetesPath{}
var _ HasHash = &KubernetesPath{}

func newKubernetesPath(k8sContext *KubernetesContext, host string, key string) *KubernetesPath {
	host = strings.TrimSuffix(host, "/")
	key = strings.TrimPrefix(key, "/")

	return &KubernetesPath{
		k8sContext: k8sContext,
		host:       host,
		key:        key,
	}
}

func (p *KubernetesPath) Path() string {
	return "k8s://" + p.host + "/" + p.key
}

func (p *KubernetesPath) Host() string {
	return p.host
}

func (p *KubernetesPath) Key() string {
	return p.key
}

func (p *KubernetesPath) String() string {
	return p.Path()
}

func (p *KubernetesPath) Remove() error {
	return fmt.Errorf("KubernetesPath::Remove not supported")
}

func (p *KubernetesPath) RemoveAll() error {
	return p.Remove()
}

func (p *KubernetesPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &KubernetesPath{
		k8sContext: p.k8sContext,
		host:       p.host,
		key:        joined,
	}
}

func (p *KubernetesPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	return fmt.Errorf("KubernetesPath::WriteFile not supported")
}

func (p *KubernetesPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	return fmt.Errorf("KubernetesPath::CreateFile not supported")
}

// ReadFile implements Path::ReadFile
func (p *KubernetesPath) ReadFile() ([]byte, error) {
	return nil, fmt.Errorf("KubernetesPath::ReadFile not supported")
}

func (p *KubernetesPath) ReadDir() ([]Path, error) {
	return nil, fmt.Errorf("KubernetesPath::ReadDir not supported")
}

func (p *KubernetesPath) ReadTree() ([]Path, error) {
	return nil, fmt.Errorf("KubernetesPath::ReadTree not supported")
}

func (p *KubernetesPath) Base() string {
	return path.Base(p.key)
}

func (p *KubernetesPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *KubernetesPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	return nil, fmt.Errorf("KubernetesPath::Hash not supported")
}
