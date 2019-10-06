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

// IAM Documentation: /docs/iam_roles.md

// TODO: We have a couple different code paths until we do lifecycles, and
// TODO: when we have a cluster or refactor some s3 code.  The only code that
// TODO: is not shared by the different path is the s3 / state store stuff.

// TODO: Initial work has been done to lock down IAM actions based on resources
// TODO: and condition keys, but this can be extended further (with thorough testing).

package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/util/pkg/vfs"
)

// PolicyDefaultVersion is the default version included in all policy documents
const PolicyDefaultVersion = "2012-10-17"

// Policy Struct is a collection of fields that form a valid AWS policy document
type Policy struct {
	Version   string
	Statement []*Statement
}

// AsJSON converts the policy document to JSON format (parsable by AWS)
func (p *Policy) AsJSON() (string, error) {
	j, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling policy to JSON: %v", err)
	}
	return string(j), nil
}

// StatementEffect is required and specifies what type of access the statement results in
type StatementEffect string

// StatementEffectAllow allows access for the given resources in the statement (based on conditions)
const StatementEffectAllow StatementEffect = "Allow"

// StatementEffectDeny allows access for the given resources in the statement (based on conditions)
const StatementEffectDeny StatementEffect = "Deny"

// Condition is a map of Conditions to be evaluated for a given IAM Statement
type Condition map[string]interface{}

// Statement is an AWS IAM Policy Statement Object:
// http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements.html#Statement
type Statement struct {
	Effect    StatementEffect
	Action    stringorslice.StringOrSlice
	Resource  stringorslice.StringOrSlice
	Condition Condition `json:",omitempty"`
}

// Equal compares two IAM Statements and returns a bool
// TODO: Extend to support Condition Keys
func (l *Statement) Equal(r *Statement) bool {
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

// PolicyBuilder struct defines all valid fields to be used when building the
// AWS IAM policy document for a given instance group role.
type PolicyBuilder struct {
	Cluster      *kops.Cluster
	HostedZoneID string
	KMSKeys      []string
	Region       string
	ResourceARN  *string
	Role         kops.InstanceGroupRole
}

// BuildAWSPolicy builds a set of IAM policy statements based on the
// instance group type and IAM Legacy flag within the Cluster Spec
func (b *PolicyBuilder) BuildAWSPolicy() (*Policy, error) {
	var p *Policy
	var err error

	// Retrieve all the KMS Keys in use
	for _, e := range b.Cluster.Spec.EtcdClusters {
		for _, m := range e.Members {
			if m.KmsKeyId != nil {
				b.KMSKeys = append(b.KMSKeys, *m.KmsKeyId)
			}
		}
	}

	switch b.Role {
	case kops.InstanceGroupRoleBastion:
		p, err = b.BuildAWSPolicyBastion()
		if err != nil {
			return nil, fmt.Errorf("failed to generate AWS IAM Policy for Bastion Instance Group: %v", err)
		}
	case kops.InstanceGroupRoleNode:
		p, err = b.BuildAWSPolicyNode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate AWS IAM Policy for Node Instance Group: %v", err)
		}
	case kops.InstanceGroupRoleMaster:
		p, err = b.BuildAWSPolicyMaster()
		if err != nil {
			return nil, fmt.Errorf("failed to generate AWS IAM Policy for Master Instance Group: %v", err)
		}
	default:
		return nil, fmt.Errorf("unrecognised instance group type: %s", b.Role)
	}

	return p, nil
}

// BuildAWSPolicyMaster generates a custom policy for a Kubernetes master.
func (b *PolicyBuilder) BuildAWSPolicyMaster() (*Policy, error) {
	resource := createResource(b)

	p := &Policy{
		Version: PolicyDefaultVersion,
	}

	addMasterEC2Policies(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName())
	addMasterASPolicies(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName())
	addMasterELBPolicies(p, resource, b.Cluster.Spec.IAM.Legacy)
	addCertIAMPolicies(p, resource)

	var err error
	if p, err = b.AddS3Permissions(p); err != nil {
		return nil, fmt.Errorf("failed to generate AWS IAM S3 access statements: %v", err)
	}

	if b.KMSKeys != nil && len(b.KMSKeys) != 0 {
		addKMSIAMPolicies(p, stringorslice.Slice(b.KMSKeys), b.Cluster.Spec.IAM.Legacy)
	}

	if b.HostedZoneID != "" {
		b.addRoute53Permissions(p, b.HostedZoneID)
	}

	if b.Cluster.Spec.IAM.Legacy {
		addRoute53ListHostedZonesPermission(p)
	}

	if b.Cluster.Spec.IAM.Legacy || b.Cluster.Spec.IAM.AllowContainerRegistry {
		addECRPermissions(p)
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Romana != nil {
		addRomanaCNIPermissions(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName())
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
		addAmazonVPCCNIPermissions(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName(), b.IAMPrefix())
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.LyftVPC != nil {
		addLyftVPCPermissions(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName())
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.Cilium != nil && b.Cluster.Spec.Networking.Cilium.Ipam == kops.CiliumIpamEni {
		addCiliumEniPermissions(p, resource, b.Cluster.Spec.IAM.Legacy)
	}

	return p, nil
}

// BuildAWSPolicyNode generates a custom policy for a Kubernetes node.
func (b *PolicyBuilder) BuildAWSPolicyNode() (*Policy, error) {
	resource := createResource(b)

	p := &Policy{
		Version: PolicyDefaultVersion,
	}

	addNodeEC2Policies(p, resource)

	var err error
	if p, err = b.AddS3Permissions(p); err != nil {
		return nil, fmt.Errorf("failed to generate AWS IAM S3 access statements: %v", err)
	}

	if b.Cluster.Spec.IAM.Legacy {
		if b.HostedZoneID != "" {
			b.addRoute53Permissions(p, b.HostedZoneID)
		}
		addRoute53ListHostedZonesPermission(p)
	}

	if b.Cluster.Spec.IAM.Legacy || b.Cluster.Spec.IAM.AllowContainerRegistry {
		addECRPermissions(p)
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.AmazonVPC != nil {
		addAmazonVPCCNIPermissions(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName(), b.IAMPrefix())
	}

	if b.Cluster.Spec.Networking != nil && b.Cluster.Spec.Networking.LyftVPC != nil {
		addLyftVPCPermissions(p, resource, b.Cluster.Spec.IAM.Legacy, b.Cluster.GetName())
	}

	return p, nil
}

// BuildAWSPolicyBastion generates a custom policy for a bastion host.
func (b *PolicyBuilder) BuildAWSPolicyBastion() (*Policy, error) {
	resource := createResource(b)

	p := &Policy{
		Version: PolicyDefaultVersion,
	}

	// Bastion hosts currently don't require any specific permissions.
	// A trivial permission is granted, because empty policies are not allowed.
	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"ec2:DescribeRegions"}),
		Resource: resource,
	})

	return p, nil
}

// IAMPrefix returns the prefix for AWS ARNs in the current region, for use with IAM
// it is arn:aws everywhere but in cn-north and us-gov-west-1
func (b *PolicyBuilder) IAMPrefix() string {
	switch b.Region {
	case "cn-north-1":
		return "arn:aws-cn"
	case "cn-northwest-1":
		return "arn:aws-cn"
	case "us-gov-east-1":
		return "arn:aws-us-gov"
	case "us-gov-west-1":
		return "arn:aws-us-gov"
	default:
		return "arn:aws"
	}
}

// AddS3Permissions updates an IAM Policy with statements granting tailored
// access to S3 assets, depending on the instance group role
func (b *PolicyBuilder) AddS3Permissions(p *Policy) (*Policy, error) {
	// For S3 IAM permissions we grant permissions to subtrees, so find the parents;
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
					klog.V(4).Infof("Ignoring location %q because found parent %q", l, locations[j])
					isTopLevel = false
				}
			}
			if isTopLevel {
				klog.V(4).Infof("Found root location %q", l)
				roots = append(roots, l)
			}
		}
	}

	sort.Strings(roots)

	for _, root := range roots {
		vfsPath, err := vfs.Context.BuildVfsPath(root)
		if err != nil {
			return nil, fmt.Errorf("cannot parse VFS path %q: %v", root, err)
		}

		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			iamS3Path := s3Path.Bucket() + "/" + s3Path.Key()
			iamS3Path = strings.TrimSuffix(iamS3Path, "/")

			p.Statement = append(p.Statement, &Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Of("s3:GetBucketLocation", "s3:GetEncryptionConfiguration", "s3:ListBucket"),
				Resource: stringorslice.Slice([]string{
					strings.Join([]string{b.IAMPrefix(), ":s3:::", s3Path.Bucket()}, ""),
				}),
			})

			if b.Cluster.Spec.IAM.Legacy {
				p.Statement = append(p.Statement, &Statement{
					Effect: StatementEffectAllow,
					Action: stringorslice.Slice([]string{"s3:*"}),
					Resource: stringorslice.Of(
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/*"}, ""),
					),
				})
			} else {
				if b.Role == kops.InstanceGroupRoleMaster {
					p.Statement = append(p.Statement, &Statement{
						Effect: StatementEffectAllow,
						Action: stringorslice.Slice([]string{"s3:Get*"}),
						Resource: stringorslice.Of(
							strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/*"}, ""),
						),
					})
				} else if b.Role == kops.InstanceGroupRoleNode {
					resources := []string{
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/addons/*"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/cluster.spec"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/config"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/instancegroup/*"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/issued/*"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/private/kube-proxy/*"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/ssh/*"}, ""),
						strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/secrets/dockerconfig"}, ""),
					}

					// @check if bootstrap tokens are enabled and if so enable access to client certificate
					if model.UseKopsControllerForKubeletBootstrap(b.Cluster) {
						// no additional permissions
					} else if b.UseBootstrapTokens() {
						resources = append(resources, strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/private/node-authorizer-client/*"}, ""))
					} else {
						resources = append(resources, strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/private/kubelet/*"}, ""))
					}

					sort.Strings(resources)

					p.Statement = append(p.Statement, &Statement{
						Effect:   StatementEffectAllow,
						Action:   stringorslice.Slice([]string{"s3:Get*"}),
						Resource: stringorslice.Of(resources...),
					})

					if b.Cluster.Spec.Networking != nil {
						// @check if kuberoute is enabled and permit access to the private key
						if b.Cluster.Spec.Networking.Kuberouter != nil {
							p.Statement = append(p.Statement, &Statement{
								Effect: StatementEffectAllow,
								Action: stringorslice.Slice([]string{"s3:Get*"}),
								Resource: stringorslice.Of(
									strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/private/kube-router/*"}, ""),
								),
							})
						}

						// @check if calico is enabled as the CNI provider and permit access to the client TLS certificate by default
						if b.Cluster.Spec.Networking.Calico != nil {
							p.Statement = append(p.Statement, &Statement{
								Effect: StatementEffectAllow,
								Action: stringorslice.Slice([]string{"s3:Get*"}),
								Resource: stringorslice.Of(
									strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/pki/private/calico-client/*"}, ""),
								),
							})
						}
					}
				}
			}
		} else if _, ok := vfsPath.(*vfs.MemFSPath); ok {
			// Tests -ignore - nothing we can do in terms of IAM policy
			klog.Warningf("ignoring memfs path %q for IAM policy builder", vfsPath)
		} else {
			// We could implement this approach, but it seems better to
			// get all clouds using cluster-readable storage
			return nil, fmt.Errorf("path is not cluster readable: %v", root)
		}
	}

	writeablePaths, err := WriteableVFSPaths(b.Cluster, b.Role)
	if err != nil {
		return nil, err
	}

	for _, vfsPath := range writeablePaths {
		if s3Path, ok := vfsPath.(*vfs.S3Path); ok {
			iamS3Path := s3Path.Bucket() + "/" + s3Path.Key()
			iamS3Path = strings.TrimSuffix(iamS3Path, "/")

			p.Statement = append(p.Statement, &Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Slice([]string{"s3:GetObject", "s3:DeleteObject", "s3:PutObject"}),
				Resource: stringorslice.Of(
					strings.Join([]string{b.IAMPrefix(), ":s3:::", iamS3Path, "/*"}, ""),
				),
			})
		} else {
			klog.Warningf("unknown writeable path, can't apply IAM policy: %q", vfsPath)
		}
	}

	return p, nil
}

func WriteableVFSPaths(cluster *kops.Cluster, role kops.InstanceGroupRole) ([]vfs.Path, error) {
	var paths []vfs.Path

	// On the master, grant IAM permissions to the backup store, if it is configured
	if role == kops.InstanceGroupRoleMaster {
		backupStores := sets.NewString()
		for _, c := range cluster.Spec.EtcdClusters {
			if c.Backups == nil || c.Backups.BackupStore == "" || backupStores.Has(c.Backups.BackupStore) {
				continue
			}
			backupStore := c.Backups.BackupStore

			vfsPath, err := vfs.Context.BuildVfsPath(backupStore)
			if err != nil {
				return nil, fmt.Errorf("cannot parse VFS path %q: %v", backupStore, err)
			}

			paths = append(paths, vfsPath)

			backupStores.Insert(backupStore)
		}
	}
	return paths, nil
}

// PolicyResource defines the PolicyBuilder and DNSZone to use when building the
// IAM policy document for a given instance group role
type PolicyResource struct {
	Builder *PolicyBuilder
	DNSZone *awstasks.DNSZone
}

var _ fi.Resource = &PolicyResource{}
var _ fi.HasDependencies = &PolicyResource{}

// GetDependencies adds the DNSZone task to the list of dependencies if set
func (b *PolicyResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	if b.DNSZone != nil {
		deps = append(deps, b.DNSZone)
	}
	return deps
}

// Open produces the AWS IAM policy for the given role
func (b *PolicyResource) Open() (io.Reader, error) {
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

	policy, err := pb.BuildAWSPolicy()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	j, err := policy.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	return bytes.NewReader([]byte(j)), nil
}

// UseBootstrapTokens check if we are using bootstrap tokens - @TODO, i don't like this we should probably pass in
// the kops model into the builder rather than duplicating the code. I'll leave for another PR
func (b *PolicyBuilder) UseBootstrapTokens() bool {
	if b.Cluster.Spec.KubeAPIServer == nil {
		return false
	}

	return fi.BoolValue(b.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken)
}

func addECRPermissions(p *Policy) {
	// TODO - I think we can just have GetAuthorizationToken here, as we are not
	// TODO - making any API calls except for GetAuthorizationToken.

	// We provide ECR access on the nodes (naturally), but we also provide access on the master.
	// We shouldn't be running lots of pods on the master, but it is perfectly reasonable to run
	// a private logging pod or similar.
	// At this point we allow all regions with ECR, since ECR is region specific.
	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
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

func (b *PolicyBuilder) addRoute53Permissions(p *Policy, hostedZoneID string) {

	// TODO: Route53 currently not supported in China, need to check and fail/return
	// Remove /hostedzone/ prefix (if present)
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "/")
	hostedZoneID = strings.TrimPrefix(hostedZoneID, "hostedzone/")

	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
		Action: stringorslice.Of("route53:ChangeResourceRecordSets",
			"route53:ListResourceRecordSets",
			"route53:GetHostedZone"),
		Resource: stringorslice.Slice([]string{b.IAMPrefix() + ":route53:::hostedzone/" + hostedZoneID}),
	})

	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:GetChange"}),
		Resource: stringorslice.Slice([]string{b.IAMPrefix() + ":route53:::change/*"}),
	})

	wildcard := stringorslice.Slice([]string{"*"})
	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:ListHostedZones"}),
		Resource: wildcard,
	})
}

func addKMSIAMPolicies(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool) {
	if legacyIAM {
		p.Statement = append(p.Statement, &Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Of(
				"kms:ListGrants",
				"kms:RevokeGrant",
			),
			Resource: resource,
		})
	}

	// TODO could use "kms:ViaService" Condition Key here?
	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
		Action: stringorslice.Of(
			"kms:CreateGrant",
			"kms:Decrypt",
			"kms:DescribeKey",
			"kms:Encrypt",
			"kms:GenerateDataKey*",
			"kms:ReEncrypt*",
		),
		Resource: resource,
	})
}

func addNodeEC2Policies(p *Policy, resource stringorslice.StringOrSlice) {
	// Protokube makes a DescribeInstances call, DescribeRegions when finding S3 State Bucket
	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"ec2:DescribeInstances", "ec2:DescribeRegions"}),
		Resource: resource,
	})
}

func addMasterEC2Policies(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool, clusterName string) {
	// The legacy IAM policy grants full ec2 API access
	if legacyIAM {
		p.Statement = append(p.Statement,
			&Statement{
				Effect:   StatementEffectAllow,
				Action:   stringorslice.Slice([]string{"ec2:*"}),
				Resource: resource,
			},
		)
	} else {

		// Describe* calls don't support any additional IAM restrictions
		// The non-Describe* ec2 calls support different types of filtering:
		// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/ec2-api-permissions.html
		// We try to lock down the permissions here in non-legacy mode,
		// but there are still some improvements we can make:

		// CreateVolume - supports filtering on tags, but we need to switch to pass tags to CreateVolume
		// CreateTags - supports filtering on existing tags. Also supports filtering on VPC for some resources (e.g. security groups)
		// Network Routing Permissions - May not be required with the CNI Networking provider

		// Comments are which cloudprovider code file makes the call
		p.Statement = append(p.Statement,
			&Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Slice([]string{
					"ec2:DescribeAccountAttributes", // aws.go
					"ec2:DescribeInstances",         // aws.go
					"ec2:DescribeInternetGateways",  // aws.go
					"ec2:DescribeRegions",           // s3context.go
					"ec2:DescribeRouteTables",       // aws.go
					"ec2:DescribeSecurityGroups",    // aws.go
					"ec2:DescribeSubnets",           // aws.go
					"ec2:DescribeVolumes",           // aws.go
				}),
				Resource: resource,
			},
			&Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Slice([]string{
					"ec2:CreateSecurityGroup",          // aws.go
					"ec2:CreateTags",                   // aws.go, tag.go
					"ec2:CreateVolume",                 // aws.go
					"ec2:DescribeVolumesModifications", // aws.go
					"ec2:ModifyInstanceAttribute",      // aws.go
					"ec2:ModifyVolume",                 // aws.go
				}),
				Resource: resource,
			},
			&Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Of(
					"ec2:AttachVolume",                  // aws.go
					"ec2:AuthorizeSecurityGroupIngress", // aws.go
					"ec2:CreateRoute",                   // aws.go
					"ec2:DeleteRoute",                   // aws.go
					"ec2:DeleteSecurityGroup",           // aws.go
					"ec2:DeleteVolume",                  // aws.go
					"ec2:DetachVolume",                  // aws.go
					"ec2:RevokeSecurityGroupIngress",    // aws.go
				),
				Resource: resource,
				Condition: Condition{
					"StringEquals": map[string]string{
						"ec2:ResourceTag/KubernetesCluster": clusterName,
					},
				},
			},
		)
	}
}

func addMasterELBPolicies(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool) {
	if legacyIAM {
		p.Statement = append(p.Statement, &Statement{
			Effect:   StatementEffectAllow,
			Action:   stringorslice.Slice([]string{"elasticloadbalancing:*"}),
			Resource: resource,
		})
	} else {
		// Comments are which cloudprovider code file makes the call
		p.Statement = append(p.Statement, &Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Of(
				"elasticloadbalancing:AddTags",                                 // aws_loadbalancer.go
				"elasticloadbalancing:AttachLoadBalancerToSubnets",             // aws_loadbalancer.go
				"elasticloadbalancing:ApplySecurityGroupsToLoadBalancer",       // aws_loadbalancer.go
				"elasticloadbalancing:CreateLoadBalancer",                      // aws_loadbalancer.go
				"elasticloadbalancing:CreateLoadBalancerPolicy",                // aws_loadbalancer.go
				"elasticloadbalancing:CreateLoadBalancerListeners",             // aws_loadbalancer.go
				"elasticloadbalancing:ConfigureHealthCheck",                    // aws_loadbalancer.go
				"elasticloadbalancing:DeleteLoadBalancer",                      // aws.go
				"elasticloadbalancing:DeleteLoadBalancerListeners",             // aws_loadbalancer.go
				"elasticloadbalancing:DescribeLoadBalancers",                   // aws.go
				"elasticloadbalancing:DescribeLoadBalancerAttributes",          // aws.go
				"elasticloadbalancing:DetachLoadBalancerFromSubnets",           // aws_loadbalancer.go
				"elasticloadbalancing:DeregisterInstancesFromLoadBalancer",     // aws_loadbalancer.go
				"elasticloadbalancing:ModifyLoadBalancerAttributes",            // aws_loadbalancer.go
				"elasticloadbalancing:RegisterInstancesWithLoadBalancer",       // aws_loadbalancer.go
				"elasticloadbalancing:SetLoadBalancerPoliciesForBackendServer", // aws_loadbalancer.go
			),
			Resource: resource,
		})

		p.Statement = append(p.Statement, &Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Of(
				"ec2:DescribeVpcs",                                       // aws_loadbalancer.go
				"elasticloadbalancing:AddTags",                           // aws_loadbalancer.go
				"elasticloadbalancing:CreateListener",                    // aws_loadbalancer.go
				"elasticloadbalancing:CreateTargetGroup",                 // aws_loadbalancer.go
				"elasticloadbalancing:DeleteListener",                    // aws_loadbalancer.go
				"elasticloadbalancing:DeleteTargetGroup",                 // aws_loadbalancer.go
				"elasticloadbalancing:DeregisterTargets",                 // aws_loadbalancer.go
				"elasticloadbalancing:DescribeListeners",                 // aws_loadbalancer.go
				"elasticloadbalancing:DescribeLoadBalancerPolicies",      // aws_loadbalancer.go
				"elasticloadbalancing:DescribeTargetGroups",              // aws_loadbalancer.go
				"elasticloadbalancing:DescribeTargetHealth",              // aws_loadbalancer.go
				"elasticloadbalancing:ModifyListener",                    // aws_loadbalancer.go
				"elasticloadbalancing:ModifyTargetGroup",                 // aws_loadbalancer.go
				"elasticloadbalancing:RegisterTargets",                   // aws_loadbalancer.go
				"elasticloadbalancing:SetLoadBalancerPoliciesOfListener", // aws_loadbalancer.go
			),
			Resource: resource,
		})
	}
}

func addMasterASPolicies(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool, clusterName string) {
	if legacyIAM {
		p.Statement = append(p.Statement, &Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"autoscaling:DescribeAutoScalingGroups",
				"autoscaling:DescribeAutoScalingInstances",
				"autoscaling:DescribeLaunchConfigurations",
				"autoscaling:DescribeTags",
				"autoscaling:SetDesiredCapacity",
				"autoscaling:TerminateInstanceInAutoScalingGroup",
				"autoscaling:UpdateAutoScalingGroup",
				"ec2:DescribeLaunchTemplateVersions",
			}),
			Resource: resource,
		})
	} else {
		// Comments are which cloudprovider / autoscaler code file makes the call
		// TODO: Make optional only if using autoscalers
		p.Statement = append(p.Statement,
			&Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Of(
					"autoscaling:DescribeAutoScalingGroups",    // aws_instancegroups.go
					"autoscaling:DescribeLaunchConfigurations", // aws.go
					"autoscaling:DescribeTags",                 // auto_scaling.go
					"ec2:DescribeLaunchTemplateVersions",
				),
				Resource: resource,
			},
			&Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Of(
					"autoscaling:SetDesiredCapacity",                  // aws_manager.go
					"autoscaling:TerminateInstanceInAutoScalingGroup", // aws_manager.go
					"autoscaling:UpdateAutoScalingGroup",              // aws_instancegroups.go
				),
				Resource: resource,
				Condition: Condition{
					"StringEquals": map[string]string{
						"autoscaling:ResourceTag/KubernetesCluster": clusterName,
					},
				},
			},
		)
	}
}

func addCertIAMPolicies(p *Policy, resource stringorslice.StringOrSlice) {
	// TODO: Make optional only if using IAM SSL Certs on ELBs
	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
		Action: stringorslice.Of(
			"iam:ListServerCertificates",
			"iam:GetServerCertificate",
		),
		Resource: resource,
	})
}

func addRoute53ListHostedZonesPermission(p *Policy) {
	wildcard := stringorslice.Slice([]string{"*"})
	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"route53:ListHostedZones"}),
		Resource: wildcard,
	})
}

func addRomanaCNIPermissions(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool, clusterName string) {
	if legacyIAM {
		// Legacy IAM provides ec2:*, so no additional permissions required
		return
	}

	// Romana requires additional Describe permissions
	// Comments are which Romana component makes the call
	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:DescribeAvailabilityZones", // vpcrouter
				"ec2:DescribeVpcs",              // vpcrouter
			}),
			Resource: resource,
		},
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:CreateRoute",  // vpcrouter
				"ec2:DeleteRoute",  // vpcrouter
				"ec2:ReplaceRoute", // vpcrouter
			}),
			Resource: resource,
			Condition: Condition{
				"StringEquals": map[string]string{
					"ec2:ResourceTag/KubernetesCluster": clusterName,
				},
			},
		},
	)
}

func addLyftVPCPermissions(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool, clusterName string) {
	if legacyIAM {
		// Legacy IAM provides ec2:*, so no additional permissions required
		return
	}

	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:DescribeSubnets",
				"ec2:AttachNetworkInterface",
				"ec2:AssignPrivateIpAddresses",
				"ec2:UnassignPrivateIpAddresses",
				"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DescribeVpcPeeringConnections",
				"ec2:DescribeSecurityGroups",
				"ec2:DetachNetworkInterface",
				"ec2:DeleteNetworkInterface",
				"ec2:ModifyNetworkInterfaceAttribute",
				"ec2:DescribeVpcs",
			}),
			Resource: resource,
		},
	)
}

func addCiliumEniPermissions(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool) {
	if legacyIAM {
		// Legacy IAM provides ec2:*, so no additional permissions required
		return
	}

	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:DescribeSubnets",
				"ec2:AttachNetworkInterface",
				"ec2:AssignPrivateIpAddresses",
				"ec2:UnassignPrivateIpAddresses",
				"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DescribeVpcPeeringConnections",
				"ec2:DescribeSecurityGroups",
				"ec2:DetachNetworkInterface",
				"ec2:DeleteNetworkInterface",
				"ec2:ModifyNetworkInterfaceAttribute",
				"ec2:DescribeVpcs",
			}),
			Resource: resource,
		},
	)
}

func addAmazonVPCCNIPermissions(p *Policy, resource stringorslice.StringOrSlice, legacyIAM bool, clusterName string, iamPrefix string) {
	if legacyIAM {
		// Legacy IAM provides ec2:*, so no additional permissions required
		return
	}

	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:AssignPrivateIpAddresses",
				"ec2:AttachNetworkInterface",
				"ec2:CreateNetworkInterface",
				"ec2:DeleteNetworkInterface",
				"ec2:DescribeInstances",
				"ec2:DescribeInstanceTypes",
				"ec2:DescribeTags",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DetachNetworkInterface",
				"ec2:ModifyNetworkInterfaceAttribute",
				"ec2:UnassignPrivateIpAddresses",
			}),
			Resource: resource,
		},
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ec2:CreateTags",
			}),
			Resource: stringorslice.Slice([]string{
				strings.Join([]string{iamPrefix, ":ec2:*:*:network-interface/*"}, ""),
			})},
	)
}

func createResource(b *PolicyBuilder) stringorslice.StringOrSlice {
	var resource stringorslice.StringOrSlice
	if b.ResourceARN != nil {
		resource = stringorslice.Slice([]string{*b.ResourceARN})
	} else {
		resource = stringorslice.Slice([]string{"*"})
	}
	return resource
}
