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
	"k8s.io/kops/pkg/util/stringorset"
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
		Effect:   iam.StatementEffectAllow,
		Action:   stringorset.Of("route53:ListResourceRecordSets"),
		Resource: stringorset.Set(zoneResources),
	})

	p.Statement = append(p.Statement, &iam.Statement{
		Effect:   iam.StatementEffectAllow,
		Action:   stringorset.Of("route53:ChangeResourceRecordSets"),
		Resource: stringorset.Set(zoneResources),
		Condition: iam.Condition{
			"ForAllValues:StringLike": map[string]interface{}{
				"route53:ChangeResourceRecordSetsNormalizedRecordNames": []string{"_acme-challenge.*"},
			},
			"ForAllValues:StringEquals": map[string]interface{}{
				"route53:ChangeResourceRecordSetsRecordTypes": []string{"TXT"},
			},
		},
	})

	p.Statement = append(p.Statement, &iam.Statement{
		Effect:   iam.StatementEffectAllow,
		Action:   stringorset.Set([]string{"route53:GetChange"}),
		Resource: stringorset.Set([]string{fmt.Sprintf("arn:%v:route53:::change/*", b.Partition)}),
	})

	wildcard := stringorset.Set([]string{"*"})
	p.Statement = append(p.Statement, &iam.Statement{
		Effect:   iam.StatementEffectAllow,
		Action:   stringorset.Set([]string{"route53:ListHostedZonesByName"}),
		Resource: wildcard,
	})
}
