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

package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/golang/glog"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// IAMModelBuilder configures IAM objects
type IAMModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &IAMModelBuilder{}

const RolePolicyTemplate = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "{{ IAMServiceEC2 }}"},
      "Action": "sts:AssumeRole"
    }
  ]
}`

func (b *IAMModelBuilder) Build(c *fi.ModelBuilderContext) error {

	// Collect the roles in use
	var roles []kops.InstanceGroupRole
	for _, ig := range b.InstanceGroups {
		found := false
		for _, r := range roles {
			if r == ig.Spec.Role {
				found = true
			}
		}
		if !found {
			roles = append(roles, ig.Spec.Role)
		}
	}

	// Generate IAM objects etc for each role
	for _, role := range roles {
		name := b.IAMName(role)

		var iamRole *awstasks.IAMRole

		arn := ""

		// Want to use a FeatureFlag in front of this to allow the validation to harden
		if b.Cluster.Spec.AuthRole != nil && featureflag.CustomRoleSupport.Enabled() {

			roleAsString := string(role)

			if role == kops.InstanceGroupRoleMaster && b.Cluster.Spec.AuthRole.Master != nil {
				arn = *b.Cluster.Spec.AuthRole.Master
				glog.Warningf("Custom Role Support is enabled, kops will use %s, for %s role, this is an advanced feature please use with great care", arn, roleAsString)
			} else if role == kops.InstanceGroupRoleNode && b.Cluster.Spec.AuthRole.Node != nil {
				arn = *b.Cluster.Spec.AuthRole.Node
				glog.Warningf("Custom Role Support is enabled, kops will use %s, for %s role, this is an advanced feature please use with great care", arn, roleAsString)
			}

		}

		// If we've specified an IAMRoleArn for this cluster role,
		// do not create a new one
		if arn != "" {
			{

				rs := strings.Split(arn, "/")

				length := len(rs)

				if length != 2 {
					return fmt.Errorf("unable to parse role arn %q", arn)
				}

				roleName := rs[length-1]
				iamRole = &awstasks.IAMRole{
					Name: &roleName,
					ID:   &arn,

					// We set Policy Document to nil as this role will be managed externally
					RolePolicyDocument: nil,
				}

				// TODO Validate role against role that kops generates
				// TODO Where is out entry point to validate the role?
				// TODO We need to run the aws finder and we do not have access to the cloud context.
				// TODO We need to validate somewhere else.

				// Steps to validate
				// 1. get the role
				// 2. generate the role out of iam_builder
				// 3. diff the two

				c.AddTask(iamRole)
			}
		} else {

			{
				rolePolicy, err := b.buildAWSIAMRolePolicy()
				if err != nil {
					return err
				}

				iamRole = &awstasks.IAMRole{
					Name:               s(name),
					RolePolicyDocument: fi.WrapResource(rolePolicy),
					ExportWithID:       s(strings.ToLower(string(role)) + "s"),
				}
				c.AddTask(iamRole)

			}

			{
				iamPolicy := &iam.IAMPolicyResource{
					Builder: &iam.IAMPolicyBuilder{
						Cluster: b.Cluster,
						Role:    role,
						Region:  b.Region,
					},
				}

				// This is slightly tricky; we need to know the hosted zone id,
				// but we might be creating the hosted zone dynamically.

				// TODO: I don't love this technique for finding the task by name & modifying it
				dnsZoneTask, found := c.Tasks["DNSZone/"+b.NameForDNSZone()]
				if found {
					iamPolicy.DNSZone = dnsZoneTask.(*awstasks.DNSZone)
				} else {
					glog.V(2).Infof("Task %q not found; won't set route53 permissions in IAM", "DNSZone/"+b.NameForDNSZone())
				}

				t := &awstasks.IAMRolePolicy{
					Name:           s(name),
					Role:           iamRole,
					PolicyDocument: iamPolicy,
				}
				c.AddTask(t)
			}
		}

		var iamInstanceProfile *awstasks.IAMInstanceProfile
		{
			iamInstanceProfile = &awstasks.IAMInstanceProfile{
				Name: s(name),
			}
			c.AddTask(iamInstanceProfile)
		}

		{
			iamInstanceProfileRole := &awstasks.IAMInstanceProfileRole{
				Name: s(name),

				InstanceProfile: iamInstanceProfile,
				Role:            iamRole,
			}
			c.AddTask(iamInstanceProfileRole)
		}

		// Generate additional policies if needed, and attach to existing role
		{
			additionalPolicy := ""
			if b.Cluster.Spec.AdditionalPolicies != nil {
				roleAsString := reflect.ValueOf(role).String()
				additionalPolicies := *(b.Cluster.Spec.AdditionalPolicies)

				additionalPolicy = additionalPolicies[strings.ToLower(roleAsString)]
			}

			additionalPolicyName := "additional." + name

			t := &awstasks.IAMRolePolicy{
				Name: s(additionalPolicyName),
				Role: iamRole,
			}

			if additionalPolicy != "" {
				p := &iam.IAMPolicy{
					Version: iam.IAMPolicyDefaultVersion,
				}

				statements := make([]*iam.IAMStatement, 0)
				json.Unmarshal([]byte(additionalPolicy), &statements)
				p.Statement = append(p.Statement, statements...)

				policy, err := p.AsJSON()
				if err != nil {
					return fmt.Errorf("error building IAM policy: %v", err)
				}

				t.PolicyDocument = fi.WrapResource(fi.NewStringResource(policy))
			} else {
				t.PolicyDocument = fi.WrapResource(fi.NewStringResource(""))
			}

			c.AddTask(t)
		}
	}

	return nil
}

// buildAWSIAMRolePolicy produces the AWS IAM role policy for the given role
func (b *IAMModelBuilder) buildAWSIAMRolePolicy() (fi.Resource, error) {
	functions := template.FuncMap{
		"IAMServiceEC2": func() string {
			// IAMServiceEC2 returns the name of the IAM service for EC2 in the current region
			// it is ec2.amazonaws.com everywhere but in cn-north, where it is ec2.amazonaws.com.cn
			switch b.Region {
			case "cn-north-1":
				return "ec2.amazonaws.com.cn"
			default:
				return "ec2.amazonaws.com"
			}
		},
	}

	templateResource, err := NewTemplateResource("AWSIAMRolePolicy", RolePolicyTemplate, functions, nil)
	if err != nil {
		return nil, err
	}
	return templateResource, nil
}
