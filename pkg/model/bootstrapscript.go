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
	"crypto/sha1"
	"encoding/base64"
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

// KubeEnv returns the nodeup config for the instance group
func (b *BootstrapScript) KubeEnv(ig *kops.InstanceGroup) (string, error) {
	config, err := b.NodeUpConfigBuilder(ig)
	if err != nil {
		return "", err
	}

	data, err := kops.ToRawYaml(config)
	if err != nil {
		return "", err
	}

	return string(data), nil
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
			return b.KubeEnv(ig)
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
				etcdSum, err := b.computeFingerprintOnStruct(cs.EtcdClusters)
				if err != nil {
					return "", err
				}
				spec["etcdFingerprint"] = etcdSum
				spec["encryptionConfig"] = cs.EncryptionConfig
				spec["kubeAPIServer"] = cs.KubeAPIServer
				spec["kubeControllerManager"] = cs.KubeControllerManager
				spec["kubeScheduler"] = cs.KubeScheduler
				spec["masterKubelet"] = cs.MasterKubelet
				spec["etcdClusters"] = make(map[string]kops.EtcdClusterSpec, 0)

				for _, etcdCluster := range cs.EtcdClusters {
					spec["etcdClusters"].(map[string]kops.EtcdClusterSpec)[etcdCluster.Name] = kops.EtcdClusterSpec{
						Image:   etcdCluster.Image,
						Version: etcdCluster.Version,
					}
				}
			}

			hooks, err := b.getRelevantHooks(cs.Hooks, ig.Spec.Role)
			if err != nil {
				return "", err
			}
			if len(hooks) > 0 {
				spec["hooks"] = hooks
			}

			fileAssets, err := b.getRelevantFileAssets(cs.FileAssets, ig.Spec.Role)
			if err != nil {
				return "", err
			}
			if len(fileAssets) > 0 {
				spec["fileAssets"] = fileAssets
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

			hooks, err := b.getRelevantHooks(ig.Spec.Hooks, ig.Spec.Role)
			if err != nil {
				return "", err
			}
			if len(hooks) > 0 {
				spec["hooks"] = hooks
			}

			fileAssets, err := b.getRelevantFileAssets(ig.Spec.FileAssets, ig.Spec.Role)
			if err != nil {
				return "", err
			}
			if len(fileAssets) > 0 {
				spec["fileAssets"] = fileAssets
			}

			content, err := yaml.Marshal(spec)
			if err != nil {
				return "", fmt.Errorf("error converting instancegroup spec to yaml for inclusion within bootstrap script: %v", err)
			}
			return string(content), nil
		},
	}

	awsNodeUpTemplate, err := resources.AWSNodeUpTemplate(ig)
	if err != nil {
		return nil, err
	}

	templateResource, err := NewTemplateResource("nodeup", awsNodeUpTemplate, functions, nil)
	if err != nil {
		return nil, err
	}

	return fi.WrapResource(templateResource), nil
}

// getRelevantHooks returns a list of hooks to be applied to the instance group,
// with the Manifest and ExecContainer Commands fingerprinted to reduce size
func (b *BootstrapScript) getRelevantHooks(allHooks []kops.HookSpec, role kops.InstanceGroupRole) ([]kops.HookSpec, error) {
	relevantHooks := []kops.HookSpec{}
	for _, hook := range allHooks {
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

	hooks := []kops.HookSpec{}
	if len(relevantHooks) > 0 {
		for _, hook := range relevantHooks {
			if hook.Manifest != "" {
				manifestFingerprint, err := b.computeFingerprint(hook.Manifest)
				if err != nil {
					return nil, err
				}
				hook.Manifest = manifestFingerprint + " (fingerprint)"
			}

			if hook.ExecContainer != nil && hook.ExecContainer.Command != nil {
				execContainerCommandFingerprint, err := b.computeFingerprint(strings.Join(hook.ExecContainer.Command[:], " "))
				if err != nil {
					return nil, err
				}

				execContainerAction := &kops.ExecContainerAction{
					Command:     []string{execContainerCommandFingerprint + " (fingerprint)"},
					Environment: hook.ExecContainer.Environment,
					Image:       hook.ExecContainer.Image,
				}
				hook.ExecContainer = execContainerAction
			}

			hook.Roles = nil
			hooks = append(hooks, hook)
		}
	}

	return hooks, nil
}

// getRelevantFileAssets returns a list of file assets to be applied to the
// instance group, with the Content fingerprinted to reduce size
func (b *BootstrapScript) getRelevantFileAssets(allFileAssets []kops.FileAssetSpec, role kops.InstanceGroupRole) ([]kops.FileAssetSpec, error) {
	relevantFileAssets := []kops.FileAssetSpec{}
	for _, fileAsset := range allFileAssets {
		if len(fileAsset.Roles) == 0 {
			relevantFileAssets = append(relevantFileAssets, fileAsset)
			continue
		}
		for _, fileAssetRole := range fileAsset.Roles {
			if role == fileAssetRole {
				relevantFileAssets = append(relevantFileAssets, fileAsset)
				break
			}
		}
	}

	fileAssets := []kops.FileAssetSpec{}
	if len(relevantFileAssets) > 0 {
		for _, fileAsset := range relevantFileAssets {
			if fileAsset.Content != "" {
				contentFingerprint, err := b.computeFingerprint(fileAsset.Content)
				if err != nil {
					return nil, err
				}
				fileAsset.Content = contentFingerprint + " (fingerprint)"
			}

			fileAsset.Roles = nil
			fileAssets = append(fileAssets, fileAsset)
		}
	}

	return fileAssets, nil
}

// computeFingerprintOnStruct is computed on the struct pointer
func (b *BootstrapScript) computeFingerprintOnStruct(v interface{}) (string, error) {
	content, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	return b.computeFingerprint(string(content))
}

// computeFingerprint takes a string and returns a base64 encoded fingerprint
func (b *BootstrapScript) computeFingerprint(content string) (string, error) {
	hasher := sha1.New()

	if _, err := hasher.Write([]byte(content)); err != nil {
		return "", fmt.Errorf("error computing fingerprint hash: %v", err)
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
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

		// Set env variables for base environment
		buffer.WriteString(`echo "http_proxy=` + httpProxyURL + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "https_proxy=` + httpProxyURL + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "no_proxy=` + ps.ProxyExcludes + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "NO_PROXY=` + ps.ProxyExcludes + `" >> /etc/environment` + "\n")

		// Load the proxy environment variables
		buffer.WriteString("while read in; do export $in; done < /etc/environment\n")

		// Set env variables for package manager depending on OS Distribution (N/A for CoreOS)
		// Note: Nodeup will source the `/etc/environment` file within docker config in the correct location
		buffer.WriteString("case `cat /proc/version` in\n")
		buffer.WriteString("*[Dd]ebian*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"${http_proxy}\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Uu]buntu*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"${http_proxy}\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Rr]ed[Hh]at*)\n")
		buffer.WriteString(`  echo "http_proxy=${http_proxy}" >> /etc/yum.conf ;;` + "\n")
		buffer.WriteString("esac\n")

		// Set env variables for systemd
		buffer.WriteString(`echo "DefaultEnvironment=\"http_proxy=${http_proxy}\" \"https_proxy=${http_proxy}\"`)
		buffer.WriteString(` \"NO_PROXY=${no_proxy}\" \"no_proxy=${no_proxy}\""`)
		buffer.WriteString(" >> /etc/systemd/system.conf\n")

		// Restart stuff
		buffer.WriteString("systemctl daemon-reload\n")
		buffer.WriteString("systemctl daemon-reexec\n")
	}
	return buffer.String()
}
