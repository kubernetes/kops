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

package iam

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/pkg/testutils/golden"
	"k8s.io/kops/pkg/util/stringorset"
	"k8s.io/kops/upup/pkg/fi"
)

func TestRoundTrip(t *testing.T) {
	grid := []struct {
		IAM  *Statement
		JSON string
	}{
		{
			IAM: &Statement{
				Effect:   StatementEffectAllow,
				Action:   stringorset.Of("ec2:DescribeRegions"),
				Resource: stringorset.Of("*"),
			},
			JSON: "{\"Action\":\"ec2:DescribeRegions\",\"Effect\":\"Allow\",\"Resource\":\"*\"}",
		},
		{
			IAM: &Statement{
				Effect:   StatementEffectDeny,
				Action:   stringorset.Of("ec2:DescribeRegions", "ec2:DescribeInstances"),
				Resource: stringorset.Of("a", "b"),
			},
			JSON: "{\"Action\":[\"ec2:DescribeInstances\",\"ec2:DescribeRegions\"],\"Effect\":\"Deny\",\"Resource\":[\"a\",\"b\"]}",
		},
		{
			IAM: &Statement{
				Effect:    StatementEffectDeny,
				Principal: Principal{Federated: "federated"},
				Condition: map[string]interface{}{
					"foo": 1,
				},
			},
			JSON: "{\"Condition\":{\"foo\":1},\"Effect\":\"Deny\",\"Principal\":{\"Federated\":\"federated\"}}",
		},
		{
			IAM: &Statement{
				Effect:    StatementEffectDeny,
				Principal: Principal{Service: new(stringorset.Of("service"))},
				Condition: map[string]interface{}{
					"bar": "baz",
				},
			},
			JSON: "{\"Condition\":{\"bar\":\"baz\"},\"Effect\":\"Deny\",\"Principal\":{\"Service\":\"service\"}}",
		},
	}
	for _, g := range grid {
		actualJSON, err := json.Marshal(g.IAM)
		if err != nil {
			t.Errorf("error encoding IAM %v to json: %v", g.IAM, err)
		}

		if g.JSON != string(actualJSON) {
			t.Errorf("Unexpected JSON encoding.  Actual=%q, Expected=%q", string(actualJSON), g.JSON)
		}

		parsed := &Statement{}
		err = json.Unmarshal([]byte(g.JSON), parsed)
		if err != nil {
			t.Errorf("error decoding IAM %s to json: %v", g.JSON, err)
		}

		if !parsed.Equal(g.IAM) {
			t.Errorf("Unexpected JSON decoded value.  Actual=%v, Expected=%v", parsed, g.IAM)
		}

	}
}

func TestPolicyGeneration(t *testing.T) {
	grid := []struct {
		Role                   Subject
		AllowContainerRegistry bool
		NLBSecurityGroupMode   *string
		Policy                 string
	}{
		{
			Role:                   &NodeRoleMaster{},
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_master_strict.json",
		},
		{
			Role:                   &NodeRoleMaster{},
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_master_strict_ecr.json",
		},
		{
			Role:                   &NodeRoleMaster{},
			AllowContainerRegistry: false,
			NLBSecurityGroupMode:   new("Managed"),
			Policy:                 "tests/iam_builder_master_nlb_sg_managed.json",
		},
		{
			Role:                   &NodeRoleNode{},
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_node_strict.json",
		},
		{
			Role:                   &NodeRoleNode{},
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_node_strict_ecr.json",
		},
		{
			Role:                   &NodeRoleBastion{},
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_bastion.json",
		},
		{
			Role:                   &NodeRoleBastion{},
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_bastion.json",
		},
	}

	for i, x := range grid {
		b := &PolicyBuilder{
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ConfigStore: kops.ConfigStoreSpec{
						Base: "s3://kops-tests/iam-builder-test.k8s.local",
					},
					IAM: &kops.IAMSpec{
						AllowContainerRegistry: x.AllowContainerRegistry,
					},
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Members: []kops.EtcdMemberSpec{
								{
									KmsKeyID: aws.String("key-id-1"),
								},
								{
									KmsKeyID: aws.String("key-id-2"),
								},
							},
						},
						{
							Members: []kops.EtcdMemberSpec{},
						},
						{
							Members: []kops.EtcdMemberSpec{
								{
									KmsKeyID: aws.String("key-id-3"),
								},
							},
						},
					},
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{
							EBSCSIDriver: &kops.EBSCSIDriverSpec{
								Enabled: new(true),
							},
							NLBSecurityGroupMode: x.NLBSecurityGroupMode,
						},
					},
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
					Networking: kops.NetworkingSpec{
						Kubenet: &kops.KubenetNetworkingSpec{},
					},
				},
			},
			Role:      x.Role,
			Region:    "us-test-1",
			Partition: "aws-test",
		}
		b.Cluster.SetName("iam-builder-test.nonexistant")

		p, err := b.BuildAWSPolicy()
		if err != nil {
			t.Errorf("case %d failed to build an AWS IAM policy. Error: %v", i, err)
			continue
		}

		actualPolicy, err := p.AsJSON()
		if err != nil {
			t.Errorf("case %d failed to convert generated IAM Policy to JSON. Error: %v", i, err)
			continue
		}

		golden.AssertMatchesFile(t, actualPolicy, x.Policy)
	}
}

func TestEmptyPolicy(t *testing.T) {
	role := &GenericServiceAccount{
		NamespacedName: types.NamespacedName{
			Name:      "myaccount",
			Namespace: "default",
		},
		Policy: nil,
	}

	cluster := testutils.BuildMinimalClusterAWS("irsa.example.com")
	b := &PolicyBuilder{
		Cluster: cluster,
		Role:    role,
	}

	pr := &PolicyResource{
		Builder: b,
	}

	policy, err := fi.ResourceAsString(pr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if policy != "" {
		t.Errorf("empty policy should result in empty string, but was %q", policy)
	}
}

func TestAddKMSIAMPolicies(t *testing.T) {
	dataActions := []string{
		"kms:Decrypt",
		"kms:Encrypt",
		"kms:GenerateDataKey*",
		"kms:ReEncrypt*",
	}
	alwaysUnconditional := []string{"kms:CreateGrant", "kms:DescribeKey"}

	tests := []struct {
		name             string
		region           string
		bypassViaService bool
		wantConditional  bool
	}{
		{name: "region set, bypass off", region: "us-east-1", bypassViaService: false, wantConditional: true},
		{name: "region set, bypass on", region: "us-east-1", bypassViaService: true, wantConditional: false},
		{name: "region unset, bypass off", region: "", bypassViaService: false, wantConditional: false},
		{name: "region unset, bypass on", region: "", bypassViaService: true, wantConditional: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPolicy("c.example.com", "aws", tc.region)
			addKMSIAMPolicies(p, tc.bypassViaService)
			for _, action := range dataActions {
				inConditional := p.kmsDataPlaneAction.Has(action)
				inUnconditional := p.unconditionalAction.Has(action)
				wantInConditional := tc.wantConditional
				wantInUnconditional := !tc.wantConditional
				if inConditional != wantInConditional || inUnconditional != wantInUnconditional {
					t.Errorf("addKMSIAMPolicies(region=%q, bypass=%v) action %q: conditional=%v unconditional=%v, want conditional=%v unconditional=%v",
						tc.region, tc.bypassViaService, action, inConditional, inUnconditional, wantInConditional, wantInUnconditional)
				}
			}
			// CreateGrant and DescribeKey are always unconditional regardless of
			// region or bypass, since the EBS CSI driver calls them directly.
			for _, action := range alwaysUnconditional {
				if !p.unconditionalAction.Has(action) {
					t.Errorf("addKMSIAMPolicies(region=%q, bypass=%v) missing unconditional %q", tc.region, tc.bypassViaService, action)
				}
				if p.kmsDataPlaneAction.Has(action) {
					t.Errorf("addKMSIAMPolicies(region=%q, bypass=%v) %q unexpectedly in kmsDataPlaneAction", tc.region, tc.bypassViaService, action)
				}
			}
		})
	}
}

func TestKmsViaServices(t *testing.T) {
	// Per AWS KMS docs, kms:ViaService uses the .amazonaws.com suffix in all
	// partitions, so the result depends only on the region.
	tests := []struct {
		name   string
		region string
		want   []string
	}{
		{name: "empty region returns nil", region: "", want: nil},
		{name: "commercial region", region: "us-east-1", want: []string{"ec2.us-east-1.amazonaws.com", "s3.*.amazonaws.com"}},
		{name: "gov region keeps amazonaws.com suffix", region: "us-gov-west-1", want: []string{"ec2.us-gov-west-1.amazonaws.com", "s3.*.amazonaws.com"}},
		{name: "china region keeps amazonaws.com suffix", region: "cn-north-1", want: []string{"ec2.cn-north-1.amazonaws.com", "s3.*.amazonaws.com"}},
		{name: "iso region keeps amazonaws.com suffix", region: "us-iso-east-1", want: []string{"ec2.us-iso-east-1.amazonaws.com", "s3.*.amazonaws.com"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kmsViaServices(tc.region)
			if !slices.Equal(got, tc.want) {
				t.Errorf("kmsViaServices(%q) = %v, want %v", tc.region, got, tc.want)
			}
		})
	}
}

func TestIAMServiceEC2(t *testing.T) {
	expectations := map[string]string{
		"us-east-1":      "ec2.amazonaws.com",
		"randomunknown":  "ec2.amazonaws.com",
		"us-gov-east-1":  "ec2.amazonaws.com",
		"cn-north-1":     "ec2.amazonaws.com.cn",
		"cn-northwest-1": "ec2.amazonaws.com.cn",
	}

	for region, expect := range expectations {
		principal, err := IAMServiceEC2(region)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if principal != expect {
			t.Errorf("expected %s for %s, but received %s", expect, region, principal)
		}
	}
}

func TestAddKarpenterPermissions(t *testing.T) {
	tests := []struct {
		name                      string
		useCustomInstanceProfiles bool
		wantPassRoleResource      string
	}{
		{name: "managed instance profiles", useCustomInstanceProfiles: false, wantPassRoleResource: "arn:aws:iam::*:role/nodes.c.example.com"},
		{name: "custom instance profiles", useCustomInstanceProfiles: true, wantPassRoleResource: "arn:aws:iam::*:role/*"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPolicy("c.example.com", "aws", "us-east-1")
			if err := AddKarpenterPermissions(p, tc.useCustomInstanceProfiles); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var passRole *Statement
			for _, s := range p.Statement {
				if slices.Contains(s.Action.Value(), "iam:PassRole") {
					if passRole != nil {
						t.Fatalf("found multiple iam:PassRole statements")
					}
					passRole = s
				}
			}
			if passRole == nil {
				t.Fatalf("no iam:PassRole statement found")
			}
			if got := passRole.Resource.Value(); len(got) != 1 || got[0] != tc.wantPassRoleResource {
				t.Errorf("iam:PassRole resource = %v, want %q", got, tc.wantPassRoleResource)
			}
			if _, ok := passRole.Condition["StringEquals"]; !ok {
				t.Errorf("iam:PassRole statement missing StringEquals condition")
			}
		})
	}
}
