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

package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
)

// IAMModelBuilder configures IAM objects
type IAMModelBuilder struct {
	*ModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &IAMModelBuilder{}

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
		{
			rolePolicy, err := b.buildAWSIAMRolePolicy()
			if err != nil {
				return err
			}

			iamRole = &awstasks.IAMRole{
				Name:               fi.String(name),
				Lifecycle:          b.Lifecycle,
				RolePolicyDocument: fi.WrapResource(rolePolicy),
				ExportWithID:       fi.String(strings.ToLower(string(role)) + "s"),
			}
			c.AddTask(iamRole)

		}

		{
			iamPolicy := &IAMPolicyResource{
				Builder: &iam.PolicyBuilder{
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
				Name:           fi.String(name),
				Lifecycle:      b.Lifecycle,
				Role:           iamRole,
				PolicyDocument: iamPolicy,
			}
			c.AddTask(t)
		}

		var iamInstanceProfile *awstasks.IAMInstanceProfile
		{
			iamInstanceProfile = &awstasks.IAMInstanceProfile{
				Name:      fi.String(name),
				Lifecycle: b.Lifecycle,
			}
			c.AddTask(iamInstanceProfile)
		}

		{
			iamInstanceProfileRole := &awstasks.IAMInstanceProfileRole{
				Name:            fi.String(name),
				Lifecycle:       b.Lifecycle,
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
				Name:      fi.String(additionalPolicyName),
				Lifecycle: b.Lifecycle,
				Role:      iamRole,
			}

			if additionalPolicy != "" {
				p := &iam.Policy{
					Version: iam.PolicyDefaultVersion,
				}

				statements := make([]*iam.Statement, 0)
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

	templateResource, err := model.NewTemplateResource("AWSIAMRolePolicy", model.RolePolicyTemplate, functions, nil)
	if err != nil {
		return nil, err
	}
	return templateResource, nil
}

type IAMPolicyResource struct {
	Builder *iam.PolicyBuilder
	DNSZone *awstasks.DNSZone
}

var _ fi.Resource = &IAMPolicyResource{}
var _ fi.HasDependencies = &IAMPolicyResource{}

func (b *IAMPolicyResource) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	if b.DNSZone != nil {
		deps = append(deps, b.DNSZone)
	}
	return deps
}

// Open produces the AWS IAM policy for the given role
func (b *IAMPolicyResource) Open() (io.Reader, error) {
	// Defensive copy before mutation
	pb := *b.Builder

	if b.DNSZone != nil {
		hostedZoneID := fi.StringValue(b.DNSZone.ZoneID)
		if hostedZoneID == "" {
			// Dependency analysis failure?
			return nil, fmt.Errorf("DNS ZoneID not set")
		}
		pb.HostedZoneID = hostedZoneID
	}

	policy, err := pb.BuildAWSPolicy()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	json, err := policy.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("error building IAM policy: %v", err)
	}
	return bytes.NewReader([]byte(json)), nil
}
