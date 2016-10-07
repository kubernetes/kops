package cloudup

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/api"
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

func (tf *TemplateFunctions) AddTo(dest template.FuncMap) {
	dest["EtcdClusterMemberTags"] = tf.EtcdClusterMemberTags
	dest["SharedVPC"] = tf.SharedVPC

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

// SharedZone is a simple helper function which makes the templates for a shared Zone clearer
func (tf *TemplateFunctions) SharedZone(zone *api.ClusterZoneSpec) bool {
	return zone.ProviderID != ""
}

// AdminCIDR returns the single CIDR that is allowed access to the admin ports of the cluster (22, 443 on master)
func (tf *TemplateFunctions) AdminCIDR() (string, error) {
	if len(tf.cluster.Spec.AdminAccess) == 0 {
		return "0.0.0.0/0", nil
	}
	if len(tf.cluster.Spec.AdminAccess) == 1 {
		return tf.cluster.Spec.AdminAccess[0], nil
	}
	return "", fmt.Errorf("Multiple AdminAccess rules are not (currently) supported")
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
