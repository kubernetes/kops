/*
Copyright 2020 The Kubernetes Authors.

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

package nodetasks

import (
	"fmt"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
)

type KubeConfig struct {
	Name      string
	Cert      fi.Resource
	Key       fi.Resource
	CA        fi.Resource
	ServerURL string

	config *fi.TaskDependentResource
}

var _ fi.Task = &KubeConfig{}
var _ fi.HasName = &KubeConfig{}
var _ fi.HasDependencies = &KubeConfig{}

func (k *KubeConfig) GetName() *string {
	return &k.Name
}

// String returns a string representation, implementing the Stringer interface
func (k *KubeConfig) String() string {
	return fmt.Sprintf("KubeConfig: %s", k.Name)
}

func (k *KubeConfig) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if hasDep, ok := k.Cert.(fi.HasDependencies); ok {
		deps = append(deps, hasDep.GetDependencies(tasks)...)
	}
	if hasDep, ok := k.Key.(fi.HasDependencies); ok {
		deps = append(deps, hasDep.GetDependencies(tasks)...)
	}
	if hasDep, ok := k.CA.(fi.HasDependencies); ok {
		deps = append(deps, hasDep.GetDependencies(tasks)...)
	}

	return deps
}

func (k *KubeConfig) GetConfig() *fi.TaskDependentResource {
	if k.config == nil {
		k.config = &fi.TaskDependentResource{Task: k}
	}
	return k.config
}

func (k *KubeConfig) Run(_ *fi.Context) error {
	cert, err := fi.ResourceAsBytes(k.Cert)
	if err != nil {
		return err
	}
	key, err := fi.ResourceAsBytes(k.Key)
	if err != nil {
		return err
	}
	ca, err := fi.ResourceAsBytes(k.CA)
	if err != nil {
		return err
	}

	user := kubeconfig.KubectlUser{
		ClientCertificateData: cert,
		ClientKeyData:         key,
	}
	cluster := kubeconfig.KubectlCluster{
		CertificateAuthorityData: ca,
		Server:                   k.ServerURL,
	}

	config := &kubeconfig.KubectlConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Users: []*kubeconfig.KubectlUserWithName{
			{
				Name: k.Name,
				User: user,
			},
		},
		Clusters: []*kubeconfig.KubectlClusterWithName{
			{
				Name:    "local",
				Cluster: cluster,
			},
		},
		Contexts: []*kubeconfig.KubectlContextWithName{
			{
				Name: "service-account-context",
				Context: kubeconfig.KubectlContext{
					Cluster: "local",
					User:    k.Name,
				},
			},
		},
		CurrentContext: "service-account-context",
	}

	yaml, err := kops.ToRawYaml(config)
	if err != nil {
		return fmt.Errorf("error marshaling kubeconfig to yaml: %v", err)
	}

	output := k.GetConfig()
	output.Resource = fi.NewBytesResource(yaml)

	return nil
}
