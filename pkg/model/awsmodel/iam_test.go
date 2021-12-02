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

	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/util/stringorslice"
)

func TestIAMServiceEC2(t *testing.T) {
	expectations := map[string]string{
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
				Action: stringorslice.String("sts:AssumeRoleWithWebIdentity"),
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
				Action: stringorslice.String("sts:AssumeRoleWithWebIdentity"),
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
