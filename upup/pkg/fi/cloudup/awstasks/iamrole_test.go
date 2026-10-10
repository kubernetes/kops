/*
Copyright 2026 The Kubernetes Authors.

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

package awstasks

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"k8s.io/kops/cloudmock/aws/mockiam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func TestIAMRoleDeleteOwnedInstanceProfile(t *testing.T) {
	clusterName := "me.example.com"
	ownershipTagKey := "kubernetes.io/cluster/" + clusterName
	name := "bastions." + clusterName

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockiam.MockIAM{
		InstanceProfiles: make(map[string]*iamtypes.InstanceProfile),
		Roles:            make(map[string]*iamtypes.Role),
	}
	cloud.MockIAM = c

	tags := []iamtypes.Tag{
		{
			Key:   &ownershipTagKey,
			Value: new("owned"),
		},
	}

	role := &iamtypes.Role{
		RoleName: &name,
		RoleId:   new("AROA1"),
		Tags:     tags,
	}
	c.Roles[name] = role
	c.InstanceProfiles[name] = &iamtypes.InstanceProfile{
		InstanceProfileName: &name,
		Roles:               []iamtypes.Role{*role},
		Tags:                tags,
	}

	allTasks := map[string]fi.CloudupTask{
		"IAMRole/" + name: &IAMRole{
			ID:        role.RoleId,
			Name:      &name,
			Lifecycle: fi.LifecycleSync,
		},
	}
	runTasks(t, cloud.WithTags(map[string]string{awsup.TagClusterName: clusterName}), allTasks)

	if _, ok := c.Roles[name]; ok {
		t.Errorf("IAM role %q was not deleted", name)
	}
	if _, ok := c.InstanceProfiles[name]; ok {
		t.Errorf("IAM instance profile %q was not deleted", name)
	}
}

func TestIAMRoleDeleteUnownedInstanceProfile(t *testing.T) {
	clusterName := "me.example.com"
	ownershipTagKey := "kubernetes.io/cluster/" + clusterName
	name := "nodes." + clusterName

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockiam.MockIAM{
		InstanceProfiles: make(map[string]*iamtypes.InstanceProfile),
		Roles:            make(map[string]*iamtypes.Role),
	}
	cloud.MockIAM = c

	role := &iamtypes.Role{
		RoleName: &name,
		RoleId:   new("AROA1"),
		Tags: []iamtypes.Tag{
			{
				Key:   &ownershipTagKey,
				Value: new("owned"),
			},
		},
	}
	c.Roles[name] = role
	c.InstanceProfiles[name] = &iamtypes.InstanceProfile{
		InstanceProfileName: &name,
		Roles:               []iamtypes.Role{*role},
	}

	allTasks := map[string]fi.CloudupTask{
		"IAMRole/" + name: &IAMRole{
			ID:        role.RoleId,
			Name:      &name,
			Lifecycle: fi.LifecycleSync,
		},
	}
	runTasks(t, cloud.WithTags(map[string]string{awsup.TagClusterName: clusterName}), allTasks)

	if _, ok := c.Roles[name]; !ok {
		t.Errorf("IAM role %q was deleted", name)
	}
	ip, ok := c.InstanceProfiles[name]
	if !ok {
		t.Fatalf("IAM instance profile %q was deleted", name)
	}
	if len(ip.Roles) != 1 {
		t.Errorf("Unexpected number of roles in IAM instance profile. Expected 1, got %d", len(ip.Roles))
	}
}

func TestIAMRoleDeleteWithoutInstanceProfile(t *testing.T) {
	ctx := context.TODO()

	clusterName := "me.example.com"
	ownershipTagKey := "kubernetes.io/cluster/" + clusterName
	name := "dns-controller.kube-system.sa." + clusterName

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockiam.MockIAM{
		InstanceProfiles: make(map[string]*iamtypes.InstanceProfile),
		Roles:            make(map[string]*iamtypes.Role),
		AttachedPolicies: make(map[string][]iamtypes.AttachedPolicy),
	}
	cloud.MockIAM = c

	role := &iamtypes.Role{
		RoleName: &name,
		RoleId:   new("AROA1"),
		Tags: []iamtypes.Tag{
			{
				Key:   &ownershipTagKey,
				Value: new("owned"),
			},
		},
	}
	c.Roles[name] = role
	c.AttachedPolicies[name] = []iamtypes.AttachedPolicy{
		{
			PolicyArn: new("arn:aws:iam::123456789012:policy/example"),
		},
	}
	if _, err := c.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(name),
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String("{}"),
	}); err != nil {
		t.Fatalf("error creating test role policy: %v", err)
	}

	allTasks := map[string]fi.CloudupTask{
		"IAMRole/" + name: &IAMRole{
			ID:        role.RoleId,
			Name:      &name,
			Lifecycle: fi.LifecycleSync,
		},
	}
	runTasks(t, cloud.WithTags(map[string]string{awsup.TagClusterName: clusterName}), allTasks)

	if _, ok := c.Roles[name]; ok {
		t.Errorf("IAM role %q was not deleted", name)
	}
	if len(c.AttachedPolicies[name]) != 0 {
		t.Errorf("Unexpected number of attached policies. Expected 0, got %d", len(c.AttachedPolicies[name]))
	}
}
