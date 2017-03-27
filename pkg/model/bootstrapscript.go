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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"os"
	"text/template"
)

// BootstrapScript creates the bootstrap script
type BootstrapScript struct {
	NodeUpSource     string
	NodeUpSourceHash string
	// TODO temporary field to enable workflow for vSphere cloud provider.
	AddAwsEnvironmentVariables bool

	NodeUpConfigBuilder func(ig *kops.InstanceGroup) (*nodeup.NodeUpConfig, error)
}

func (b *BootstrapScript) ResourceNodeUp(ig *kops.InstanceGroup) (*fi.ResourceHolder, error) {
	if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		// Bastions are just bare machines (currently), used as SSH jump-hosts
		return nil, nil
	}

	functions := template.FuncMap{
		"NodeUpSource": func() string {
			return b.NodeUpSource
		},
		"NodeUpSourceHash": func() string {
			return b.NodeUpSourceHash
		},
		"KubeEnv": func() (string, error) {
			config, err := b.NodeUpConfigBuilder(ig)
			if err != nil {
				return "", err
			}

			data, err := kops.ToRawYaml(config)
			if err != nil {
				return "", err
			}

			return string(data), nil
		},
		// TODO temporary code, till vsphere cloud provider gets its own VFS implementation.
		"Env1": func() string {
			if b.AddAwsEnvironmentVariables && os.Getenv("AWS_REGION") != "" {
				return "export AWS_REGION=" + os.Getenv("AWS_REGION")
			}
			return ""
		},
		"Env2": func() string {
			if b.AddAwsEnvironmentVariables && os.Getenv("AWS_ACCESS_KEY_ID") != "" {
				return "export AWS_ACCESS_KEY_ID=" + os.Getenv("AWS_ACCESS_KEY_ID")
			}
			return ""
		},
		"Env3": func() string {
			if b.AddAwsEnvironmentVariables && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
				return "export AWS_SECRET_ACCESS_KEY=" + os.Getenv("AWS_SECRET_ACCESS_KEY")
			}
			return ""
		},
	}

	nodeupTemplate := resources.AWSNodeUpTemplate
	if b.AddAwsEnvironmentVariables {
		nodeupTemplate = resources.VsphereNodeUpTemplate
	}

	templateResource, err := NewTemplateResource("nodeup", nodeupTemplate, functions, nil)
	if err != nil {
		return nil, err
	}
	return fi.WrapResource(templateResource), nil
}
