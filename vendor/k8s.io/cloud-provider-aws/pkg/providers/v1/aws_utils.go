/*
Copyright 2014 The Kubernetes Authors.

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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"

	"k8s.io/apimachinery/pkg/util/sets"
)

func stringSetToPointers(in sets.String) []*string {
	if in == nil {
		return nil
	}
	out := make([]*string, 0, len(in))
	for k := range in {
		out = append(out, aws.String(k))
	}
	return out
}

func stringSetFromPointers(in []*string) sets.String {
	if in == nil {
		return nil
	}
	out := sets.NewString()
	for i := range in {
		out.Insert(aws.StringValue(in[i]))
	}
	return out
}

// GetSourceAccount constructs source acct and return them for use
func GetSourceAccount(roleARN string) (string, error) {
	// ARN format (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html)
	// arn:partition:service:region:account-id:resource-type/resource-id
	// IAM format, region is always blank
	// arn:aws:iam::account:role/role-name-with-path
	if !arn.IsARN(roleARN) {
		return "", fmt.Errorf("incorrect ARN format for role %s", roleARN)
	}

	parsedArn, err := arn.Parse(roleARN)
	if err != nil {
		return "", err
	}

	return parsedArn.AccountID, nil
}
