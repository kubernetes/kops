/*
Copyright 2019 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// IAMModelBuilder configures IAM objects
type IAMModelBuilder struct {
	*KopsModelContext

	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &IAMModelBuilder{}

const NodeRolePolicyTemplate = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "{{ IAMServiceEC2 }}"},
      "Action": "sts:AssumeRole"
    }
  ]
}`

func (b *IAMModelBuilder) Build(c *fi.ModelBuilderContext) error {
	shared := sets.NewString()
	// Generate IAM tasks for each shared and managed role
	for _, ig := range b.InstanceGroups {
		role, err := iam.BuildNodeRoleSubject(ig.Spec.Role)
		if err != nil {
			return err
		}
		if ig.Spec.IAM != nil && ig.Spec.IAM.Profile != nil {
			// Shared role case
			profileARN := *ig.Spec.IAM.Profile
			iamName, err := findCustomAuthNameFromArn(profileARN)
			if err != nil {
				return fmt.Errorf("unable to parse instance profile name from arn %q: %v", profileARN, err)
			}
			if !shared.Has(iamName) {
				if err := b.buildIAMTasks(role, iamName, ig, c, true); err != nil {
					return err
				}
				shared.Insert(iamName)
			}
		} else {
			// Managed role case
			iamName := b.IAMName(ig)
			if !shared.Has(iamName) {
				if err := b.buildIAMTasks(role, iamName, ig, c, false); err != nil {
					return err
				}
				shared.Insert(iamName)
			}
		}
	}

	return nil
}

// BuildServiceAccountRoleTasks build tasks specifically for the ServiceAccount role.
func (b *IAMModelBuilder) BuildServiceAccountRoleTasks(role iam.Subject, c *fi.ModelBuilderContext) error {
	iamName, err := b.IAMNameForServiceAccountRole(role)
	if err != nil {
		return err
	}

	iamRole, err := b.buildIAMRole(role, iamName, "", c)
	if err != nil {
		return err
	}

	if err := b.buildIAMRolePolicy(role, iamName, iamRole, c); err != nil {
		return err
	}

	return nil
}

func (b *IAMModelBuilder) buildIAMRole(role iam.Subject, iamName, exportId string, c *fi.ModelBuilderContext) (*awstasks.IAMRole, error) {
	roleKey, isServiceAccount := b.roleKey(role)

	rolePolicy, err := b.buildAWSIAMRolePolicy(role)
	if err != nil {
		return nil, err
	}

	iamRole := &awstasks.IAMRole{
		Name:               s(iamName),
		Lifecycle:          b.Lifecycle,
		RolePolicyDocument: fi.WrapResource(rolePolicy),
	}

	if isServiceAccount {
		// e.g. kube-system-dns-controller
		iamRole.ExportWithID = s(roleKey)
	} else {
		// e.g. nodes
		iamRole.ExportWithID = s(exportId)
	}

	if b.Cluster.Spec.IAM != nil && b.Cluster.Spec.IAM.PermissionsBoundary != nil {
		iamRole.PermissionsBoundary = b.Cluster.Spec.IAM.PermissionsBoundary
	}

	c.AddTask(iamRole)

	return iamRole, nil
}

func (b *IAMModelBuilder) buildIAMRolePolicy(role iam.Subject, iamName string, iamRole *awstasks.IAMRole, c *fi.ModelBuilderContext) error {
	iamPolicy := &iam.PolicyResource{
		Builder: &iam.PolicyBuilder{
			Cluster:              b.Cluster,
			Role:                 role,
			Region:               b.Region,
			UseServiceAccountIAM: b.UseServiceAccountIAM(),
		},
	}

	if !dns.IsGossipHostname(b.Cluster.ObjectMeta.Name) {
		// This is slightly tricky; we need to know the hosted zone id,
		// but we might be creating the hosted zone dynamically.
		// We create a stub-reference which will be combined by the execution engine.
		iamPolicy.DNSZone = &awstasks.DNSZone{
			Name: fi.String(b.NameForDNSZone()),
		}
	}

	t := &awstasks.IAMRolePolicy{
		Name:           s(iamName),
		Lifecycle:      b.Lifecycle,
		Role:           iamRole,
		PolicyDocument: iamPolicy,
	}
	c.AddTask(t)

	return nil
}

// roleKey builds a string to represent the role uniquely.  It returns true if this is a service account role.
func (b *IAMModelBuilder) roleKey(role iam.Subject) (string, bool) {
	serviceAccount, ok := role.ServiceAccount()
	if ok {
		return strings.ToLower(serviceAccount.Namespace + "-" + serviceAccount.Name), true
	}

	// This isn't great, but we have to be backwards compatible with the old names.
	switch role.(type) {
	case *iam.NodeRoleMaster:
		return strings.ToLower(string(kops.InstanceGroupRoleMaster)), false
	case *iam.NodeRoleNode:
		return strings.ToLower(string(kops.InstanceGroupRoleNode)), false
	case *iam.NodeRoleBastion:
		return strings.ToLower(string(kops.InstanceGroupRoleBastion)), false

	default:
		klog.Fatalf("unknown node role type: %T", role)
		return "", false
	}
}

func (b *IAMModelBuilder) createExternalPoliciesTask(name string, iamRole *awstasks.IAMRole, externalPolicies []string, c *fi.ModelBuilderContext) fi.Task {
	return &awstasks.IAMRolePolicy{
		Name:             s(name),
		Lifecycle:        b.Lifecycle,
		Role:             iamRole,
		Managed:          true,
		ExternalPolicies: &externalPolicies,
	}
}

func (b *IAMModelBuilder) createAdditionalPolicyTask(name string, iamRole *awstasks.IAMRole, additionalPolicy string, c *fi.ModelBuilderContext) (fi.Task, error) {
	t := &awstasks.IAMRolePolicy{
		Name:      s(name),
		Lifecycle: b.Lifecycle,
		Role:      iamRole,
	}

	if additionalPolicy != "" {
		p := &iam.Policy{
			Version: iam.PolicyDefaultVersion,
		}

		statements, err := iam.ParseStatements(additionalPolicy)
		if err != nil {
			return nil, fmt.Errorf("additionalPolicy %q is invalid: %v", name, err)
		}

		p.Statement = append(p.Statement, statements...)

		policy, err := p.AsJSON()
		if err != nil {
			return nil, fmt.Errorf("error building IAM policy: %v", err)
		}

		t.PolicyDocument = fi.WrapResource(fi.NewStringResource(policy))
	} else {
		t.PolicyDocument = fi.WrapResource(fi.NewStringResource(""))
	}

	return t, nil
}

func (b *IAMModelBuilder) buildIAMTasks(role iam.Subject, iamName string, ig *kops.InstanceGroup, c *fi.ModelBuilderContext, shared bool) error {
	roleKey, _ := b.roleKey(role)

	iamRole, err := b.buildIAMRole(role, iamName, ig.Name, c)
	if err != nil {
		return err
	}

	if err := b.buildIAMRolePolicy(role, iamName, iamRole, c); err != nil {
		return err
	}

	{
		// To minimize diff for easier code review

		var iamInstanceProfile *awstasks.IAMInstanceProfile
		{
			iamInstanceProfile = &awstasks.IAMInstanceProfile{
				Name:      s(iamName),
				Lifecycle: b.Lifecycle,
				Shared:    fi.Bool(shared),
			}
			c.AddTask(iamInstanceProfile)
		}
		{
			iamInstanceProfileRole := &awstasks.IAMInstanceProfileRole{
				Name:            s(iamName),
				Lifecycle:       b.Lifecycle,
				InstanceProfile: iamInstanceProfile,
				Role:            iamRole,
			}
			c.AddTask(iamInstanceProfileRole)
		}

		// Create External Policy defined at the cluster level task
		if !shared {
			var externalPolicies []string
			if b.Cluster.Spec.ExternalPolicies != nil {
				p := *(b.Cluster.Spec.ExternalPolicies)
				externalPolicies = append(externalPolicies, p[roleKey]...)
			}
			c.AddTask(b.createExternalPoliciesTask(
				fmt.Sprintf("%s-policyoverride", iamName),
				iamRole,
				externalPolicies,
				c,
			))
		}

		// Create External Policy defined at the instance group level task
		if !shared {
			var externalPolicies []string
			if ig != nil && ig.Spec.ExternalPolicies != nil {
				externalPolicies = *ig.Spec.ExternalPolicies
			}
			c.AddTask(b.createExternalPoliciesTask(
				fmt.Sprintf("%s-policyoverride-ig", iamName),
				iamRole,
				externalPolicies,
				c,
			))
		}

		// Generate additional policies defined at the cluster level if needed, and attach to existing role
		if !shared {
			additionalPolicy := ""
			if b.Cluster.Spec.AdditionalPolicies != nil {
				additionalPolicies := *(b.Cluster.Spec.AdditionalPolicies)
				additionalPolicy = additionalPolicies[roleKey]
			}
			t, err := b.createAdditionalPolicyTask(
				"additional."+iamName,
				iamRole,
				additionalPolicy,
				c,
			)
			if err != nil {
				return err
			}
			c.AddTask(t)
		}

		// Generate additional policies defined at the instance group level if needed, and attach to existing role
		if !shared {
			additionalPolicy := ""
			if ig != nil && ig.Spec.AdditionalPolicy != nil {
				additionalPolicy = *ig.Spec.AdditionalPolicy
			}
			t, err := b.createAdditionalPolicyTask(
				"additional-ig."+iamName,
				iamRole,
				additionalPolicy,
				c,
			)
			if err != nil {
				return err
			}
			c.AddTask(t)
		}
	}

	return nil
}

// IAMServiceEC2 returns the name of the IAM service for EC2 in the current region.
// It is ec2.amazonaws.com everywhere but in cn-north / cn-northwest, where it is ec2.amazonaws.com.cn
func IAMServiceEC2(region string) string {
	switch region {
	case "cn-north-1":
		return "ec2.amazonaws.com.cn"
	case "cn-northwest-1":
		return "ec2.amazonaws.com.cn"
	default:
		return "ec2.amazonaws.com"
	}
}

// buildAWSIAMRolePolicy produces the AWS IAM role policy for the given role.
func (b *IAMModelBuilder) buildAWSIAMRolePolicy(role iam.Subject) (fi.Resource, error) {
	var policy string
	serviceAccount, ok := role.ServiceAccount()
	if ok {
		serviceAccountIssuer, err := iam.ServiceAccountIssuer(b.ClusterName(), &b.Cluster.Spec)
		if err != nil {
			return nil, err
		}
		oidcProvider := strings.TrimPrefix(serviceAccountIssuer, "https://")

		iamPolicy := &iam.Policy{
			Version: iam.PolicyDefaultVersion,
			Statement: []*iam.Statement{
				{
					Effect: "Allow",
					Principal: iam.Principal{
						Federated: "arn:aws:iam::" + b.AWSAccountID + ":oidc-provider/" + oidcProvider,
					},
					Action: stringorslice.String("sts:AssumeRoleWithWebIdentity"),
					Condition: map[string]interface{}{
						"StringEquals": map[string]interface{}{
							oidcProvider + ":sub": "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name,
						},
					},
				},
			},
		}
		s, err := iamPolicy.AsJSON()
		if err != nil {
			return nil, err
		}
		policy = s
	} else {
		// We don't generate using json.Marshal here, it would create whitespace changes in the policy for existing clusters.

		policy = strings.ReplaceAll(NodeRolePolicyTemplate, "{{ IAMServiceEC2 }}", IAMServiceEC2(b.Region))
	}

	return fi.NewStringResource(policy), nil
}
