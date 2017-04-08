/*
Copyright 2016 The Kubernetes Authors.

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

package kubernetestarget

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/kutil"
)

type KubernetesTarget struct {
	KubernetesClient kubernetes.Interface
	cluster          *kopsapi.Cluster
}

func NewKubernetesTarget(clientset simple.Clientset, keyStore fi.Keystore, cluster *kopsapi.Cluster) (*KubernetesTarget, error) {
	var secretStore fi.SecretStore
	status := &kopsapi.NoopStatusStore{}
	kubeconfig, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, status)
	if err != nil {
		return nil, fmt.Errorf("error building credentials for cluster %q: %v", cluster.ObjectMeta.Name, err)
	}

	clientConfig, err := kubeconfig.BuildRestConfig()
	if err != nil {
		return nil, fmt.Errorf("error building configuration for cluster %q: %v", cluster.ObjectMeta.Name, err)
	}

	k8sClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot build k8s client: %v", err)
	}

	target := &KubernetesTarget{
		cluster:          cluster,
		KubernetesClient: k8sClient,
	}
	return target, nil

}

var _ fi.Target = &KubernetesTarget{}

func (t *KubernetesTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *KubernetesTarget) ProcessDeletions() bool {
	// We don't expect any, but it would be up to us to process
	return true
}

func (t *KubernetesTarget) Apply(manifest []byte) error {
	context := t.cluster.ObjectMeta.Name

	// Would be nice if we could use RunApply from kubectl's code directly...
	// ... but that seems really hard

	kubectl := &kutil.Kubectl{}
	err := kubectl.Apply(context, manifest)
	if err != nil {
		return err
	}

	return nil
}
