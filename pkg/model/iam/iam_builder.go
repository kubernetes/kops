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
	Cluster      *api.Cluster
	ClusterName  string
	Role         api.InstanceGroupRole
	Region       string
	HostedZoneID string
	ResourceARN  *string
	// We probably implement this
	// have the capability to shut off ECR perms
	//CreateECRPerms        bool
}

// BuildAWSIAMPolicy generates the IAM policies for a bastion, node or master
func (b *IAMPolicyBuilder) BuildAWSIAMPolicy() (*IAMPolicy, error) {
	resource := b.createResource()

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

	if b.Role == api.InstanceGroupRoleNode {
		// protokube makes a describe instance call
		p.Statement = append(p.Statement, &IAMStatement{
			Sid:      "kopsK8sNodeEC2Perms",
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"ec2:DescribeInstances"}),
			Resource: resource,
		})
	}

	if b.Role == api.InstanceGroupRoleMaster {

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

		if b.HostedZoneID != "" {
			// TODO we should test if we are in China, and not just return
			// TODO no Route53 in China

			// Remove /hostedzone/ prefix (if present)
			hostedZoneID := strings.TrimPrefix(b.HostedZoneID, "/")
			hostedZoneID = strings.TrimPrefix(hostedZoneID, "hostedzone/")

			p.Statement = append(p.Statement, &IAMStatement{
				Sid:    "kopsK8sRoute53Change",
				Effect: IAMStatementEffectAllow,
				Action: stringorslice.Of("route53:ChangeResourceRecordSets",
					"route53:ListResourceRecordSets",
					"route53:ListHostedZones",
					"route53:ListHostedZonesByName",
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

func (b *IAMPolicyBuilder) createResource() stringorslice.StringOrSlice {
	var resource stringorslice.StringOrSlice
	if b.ResourceARN != nil {
		resource = stringorslice.Slice([]string{*b.ResourceARN})
	} else {
		resource = stringorslice.Slice([]string{"*"})
	}
	return resource
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
