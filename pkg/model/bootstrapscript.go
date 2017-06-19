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
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"os"
	"strconv"
	"text/template"
)

// BootstrapScript creates the bootstrap script
type BootstrapScript struct {
	NodeUpSource        string
	NodeUpSourceHash    string
	NodeUpConfigBuilder func(ig *kops.InstanceGroup) (*nodeup.NodeUpConfig, error)
}

func (b *BootstrapScript) ResourceNodeUp(ig *kops.InstanceGroup, ps *kops.EgressProxySpec) (*fi.ResourceHolder, error) {
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

		// Pass in extra environment variables for user-defined S3 service
		"S3Env": func() string {
			if os.Getenv("S3_ENDPOINT") != "" {
				return fmt.Sprintf("export S3_ENDPOINT=%s\nexport S3_REGION=%s\nexport S3_ACCESS_KEY_ID=%s\nexport S3_SECRET_ACCESS_KEY=%s\n",
					os.Getenv("S3_ENDPOINT"),
					os.Getenv("S3_REGION"),
					os.Getenv("S3_ACCESS_KEY_ID"),
					os.Getenv("S3_SECRET_ACCESS_KEY"))
			}
			return ""
		},

		"ProxyEnv": func() string {
			scriptSnippet := ""

			if ps != nil && ps.HTTPProxy.Host != "" {
				httpProxyUrl := "http://"
				if ps.HTTPProxy.User != "" {

					httpProxyUrl += ps.HTTPProxy.User
					if ps.HTTPProxy.Password != "" {
						httpProxyUrl += "@" + ps.HTTPProxy.Password
					}
				}
				httpProxyUrl += ps.HTTPProxy.Host + ":" + strconv.Itoa(ps.HTTPProxy.Port)
				scriptSnippet =
					"export http_proxy=" + httpProxyUrl + "\n" +
						"export https_proxy=${http_proxy}\n" +
						"export ftp_proxy=${http_proxy}\n" +
						"export no_proxy=" + ps.ProxyExcludes + "\n" +
						"echo \"export http_proxy=${http_proxy}\" >> /etc/default/docker\n" +
						"echo \"export https_proxy=${http_proxy}\" >> /etc/default/docker\n" +
						"echo \"export ftp_proxy=${http_proxy}\" >> /etc/default/docker\n" +
						"echo \"export no_proxy=${no_proxy}\" >> /etc/default/docker\n" +
						"echo \"export http_proxy=${http_proxy}\" >> /etc/environment\n" +
						"echo \"export https_proxy=${http_proxy}\" >> /etc/environment\n" +
						"echo \"export ftp_proxy=${http_proxy}\" >> /etc/environment\n" +
						"echo \"export no_proxy=${no_proxy}\" >> /etc/environment\n" +
						"echo DefaultEnvironment=\\\"http_proxy=${http_proxy}\\\" \\\"https_proxy=${http_proxy}\\\" \\\"ftp_proxy=${http_proxy}\\\" \\\"no_proxy=${no_proxy}\\\" >> /etc/systemd/system.conf\n" +
						"systemctl daemon-reload\n" +
						"systemctl daemon-reexec\n" +
						"echo \"Acquire::http::Proxy \\\"${http_proxy}\\\";\" > /etc/apt/apt.conf.d/30proxy\n\n"
			}
			return scriptSnippet
		},
	}

	templateResource, err := NewTemplateResource("nodeup", resources.AWSNodeUpTemplate, functions, nil)
	if err != nil {
		return nil, err
	}
	return fi.WrapResource(templateResource), nil
}
