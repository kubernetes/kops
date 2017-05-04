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

// TODO: We have a couple different code paths until with do lifecycles, and
// TODO: when we have a cluster or refactor some s3 code.  The only code that
// TODO: is not shared by the different path is the s3 / state store stuff

// TODO: We may want to look at https://aws.amazon.com/blogs/security/how-to-help-lock-down-a-users-amazon-ec2-capabilities-to-a-single-vpc/
// TODO: But that gets complicated fast.  I would like to lock the policy down to a single VPC.

package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/util/pkg/vfs"
)

const IAMPolicyDefaultVersion = "2012-10-17"

type IAMPolicy struct {
	Version   string
	Statement []*IAMStatement
}

func (p *IAMPolicy) AsJSON() (string, error) {
	j, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling policy to JSON: %v", err)
	}
	return string(j), nil
}

type IAMStatementEffect string
type IAMSid string

const IAMStatementEffectAllow IAMStatementEffect = "Allow"
const IAMStatementEffectDeny IAMStatementEffect = "Deny"

type IAMStatement struct {
	Effect   IAMStatementEffect
	Action   stringorslice.StringOrSlice
	Resource stringorslice.StringOrSlice
	Sid      IAMSid
}

func (l *IAMStatement) Equal(r *IAMStatement) bool {
	if l.Effect != r.Effect {
		return false
	}
	if !l.Action.Equal(r.Action) {
		return false
	}
	if !l.Resource.Equal(r.Resource) {
		return false
	}
	return true
}

type IAMPolicyBuilder struct {
	Cluster               *api.Cluster
	ClusterName           string
	Role                  api.InstanceGroupRole
	Region                string
	HostedZoneID          string
	S3Bucket              string
	ResourceARN           *string
	CreateECRPerms        bool
	CreateNetworkingPerms bool
	CloudFormationPerms   bool
	CreatePolicyPerms     bool
	KMSKeys               []string
}

// BuildAWSIAMPolicy generates the IAM policies for a bastion, node or master
func (b *IAMPolicyBuilder) BuildAWSIAMPolicy() (*IAMPolicy, error) {
	resource := createResource(b)

	iamPrefix := b.IAMPrefix()

	p := &IAMPolicy{
		Version: IAMPolicyDefaultVersion,
	}

	// Don't give bastions any permissions (yet)
	if b.Role == api.InstanceGroupRoleBastion {
		p.Statement = append(p.Statement, &IAMStatement{
			// We grant a trivial (?) permission (DescribeRegions), because empty policies are not allowed
			Sid:      "kopsK8sBastion",
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"ec2:DescribeRegions"}),
			Resource: resource,
		})

		return p, nil
	}

	if b.Role == api.InstanceGroupRoleNode {
		addNodeEC2Policies(p, resource)
	}

	addECRPermissions(p)

	if b.Role == api.InstanceGroupRoleMaster {

		addMasterEC2Policies(p, resource)

		addMasterELBPolicies(p, resource)

		addMasterASPolicies(p, resource)

		addCertIAMPolicies(p, resource)

		// Restrict the KMS permissions to only the keys that are being used
		kmsKeyIDs := sets.NewString()
		for _, e := range b.Cluster.Spec.EtcdClusters {
			for _, m := range e.Members {
				if m.KmsKeyId != nil {
					kmsKeyIDs.Insert(*m.KmsKeyId)
				}
			}
		}

		if kmsKeyIDs.Len() > 0 {
			addKMSIAMPolicies(p, stringorslice.Slice(kmsKeyIDs.List()))
		}

		if b.HostedZoneID != "" {
			addRoute53Permissions(p, b.HostedZoneID)
		}
	}

	// For S3 IAM permissions, we grant permissions to subtrees.  So find the parents;
	// we don't need to grant mypath and mypath/child.
	var roots []string
	{
		var locations []string

		for _, p := range []string{
			b.Cluster.Spec.KeyStore,
			b.Cluster.Spec.SecretStore,
			b.Cluster.Spec.ConfigStore,
		} {
			if p == "" {
				continue
			}

			if !strings.HasSuffix(p, "/") {
				p = p + "/"
			}
			locations = append(locations, p)
		}

		for i, l := range locations {
			isTopLevel := true
			for j := range locations {
				if i == j {
					continue
				}
				if strings.HasPrefix(l, locations[j]) {
					glog.V(4).Infof("Ignoring location %q because found parent %q", l, locations[j])
					isTopLevel = false
				}
			}
			if isTopLevel {
				glog.V(4).Infof("Found root location %q", l)
				roots = append(roots, l)
			}
		}
	}

	for _, root := range roots {
		vfsPath, err := vfs.Context.BuildVfsPath(root)
		if err != nil {
			return nil, fmt.Errorf("cannot parse VFS path %q: %v", root, err)
		}

		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			// Note that the config store may itself be a subdirectory of a bucket
			iamS3Path := s3Path.Bucket() + "/" + s3Path.Key()
			iamS3Path = strings.TrimSuffix(iamS3Path, "/")

			p.Statement = append(p.Statement, &IAMStatement{
				Sid:    "kopsK8sStateStoreAccess",
				Effect: IAMStatementEffectAllow,
				Action: stringorslice.Of(
					"s3:GetObject",
					"s3:ListObject",
				),
				Resource: stringorslice.Of(
					iamPrefix+":s3:::"+iamS3Path,
					iamPrefix+":s3:::"+iamS3Path+"/*",
				),
			})

			p.Statement = append(p.Statement, &IAMStatement{
				Sid:    "kopsK8sStateStoreAccessList",
				Effect: IAMStatementEffectAllow,
				Action: stringorslice.Of("s3:GetBucketLocation", "s3:ListBucket"),
				Resource: stringorslice.Slice([]string{
					iamPrefix + ":s3:::" + s3Path.Bucket(),
				}),
			})
		} else if _, ok := vfsPath.(*vfs.MemFSPath); ok {
			// Tests -ignore - nothing we can do in terms of IAM policy
			glog.Warningf("ignoring memfs path %q for IAM policy builder", vfsPath)
		} else {
			// We could implement this approach, but it seems better to get all clouds using cluster-readable storage
			return nil, fmt.Errorf("path is not cluster readable: %v", root)
		}
	}

	return p, nil
}

// BuildAWSIAMPolicyNode generates a custom policy for a Kubernetes master.
func (b *IAMPolicyBuilder) BuildAWSIAMPolicyMaster() (*IAMPolicy, error) {

	resource := createResource(b)

	p := &IAMPolicy{
		Version: IAMPolicyDefaultVersion,
	}

	addMasterEC2Policies(p, resource)

	addMasterELBPolicies(p, resource)

	addMasterASPolicies(p, resource)

	addCertIAMPolicies(p, resource)

	b.addIAMCustomPolicies(p, resource)

	if b.KMSKeys != nil && len(b.KMSKeys) != 0 {
		addKMSIAMPolicies(p, stringorslice.Slice(b.KMSKeys))
	}

	// We provide ECR access on the nodes (naturally), but we also provide access on the master.
	// We shouldn't be running lots of pods on the master, but it is perfectly reasonable to run
	// a private logging pod or similar.
	if b.CreateECRPerms {
		addECRPermissions(p)
	}

	if b.HostedZoneID != "" {
		addRoute53Permissions(p, b.HostedZoneID)
	}

	return p, nil

}

// BuildAWSIAMPolicyNode generates a custom policy for a Kubernetes node.
func (b *IAMPolicyBuilder) BuildAWSIAMPolicyNode() (*IAMPolicy, error) {
	resource := createResource(b)

	p := &IAMPolicy{
		Version: IAMPolicyDefaultVersion,
	}

	if b.CreateECRPerms {
		addECRPermissions(p)
	}

	addNodeEC2Policies(p, resource)

	b.addIAMCustomPolicies(p, resource)

	return p, nil

}

// BuildAWSIAMPolicyInstaller generates an installer/admin policy that can be used to execute kops.
// This method is used to create an example user for users who want a specific user for kops.
func (b *IAMPolicyBuilder) BuildAWSIAMPolicyInstaller() (*IAMPolicy, error) {

	resource := createResource(b)
	iamPrefix := b.IAMPrefix()

	p := &IAMPolicy{
		Version: IAMPolicyDefaultVersion,
	}
	if b.CreateNetworkingPerms {
		p.Statement = append(p.Statement, &IAMStatement{
			Sid:    "kopsAdminVpcNetworking",
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Of(
				"ec2:AllocateAddress",
				"ec2:AssociateAddress",
				"ec2:AssociateDhcpOptions",
				"ec2:AssociateRouteTable",
				"ec2:AttachInternetGateway",
				"ec2:CreateDhcpOptions",
				"ec2:CreateInternetGateway",
				"ec2:CreateNatGateway",
				"ec2:CreateRoute",
				"ec2:CreateRouteTable",
				"ec2:CreateSubnet",
				"ec2:CreateVpc",
				"ec2:DeleteDhcpOptions",
				"ec2:DeleteInternetGateway",
				"ec2:DeleteNatGateway",
				"ec2:DeleteRoute",
				"ec2:DeleteRouteTable",
				"ec2:DetachInternetGateway",
				"ec2:DisassociateRouteTable",
				"ec2:ModifyVpcAttribute",
				"ec2:ReplaceRoute",
			),
			Resource: resource,
		})
	}

	// ec2
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminEc2",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"ec2:AttachVolume",
			"ec2:AuthorizeSecurityGroupEgress",
			"ec2:AuthorizeSecurityGroupIngress",
			"ec2:CreateSecurityGroup",
			"ec2:CreateTags",
			"ec2:CreateVolume",
			"ec2:DeleteKeyPair",
			"ec2:DeleteSecurityGroup",
			"ec2:DeleteTags",
			"ec2:DeleteVolume",
			"ec2:DescribeAddresses",
			"ec2:DescribeAvailabilityZones",
			"ec2:DescribeDhcpOptions",
			"ec2:DescribeHosts",
			"ec2:DescribeImages",
			"ec2:DescribeImageAttributes",
			"ec2:DescribeInstances",
			"ec2:DescribeInternetGateways",
			"ec2:DescribeKeyPairs",
			"ec2:DescribeNatGateways",
			"ec2:DescribeRegions",
			"ec2:DescribeRouteTables",
			"ec2:DescribeSecurityGroups",
			"ec2:DescribeSubnets",
			"ec2:DescribeTags",
			"ec2:DescribeVolumes",
			"ec2:DescribeVpcAttribute",
			"ec2:DescribeVpc",
			"ec2:DescribeVpcs",
			"ec2:DetachVolume",
			"ec2:ImportKeyPair",
			"ec2:RevokeSecurityGroupEgress",
			"ec2:RevokeSecurityGroupIngress",
			"ec2:StopInstances",
			"ec2:TerminateInstances",
		),
		Resource: resource,
	})

	// elasticloadbalancing
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminElb",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"elasticloadbalancing:AddTags",
			"elasticloadbalancing:ApplySec*",
			"elasticloadbalancing:ConfigureHealthCheck",
			"elasticloadbalancing:CreateLoadBalancer",
			"elasticloadbalancing:CreateLoadBalancerListeners",
			"elasticloadbalancing:DescribeLoadBalancers",
			"elasticloadbalancing:DescribeTags",
			"elasticloadbalancing:DeleteLoadBalancer",
			"elasticloadbalancing:DescribeLoadBalancerAttributes",
			"elasticloadbalancing:ModifyLoadBalancerAttributes",
			"elasticloadbalancing:RegisterInstancesWithLoadBalancer",
			"elasticloadbalancing:RemoveTags",
			"elasticloadbalancing:SetSecurityGroups",
		),
		Resource: resource,
	})

	// autoscaling
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminAsg",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"autoscaling:AttachInstances",
			"autoscaling:AttachLoadBalancers",
			"autoscaling:CreateAutoScalingGroup",
			"autoscaling:CreateLaunchConfiguration",
			"autoscaling:CreateOrUpdateTags",
			"autoscaling:DeleteAutoScalingGroup",
			"autoscaling:DeleteLaunchConfiguration",
			"autoscaling:DeleteTags",
			"autoscaling:Describe*",
			"autoscaling:SetDesiredCapacity",
			"autoscaling:TerminateInstanceInAutoScalingGroup",
			"autoscaling:UpdateAutoScalingGroup",
		),
		Resource: resource,
	})

	if b.CloudFormationPerms {
		p.Statement = append(p.Statement, &IAMStatement{
			Sid:    "kopsAdminCF",
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Of(
				"cloudformation:*",
			),
			Resource: resource,
		})
	}

	// Installer does not need role policy information if reusing policies
	if b.CreatePolicyPerms {
		p.Statement = append(p.Statement, &IAMStatement{
			Sid:    "kopsAdminRolePolicies",
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Of(
				"iam:DeleteRolePolicy",
				"iam:PutRolePolicy",
			),
			Resource: resource,
		})
	}

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminIAM",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"iam:AddRoleToInstanceProfile",
			"iam:CreateInstanceProfile",
			"iam:CreateRole",
			"iam:DeleteInstanceProfile",
			"iam:DeleteRole",
			"iam:GetRole",
			"iam:GetRolePolicy",
			"iam:GetInstanceProfile",
			"iam:ListInstanceProfiles",
			"iam:ListRolePolicies",
			"iam:ListRoles",
			"iam:ListInstanceProfiles",
			"iam:PassRole",
			"iam:RemoveRoleFromInstanceProfile",
			"iam:UpdateAssumeRolePolicy",
		),
		Resource: resource,
	})

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminKMS",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"kms:Encrypt",
			"kms:Decrypt",
			"kms:ReEncrypt*",
			"kms:GenerateDataKey*",
			"kms:DescribeKey",
			"kms:CreateGrant",
			"kms:ListGrants",
			"kms:RevokeGrant",
		),
		Resource: resource,
	})

	if b.HostedZoneID != "" {
		addRoute53Permissions(p, b.HostedZoneID)
	}

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsAdminStatestoreAccess",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"s3:PutObject",
			"s3:CreateBucket",
			"s3:DeleteBucket",
			"s3:DeleteObject",
			"s3:GetObject",
			"s3:GetBucketLocation",
			"s3:ListBucket",
		),
		Resource: stringorslice.Of(
			iamPrefix+":s3:::"+b.S3Bucket,
			iamPrefix+":s3:::"+b.S3Bucket+"/*",
		),
	})

	return p, nil
}

// IAMPrefix returns the prefix for AWS ARNs in the current region, for use with IAM
// it is arn:aws everywhere but in cn-north, where it is arn:aws-cn
func (b *IAMPolicyBuilder) IAMPrefix() string {
	switch b.Region {
	case "cn-north-1":
		return "arn:aws-cn"
	default:
		return "arn:aws"
	}
}

func (b *IAMPolicyBuilder) addIAMCustomPolicies(p *IAMPolicy, resource stringorslice.StringOrSlice) {

	// For S3 IAM permissions, we grant permissions to subtrees.  So find the parents;
	// we don't need to grant mypath and mypath/child.

	iamPrefix := b.IAMPrefix()

	bucket := iamPrefix + ":s3:::" + b.S3Bucket + "/" + b.ClusterName
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsK8sStatestoreAccessBucket",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"s3:GetBucket",
			"s3:ListBucket",
		),
		Resource: stringorslice.Of(
			bucket,
			bucket+"/*",
		),
	})

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "KopsK8sStatestoreAccessGet",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of("s3:GetBucketLocation", "s3:ListBucket"),
		Resource: stringorslice.Slice([]string{
			iamPrefix + ":s3:::" + b.S3Bucket,
		}),
	})

}

type IAMPolicyResource struct {
	Builder *IAMPolicyBuilder
	DNSZone *awstasks.DNSZone
}

var _ fi.Resource = &IAMPolicyResource{}
var _ fi.HasDependencies = &IAMPolicyResource{}

func (b *IAMPolicyResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return []fi.Task{b.DNSZone}
}

// Open produces the AWS IAM policy for the given role
func (b *IAMPolicyResource) Open() (io.Reader, error) {
	// Defensive copy before mutation
	pb := *b.Builder

	if b.DNSZone != nil {
		hostedZoneID := fi.StringValue(b.DNSZone.ZoneID)
		if hostedZoneID == "" {
			// Dependency analysis failure?
			return nil, fmt.Errorf("DNS ZoneID not set")
		}
		pb.HostedZoneID = hostedZoneID
	}

	policy, err := pb.BuildAWSIAMPolicy()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	j, err := policy.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	return bytes.NewReader([]byte(j)), nil
}

func addECRPermissions(p *IAMPolicy) {
	// TODO - I think we can just have GetAuthorizationToken here, as we are not
	// TODO - making any API calls except for GetAuthorizationToken.

	// We provide ECR access on the nodes (naturally), but we also provide access on the master.
	// We shouldn't be running lots of pods on the master, but it is perfectly reasonable to run
	// a private logging pod or similar.
	// At this point we allow all regions with ECR, since ECR is region specific.
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsK8sECR",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"ecr:GetAuthorizationToken",
			"ecr:BatchCheckLayerAvailability",
			"ecr:GetDownloadUrlForLayer",
			"ecr:GetRepositoryPolicy",
			"ecr:DescribeRepositories",
			"ecr:ListImages",
			"ecr:BatchGetImage",
		),
		Resource: stringorslice.Slice([]string{"*"}),
	})
}

func addRoute53Permissions(p *IAMPolicy, hostedZoneID string) {

	// TODO we should test if we are in China, and not just return
	// TODO no Route53 in China

	// Remove /hostedzone/ prefix (if present)
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "/")
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "hostedzone/")

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsK8sRoute53Change",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of("route53:ChangeResourceRecordSets",
			"route53:ListResourceRecordSets",
			"route53:GetHostedZone"),
		Resource: stringorslice.Slice([]string{"arn:aws:route53:::hostedzone/" + hostedZoneID}),
	})

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:      "kopsK8sRoute53GetChanges",
		Effect:   IAMStatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:GetChange"}),
		Resource: stringorslice.Slice([]string{"arn:aws:route53:::change/*"}),
	})

	wildcard := stringorslice.Slice([]string{"*"})
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:      "kopsK8sRoute53ListZones",
		Effect:   IAMStatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:ListHostedZones"}),
		Resource: wildcard,
	})
}

func addKMSIAMPolicies(p *IAMPolicy, resource stringorslice.StringOrSlice) {

	// TODO should we add conditions?
	//	"Condition": {
	//	    "StringEquals": {
	//	      "kms:ViaService": [
	//	        "ec2.us-west-2.amazonaws.com",
	//	      ]
	//	    }
	//	  }

	// I removed these perms and testing is fine with encrypted volumes
	//			"kms:ListGrants",
	//			"kms:RevokeGrant",

	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsK8sKMSEncryptedVolumes",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"kms:Encrypt",
			"kms:Decrypt",
			"kms:ReEncrypt*",
			"kms:GenerateDataKey*",
			"kms:DescribeKey",
			"kms:CreateGrant",
		),
		Resource: resource,
	})
}

func addNodeEC2Policies(p *IAMPolicy, resource stringorslice.StringOrSlice) {

	// protokube makes a describe instance call
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:      "kopsK8sNodeEC2Perms",
		Effect:   IAMStatementEffectAllow,
		Action:   stringorslice.Slice([]string{"ec2:DescribeInstances"}),
		Resource: resource,
	})
}

func addMasterEC2Policies(p *IAMPolicy, resource stringorslice.StringOrSlice) {

	// comments are which cloudprovider code file makes the call
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsK8sMasterEC2Perms",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"ec2:AttachVolume",                  // aws.go
			"ec2:AuthorizeSecurityGroupIngress", // aws.go
			"ec2:CreateTags",                    // tag.go
			"ec2:CreateVolume",                  // aws.go
			"ec2:CreateRoute",                   // aws.go
			"ec2:CreateSecurityGroup",           // aws.go
			"ec2:DeleteSecurityGroup",           // aws.go
			"ec2:DeleteRoute",                   // aws.go
			"ec2:DeleteVolume",                  // aws.go
			"ec2:DescribeInstances",             // aws.go
			"ec2:DescribeRouteTables",           // aws.go
			"ec2:DescribeSubnets",               // aws.go
			"ec2:DescribeSecurityGroups",        // aws.go
			"ec2:DescribeVolumes",               // aws.go
			"ec2:DetachVolume",                  // aws.go
			"ec2:ModifyInstanceAttribute",       // aws.go
			"ec2:RevokeSecurityGroupIngress",    // aws.go
		),
		Resource: resource,
	})

}

func addMasterELBPolicies(p *IAMPolicy, resource stringorslice.StringOrSlice) {

	// comments are which cloudprovider code file makes the call
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsElbPerms",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"elasticloadbalancing:AttachLoadBalancerToSubnets",             // aws_loadbalanacer.go
			"elasticloadbalancing:ApplySecurityGroupsToLoadBalancer",       // aws_loadbalanacer.go
			"elasticloadbalancing:CreateLoadBalancer",                      // aws_loadbalanacer.go
			"elasticloadbalancing:CreateLoadBalancerPolicy",                // aws_loadbalanacer.go
			"elasticloadbalancing:CreateLoadBalancerListeners",             // aws_loadbalanacer.go
			"elasticloadbalancing:ConfigureHealthCheck",                    // aws_loadbalanacer.go
			"elasticloadbalancing:DeleteLoadBalancer",                      // aws.go
			"elasticloadbalancing:DeleteLoadBalancerListeners",             // aws_loadbalanacer.go
			"elasticloadbalancing:DescribeLoadBalancers",                   // aws.go
			"elasticloadbalancing:DescribeLoadBalancerAttributes",          // aws.go
			"elasticloadbalancing:DetachLoadBalancerFromSubnets",           // aws_loadbalancer.go
			"elasticloadbalancing:DeregisterInstancesFromLoadBalancer",     // aws_loadbalanacer.go
			"elasticloadbalancing:ModifyLoadBalancerAttributes",            // aws_loadbalanacer.go
			"elasticloadbalancing:RegisterInstancesWithLoadBalancer",       // aws_loadbalanacer.go
			"elasticloadbalancing:SetLoadBalancerPoliciesForBackendServer", // aws_loadbalanacer.go
		),
		Resource: resource,
	})
}

func addMasterASPolicies(p *IAMPolicy, resource stringorslice.StringOrSlice) {
	// comments are which cloudprovider / autoscaler code file makes the call
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsMasterASPerms",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"autoscaling:DescribeAutoScalingGroups",           // aws_instancegroups.go
			"autoscaling:GetAsgForInstance",                   // aws_manager.go
			"autoscaling:SetDesiredCapacity",                  // aws_manager.go
			"autoscaling:TerminateInstanceInAutoScalingGroup", // aws_manager.go
			"autoscaling:UpdateAutoScalingGroup",              // aws_instancegroups.go
		),
		Resource: resource,
	})
}

func addCertIAMPolicies(p *IAMPolicy, resource stringorslice.StringOrSlice) {
	// This is needed if we are using iam ssl certs
	// on ELBs
	// TODO need to test this
	p.Statement = append(p.Statement, &IAMStatement{
		Sid:    "kopsMasterCertIAMPerms",
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of(
			"iam:ListServerCertificates",
			"iam:GetServerCertificate",
		),
		Resource: resource,
	})

}

func createResource(b *IAMPolicyBuilder) stringorslice.StringOrSlice {
	var resource stringorslice.StringOrSlice
	if b.ResourceARN != nil {
		resource = stringorslice.Slice([]string{*b.ResourceARN})
	} else {
		resource = stringorslice.Slice([]string{"*"})
	}
	return resource
}
