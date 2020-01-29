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

package alimodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// PolicyDefaultVersion is the default version included in all policy documents
const PolicyDefaultVersion = "1"

// Policy Struct is a collection of fields that form a valid Alicloud policy document
type Policy struct {
	Version   string
	Statement []*Statement
}

// AsJSON converts the policy document to JSON format (parsable by Alicloud)
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

// Condition is a map of Conditions to be evaluated for a given RAM Statement
type Condition map[string]interface{}

// Statement is an Alicloud RAM Policy Statement Object:
// https://https://help.aliyun.com/document_detail/93739.html
type Statement struct {
	Effect    StatementEffect
	Action    stringorslice.StringOrSlice
	Resource  stringorslice.StringOrSlice
	Condition Condition `json:",omitempty"`
}

// PolicyResource defines the PolicyBuilder and DNSZone to use when building the
// RAM policy document for a given instance group role
type PolicyResource struct {
	Builder *PolicyBuilder
}

var _ fi.Resource = &PolicyResource{}
var _ fi.HasDependencies = &PolicyResource{}

// GetDependencies adds the DNSZone task to the list of dependencies if set
func (b *PolicyResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

// Open produces the Alicloud RAM policy for the given role
func (b *PolicyResource) Open() (io.Reader, error) {
	// Defensive copy before mutation
	pb := *b.Builder

	policy, err := pb.BuildAlicloudPolicy()
	if err != nil {
		return nil, fmt.Errorf("error building RAM policy: %v", err)
	}
	j, err := policy.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("error building RAM policy: %v", err)
	}
	return bytes.NewReader([]byte(j)), nil
}

// PolicyBuilder struct defines all valid fields to be used when building the
// Alicloud RAM policy document for a given instance group role.
type PolicyBuilder struct {
	Cluster      *kops.Cluster
	HostedZoneID string
	KMSKeys      []string
	Region       string
	ResourceARN  *string
	Role         kops.InstanceGroupRole
}

// BuildAlicloudPolicy builds a set of RAM policy statements based on the
// instance group type.
func (b *PolicyBuilder) BuildAlicloudPolicy() (*Policy, error) {
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
	// case kops.InstanceGroupRoleBastion:
	// 	p, err = b.BuildAlicloudPolicyBastion()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to generate Alicloud RAM Policy for Bastion Instance Group: %v", err)
	// 	}
	case kops.InstanceGroupRoleNode:
		p, err = b.BuildAlicloudPolicyNode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Alicloud RAM Policy for Node Instance Group: %v", err)
		}
	case kops.InstanceGroupRoleMaster:
		p, err = b.BuildAlicloudPolicyMaster()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Alicloud RAM Policy for Master Instance Group: %v", err)
		}
	default:
		return nil, fmt.Errorf("unrecognised instance group type: %s", b.Role)
	}

	return p, nil
}

// BuildAlicloudPolicyMaster generates a custom policy for a Kubernetes master.
func (b *PolicyBuilder) BuildAlicloudPolicyMaster() (*Policy, error) {
	resource := createResource(b)

	p := &Policy{
		Version: PolicyDefaultVersion,
	}

	addMasterECSPolicies(p, resource, b.Cluster.GetName())
	addMasterESSPolicies(p, resource, b.Cluster.GetName())
	addMasterSLBPolicies(p, resource)
	addVPCPermissions(p, resource, b.Cluster.GetName())

	var err error
	if p, err = b.AddOSSPermissions(p); err != nil {
		return nil, fmt.Errorf("failed to generate Alicloud RAM OSS access statements: %v", err)
	}

	if b.Cluster.Spec.IAM.AllowContainerRegistry {
		addCRPermissions(p)
	}

	return p, nil
}

// BuildAlicloudPolicyNode generates a custom policy for a Kubernetes node.
func (b *PolicyBuilder) BuildAlicloudPolicyNode() (*Policy, error) {
	resource := createResource(b)

	p := &Policy{
		Version: PolicyDefaultVersion,
	}

	addNodeECSPolicies(p, resource)

	var err error
	if p, err = b.AddOSSPermissions(p); err != nil {
		return nil, fmt.Errorf("failed to generate Alicloud RAM OSS access statements: %v", err)
	}

	if b.Cluster.Spec.IAM.AllowContainerRegistry {
		addCRPermissions(p)
	}

	return p, nil
}

// RAMPrefix returns the prefix for Alicloud ARNs in the current region, for use with RAM
// It is arn everywhere for now
func (b *PolicyBuilder) RAMPrefix() string {
	return "acs"
}

// AddOSSPermissions updates an RAM Policy with statements granting tailored
// access to OSS assets, depending on the instance group role
func (b *PolicyBuilder) AddOSSPermissions(p *Policy) (*Policy, error) {
	// For OSS RAM permissions we grant permissions to subtrees, so find the parents;
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

		if ossPath, ok := vfsPath.(*vfs.OSSPath); ok {
			ramOSSPath := ossPath.Bucket() + "/" + ossPath.Key()
			ramOSSPath = strings.TrimSuffix(ramOSSPath, "/")

			p.Statement = append(p.Statement, &Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Of("oss:GetBucketLocation", "oss:List*"),
				Resource: stringorslice.Slice([]string{
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ossPath.Bucket()}, ""),
				}),
			})

			if b.Role == kops.InstanceGroupRoleMaster {
				p.Statement = append(p.Statement, &Statement{
					Effect: StatementEffectAllow,
					Action: stringorslice.Slice([]string{"oss:Get*"}),
					Resource: stringorslice.Of(
						strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/*"}, ""),
					),
				})
			} else if b.Role == kops.InstanceGroupRoleNode {
				resources := []string{
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/addons/*"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/cluster.spec"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/config"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/instancegroup/*"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/issued/*"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/private/kube-proxy/*"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/ssh/*"}, ""),
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/secrets/dockerconfig"}, ""),
				}

				resources = append(resources, strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/private/kubelet/*"}, ""))

				sort.Strings(resources)

				p.Statement = append(p.Statement, &Statement{
					Effect:   StatementEffectAllow,
					Action:   stringorslice.Slice([]string{"oss:Get*"}),
					Resource: stringorslice.Of(resources...),
				})

				if b.Cluster.Spec.Networking != nil {
					// @check if kuberoute is enabled and permit access to the private key
					if b.Cluster.Spec.Networking.Kuberouter != nil {
						p.Statement = append(p.Statement, &Statement{
							Effect: StatementEffectAllow,
							Action: stringorslice.Slice([]string{"oss:Get*"}),
							Resource: stringorslice.Of(
								strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/private/kube-router/*"}, ""),
							),
						})
					}

					// @check if calico is enabled as the CNI provider and permit access to the client TLS certificate by default
					if b.Cluster.Spec.Networking.Calico != nil {
						p.Statement = append(p.Statement, &Statement{
							Effect: StatementEffectAllow,
							Action: stringorslice.Slice([]string{"oss:Get*"}),
							Resource: stringorslice.Of(
								strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/pki/private/calico-client/*"}, ""),
							),
						})
					}
				}
			}
		} else if _, ok := vfsPath.(*vfs.MemFSPath); ok {
			// Tests -ignore - nothing we can do in terms of RAM policy
			klog.Warningf("ignoring memfs path %q for RAM policy builder", vfsPath)
		} else {
			// We could implement this approach, but it seems better to
			// get all clouds using cluster-readable storage
			return nil, fmt.Errorf("path is not cluster readable: %v", root)
		}
	}

	writeablePaths, err := iam.WriteableVFSPaths(b.Cluster, b.Role)
	if err != nil {
		return nil, err
	}

	for _, vfsPath := range writeablePaths {
		if ossPath, ok := vfsPath.(*vfs.OSSPath); ok {
			ramOSSPath := ossPath.Bucket() + "/" + ossPath.Key()
			ramOSSPath = strings.TrimSuffix(ramOSSPath, "/")

			p.Statement = append(p.Statement, &Statement{
				Effect: StatementEffectAllow,
				Action: stringorslice.Slice([]string{"oss:GetObject", "oss:DeleteObject", "oss:PutObject"}),
				Resource: stringorslice.Of(
					strings.Join([]string{b.RAMPrefix(), ":oss:*:*:", ramOSSPath, "/*"}, ""),
				),
			})
		} else {
			klog.Warningf("unknown writeable path, can't apply RAM policy: %q", vfsPath)
		}
	}

	return p, nil
}

func addMasterECSPolicies(p *Policy, resource stringorslice.StringOrSlice, clusterName string) {
	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"ecs:Describe*",
				"ecs:AttachDisk",
				"ecs:CreateDisk",
				"ecs:CreateSnapshot",
				"ecs:DeleteDisk",
				"ecs:DeleteSnapshot",
				"ecs:DetachDisk",
				"ecs:ModifyAutoSnapshotPolicy",
				"ecs:ModifyDiskAttribute",
			}),
			Resource: resource,
		})
}

func addNodeECSPolicies(p *Policy, resource stringorslice.StringOrSlice) {
	p.Statement = append(p.Statement, &Statement{
		Effect:   StatementEffectAllow,
		Action:   stringorslice.Slice([]string{"ecs:DescribeInstances"}),
		Resource: resource,
	})
}

func addMasterSLBPolicies(p *Policy, resource stringorslice.StringOrSlice) {
	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
		Action: stringorslice.Of(
			"slb:*",
		),
		Resource: resource,
	})

}

func addMasterESSPolicies(p *Policy, resource stringorslice.StringOrSlice, clusterName string) {
	// Comments are which cloudprovider / autoscaler code file makes the call
	// TODO: Make optional only if using autoscalers
	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Of(
				"ess:Describe*",
			),
			Resource: resource,
		},
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Of(
				"ess:ModifyScalingGroup",
			),
			Resource: resource,
			Condition: Condition{
				"StringEquals": map[string]string{
					"ess:ResourceTag/KubernetesCluster": clusterName,
				},
			},
		},
	)
}

func addVPCPermissions(p *Policy, resource stringorslice.StringOrSlice, clusterName string) {
	p.Statement = append(p.Statement,
		&Statement{
			Effect: StatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"vpc:*",
			}),
			Resource: resource,
		},
	)
}

func addCRPermissions(p *Policy) {
	p.Statement = append(p.Statement, &Statement{
		Effect: StatementEffectAllow,
		Action: stringorslice.Of(
			"cr:Get*",
			"cr:List*",
			"cr:PullRepository",
		),
		Resource: stringorslice.Slice([]string{"*"}),
	})
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
