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

package cloudup

import (
	api "k8s.io/kops/pkg/apis/kops"
	"testing"
	"strings"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/pkg/model/iam"
	"regexp"
	"log"
	"fmt"
)

func TestBuildIAMPolicy_AdditionalMasterPermissions(t *testing.T) {

	c := buildCluster(nil)
	additionalMasterPermissions := `{
		"Effect": "Allow",
		"Action": ["dynamodb:*"],
		"Resource": ["*"]
	}`

	c.Spec.AdditionalMasterPermissions = fi.String(fmt.Sprintf("[%s]", additionalMasterPermissions))

	iamPolicyBuilder := &iam.IAMPolicyBuilder{
		Cluster: c,
		Role: api.InstanceGroupRoleMaster,
		Region: "us-east-1",

	}

	iamPolicy, err := iamPolicyBuilder.BuildAWSIAMPolicy()
	if err != nil {
		t.Fatalf("BuildAWSIAMPolicy error: %v", err)
	}

	json, err := iamPolicy.AsJSON()
	if err != nil {
		t.Fatalf("Error marshaling IAM policy: %v", err)
	}

	reg, err := regexp.Compile("\\s")
	if err != nil {
		log.Fatal(err)
	}

	jsonWhitespaceLess := reg.ReplaceAllString(json, "")
	additionalMasterWhiteSpaceLess := reg.ReplaceAllString(additionalMasterPermissions, "")
	if !strings.Contains(jsonWhitespaceLess, additionalMasterWhiteSpaceLess) {
		t.Fatalf("IAM Policy did not contain additionalMasterPermissions: %v and %v", jsonWhitespaceLess, additionalMasterWhiteSpaceLess)
	}

}

func TestBuildIAMPolicy_AdditionalNodePermissions(t *testing.T) {

	c := buildCluster(nil)
	additionalNodePermissions := `{
		"Effect": "Allow",
		"Action": ["es:*"],
		"Resource": ["*"]
	}`

	c.Spec.AdditionalNodePermissions = fi.String(fmt.Sprintf("[%s]", additionalNodePermissions))

	iamPolicyBuilder := &iam.IAMPolicyBuilder{
		Cluster: c,
		Role: api.InstanceGroupRoleNode,
		Region: "us-west-2",

	}

	iamPolicy, err := iamPolicyBuilder.BuildAWSIAMPolicy()
	if err != nil {
		t.Fatalf("BuildAWSIAMPolicy error: %v", err)
	}

	json, err := iamPolicy.AsJSON()
	if err != nil {
		t.Fatalf("Error marshaling IAM policy: %v", err)
	}

	reg, err := regexp.Compile("\\s")
	if err != nil {
		log.Fatal(err)
	}

	jsonWhitespaceLess := reg.ReplaceAllString(json, "")
	additionalNodeWhiteSpaceLess := reg.ReplaceAllString(additionalNodePermissions, "")
	if !strings.Contains(jsonWhitespaceLess, additionalNodeWhiteSpaceLess) {
		t.Fatalf("IAM Policy did not contain additionalNodePermissions: %v and %v", jsonWhitespaceLess, additionalNodeWhiteSpaceLess)
	}

}
