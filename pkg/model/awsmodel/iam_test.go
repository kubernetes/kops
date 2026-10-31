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
	"reflect"
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cloudmock/aws/mockiam"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func Test_formatAWSIAMStatement(t *testing.T) {
	type args struct {
		acountId     string
		partition    string
		oidcProvider string
		namespace    string
		name         string
	}
	tests := []struct {
		name    string
		args    args
		want    *iam.Statement
		wantErr bool
	}{
		{
			name: "namespace and name without wildcard",
			args: args{
				acountId:     "0123456789",
				partition:    "aws-test",
				oidcProvider: "oidc-test",
				namespace:    "test",
				name:         "test",
			},
			wantErr: false,
			want: &iam.Statement{
				Effect: "Allow",
				Principal: iam.Principal{
					Federated: "arn:aws-test:iam::0123456789:oidc-provider/oidc-test",
				},
				Action: stringorset.String("sts:AssumeRoleWithWebIdentity"),
				Condition: map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"oidc-test:sub": "system:serviceaccount:test:test",
					},
				},
			},
		},
		{
			name: "name contains wildcard",
			args: args{
				acountId:     "0123456789",
				partition:    "aws-test",
				oidcProvider: "oidc-test",
				namespace:    "test",
				name:         "test-*",
			},
			wantErr: true,
		},
		{
			name: "namespace contains wildcard",
			args: args{
				acountId:     "0123456789",
				partition:    "aws-test",
				oidcProvider: "oidc-test",
				namespace:    "test-*",
				name:         "test",
			},
			wantErr: false,
			want: &iam.Statement{
				Effect: "Allow",
				Principal: iam.Principal{
					Federated: "arn:aws-test:iam::0123456789:oidc-provider/oidc-test",
				},
				Action: stringorset.String("sts:AssumeRoleWithWebIdentity"),
				Condition: map[string]interface{}{
					"StringLike": map[string]interface{}{
						"oidc-test:sub": "system:serviceaccount:test-*:test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatAWSIAMStatement(tt.args.acountId, tt.args.partition, tt.args.oidcProvider, tt.args.namespace, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatAWSIAMStatement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatAWSIAMStatement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIAMFindDeletionsSharedInstanceProfile(t *testing.T) {
	clusterName := "me.example.com"
	ownershipTagKey := "kubernetes.io/cluster/" + clusterName

	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	c := &mockiam.MockIAM{
		Roles: make(map[string]*iamtypes.Role),
	}
	cloud.MockIAM = c

	tags := []iamtypes.Tag{
		{
			Key:   &ownershipTagKey,
			Value: new("owned"),
		},
	}

	for _, name := range []string{"nodes." + clusterName, "bastions." + clusterName} {
		c.Roles[name] = &iamtypes.Role{
			RoleName: new(name),
			RoleId:   new(name),
			Tags:     tags,
		}
	}

	b := &IAMModelBuilder{
		Cluster:   &kops.Cluster{ObjectMeta: metav1.ObjectMeta{Name: clusterName}},
		Lifecycle: fi.LifecycleSync,
	}
	context := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
	}
	context.AddTask(&awstasks.IAMInstanceProfile{
		Name:      new("nodes." + clusterName),
		Lifecycle: fi.LifecycleSync,
		Shared:    new(true),
	})

	if err := b.FindDeletions(context, cloud); err != nil {
		t.Fatalf("error finding deletions: %v", err)
	}

	if _, ok := context.Tasks["IAMRole/nodes."+clusterName]; ok {
		t.Errorf("IAM role %q was queued for deletion", "nodes."+clusterName)
	}
	if _, ok := context.Tasks["IAMRole/bastions."+clusterName]; !ok {
		t.Errorf("IAM role %q was not queued for deletion", "bastions."+clusterName)
	}
}
