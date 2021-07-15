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

package awscloudcontrollermanager

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/pkg/model/iam"
)

// ServiceAccount represents the service-account used by the AWS Cloud Controller Manager.
// It implements iam.Subject to get AWS IAM permissions.
type ServiceAccount struct {
}

var _ iam.Subject = &ServiceAccount{}

// BuildAWSPolicy generates a custom policy for a ServiceAccount IAM role.
func (r *ServiceAccount) BuildAWSPolicy(b *iam.PolicyBuilder) (*iam.Policy, error) {
	clusterName := b.Cluster.ObjectMeta.Name
	p := iam.NewPolicy(clusterName)
	iam.AddCCMPermissions(p, clusterName, b.Cluster.Spec.Networking.Kubenet != nil)

	return p, nil
}

// ServiceAccount returns the kubernetes service account used.
func (r *ServiceAccount) ServiceAccount() (types.NamespacedName, bool) {
	return types.NamespacedName{
		Namespace: "kube-system",
		Name:      "aws-cloud-controller-manager",
	}, true
}
