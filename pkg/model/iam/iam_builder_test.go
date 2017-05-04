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

	"k8s.io/kops/pkg/util/stringorslice"
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
				Sid:      "foo",
			},
			JSON: "{\"Effect\":\"Allow\",\"Action\":\"ec2:DescribeRegions\",\"Resource\":\"*\",\"Sid\":\"foo\"}",
		},
		{
			IAM: &IAMStatement{
				Effect:   IAMStatementEffectDeny,
				Action:   stringorslice.Of("ec2:DescribeRegions", "ec2:DescribeInstances"),
				Resource: stringorslice.Of("a", "b"),
				Sid:      "foo",
			},
			JSON: "{\"Effect\":\"Deny\",\"Action\":[\"ec2:DescribeRegions\",\"ec2:DescribeInstances\"],\"Resource\":[\"a\",\"b\"],\"Sid\":\"foo\"}",
		},
	}
	for _, g := range grid {
		actualJson, err := json.Marshal(g.IAM)
		if err != nil {
			t.Errorf("error encoding IAM %s to json: %v", g.IAM, err)
		}

		if g.JSON != string(actualJson) {
			t.Errorf("Unexpected JSON encoding.  Actual=%q, Expected=%q", string(actualJson), g.JSON)
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
