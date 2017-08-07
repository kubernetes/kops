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

/******************************************************************************
Template Functions are what map functions in the models, to internal logic in
kops. This is the point where we connect static YAML configuration to dynamic
runtime values in memory.

When defining a new function:
	- Build the new function here
	- Define the new function in AddTo()
		dest["MyNewFunction"] = MyNewFunction // <-- Function Pointer
******************************************************************************/

package cloudup

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type TemplateFunctions struct {
	cluster        *kops.Cluster
	instanceGroups []*kops.InstanceGroup

	tags   sets.String
	region string

	modelContext *model.KopsModelContext
}

// This will define the available functions we can use in our YAML models
// If we are trying to get a new function implemented it MUST
// be defined here.
func (tf *TemplateFunctions) AddTo(dest template.FuncMap) {
	dest["SharedVPC"] = tf.SharedVPC

	// Remember that we may be on a different arch from the target.  Hard-code for now.
	dest["Arch"] = func() string { return "amd64" }

	dest["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	dest["replace"] = func(s, find, replace string) string {
		return strings.Replace(s, find, replace, -1)
	}
	dest["join"] = func(a []string, sep string) string {
		return strings.Join(a, sep)
	}

	dest["ClusterName"] = tf.modelContext.ClusterName

	dest["HasTag"] = tf.HasTag

	dest["Image"] = tf.Image

	dest["WithDefaultBool"] = func(v *bool, defaultValue bool) bool {
		if v != nil {
			return *v
		}
		return defaultValue
	}

	dest["GetInstanceGroup"] = tf.GetInstanceGroup

	dest["CloudTags"] = tf.modelContext.CloudTagsForInstanceGroup

	dest["KubeDNS"] = func() *kops.KubeDNSConfig {
		return tf.cluster.Spec.KubeDNS
	}

	dest["DnsControllerArgv"] = tf.DnsControllerArgv

	dest["ExternalDnsArgv"] = tf.ExternalDnsArgv

	// TODO: Only for GCE?
	dest["EncodeGCELabel"] = gce.EncodeGCELabel

	dest["Region"] = func() string {
		return tf.region
	}

	dest["ProxyEnv"] = tf.ProxyEnv
}

// SharedVPC is a simple helper function which makes the templates for a shared VPC clearer
func (tf *TemplateFunctions) SharedVPC() bool {
	return tf.cluster.SharedVPC()
}

// Image returns the docker image name for the specified component
func (tf *TemplateFunctions) Image(component string) (string, error) {
	return components.Image(component, &tf.cluster.Spec)
}

// HasTag returns true if the specified tag is set
func (tf *TemplateFunctions) HasTag(tag string) bool {
	_, found := tf.tags[tag]
	return found
}

// GetInstanceGroup returns the instance group with the specified name
func (tf *TemplateFunctions) GetInstanceGroup(name string) (*kops.InstanceGroup, error) {
	for _, ig := range tf.instanceGroups {
		if ig.ObjectMeta.Name == name {
			return ig, nil
		}
	}
	return nil, fmt.Errorf("InstanceGroup %q not found", name)
}

func (tf *TemplateFunctions) DnsControllerArgv() ([]string, error) {
	var argv []string

	argv = append(argv, "/usr/bin/dns-controller")

	argv = append(argv, "--watch-ingress=false")

	switch kops.CloudProviderID(tf.cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		if strings.HasPrefix(os.Getenv("AWS_REGION"), "cn-") {
			argv = append(argv, "--dns=gossip")
		} else {
			argv = append(argv, "--dns=aws-route53")
		}
	case kops.CloudProviderGCE:
		argv = append(argv, "--dns=google-clouddns")
	case kops.CloudProviderVSphere:
		argv = append(argv, "--dns=coredns")
		argv = append(argv, "--dns-server="+*tf.cluster.Spec.CloudConfig.VSphereCoreDNSServer)

	default:
		return nil, fmt.Errorf("unhandled cloudprovider %q", tf.cluster.Spec.CloudProvider)
	}

	if dns.IsGossipHostname(tf.cluster.Spec.MasterInternalName) {
		argv = append(argv, "--gossip-seed=127.0.0.1:3999")
	}

	zone := tf.cluster.Spec.DNSZone
	if zone != "" {
		if strings.Contains(zone, ".") {
			// match by name
			argv = append(argv, "--zone="+zone)
		} else {
			// match by id
			argv = append(argv, "--zone=*/"+zone)
		}
	}
	// permit wildcard updates
	argv = append(argv, "--zone=*/*")

	// Verbose, but not crazy logging
	argv = append(argv, "-v=2")

	return argv, nil
}

func (tf *TemplateFunctions) ExternalDnsArgv() ([]string, error) {
	var argv []string

	cloudProvider := tf.cluster.Spec.CloudProvider

	switch kops.CloudProviderID(cloudProvider) {
	case kops.CloudProviderAWS:
		argv = append(argv, "--provider=aws")
	case kops.CloudProviderGCE:
		project := tf.cluster.Spec.Project
		argv = append(argv, "--provider=google")
		argv = append(argv, "--google-project="+project)
	default:
		return nil, fmt.Errorf("unhandled cloudprovider %q", tf.cluster.Spec.CloudProvider)
	}

	argv = append(argv, "--source=ingress")

	return argv, nil
}

func (tf *TemplateFunctions) ProxyEnv() map[string]string {
	envs := map[string]string{}
	proxies := tf.cluster.Spec.EgressProxy
	if proxies == nil {
		return envs
	}
	httpProxy := proxies.HTTPProxy
	if httpProxy.Host != "" {
		var portSuffix string
		if httpProxy.Port != 0 {
			portSuffix = ":" + strconv.Itoa(httpProxy.Port)
		} else {
			portSuffix = ""
		}
		url := "http://" + httpProxy.Host + portSuffix
		envs["http_proxy"] = url
		envs["https_proxy"] = url
	}
	if proxies.ProxyExcludes != "" {
		envs["no_proxy"] = proxies.ProxyExcludes
		envs["NO_PROXY"] = proxies.ProxyExcludes
	}
	return envs
}
