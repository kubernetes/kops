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
	"text/template"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/iam"
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

const PodRolePolicyTemplate = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Federated": "{{ FederatedPrincipal }}" },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": { "{{OIDCProvider}}:sub": "{{ OIDCSub }}" }
      }
    }
  ]
}`

func (b *IAMModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// Collect managed Instance Group roles
	managedRoles := make(map[kops.InstanceGroupRole]bool)

	// Collect Instance Profile ARNs and their associated Instance Group roles
	sharedProfileARNsToIGRole := make(map[string]kops.InstanceGroupRole)
	for _, ig := range b.InstanceGroups {
		if ig.Spec.IAM != nil && ig.Spec.IAM.Profile != nil {
			specProfile := fi.StringValue(ig.Spec.IAM.Profile)
			if matchingRole, ok := sharedProfileARNsToIGRole[specProfile]; ok {
				if matchingRole != ig.Spec.Role {
					return fmt.Errorf("found IAM instance profile assigned to multiple Instance Group roles %v and %v: %v",
						ig.Spec.Role, sharedProfileARNsToIGRole[specProfile], specProfile)
				}
			} else {
				sharedProfileARNsToIGRole[specProfile] = ig.Spec.Role
			}
		} else {
			managedRoles[ig.Spec.Role] = true
		}
	}

	// Generate IAM tasks for each shared role
	for profileARN, igRole := range sharedProfileARNsToIGRole {
		role := iam.PodOrNodeRole{NodeRole: igRole}

		iamName, err := findCustomAuthNameFromArn(profileARN)
		if err != nil {
			return fmt.Errorf("unable to parse instance profile name from arn %q: %v", profileARN, err)
		}
		err = b.buildIAMTasks(role, iamName, c, true)
		if err != nil {
			return err
		}
	}

	// Generate IAM tasks for each managed role
	for igRole := range managedRoles {
		role := iam.PodOrNodeRole{NodeRole: igRole}
		iamName := b.IAMName(igRole)
		err := b.buildIAMTasks(role, iamName, c, false)
		if err != nil {
			return err
		}
	}

	// Generate IAM tasks for pod roles
	if featureflag.UsePodIAM.Enabled() {
		podRoles := []iam.PodRole{iam.PodRoleKopsController}
		for _, podRole := range podRoles {
			role := iam.PodOrNodeRole{PodRole: podRole}

			iamName := b.IAMNameForPodRole(podRole)
			err := b.buildIAMTasks(role, iamName, c, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *IAMModelBuilder) buildIAMTasks(role iam.PodOrNodeRole, iamName string, c *fi.ModelBuilderContext, shared bool) error {
	var roleKey string
	if role.PodRole != iam.PodRoleEmpty {
		roleKey = strings.ToLower(string(role.PodRole))
	} else {
		roleKey = strings.ToLower(string(role.NodeRole) + "s")
	}

	{ // To minimize diff for easier code review
		var iamRole *awstasks.IAMRole
		{
			rolePolicy, err := b.buildAWSIAMRolePolicy(role)
			if err != nil {
				return err
			}

			{
				policy, err := fi.ResourceAsString(rolePolicy)
				klog.Infof("policy %s err %v", policy, err)
			}
			
			iamRole = &awstasks.IAMRole{
				Name:      s(iamName),
				Lifecycle: b.Lifecycle,

				RolePolicyDocument: fi.WrapResource(rolePolicy),
				ExportWithID:       s(roleKey + "s"),
			}
			c.AddTask(iamRole)

		}

		{
			iamPolicy := &iam.PolicyResource{
				Builder: &iam.PolicyBuilder{
					Cluster: b.Cluster,
					Role:    role,
					Region:  b.Region,
				},
			}

			// This is slightly tricky; we need to know the hosted zone id,
			// but we might be creating the hosted zone dynamically.

			// TODO: I don't love this technique for finding the task by name & modifying it
			dnsZoneTask, found := c.Tasks["DNSZone/"+b.NameForDNSZone()]
			if found {
				iamPolicy.DNSZone = dnsZoneTask.(*awstasks.DNSZone)
			} else {
				klog.V(2).Infof("Task %q not found; won't set route53 permissions in IAM", "DNSZone/"+b.NameForDNSZone())
			}

			t := &awstasks.IAMRolePolicy{
				Name:      s(iamName),
				Lifecycle: b.Lifecycle,

				Role:           iamRole,
				PolicyDocument: iamPolicy,
			}
			c.AddTask(t)
		}

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
				Name:      s(iamName),
				Lifecycle: b.Lifecycle,

				InstanceProfile: iamInstanceProfile,
				Role:            iamRole,
			}
			c.AddTask(iamInstanceProfileRole)
		}

		// Create External Policy tasks
		{
			var externalPolicies []string

			if b.Cluster.Spec.ExternalPolicies != nil {
				p := *(b.Cluster.Spec.ExternalPolicies)
				externalPolicies = append(externalPolicies, p[roleKey]...)
			}

			name := fmt.Sprintf("%s-policyoverride", roleKey)
			t := &awstasks.IAMRolePolicy{
				Name:             s(name),
				Lifecycle:        b.Lifecycle,
				Role:             iamRole,
				Managed:          true,
				ExternalPolicies: &externalPolicies,
			}

			c.AddTask(t)
		}

		// Generate additional policies if needed, and attach to existing role
		{
			additionalPolicy := ""
			if b.Cluster.Spec.AdditionalPolicies != nil {
				additionalPolicies := *(b.Cluster.Spec.AdditionalPolicies)

				additionalPolicy = additionalPolicies[roleKey]
			}

			additionalPolicyName := "additional." + iamName

			t := &awstasks.IAMRolePolicy{
				Name:      s(additionalPolicyName),
				Lifecycle: b.Lifecycle,

				Role: iamRole,
			}

			if additionalPolicy != "" {
				p := &iam.Policy{
					Version: iam.PolicyDefaultVersion,
				}

				statements, err := iam.ParseStatements(additionalPolicy)
				if err != nil {
					return fmt.Errorf("additionalPolicy %q is invalid: %v", roleKey, err)
				}

				p.Statement = append(p.Statement, statements...)

				policy, err := p.AsJSON()
				if err != nil {
					return fmt.Errorf("error building IAM policy: %v", err)
				}

				t.PolicyDocument = fi.WrapResource(fi.NewStringResource(policy))
			} else {
				t.PolicyDocument = fi.WrapResource(fi.NewStringResource(""))
			}

			c.AddTask(t)
		}
	}

	return nil
}

// buildAWSIAMRolePolicy produces the AWS IAM role policy for the given role
func (b *IAMModelBuilder) buildAWSIAMRolePolicy(role iam.PodOrNodeRole) (fi.Resource, error) {
	functions := template.FuncMap{
		"IAMServiceEC2": func() string {
			// IAMServiceEC2 returns the name of the IAM service for EC2 in the current region
			// it is ec2.amazonaws.com everywhere but in cn-north, where it is ec2.amazonaws.com.cn
			switch b.Region {
			case "cn-north-1":
				return "ec2.amazonaws.com.cn"
			case "cn-northwest-1":
				return "ec2.amazonaws.com.cn"
			default:
				return "ec2.amazonaws.com"
			}
		},
	}

	var template string
	if role.PodRole != iam.PodRoleEmpty {
		serviceAccount := iam.ServiceAccountForPodRole(role.PodRole)

		serviceAccountIssuer, err := iam.ServiceAccountIssuer(b.Cluster.ClusterName, &b.Cluster.Spec)
		if err != nil {
			return nil, err
		}
		oidcProvider := strings.TrimPrefix(serviceAccountIssuer, "https://")

		functions["OIDCProvider"] = func() string { return oidcProvider }
		functions["OIDCSub"] = func() string { return "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name }
		functions["FederatedPrincipal"] = func() string { return "arn:aws:iam::" + b.AWSAccountID + ":oidc-provider/" + oidcProvider }

		template = PodRolePolicyTemplate
	} else {
		template = NodeRolePolicyTemplate
	}

	templateResource, err := NewTemplateResource("AWSIAMRolePolicy", template, functions, nil)
	if err != nil {
		return nil, err
	}
	return templateResource, nil
}
