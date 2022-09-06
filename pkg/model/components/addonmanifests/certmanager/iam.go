/*
Copyright 2022 The Kubernetes Authors.

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

package certmanager

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorslice"
)

// ServiceAccount represents the service-account used by cert-manager.
// It implements iam.Subject to get AWS IAM permissions.
type ServiceAccount struct{}

var _ iam.Subject = &ServiceAccount{}

// BuildAWSPolicy generates a custom policy for a ServiceAccount IAM role.
func (r *ServiceAccount) BuildAWSPolicy(b *iam.PolicyBuilder) (*iam.Policy, error) {
	clusterName := b.Cluster.ObjectMeta.Name
	p := iam.NewPolicy(clusterName, b.Partition)

	addCertManagerPermissions(b, p)

	return p, nil
}

// ServiceAccount returns the kubernetes service account used.
func (r *ServiceAccount) ServiceAccount() (types.NamespacedName, bool) {
	return types.NamespacedName{
		Namespace: "kube-system",
		Name:      "cert-manager",
	}, true
}

func addCertManagerPermissions(b *iam.PolicyBuilder, p *iam.Policy) {
	var zoneResources []string
	for _, id := range b.Cluster.Spec.CertManager.HostedZoneIDs {
		zoneResources = append(zoneResources, fmt.Sprintf("arn:%v:route53:::hostedzone/%v", b.Partition, id))
	}

	p.Statement = append(p.Statement, &iam.Statement{
		Effect: iam.StatementEffectAllow,
		Action: stringorslice.Of("route53:ChangeResourceRecordSets",
			"route53:ListResourceRecordSets",
		),
		Resource: stringorslice.Slice(zoneResources),
	})

	p.Statement = append(p.Statement, &iam.Statement{
		Effect:   iam.StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:GetChange"}),
		Resource: stringorslice.Slice([]string{fmt.Sprintf("arn:%v:route53:::change/*", b.Partition)}),
	})

	wildcard := stringorslice.Slice([]string{"*"})
	p.Statement = append(p.Statement, &iam.Statement{
		Effect:   iam.StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:ListHostedZonesByName"}),
		Resource: wildcard,
	})
}
