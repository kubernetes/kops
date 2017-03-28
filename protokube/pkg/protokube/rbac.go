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

package protokube

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/apiconfig"
)

func InitializeRBAC(kubeContext *KubernetesContext) error {
	k8sClient, err := kubeContext.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error connecting to kubernetes: %v", err)
	}
	clientset := k8sClient.(*kubernetes.Clientset)

	var errors []error
	if err := apiconfig.CreateServiceAccounts(clientset); err != nil {
		errors = append(errors, fmt.Errorf("error creating service accounts: %v", err))
	}
	if err := apiconfig.CreateClusterRoleBindings(clientset); err != nil {
		errors = append(errors, fmt.Errorf("error creating cluster role bindings: %v", err))
	}

	if len(errors) != 0 {
		if len(errors) != 1 {
			for _, err := range errors {
				glog.Warningf("Error configuring RBAC: %v", err)
			}
		}
		return errors[0]
	}

	return nil
}
