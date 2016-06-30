package cloudup

import (
	"encoding/binary"
	"fmt"
	"k8s.io/kops/upup/pkg/api"
	"math/big"
	"net"
	"sort"
	"strings"
	"text/template"
)

type TemplateFunctions struct {
	cluster *api.Cluster
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
	dest["WellKnownServiceIP"] = tf.WellKnownServiceIP
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
	return tf.cluster.Spec.NetworkID != ""
}
