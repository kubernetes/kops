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
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/pkg/util/stringorslice"
	"k8s.io/kops/util/pkg/vfs"
)

func TestRoundTrip(t *testing.T) {
	grid := []struct {
		IAM  *IAMStatement
		JSON string
	}{
		{
			IAM: &IAMStatement{
				Effect:   IAMStatementEffectAllow,
				Action:   stringorslice.Of("ec2:DescribeRegions"),
				Resource: stringorslice.Of("*"),
			},
			JSON: "{\"Effect\":\"Allow\",\"Action\":\"ec2:DescribeRegions\",\"Resource\":\"*\"}",
		},
		{
			IAM: &IAMStatement{
				Effect:   IAMStatementEffectDeny,
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

		parsed := &IAMStatement{}
		err = json.Unmarshal([]byte(g.JSON), parsed)
		if err != nil {
			t.Errorf("error decoding IAM %s to json: %v", g.JSON, err)
		}

		if !parsed.Equal(g.IAM) {
			t.Errorf("Unexpected JSON decoded value.  Actual=%v, Expected=%v", parsed, g.IAM)
		}

	}
}

func TestS3PolicyGeneration(t *testing.T) {
	defaultS3Statements := []*IAMStatement{
		{
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Of(
				"s3:GetBucketLocation",
				"s3:ListBucket",
			),
			Resource: stringorslice.Slice([]string{
				"arn:aws:s3:::bucket-name",
			}),
		},
		{
			Effect: IAMStatementEffectAllow,
			Action: stringorslice.Slice([]string{
				"s3:List*",
			}),
			Resource: stringorslice.Slice([]string{
				"arn:aws:s3:::bucket-name/cluster-name.k8s.local",
				"arn:aws:s3:::bucket-name/cluster-name.k8s.local/*",
			}),
		},
	}

	grid := []struct {
		Role      kops.InstanceGroupRole
		IAMPolicy IAMPolicy
	}{
		{
			Role: "Master",
			IAMPolicy: IAMPolicy{
				Statement: append(defaultS3Statements, &IAMStatement{
					Effect: IAMStatementEffectAllow,
					Action: stringorslice.Slice([]string{
						"s3:Get*",
					}),
					Resource: stringorslice.Of(
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/*",
					),
				}),
			},
		},
		{
			Role: "Node",
			IAMPolicy: IAMPolicy{
				Statement: append(defaultS3Statements, &IAMStatement{
					Effect: IAMStatementEffectAllow,
					Action: stringorslice.Slice([]string{
						"s3:Get*",
					}),
					Resource: stringorslice.Slice([]string{
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/addons/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/instancegroup/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/issued/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/ssh/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/private/ca/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/private/master/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/private/kube-proxy/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/pki/private/kubelet/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/secrets/*",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/cluster.spec",
						"arn:aws:s3:::bucket-name/cluster-name.k8s.local/config",
					}),
				}),
			},
		},
		{
			Role: "Bastion",
			IAMPolicy: IAMPolicy{
				Statement: defaultS3Statements,
			},
		},
	}

	for i, x := range grid {
		ip := &IAMPolicy{}

		vfsPath, err := vfs.Context.BuildVfsPath("s3://bucket-name/cluster-name.k8s.local")
		if err != nil {
			t.Errorf("case %d failed to build Vfs Path. error: %s", i, err)
			continue
		}
		s3Path, ok := vfsPath.(*vfs.S3Path)
		if !ok {
			t.Errorf("case %d failed to build S3 Path.", i)
			continue
		}

		addS3Permissions(ip, "arn:aws", s3Path, x.Role)

		expectedPolicy, err := x.IAMPolicy.AsJSON()
		if err != nil {
			t.Errorf("case %d failed to convert expected IAM Policy to JSON. Error: %q", i, err)
			continue
		}
		actualPolicy, err := ip.AsJSON()
		if err != nil {
			t.Errorf("case %d failed to convert generated IAM Policy to JSON. Error: %q", i, err)
			continue
		}

		if expectedPolicy != actualPolicy {
			diffString := diff.FormatDiff(expectedPolicy, actualPolicy)
			t.Logf("diff:\n%s\n", diffString)
			t.Errorf("case %d failed, policy output differed from expected.", i)
			continue
		}
	}
}
