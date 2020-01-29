/*
Copyright 2019 The Kubernetes Authors.

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

package alimodel

import (
	"encoding/json"

	"github.com/denverdino/aliyungo/ram"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

const PolicyType = string(ram.Custom)

type RAMModelBuilder struct {
	*ALIModelContext

	Lifecycle *fi.Lifecycle
}

type AssumeRolePolicyDocument struct {
	Statement []AssumeRolePolicyItem
	Version   string
}

type AssumeRolePolicyItem struct {
	Action    string
	Effect    string
	Principal AssumeRolePolicyPrincpal
}

type AssumeRolePolicyPrincpal struct {
	Service []string
}

var _ fi.ModelBuilder = &RAMModelBuilder{}

func (b *RAMModelBuilder) Build(c *fi.ModelBuilderContext) error {
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

	// Generate RAM objects etc for each role
	for _, role := range roles {
		ramName := b.GetNameForRAM(role)
		err := b.buildRAMTasks(role, ramName, c, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *RAMModelBuilder) buildRAMTasks(igRole kops.InstanceGroupRole, ramName string, c *fi.ModelBuilderContext, shared bool) error {
	var ramRole *alitasks.RAMRole
	{
		assumeRolePolicyDocument := b.CreateAssumeRolePolicyDocument()
		ramRole = &alitasks.RAMRole{
			Name:                     s(ramName),
			Lifecycle:                b.Lifecycle,
			AssumeRolePolicyDocument: s(assumeRolePolicyDocument),
		}
		c.AddTask(ramRole)
	}

	{
		policyDocument := &PolicyResource{
			Builder: &PolicyBuilder{
				Cluster: b.Cluster,
				Role:    igRole,
				Region:  b.Region,
			},
		}

		// policyDocument := b.CreatePolicyDocument()
		policyType := PolicyType
		ramPolicy := &alitasks.RAMPolicy{
			Name:           s(ramName),
			Lifecycle:      b.Lifecycle,
			PolicyDocument: policyDocument,
			RamRole:        ramRole,
			PolicyType:     s(policyType),
		}
		c.AddTask(ramPolicy)
	}

	return nil
}

func (b *RAMModelBuilder) CreateAssumeRolePolicyDocument() string {
	princpal := AssumeRolePolicyPrincpal{Service: []string{"ecs.aliyuncs.com"}}

	policydocument := AssumeRolePolicyDocument{
		Statement: []AssumeRolePolicyItem{
			{Action: "sts:AssumeRole", Effect: "Allow", Principal: princpal},
		},
		Version: "1",
	}
	rolePolicy, _ := json.Marshal(policydocument)
	return string(rolePolicy)
}

func (b *RAMModelBuilder) CreatePolicyDocument() string {
	policydocument := ram.PolicyDocument{
		Statement: []ram.PolicyItem{
			{
				Action:   "oss:List*",
				Effect:   "Allow",
				Resource: "*",
			},

			{
				Action:   "oss:Get*",
				Effect:   "Allow",
				Resource: "*",
			},

			{
				Action:   "ecs:Describe*",
				Effect:   "Allow",
				Resource: "*",
			},

			{
				Action:   "slb:Describe*",
				Effect:   "Allow",
				Resource: "*",
			},
		},
		Version: "1",
	}

	rolePolicy, _ := json.Marshal(policydocument)
	return string(rolePolicy)
}
