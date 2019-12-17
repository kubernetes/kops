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
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/util/stringorslice"
)

func TestRoundTrip(t *testing.T) {
	grid := []struct {
		IAM  *Statement
		JSON string
	}{
		{
			IAM: &Statement{
				Effect:   StatementEffectAllow,
				Action:   stringorslice.Of("ec2:DescribeRegions"),
				Resource: stringorslice.Of("*"),
			},
			JSON: "{\"Effect\":\"Allow\",\"Action\":\"ec2:DescribeRegions\",\"Resource\":\"*\"}",
		},
		{
			IAM: &Statement{
				Effect:   StatementEffectDeny,
				Action:   stringorslice.Of("ec2:DescribeRegions", "ec2:DescribeInstances"),
				Resource: stringorslice.Of("a", "b"),
			},
			JSON: "{\"Effect\":\"Deny\",\"Action\":[\"ec2:DescribeRegions\",\"ec2:DescribeInstances\"],\"Resource\":[\"a\",\"b\"]}",
		},
	}
	for _, g := range grid {
		actualJSON, err := json.Marshal(g.IAM)
		if err != nil {
			t.Errorf("error encoding IAM %s to json: %v", g.IAM, err)
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
		Role                   kops.InstanceGroupRole
		LegacyIAM              bool
		AllowContainerRegistry bool
		Policy                 string
	}{
		{
			Role:                   "Master",
			LegacyIAM:              true,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_master_legacy.json",
		},
		{
			Role:                   "Master",
			LegacyIAM:              false,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_master_strict.json",
		},
		{
			Role:                   "Master",
			LegacyIAM:              false,
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_master_strict_ecr.json",
		},
		{
			Role:                   "Node",
			LegacyIAM:              true,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_node_legacy.json",
		},
		{
			Role:                   "Node",
			LegacyIAM:              false,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_node_strict.json",
		},
		{
			Role:                   "Node",
			LegacyIAM:              false,
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_node_strict_ecr.json",
		},
		{
			Role:                   "Bastion",
			LegacyIAM:              true,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_bastion.json",
		},
		{
			Role:                   "Bastion",
			LegacyIAM:              false,
			AllowContainerRegistry: false,
			Policy:                 "tests/iam_builder_bastion.json",
		},
		{
			Role:                   "Bastion",
			LegacyIAM:              false,
			AllowContainerRegistry: true,
			Policy:                 "tests/iam_builder_bastion.json",
		},
	}

	for i, x := range grid {
		b := &PolicyBuilder{
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ConfigStore: "s3://kops-tests/iam-builder-test.k8s.local",
					IAM: &kops.IAMSpec{
						Legacy:                 x.LegacyIAM,
						AllowContainerRegistry: x.AllowContainerRegistry,
					},
					EtcdClusters: []*kops.EtcdClusterSpec{
						{
							Members: []*kops.EtcdMemberSpec{
								{
									KmsKeyId: aws.String("key-id-1"),
								},
								{
									KmsKeyId: aws.String("key-id-2"),
								},
							},
						},
						{
							Members: []*kops.EtcdMemberSpec{},
						},
						{
							Members: []*kops.EtcdMemberSpec{
								{
									KmsKeyId: aws.String("key-id-3"),
								},
							},
						},
					},
				},
			},
			Role: x.Role,
		}
		b.Cluster.SetName("iam-builder-test.k8s.local")

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
		actualPolicy = strings.TrimSpace(actualPolicy)

		expectedPolicyBytes, err := ioutil.ReadFile(x.Policy)
		if err != nil {
			t.Fatalf("unexpected error reading IAM Policy from file %q: %v", x.Policy, err)
		}
		expectedPolicy := strings.TrimSpace(string(expectedPolicyBytes))

		if expectedPolicy != actualPolicy {
			diffString := diff.FormatDiff(expectedPolicy, actualPolicy)
			t.Logf("diff:\n%s\n", diffString)
			t.Errorf("case %d failed, policy output differed from expected (%s).", i, x.Policy)
			continue
		}
	}
}
