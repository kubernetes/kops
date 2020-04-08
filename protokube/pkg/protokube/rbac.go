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
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/klog"
)

// applyRBAC is responsible for initializing RBAC
func applyRBAC(ctx context.Context, kubeContext *KubernetesContext) error {
	k8sClient, err := kubeContext.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error connecting to kubernetes: %v", err)
	}
	clientset := k8sClient.(*kubernetes.Clientset)

	var errors []error
	// kube-dns & kube-proxy service accounts
	if err := createServiceAccounts(ctx, clientset); err != nil {
		errors = append(errors, fmt.Errorf("error creating service accounts: %v", err))
	}
	//Currently all kubeadm specific
	if err := createClusterRoleBindings(ctx, clientset); err != nil {
		errors = append(errors, fmt.Errorf("error creating cluster role bindings: %v", err))
	}

	if len(errors) != 0 {
		if len(errors) != 1 {
			for _, err := range errors {
				klog.Warningf("Error configuring RBAC: %v", err)
			}
		}
		return errors[0]
	}

	return nil
}

// The below code should mirror the code in kubeadm.
// We'll develop it here then contribute it back once they are out of core -
// otherwise it is using the wrong version of the k8s client.
const (
	// KubeProxyClusterRoleName sets the name for the kube-proxy ClusterRole
	KubeProxyClusterRoleName = "system:node-proxier"

	clusterRoleKind    = "ClusterRole"
	roleKind           = "Role"
	serviceAccountKind = "ServiceAccount"
	rbacAPIGroup       = "rbac.authorization.k8s.io"
	//anonymousUser            = "system:anonymous"

	// Constants for what we name our ServiceAccounts with limited access to the cluster in case of RBAC
	KubeDNSServiceAccountName   = "kube-dns"
	KubeProxyServiceAccountName = "kube-proxy"
)

// createServiceAccounts creates the necessary serviceaccounts that kubeadm uses/might use, if they don't already exist.
func createServiceAccounts(ctx context.Context, clientset kubernetes.Interface) error {
	serviceAccounts := []v1.ServiceAccount{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KubeDNSServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KubeProxyServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}

	for _, sa := range serviceAccounts {
		if _, err := clientset.CoreV1().ServiceAccounts(metav1.NamespaceSystem).Create(ctx, &sa, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return err
			}
		}
	}
	return nil
}

func createClusterRoleBindings(ctx context.Context, clientset *kubernetes.Clientset) error {
	clusterRoleBindings := []rbac.ClusterRoleBinding{
		//{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Name: "kubeadm:kubelet-bootstrap",
		//	},
		//	RoleRef: rbac.RoleRef{
		//		APIGroup: rbacAPIGroup,
		//		Kind:     clusterRoleKind,
		//		Name:     NodeBootstrapperClusterRoleName,
		//	},
		//	Subjects: []rbac.Subject{
		//		{
		//			Kind: "Group",
		//			Name: bootstrapapi.BootstrapGroup,
		//		},
		//	},
		//},
		//{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Name: nodeAutoApproveBootstrap,
		//	},
		//	RoleRef: rbac.RoleRef{
		//		APIGroup: rbacAPIGroup,
		//		Kind:     clusterRoleKind,
		//		Name:     nodeAutoApproveBootstrap,
		//	},
		//	Subjects: []rbac.Subject{
		//		{
		//			Kind: "Group",
		//			Name: bootstrapapi.BootstrapGroup,
		//		},
		//	},
		//},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kubeadm:node-proxier",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: rbacAPIGroup,
				Kind:     clusterRoleKind,
				Name:     KubeProxyClusterRoleName,
			},
			Subjects: []rbac.Subject{
				{
					Kind:      serviceAccountKind,
					Name:      KubeProxyServiceAccountName,
					Namespace: metav1.NamespaceSystem,
				},
			},
		},
	}

	for _, clusterRoleBinding := range clusterRoleBindings {
		if _, err := clientset.RbacV1beta1().ClusterRoleBindings().Create(ctx, &clusterRoleBinding, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return fmt.Errorf("unable to create RBAC clusterrolebinding: %v", err)
			}

			if _, err := clientset.RbacV1beta1().ClusterRoleBindings().Update(ctx, &clusterRoleBinding, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("unable to update RBAC clusterrolebinding: %v", err)
			}
		}
	}
	return nil
}
