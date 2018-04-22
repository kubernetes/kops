/*
Copyright 2018 The Kubernetes Authors.

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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func cleanAlias(s string) string {
	return "alias/" + strings.Replace(s, ".", "_", -1)
}

func ListCustomerMasterKeys(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	keyName := cleanAlias(clusterName)

	request := &kms.DescribeKeyInput{
		KeyId: aws.String(keyName),
	}

	resourceTrackers := []*resources.Resource{}

	response, err := c.KMS().DescribeKey(request)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case kms.ErrCodeNotFoundException:
				return resourceTrackers, nil
			default:
				return resourceTrackers, fmt.Errorf("error listing keys: %v", err)
			}
		}
	}

	resourceTrackers = append(resourceTrackers, &resources.Resource{
		Name:    keyName,
		ID:      *response.KeyMetadata.Arn,
		Deleter: DeleteCustomerMasterKeys,
	})

	return resourceTrackers, nil
}

func DeleteCustomerMasterKeys(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	request := &kms.DisableKeyInput{
		KeyId: aws.String(r.ID),
	}

	_, err := c.KMS().DisableKey(request)
	if err != nil {
		return err
	}

	return nil
}
