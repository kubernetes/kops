/*
Copyright 2021 The Kubernetes Authors.

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
	"testing"
)

func TestIAMServiceEC2(t *testing.T) {
	var expectations = map[string]string{
		"us-east-1":      "ec2.amazonaws.com",
		"randomunknown":  "ec2.amazonaws.com",
		"us-gov-east-1":  "ec2.amazonaws.com",
		"cn-north-1":     "ec2.amazonaws.com.cn",
		"cn-northwest-1": "ec2.amazonaws.com.cn",
	}

	for region, expect := range expectations {
		principal := IAMServiceEC2(region)
		if principal != expect {
			t.Errorf("expected %s for %s, but received %s", expect, region, principal)
		}
	}
}
