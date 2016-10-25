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
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
	"math/big"
	"net"
	"sort"
	"strings"
	"text/template"
)

type TemplateFunctions struct {
	cluster        *api.Cluster
	instanceGroups []*api.InstanceGroup

	tags   map[string]struct{}
	region string
}

func (tf *TemplateFunctions) WellKnownServiceIP(id int) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(tf.cluster.Spec.ServiceClusterIPRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing ServiceClusterIPRange %q: %v", tf.cluster.Spec.ServiceClusterIPRange, err)
	}

	ip4 := cidr.IP.To4()
	if ip4 != nil {
		n := binary.BigEndian.Uint32(ip4)
		n += uint32(id)
		serviceIP := make(net.IP, len(ip4))
		binary.BigEndian.PutUint32(serviceIP, n)
		return serviceIP, nil
	}

	ip6 := cidr.IP.To16()
	if ip6 != nil {
		baseIPInt := big.NewInt(0)
		baseIPInt.SetBytes(ip6)
		serviceIPInt := big.NewInt(0)
		serviceIPInt.Add(big.NewInt(int64(id)), baseIPInt)
		serviceIP := make(net.IP, len(ip6))
		serviceIPBytes := serviceIPInt.Bytes()
		for i := range serviceIPBytes {
			serviceIP[len(serviceIP)-len(serviceIPBytes)+i] = serviceIPBytes[i]
		}
		return serviceIP, nil
	}

	return nil, fmt.Errorf("Unexpected IP address type for ServiceClusterIPRange: %s", tf.cluster.Spec.ServiceClusterIPRange)
}

// This will define the available functions we can use in our YAML models
// If we are trying to get a new function implemented it MUST
// be defined here.
func (tf *TemplateFunctions) AddTo(dest template.FuncMap) {
	dest["EtcdClusterMemberTags"] = tf.EtcdClusterMemberTags
	dest["SharedVPC"] = tf.SharedVPC

	// Network topology definitions
	dest["IsTopologyPublic"]  = tf.IsTopologyPublic
	dest["IsTopologyPrivate"] = tf.IsTopologyPrivate
	dest["IsTopologyPrivateMasters"] = tf.IsTopologyPrivateMasters()

	dest["SharedZone"] = tf.SharedZone
	dest["WellKnownServiceIP"] = tf.WellKnownServiceIP
	dest["AdminCIDR"] = tf.AdminCIDR

	dest["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	dest["replace"] = func(s, find, replace string) string {
		return strings.Replace(s, find, replace, -1)
	}
	dest["join"] = func(a []string, sep string) string {
		return strings.Join(a, sep)
	}

	dest["ClusterName"] = func() string {
		return tf.cluster.Name
	}

	dest["HasTag"] = tf.HasTag

	dest["IAMServiceEC2"] = tf.IAMServiceEC2

	dest["Image"] = tf.Image

	dest["IAMMasterPolicy"] = func() (string, error) {
		return tf.buildAWSIAMPolicy(api.InstanceGroupRoleMaster)
	}
	dest["IAMNodePolicy"] = func() (string, error) {
		return tf.buildAWSIAMPolicy(api.InstanceGroupRoleNode)
	}
	dest["WithDefaultBool"] = func(v *bool, defaultValue bool) bool {
		if v != nil {
			return *v
		}
		return defaultValue
	}

	dest["GetInstanceGroup"] = tf.GetInstanceGroup

	dest["CloudTags"] = tf.CloudTags

	dest["APIServerCount"] = tf.APIServerCount

	dest["KubeDNS"] = func() *api.KubeDNSConfig {
		return tf.cluster.Spec.KubeDNS
	}

	dest["DnsControllerArgv"] = tf.DnsControllerArgv
}

func (tf *TemplateFunctions) EtcdClusterMemberTags(etcd *api.EtcdClusterSpec, m *api.EtcdMemberSpec) map[string]string {
	tags := make(map[string]string)

	var allMembers []string

	for _, m := range etcd.Members {
		allMembers = append(allMembers, m.Name)
	}

	sort.Strings(allMembers)

	// This is the configuration of the etcd cluster
	tags["k8s.io/etcd/"+etcd.Name] = m.Name + "/" + strings.Join(allMembers, ",")

	// This says "only mount on a master"
	tags["k8s.io/role/master"] = "1"

	return tags
}

// SharedVPC is a simple helper function which makes the templates for a shared VPC clearer
func (tf *TemplateFunctions) SharedVPC() bool {
	return tf.cluster.SharedVPC()
}

// These are the network topology functions. They are boolean logic for checking which type of
// topology this cluster is set to be deployed with.
func (tf *TemplateFunctions) IsTopologyPrivate()         bool  { return tf.cluster.IsTopologyPrivate() }
func (tf *TemplateFunctions) IsTopologyPublic()          bool  { return tf.cluster.IsTopologyPublic() }
func (tf *TemplateFunctions) IsTopologyPrivateMasters()  bool  { return tf.cluster.IsTopologyPrivateMasters() }

// SharedZone is a simple helper function which makes the templates for a shared Zone clearer
func (tf *TemplateFunctions) SharedZone(zone *api.ClusterZoneSpec) bool {
	return zone.ProviderID != ""
}

// AdminCIDR returns the CIDRs that are allowed to access the admin ports of the cluster
// (22, 443 on master and 22 on nodes)
func (tf *TemplateFunctions) AdminCIDR() []string {
	if len(tf.cluster.Spec.AdminAccess) == 0 {
		return []string{"0.0.0.0/0"}
	}
	return tf.cluster.Spec.AdminAccess
}

// IAMServiceEC2 returns the name of the IAM service for EC2 in the current region
// it is ec2.amazonaws.com everywhere but in cn-north, where it is ec2.amazonaws.com.cn
func (tf *TemplateFunctions) IAMServiceEC2() string {
	switch tf.region {
	case "cn-north-1":
		return "ec2.amazonaws.com.cn"
	default:
		return "ec2.amazonaws.com"
	}
}

// Image returns the docker image name for the specified component
func (tf *TemplateFunctions) Image(component string) (string, error) {
	if component == "kube-dns" {
		// TODO: Once we are shipping different versions, start to use them
		return "gcr.io/google_containers/kubedns-amd64:1.3", nil
	}

	if !isBaseURL(tf.cluster.Spec.KubernetesVersion) {
		return "gcr.io/google_containers/" + component + ":" + "v" + tf.cluster.Spec.KubernetesVersion, nil
	}

	baseURL := tf.cluster.Spec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	tagURL := baseURL + "/bin/linux/amd64/" + component + ".docker_tag"
	glog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	glog.V(2).Infof("Found tag %q for %q", tag, component)

	return "gcr.io/google_containers/" + component + ":" + tag, nil
}

// HasTag returns true if the specified tag is set
func (tf *TemplateFunctions) HasTag(tag string) bool {
	_, found := tf.tags[tag]
	return found
}

// buildAWSIAMPolicy produces the AWS IAM policy for the given role
func (tf *TemplateFunctions) buildAWSIAMPolicy(role api.InstanceGroupRole) (string, error) {
	b := &IAMPolicyBuilder{
		Cluster: tf.cluster,
		Role:    role,
		Region:  tf.region,
	}

	policy, err := b.BuildAWSIAMPolicy()
	if err != nil {
		return "", fmt.Errorf("error building IAM policy: %v", err)
	}
	json, err := policy.AsJSON()
	if err != nil {
		return "", fmt.Errorf("error building IAM policy: %v", err)
	}
	return json, nil
}

// CloudTags computes the tags to apply to instances in the specified InstanceGroup
func (tf *TemplateFunctions) CloudTags(ig *api.InstanceGroup) (map[string]string, error) {
	labels := make(map[string]string)

	// Apply any user-specified labels
	for k, v := range ig.Spec.CloudLabels {
		labels[k] = v
	}

	// The system tags take priority because the cluster likely breaks without them...

	if ig.Spec.Role == api.InstanceGroupRoleMaster {
		labels["k8s.io/role/master"] = "1"
	}

	if ig.Spec.Role == api.InstanceGroupRoleNode {
		labels["k8s.io/role/node"] = "1"
	}

	return labels, nil
}

// GetInstanceGroup returns the instance group with the specified name
func (tf *TemplateFunctions) GetInstanceGroup(name string) (*api.InstanceGroup, error) {
	for _, ig := range tf.instanceGroups {
		if ig.Name == name {
			return ig, nil
		}
	}
	return nil, fmt.Errorf("InstanceGroup %q not found", name)
}

// APIServerCount returns the value for the apiserver --apiserver-count flag
func (tf *TemplateFunctions) APIServerCount() int {
	count := 0
	for _, ig := range tf.instanceGroups {
		if !ig.IsMaster() {
			continue
		}
		size := fi.IntValue(ig.Spec.MaxSize)
		if size == 0 {
			size = fi.IntValue(ig.Spec.MinSize)
		}
		count += size
	}
	return count
}

func (tf *TemplateFunctions) DnsControllerArgv() ([]string, error) {
	var argv []string

	argv = append(argv, "/usr/bin/dns-controller")

	argv = append(argv, "--watch-ingress=false")
	argv = append(argv, "--dns=aws-route53")

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
	argv = append(argv, "-v=8")

	return argv, nil
}