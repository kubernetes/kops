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

package awstasks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=CMK

type CMK struct {
	Name      *string
	Arn       *string
	Lifecycle *fi.Lifecycle
	Shared    *bool
	Tags      map[string]string
}

func (e *CMK) Find(c *fi.Context) (*CMK, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	keyData, err := findCustomerMasterKey(cloud, *e.Name)
	if err != nil {
		return nil, err
	}

	if keyData == nil {
		return nil, nil
	}

	if e.Arn == nil {
		e.Arn = keyData.Arn
	}

	return &CMK{
		Name:      e.Name,
		Arn:       keyData.Arn,
		Lifecycle: e.Lifecycle,
		Shared:    e.Shared,
	}, nil
}

func isARN(name string) bool {
	return strings.Contains(name, "arn:aws:kms:")
}

func findCustomerMasterKey(cloud awsup.AWSCloud, name string) (*kms.KeyMetadata, error) {
	request := &kms.DescribeKeyInput{}

	if isARN(name) {
		request.KeyId = aws.String(name)
	} else {
		request.KeyId = cleanAlias(name)
	}

	response, err := cloud.KMS().DescribeKey(request)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case kms.ErrCodeNotFoundException:
				return nil, nil
			default:
				return nil, fmt.Errorf("error listing keys: %v", err)
			}
		}
	}

	return response.KeyMetadata, nil
}

func (k *CMK) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(k, c)
}

func (k *CMK) CheckChanges(a, e, changes *CMK) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	}

	return nil
}

func cleanAlias(s string) *string {
	return aws.String("alias/" + strings.Replace(s, ".", "_", -1))
}

func (_ *CMK) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *CMK) error {
	if a == nil {
		if isARN(*e.Name) {
			return fmt.Errorf("Key %s does not exist", *e.Name)
		} else {
			keyRequest := &kms.CreateKeyInput{}
			keyData, err := t.Cloud.KMS().CreateKey(keyRequest)
			if err != nil {
				return fmt.Errorf("error creating key: %v", err)
			}

			e.Arn = keyData.KeyMetadata.Arn

			aliasRequest := &kms.CreateAliasInput{
				AliasName:   cleanAlias(*e.Name),
				TargetKeyId: keyData.KeyMetadata.Arn,
			}

			_, err = t.Cloud.KMS().CreateAlias(aliasRequest)
			if err != nil {
				return fmt.Errorf("error creating key alias: %v", err)
			}
		}
	}

	return nil
}

type terraformCMK struct {
}

func (_ *CMK) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *CMK) error {
	return nil
}

type cloudformationCMK struct {
}

func (_ *CMK) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *CMK) error {
	return nil
}
