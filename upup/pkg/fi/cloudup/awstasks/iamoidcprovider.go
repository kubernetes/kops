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
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type IAMOIDCProvider struct {
	Lifecycle fi.Lifecycle

	ClientIDs   []*string
	Thumbprints []*string
	URL         *string

	Name *string
	Tags map[string]string

	arn *string
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

		if actualURL == fi.ValueOf(e.URL) {

			actual := &IAMOIDCProvider{
				ClientIDs:   descResp.ClientIDList,
				Thumbprints: descResp.ThumbprintList,
				URL:         &actualURL,
				Tags:        mapIAMTagsToMap(descResp.Tags),
				arn:         arn,
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
	if e.URL == nil {
		return fi.RequiredField("URL")
	}
	if e.ClientIDs == nil {
		return fi.RequiredField("ClientIDs")
	}
	if len(e.Thumbprints) == 0 {
		return fi.RequiredField("Thumbprints")
	}

	if a != nil {
		if changes.URL != nil {
			return fi.CannotChangeField("URL")
		}
	}
	return nil
}

func (p *IAMOIDCProvider) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *IAMOIDCProvider) error {
	thumbprints := e.Thumbprints

	if a == nil {
		klog.V(2).Infof("Creating IAMOIDCProvider with Name:%q", *e.Name)

		request := &iam.CreateOpenIDConnectProviderInput{
			ClientIDList:   e.ClientIDs,
			ThumbprintList: thumbprints,
			Url:            e.URL,
			Tags:           mapToIAMTags(e.Tags),
		}

		response, err := t.Cloud.IAM().CreateOpenIDConnectProvider(request)
		if err != nil {
			return fmt.Errorf("error creating IAMOIDCProvider: %v", err)
		}

		e.arn = response.OpenIDConnectProviderArn
	} else {
		if changes.Thumbprints != nil {
			klog.V(2).Infof("Updating IAMOIDCProvider Thumbprints %q", fi.ValueOf(e.arn))

			request := &iam.UpdateOpenIDConnectProviderThumbprintInput{}
			request.OpenIDConnectProviderArn = a.arn
			request.ThumbprintList = thumbprints

			_, err := t.Cloud.IAM().UpdateOpenIDConnectProviderThumbprint(request)
			if err != nil {
				return fmt.Errorf("error updating IAMOIDCProvider Thumbprints: %v", err)
			}
		}
		if changes.Tags != nil {
			if len(a.Tags) > 0 {
				existingTagKeys := make([]*string, 0)
				for k := range a.Tags {
					existingTagKeys = append(existingTagKeys, &k)
				}
				untagRequest := &iam.UntagOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: a.arn,
					TagKeys:                  existingTagKeys,
				}
				_, err := t.Cloud.IAM().UntagOpenIDConnectProvider(untagRequest)
				if err != nil {
					return fmt.Errorf("error untagging IAMOIDCProvider: %v", err)
				}
			}
			if len(e.Tags) > 0 {
				tagRequest := &iam.TagOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: a.arn,
					Tags:                     mapToIAMTags(e.Tags),
				}
				_, err := t.Cloud.IAM().TagOpenIDConnectProvider(tagRequest)
				if err != nil {
					return fmt.Errorf("error tagging IAMOIDCProvider: %v", err)
				}
			}
		}
		if changes.ClientIDs != nil {
			actual := sets.NewString()
			for _, aud := range a.ClientIDs {
				actual.Insert(*aud)
			}
			expected := sets.NewString()
			for _, aud := range e.ClientIDs {
				expected.Insert(*aud)
			}
			toRemove := actual.Difference(expected)
			for _, elem := range toRemove.List() {
				request := &iam.RemoveClientIDFromOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: a.arn,
					ClientID:                 &elem,
				}
				_, err := t.Cloud.IAM().RemoveClientIDFromOpenIDConnectProvider(request)
				if err != nil {
					return fmt.Errorf("error removing audience %s to IAMOIDCProvider: %v", elem, err)
				}
			}
			toAdd := expected.Difference(actual)
			for _, elem := range toAdd.List() {
				request := &iam.AddClientIDToOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: a.arn,
					ClientID:                 &elem,
				}
				_, err := t.Cloud.IAM().AddClientIDToOpenIDConnectProvider(request)
				if err != nil {
					return fmt.Errorf("error adding audience %s to IAMOIDCProvider: %v", elem, err)
				}
			}
		}
	}
	return nil
}

type terraformIAMOIDCProvider struct {
	URL            *string   `cty:"url"`
	ClientIDList   []*string `cty:"client_id_list"`
	ThumbprintList []*string `cty:"thumbprint_list"`

	AssumeRolePolicy *terraformWriter.Literal `cty:"assume_role_policy"`
	Tags             map[string]string        `cty:"tags"`
}

func (p *IAMOIDCProvider) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *IAMOIDCProvider) error {
	err := t.AddOutputVariable("iam_openid_connect_provider_arn", e.TerraformLink())
	if err != nil {
		return err
	}

	issuerSubs := strings.SplitAfter(aws.StringValue(e.URL), "://")
	issuer := issuerSubs[len(issuerSubs)-1]
	err = t.AddOutputVariable("iam_openid_connect_provider_issuer", terraformWriter.LiteralFromStringValue(issuer))
	if err != nil {
		return err
	}

	tf := &terraformIAMOIDCProvider{
		URL:            e.URL,
		ClientIDList:   e.ClientIDs,
		ThumbprintList: e.Thumbprints,
		Tags:           e.Tags,
	}

	return t.RenderResource("aws_iam_openid_connect_provider", *e.Name, tf)
}

func (e *IAMOIDCProvider) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("aws_iam_openid_connect_provider", *e.Name, "arn")
}
