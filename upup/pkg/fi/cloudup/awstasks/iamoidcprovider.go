/*
Copyright 2020 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

//go:generate fitask -type=IAMOIDCProvider
type IAMOIDCProvider struct {
	Lifecycle *fi.Lifecycle

	ARN         *string
	ClientIDs   []*string
	Thumbprints []*string
	URL         *string

	Name *string
}

var _ fi.CompareWithID = &IAMOIDCProvider{}

func (e *IAMOIDCProvider) CompareWithID() *string {
	return e.Name
}

func (e *IAMOIDCProvider) Find(c *fi.Context) (*IAMOIDCProvider, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	response, err := cloud.IAM().ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing oidc providers: %v", err)
	}

	providers := response.OpenIDConnectProviderList
	for _, provider := range providers {
		arn := provider.Arn
		descResp, err := cloud.IAM().GetOpenIDConnectProvider(&iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: arn,
		})
		if err != nil {
			return nil, fmt.Errorf("error describing oidc provider: %v", err)
		}
		// AWS does not return the https:// in the url
		actualURL := aws.StringValue(descResp.Url)
		if !strings.Contains(actualURL, "://") {
			actualURL = "https://" + actualURL
		}

		if actualURL == fi.StringValue(e.URL) {
			actual := &IAMOIDCProvider{
				ClientIDs:   descResp.ClientIDList,
				Thumbprints: descResp.ThumbprintList,
				URL:         &actualURL,
				ARN:         arn,
			}

			actual.Lifecycle = e.Lifecycle
			actual.Name = e.Name

			klog.V(2).Infof("found matching IAMOIDCProvider %q", aws.StringValue(arn))
			return actual, nil
		}
	}
	return nil, nil
}

func (e *IAMOIDCProvider) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *IAMOIDCProvider) CheckChanges(a, e, changes *IAMOIDCProvider) error {
	if a != nil {
		if e.URL == nil {
			return fi.RequiredField("URL")
		}
		if e.ClientIDs == nil {
			return fi.RequiredField("ClientIDs")
		}
		if len(e.Thumbprints) == 0 {
			return fi.RequiredField("Thumbprints")
		}
	} else {
		if changes.ClientIDs == nil {
			return fi.CannotChangeField("ClientIDs")
		}
		if changes.URL == nil {
			return fi.CannotChangeField("URL")
		}
	}
	return nil
}

func (p *IAMOIDCProvider) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMOIDCProvider) error {
	if a == nil {

		klog.V(2).Infof("Creating IAMOIDCProvider with Name:%q", *e.Name)

		request := &iam.CreateOpenIDConnectProviderInput{
			ClientIDList:   e.ClientIDs,
			ThumbprintList: e.Thumbprints,
			Url:            e.URL,
		}

		response, err := t.Cloud.IAM().CreateOpenIDConnectProvider(request)
		if err != nil {
			return fmt.Errorf("error creating IAMOIDCProvider: %v", err)
		}

		e.ARN = response.OpenIDConnectProviderArn
	} else {
		if changes.Thumbprints != nil {
			klog.V(2).Infof("Updating IAMOIDCProvider Thumbprints %q", *e.ARN)

			request := &iam.UpdateOpenIDConnectProviderThumbprintInput{}
			request.OpenIDConnectProviderArn = e.ARN
			request.ThumbprintList = e.Thumbprints

			_, err := t.Cloud.IAM().UpdateOpenIDConnectProviderThumbprint(request)
			if err != nil {
				return fmt.Errorf("error updating IAMOIDCProvider Thumbprints: %v", err)
			}
		}
	}
	return nil
}

type terraformIAMOIDCProvider struct {
	URL            *string   `json:"url" cty:"url"`
	ClientIDList   []*string `json:"client_id_list" cty:"client_id_list"`
	ThumbprintList []*string `json:"thumbprint_list" cty:"thumbprint_list"`

	Name             *string            `json:"name" cty:"name"`
	AssumeRolePolicy *terraform.Literal `json:"assume_role_policy" cty:"assume_role_policy"`
}

func (p *IAMOIDCProvider) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMOIDCProvider) error {

	tf := &terraformIAMOIDCProvider{
		Name:           e.Name,
		URL:            e.URL,
		ClientIDList:   e.ClientIDs,
		ThumbprintList: e.Thumbprints,
	}

	return t.RenderResource("aws_iam_openid_connect_provider", *e.Name, tf)
}

func (e *IAMOIDCProvider) TerraformLink() *terraform.Literal {
	return terraform.LiteralProperty("aws_iam_openid_connect_provider", *e.Name, "arn")
}

func (_ *IAMOIDCProvider) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *IAMOIDCProvider) error {
	return errors.New("cloudformation does not support IAM OIDC Provider")
}
