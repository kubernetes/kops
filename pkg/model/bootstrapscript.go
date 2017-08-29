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
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
)

// BootstrapScript creates the bootstrap script
type BootstrapScript struct {
	NodeUpSource        string
	NodeUpSourceHash    string
	NodeUpConfigBuilder func(ig *kops.InstanceGroup) (*nodeup.Config, error)
}

// ResourceNodeUp generates and returns a nodeup (bootstrap) script from a
// template file, substituting in specific env vars & cluster spec configuration
func (b *BootstrapScript) ResourceNodeUp(ig *kops.InstanceGroup, cs *kops.ClusterSpec) (*fi.ResourceHolder, error) {
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
			return b.createProxyEnv(cs.EgressProxy)
		},
		"AWS_REGION": func() string {
			if os.Getenv("AWS_REGION") != "" {
				return fmt.Sprintf("export AWS_REGION=%s\n",
					os.Getenv("AWS_REGION"))
			}
			return ""
		},

		"ClusterSpec": func() (string, error) {
			spec := make(map[string]interface{})
			spec["cloudConfig"] = cs.CloudConfig
			spec["docker"] = cs.Docker
			spec["kubelet"] = cs.Kubelet
			spec["kubeProxy"] = cs.KubeProxy

			if ig.IsMaster() {
				spec["kubeAPIServer"] = cs.KubeAPIServer
				spec["kubeControllerManager"] = cs.KubeControllerManager
				spec["kubeScheduler"] = cs.KubeScheduler
				spec["masterKubelet"] = cs.MasterKubelet
			}

			hooks := b.getRelevantHooks(cs.Hooks, ig.Spec.Role)
			if len(hooks) > 0 {
				spec["hooks"] = hooks
			}

			content, err := yaml.Marshal(spec)
			if err != nil {
				return "", fmt.Errorf("error converting cluster spec to yaml for inclusion within bootstrap script: %v", err)
			}
			return string(content), nil
		},

		"IGSpec": func() (string, error) {
			spec := make(map[string]interface{})
			spec["kubelet"] = ig.Spec.Kubelet
			spec["nodeLabels"] = ig.Spec.NodeLabels
			spec["taints"] = ig.Spec.Taints
			hooks := b.getRelevantHooks(ig.Spec.Hooks, ig.Spec.Role)
			if len(hooks) > 0 {
				spec["hooks"] = hooks
			}

			content, err := yaml.Marshal(spec)
			if err != nil {
				return "", fmt.Errorf("error converting instancegroup spec to yaml for inclusion within bootstrap script: %v", err)
			}
			return string(content), nil
		},
	}

	templateResource, err := NewTemplateResource("nodeup", resources.AWSNodeUpTemplate, functions, nil)
	if err != nil {
		return nil, err
	}

	return fi.WrapResource(templateResource), nil
}

// getRelevantHooks returns a list of hooks to be applied to the instance group
func (b *BootstrapScript) getRelevantHooks(hooks []kops.HookSpec, role kops.InstanceGroupRole) []kops.HookSpec {
	relevantHooks := []kops.HookSpec{}
	for _, hook := range hooks {
		if len(hook.Roles) == 0 {
			relevantHooks = append(relevantHooks, hook)
			continue
		}
		for _, hookRole := range hook.Roles {
			if role == hookRole {
				relevantHooks = append(relevantHooks, hook)
				break
			}
		}
	}
	return relevantHooks
}

func (b *BootstrapScript) createProxyEnv(ps *kops.EgressProxySpec) string {
	var buffer bytes.Buffer

	if ps != nil && ps.HTTPProxy.Host != "" {
		var httpProxyURL string

		// TODO double check that all the code does this
		// TODO move this into a validate so we can enforce the string syntax
		if !strings.HasPrefix(ps.HTTPProxy.Host, "http://") {
			httpProxyURL = "http://"
		}

		if ps.HTTPProxy.Port != 0 {
			httpProxyURL += ps.HTTPProxy.Host + ":" + strconv.Itoa(ps.HTTPProxy.Port)
		} else {
			httpProxyURL += ps.HTTPProxy.Host
		}

		// Set base env variables
		buffer.WriteString("export http_proxy=" + httpProxyURL + "\n")
		buffer.WriteString("export https_proxy=${http_proxy}\n")
		buffer.WriteString("export no_proxy=" + ps.ProxyExcludes + "\n")
		buffer.WriteString("export NO_PROXY=${no_proxy}\n")

		// TODO move the rest of this configuration work to nodeup

		// Set env variables for docker
		buffer.WriteString("echo \"export http_proxy=${http_proxy}\" >> /etc/default/docker\n")
		buffer.WriteString("echo \"export https_proxy=${http_proxy}\" >> /etc/default/docker\n")
		buffer.WriteString("echo \"export no_proxy=${no_proxy}\" >> /etc/default/docker\n")
		buffer.WriteString("echo \"export NO_PROXY=${no_proxy}\" >> /etc/default/docker\n")

		// Set env variables for base environment
		buffer.WriteString("echo \"export http_proxy=${http_proxy}\" >> /etc/environment\n")
		buffer.WriteString("echo \"export https_proxy=${http_proxy}\" >> /etc/environment\n")
		buffer.WriteString("echo \"export no_proxy=${no_proxy}\" >> /etc/environment\n")
		buffer.WriteString("echo \"export NO_PROXY=${no_proxy}\" >> /etc/environment\n")

		// Set env variables to systemd
		buffer.WriteString("echo DefaultEnvironment=\\\"http_proxy=${http_proxy}\\\" \\\"https_proxy=${http_proxy}\\\"")
		buffer.WriteString("echo DefaultEnvironment=\\\"http_proxy=${http_proxy}\\\" \\\"https_proxy=${http_proxy}\\\"")
		buffer.WriteString(" \\\"NO_PROXY=${no_proxy}\\\" \\\"no_proxy=${no_proxy}\\\"")
		buffer.WriteString(" >> /etc/systemd/system.conf\n")

		// source in the environment this step ensures that environment file is correct
		buffer.WriteString("source /etc/environment\n")

		// Restart stuff
		buffer.WriteString("systemctl daemon-reload\n")
		buffer.WriteString("systemctl daemon-reexec\n")

		// TODO do we need no_proxy in these as well??
		// TODO handle CoreOS
		// Depending on OS set package manager proxy settings
		buffer.WriteString("if [ -f /etc/lsb-release ] || [ -f /etc/debian_version ]; then\n")
		buffer.WriteString("    echo \"Acquire::http::Proxy \\\"${http_proxy}\\\";\" > /etc/apt/apt.conf.d/30proxy\n")
		buffer.WriteString("elif [ -f /etc/redhat-release ]; then\n")
		buffer.WriteString("  echo \"http_proxy=${http_proxy}\" >> /etc/yum.conf\n")
		buffer.WriteString("fi\n")
	}
	return buffer.String()
}
