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

package awsmodel

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	awsIam "github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// IAMModelBuilder configures IAM objects
type IAMModelBuilder struct {
	*AWSModelContext
	Lifecycle fi.Lifecycle
	Cluster   *kops.Cluster
}

var (
	_ fi.ModelBuilder = &IAMModelBuilder{}
	_ fi.HasDeletions = &IAMModelBuilder{}
)

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
		lchPermissions := false
		defaultWarmPool := b.Cluster.Spec.WarmPool
		for _, ig := range b.InstanceGroups {
			warmPool := defaultWarmPool.ResolveDefaults(ig)
			if ig.Spec.Role == igRole && warmPool.IsEnabled() && warmPool.EnableLifecycleHook {
				lchPermissions = true
				break

			}
		}
		role, err := iam.BuildNodeRoleSubject(igRole, lchPermissions)
		if err != nil {
			return err
		}

		iamName, err := model.FindCustomAuthNameFromArn(profileARN)
		if err != nil {
			return fmt.Errorf("unable to parse instance profile name from arn %q: %v", profileARN, err)
		}
		err = b.buildIAMTasks(role, iamName, c, true)
		if err != nil {
			return err
		}
	}

	// Generate IAM tasks for each managed role
	defaultWarmPool := b.Cluster.Spec.WarmPool
	for igRole := range managedRoles {
		haveWarmPool := false
		for _, ig := range b.InstanceGroups {
			warmPool := defaultWarmPool.ResolveDefaults(ig)
			if ig.Spec.Role == igRole && warmPool.IsEnabled() && warmPool.EnableLifecycleHook {
				haveWarmPool = true
				break

			}
		}
		role, err := iam.BuildNodeRoleSubject(igRole, haveWarmPool)
		if err != nil {
			return err
		}

		iamName := b.IAMName(igRole)
		if err := b.buildIAMTasks(role, iamName, c, false); err != nil {
			return err
		}
	}

	iamSpec := b.Cluster.Spec.IAM
	if iamSpec != nil {
		for _, sa := range iamSpec.ServiceAccountExternalPermissions {
			var p *iam.Policy
			aws := sa.AWS
			if aws.InlinePolicy != "" {
				bp, err := b.buildPolicy(aws.InlinePolicy)
				p = bp
				if err != nil {
					return fmt.Errorf("error inline policy: %w", err)
				}
			}
			serviceAccount := &iam.GenericServiceAccount{
				NamespacedName: types.NamespacedName{
					Name:      sa.Name,
					Namespace: sa.Namespace,
				},
				Policy: p,
			}
			iamRole, err := b.BuildServiceAccountRoleTasks(serviceAccount, c)
			if err != nil {
				return fmt.Errorf("error building service account role tasks: %w", err)
			}
			if len(aws.PolicyARNs) > 0 {
				name := "external-" + fi.StringValue(iamRole.Name)
				externalPolicies := aws.PolicyARNs
				c.AddTask(&awstasks.IAMRolePolicy{
					Name:             fi.String(name),
					ExternalPolicies: &externalPolicies,
					Managed:          true,
					Role:             iamRole,
					Lifecycle:        b.Lifecycle,
				})
			}
		}
	}

	return nil
}

// BuildServiceAccountRoleTasks build tasks specifically for the ServiceAccount role.
func (b *IAMModelBuilder) BuildServiceAccountRoleTasks(role iam.Subject, c *fi.ModelBuilderContext) (*awstasks.IAMRole, error) {
	iamName, err := b.IAMNameForServiceAccountRole(role)
	if err != nil {
		return nil, err
	}

	iamRole, err := b.buildIAMRole(role, iamName, c)
	if err != nil {
		return nil, err
	}

	if err := b.buildIAMRolePolicy(role, iamName, iamRole, c); err != nil {
		return nil, err
	}

	return iamRole, nil
}

func (b *IAMModelBuilder) buildIAMRole(role iam.Subject, iamName string, c *fi.ModelBuilderContext) (*awstasks.IAMRole, error) {
	roleKey, isServiceAccount := b.roleKey(role)

	rolePolicy, err := b.buildAWSIAMRolePolicy(role)
	if err != nil {
		return nil, err
	}

	iamRole := &awstasks.IAMRole{
		Name:      fi.String(iamName),
		Lifecycle: b.Lifecycle,

		RolePolicyDocument: rolePolicy,
	}

	if isServiceAccount {
		// e.g. kube-system-dns-controller
		iamRole.ExportWithID = fi.String(roleKey)
		sa, ok := role.ServiceAccount()
		if ok {
			iamRole.Tags = b.CloudTagsForServiceAccount(iamName, sa)
		}
	} else {
		// e.g. nodes
		iamRole.ExportWithID = fi.String(roleKey + "s")
		iamRole.Tags = b.CloudTags(iamName, false)
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
			Cluster:                               b.Cluster,
			Role:                                  role,
			Region:                                b.Region,
			Partition:                             b.AWSPartition,
			UseServiceAccountExternalPermisssions: b.UseServiceAccountExternalPermissions(),
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
		Name:      fi.String(iamName),
		Lifecycle: b.Lifecycle,

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
		return strings.ToLower(strings.ReplaceAll(serviceAccount.Namespace+"-"+serviceAccount.Name, "*", "wildcard")), true
	}

	// This isn't great, but we have to be backwards compatible with the old names.
	switch role.(type) {
	case *iam.NodeRoleMaster:
		return strings.ToLower(string(kops.InstanceGroupRoleMaster)), false
	case *iam.NodeRoleAPIServer:
		return strings.ToLower(string(kops.InstanceGroupRoleAPIServer)), false
	case *iam.NodeRoleNode:
		return strings.ToLower(string(kops.InstanceGroupRoleNode)), false
	case *iam.NodeRoleBastion:
		return strings.ToLower(string(kops.InstanceGroupRoleBastion)), false

	default:
		klog.Fatalf("unknown node role type: %T", role)
		return "", false
	}
}

func (b *IAMModelBuilder) buildIAMTasks(role iam.Subject, iamName string, c *fi.ModelBuilderContext, shared bool) error {
	roleKey, _ := b.roleKey(role)

	{
		// To minimize diff for easier code review

		var iamInstanceProfile *awstasks.IAMInstanceProfile
		{
			iamInstanceProfile = &awstasks.IAMInstanceProfile{
				Name:      fi.String(iamName),
				Lifecycle: b.Lifecycle,
				Shared:    fi.Bool(shared),
				Tags:      b.CloudTags(iamName, shared),
			}
			c.AddTask(iamInstanceProfile)
		}

		if !shared {

			// Create External Policy tasks
			iamRole, err := b.buildIAMRole(role, iamName, c)
			if err != nil {
				return err
			}

			{
				if err := b.buildIAMRolePolicy(role, iamName, iamRole, c); err != nil {
					return err
				}
				{
					iamInstanceProfileRole := &awstasks.IAMInstanceProfileRole{
						Name:      fi.String(iamName),
						Lifecycle: b.Lifecycle,

						InstanceProfile: iamInstanceProfile,
						Role:            iamRole,
					}
					c.AddTask(iamInstanceProfileRole)
				}

				var externalPolicies []string

				if b.Cluster.Spec.ExternalPolicies != nil {
					p := *(b.Cluster.Spec.ExternalPolicies)
					externalPolicies = append(externalPolicies, p[roleKey]...)
				}
				sort.Strings(externalPolicies)

				name := fmt.Sprintf("%s-policyoverride", roleKey)
				t := &awstasks.IAMRolePolicy{
					Name:             fi.String(name),
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
					Name:      fi.String(additionalPolicyName),
					Lifecycle: b.Lifecycle,

					Role: iamRole,
				}

				if additionalPolicy != "" {
					p, err := b.buildPolicy(additionalPolicy)
					if err != nil {
						return fmt.Errorf("additionalPolicy %q is invalid: %v", roleKey, err)
					}

					policy, err := p.AsJSON()
					if err != nil {
						return fmt.Errorf("error building IAM policy: %w", err)
					}

					t.PolicyDocument = fi.NewStringResource(policy)
				} else {
					t.PolicyDocument = fi.NewStringResource("")
				}

				c.AddTask(t)
			}
		}
	}

	return nil
}

func (b *IAMModelBuilder) buildPolicy(policyString string) (*iam.Policy, error) {
	p := &iam.Policy{
		Version: iam.PolicyDefaultVersion,
	}

	statements, err := iam.ParseStatements(policyString)
	if err != nil {
		return nil, err
	}

	p.Statement = append(p.Statement, statements...)
	return p, nil
}

// IAMServiceEC2 returns the name of the IAM service for EC2 in the current region.
// It is ec2.amazonaws.com in the default aws partition, but different in other isolated/custom partitions
func IAMServiceEC2(region string) string {
	partitions := endpoints.DefaultPartitions()
	for _, p := range partitions {
		if _, ok := p.Regions()[region]; ok {
			ep := "ec2." + p.DNSSuffix()
			return ep
		}
	}
	return "ec2.amazonaws.com"
}

func formatAWSIAMStatement(accountId, partition, oidcProvider, namespace, name string) (*iam.Statement, error) {
	// disallow wildcard in the service account name
	if strings.Contains(name, "*") {
		return nil, fmt.Errorf("service account name cannot contain a wildcard %s", name)
	}

	// if the namespace contains a wildcard, use StringLike condition instead of StringEquals
	condition := "StringEquals"
	if strings.Contains(namespace, "*") {
		condition = "StringLike"
	}

	return &iam.Statement{
			Effect: "Allow",
			Principal: iam.Principal{
				Federated: "arn:" + partition + ":iam::" + accountId + ":oidc-provider/" + oidcProvider,
			},
			Action: stringorslice.String("sts:AssumeRoleWithWebIdentity"),
			Condition: map[string]interface{}{
				condition: map[string]interface{}{
					oidcProvider + ":sub": "system:serviceaccount:" + namespace + ":" + name,
				},
			},
		},
		nil
}

// buildAWSIAMRolePolicy produces the AWS IAM role policy for the given role.
func (b *IAMModelBuilder) buildAWSIAMRolePolicy(role iam.Subject) (fi.Resource, error) {
	var policy string
	serviceAccount, ok := role.ServiceAccount()
	if ok {
		oidcProvider := strings.TrimPrefix(*b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer, "https://")

		statement, err := formatAWSIAMStatement(b.AWSAccountID, b.AWSPartition, oidcProvider, serviceAccount.Namespace, serviceAccount.Name)
		if err != nil {
			return nil, err
		}

		iamPolicy := &iam.Policy{
			Version:   iam.PolicyDefaultVersion,
			Statement: []*iam.Statement{statement},
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

func (b *IAMModelBuilder) FindDeletions(context *fi.ModelBuilderContext, cloud fi.Cloud) error {
	iamapi := cloud.(awsup.AWSCloud).IAM()
	ownershipTag := "kubernetes.io/cluster/" + b.Cluster.ObjectMeta.Name
	request := &awsIam.ListRolesInput{}
	var getRoleErr error
	err := iamapi.ListRolesPages(request, func(p *awsIam.ListRolesOutput, lastPage bool) bool {
		for _, role := range p.Roles {
			if !strings.HasSuffix(fi.StringValue(role.RoleName), "."+b.Cluster.ObjectMeta.Name) {
				continue
			}
			getRequest := &awsIam.GetRoleInput{RoleName: role.RoleName}
			roleOutput, err := iamapi.GetRole(getRequest)
			if err != nil {
				getRoleErr = fmt.Errorf("calling IAM GetRole on %s: %w", fi.StringValue(role.RoleName), err)
				return false
			}
			for _, tag := range roleOutput.Role.Tags {
				if fi.StringValue(tag.Key) == ownershipTag && fi.StringValue(tag.Value) == "owned" {
					if _, ok := context.Tasks["IAMRole/"+fi.StringValue(role.RoleName)]; !ok {
						context.AddTask(&awstasks.IAMRole{
							ID:        role.RoleId,
							Name:      role.RoleName,
							Lifecycle: b.Lifecycle,
						})
					}
				}
			}
		}
		return true
	})
	if getRoleErr != nil {
		return getRoleErr
	}
	if err != nil {
		return fmt.Errorf("listing IAM roles: %w", err)
	}
	return nil
}
