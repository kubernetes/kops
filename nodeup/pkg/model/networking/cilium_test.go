/*
Copyright 2021 The Kubernetes Authors.

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

package networking

import (
	"runtime"
	"testing"

	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
)

func TestCiliumBuilder(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("cilium nodeup test will only work on linux")
	}
	context := &model.NodeupModelContext{
		Cluster: &kops.Cluster{
			Spec: kops.ClusterSpec{
				CloudProvider: "aws",
				EtcdClusters: []kops.EtcdClusterSpec{
					{
						Name:     "cilium",
						Provider: kops.EtcdProviderTypeManager,
					},
				},
				KubernetesVersion: "1.19.0",
				Networking: &kops.NetworkingSpec{
					Cilium: &kops.CiliumNetworkingSpec{
						EtcdManaged: true,
					},
				},
			},
		},
		HasAPIServer: true,
		KeyStore:     &fakeKeyStore{},
		IsMaster:     true,
	}
	etcdBuilder := &model.EtcdManagerTLSBuilder{
		NodeupModelContext: context,
	}
	ciliumBuilder := &CiliumBuilder{
		NodeupModelContext: context,
	}

	modelContext := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	if err := etcdBuilder.Build(modelContext); err != nil {
		t.Errorf("unexpected error building etcd: %v", err)
	}

	if err := ciliumBuilder.Build(modelContext); err != nil {
		t.Errorf("unexpected error building cilium: %v", err)
	}
}

type fakeKeyStore struct {
	fi.CAStore
}

func (*fakeKeyStore) FindCert(name string) (*pki.Certificate, error) {
	return &pki.Certificate{}, nil
}

func (*fakeKeyStore) FindPrivateKey(name string) (*pki.PrivateKey, error) {
	return &pki.PrivateKey{}, nil
}
