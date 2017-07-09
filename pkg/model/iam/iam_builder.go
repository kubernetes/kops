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

const IAMStatementEffectAllow IAMStatementEffect = "Allow"
const IAMStatementEffectDeny IAMStatementEffect = "Deny"

type IAMStatement struct {
	Effect   IAMStatementEffect
	Action   stringorslice.StringOrSlice
	Resource stringorslice.StringOrSlice
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
	Role         api.InstanceGroupRole
	Region       string
	HostedZoneID string
}

func (b *IAMPolicyBuilder) BuildAWSIAMPolicy() (*IAMPolicy, error) {
	wildcard := stringorslice.Slice([]string{"*"})

	iamPrefix := b.IAMPrefix()

	p := &IAMPolicy{
		Version: IAMPolicyDefaultVersion,
	}

	// Don't give bastions any permissions (yet)
	if b.Role == api.InstanceGroupRoleBastion {
		p.Statement = append(p.Statement, &IAMStatement{
			// We grant a trivial (?) permission (DescribeRegions), because empty policies are not allowed
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"ec2:DescribeRegions"}),
			Resource: wildcard,
		})

		return p, nil
	}

	if b.Role == api.InstanceGroupRoleNode {
		p.Statement = append(p.Statement, &IAMStatement{
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"ec2:Describe*"}),
			Resource: wildcard,
		})

	}

	{
		// We provide ECR access on the nodes (naturally), but we also provide access on the master.
		// We shouldn't be running lots of pods on the master, but it is perfectly reasonable to run
		// a private logging pod or similar.
		p.Statement = append(p.Statement, &IAMStatement{
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
			Resource: wildcard,
		})
	}

	if b.Role == api.InstanceGroupRoleMaster {
		p.Statement = append(p.Statement, &IAMStatement{
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"ec2:*"}),
			Resource: wildcard,
		})

		p.Statement = append(p.Statement, &IAMStatement{
			Effect:   IAMStatementEffectAllow,
			Action:   stringorslice.Slice([]string{"elasticloadbalancing:*"}),
			Resource: wildcard,
		})

		p.Statement = append(p.Statement, &IAMStatement{
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Of(
				"autoscaling:DescribeAutoScalingGroups",
				"autoscaling:DescribeAutoScalingInstances",
				"autoscaling:SetDesiredCapacity",
				"autoscaling:TerminateInstanceInAutoScalingGroup",
			),
			Resource: wildcard,
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
			p.Statement = append(p.Statement, &IAMStatement{
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
				Resource: stringorslice.Slice(kmsKeyIDs.List()),
			})
		}
	}

	if b.HostedZoneID != "" {
		addRoute53Permissions(p, b.HostedZoneID)
	}
	// dns-controller currently assumes it can list the hosted zones, even when using gossip
	addRoute53ListHostedZonesPermission(p)

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
				Effect: IAMStatementEffectAllow,
				Action: stringorslice.Slice([]string{"s3:*"}),
				Resource: stringorslice.Of(
					iamPrefix+":s3:::"+iamS3Path,
					iamPrefix+":s3:::"+iamS3Path+"/*",
				),
			})

			p.Statement = append(p.Statement, &IAMStatement{
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

func addRoute53Permissions(p *IAMPolicy, hostedZoneID string) {
	// Remove /hostedzone/ prefix (if present)
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "/")
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "hostedzone/")

	p.Statement = append(p.Statement, &IAMStatement{
		Effect: IAMStatementEffectAllow,
		Action: stringorslice.Of("route53:ChangeResourceRecordSets",
			"route53:ListResourceRecordSets",
			"route53:GetHostedZone"),
		Resource: stringorslice.Slice([]string{"arn:aws:route53:::hostedzone/" + hostedZoneID}),
	})

	p.Statement = append(p.Statement, &IAMStatement{
		Effect:   IAMStatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:GetChange"}),
		Resource: stringorslice.Slice([]string{"arn:aws:route53:::change/*"}),
	})
}

func addRoute53ListHostedZonesPermission(p *IAMPolicy) {
	wildcard := stringorslice.Slice([]string{"*"})
	p.Statement = append(p.Statement, &IAMStatement{
		Effect:   IAMStatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:ListHostedZones"}),
		Resource: wildcard,
	})
}

// IAMPrefix returns the prefix for AWS ARNs in the current region, for use with IAM
// it is arn:aws everywhere but in cn-north and us-gov-west-1
func (b *IAMPolicyBuilder) IAMPrefix() string {
	switch b.Region {
	case "cn-north-1":
		return "arn:aws-cn"
	case "us-gov-west-1":
		return "arn:aws-us-gov"
	default:
		return "arn:aws"
	}
}

type IAMPolicyResource struct {
	Builder *IAMPolicyBuilder
	DNSZone *awstasks.DNSZone
}

var _ fi.Resource = &IAMPolicyResource{}
var _ fi.HasDependencies = &IAMPolicyResource{}

func (b *IAMPolicyResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	if b.DNSZone != nil {
		deps = append(deps, b.DNSZone)
	}
	return deps
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
	json, err := policy.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	return bytes.NewReader([]byte(json)), nil
}
