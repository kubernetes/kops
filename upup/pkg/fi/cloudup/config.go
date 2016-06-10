package cloudup

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"math/big"
	"net"
	"strconv"
)

type CloudConfig struct {
	CloudProvider string `json:",omitempty"`

	// The version of kubernetes to install
	KubernetesVersion string `json:",omitempty"`

	// The Node initializer technique to use: cloudinit or nodeup
	NodeInit string `json:",omitempty"`

	// Configuration of zones we are targeting
	Zones       []string `json:",omitempty"`
	MasterZones []string `json:",omitempty"`
	NodeZones   []string `json:",omitempty"`
	Region      string   `json:",omitempty"`
	Project     string   `json:",omitempty"`

	// The internal and external names for the master nodes
	MasterPublicName   string `json:",omitempty"`
	MasterInternalName string `json:",omitempty"`

	// The CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s overlay
	NetworkCIDR string `json:",omitempty"`

	// The DNS zone we should use when configuring DNS
	DNSZone string `json:",omitempty"`

	InstancePrefix    string `json:",omitempty"`
	ClusterName       string `json:",omitempty"`
	AllocateNodeCIDRs *bool  `json:",omitempty"`

	Multizone bool `json:",omitempty"`

	ClusterIPRange        string `json:",omitempty"`
	ServiceClusterIPRange string `json:",omitempty"`
	MasterIPRange         string `json:",omitempty"`
	NonMasqueradeCidr     string `json:",omitempty"`

	NetworkProvider string `json:",omitempty"`

	HairpinMode string `json:",omitempty"`

	OpencontrailTag           string `json:",omitempty"`
	OpencontrailKubernetesTag string `json:",omitempty"`
	OpencontrailPublicSubnet  string `json:",omitempty"`

	EnableClusterMonitoring string `json:",omitempty"`
	EnableL7LoadBalancing   string `json:",omitempty"`
	EnableClusterUI         *bool  `json:",omitempty"`

	EnableClusterDNS *bool  `json:",omitempty"`
	DNSReplicas      int    `json:",omitempty"`
	DNSServerIP      string `json:",omitempty"`
	DNSDomain        string `json:",omitempty"`

	EnableClusterLogging         *bool  `json:",omitempty"`
	EnableNodeLogging            *bool  `json:",omitempty"`
	LoggingDestination           string `json:",omitempty"`
	ElasticsearchLoggingReplicas int    `json:",omitempty"`

	EnableClusterRegistry   *bool  `json:",omitempty"`
	ClusterRegistryDisk     string `json:",omitempty"`
	ClusterRegistryDiskSize int    `json:",omitempty"`

	EnableCustomMetrics *bool `json:",omitempty"`

	RegisterMasterKubelet *bool  `json:",omitempty"`
	MasterVolumeType      string `json:",omitempty"`
	MasterVolumeSize      int    `json:",omitempty"`
	MasterTag             string `json:",omitempty"`
	MasterMachineType     string `json:",omitempty"`
	MasterImage           string `json:",omitempty"`

	NodeImage          string `json:",omitempty"`
	NodeCount          int    `json:",omitempty"`
	NodeInstancePrefix string `json:",omitempty"`
	NodeLabels         string `json:",omitempty"`
	NodeMachineType    string `json:",omitempty"`
	NodeTag            string `json:",omitempty"`

	KubeUser string `json:",omitempty"`

	// These are moved to CAStore / SecretStore
	//KubePassword			string
	//KubeletToken                  string
	//KubeProxyToken                string
	//BearerToken                   string
	//CACert                        []byte
	//CAKey                         []byte
	//KubeletCert                   []byte
	//KubeletKey                    []byte
	//MasterCert                    []byte
	//MasterKey                     []byte
	//KubecfgCert                   []byte
	//KubecfgKey                    []byte

	AdmissionControl string `json:",omitempty"`
	RuntimeConfig    string `json:",omitempty"`

	KubeImageTag       string `json:",omitempty"`
	KubeDockerRegistry string `json:",omitempty"`
	KubeAddonRegistry  string `json:",omitempty"`

	KubeletPort int `json:",omitempty"`

	KubeApiserverRequestTimeout int `json:",omitempty"`

	TerminatedPodGcThreshold string `json:",omitempty"`

	EnableManifestURL *bool  `json:",omitempty"`
	ManifestURL       string `json:",omitempty"`
	ManifestURLHeader string `json:",omitempty"`

	TestCluster string `json:",omitempty"`

	DockerOptions   string `json:",omitempty"`
	DockerStorage   string `json:",omitempty"`
	ExtraDockerOpts string `json:",omitempty"`

	E2EStorageTestEnvironment     string `json:",omitempty"`
	KubeletTestArgs               string `json:",omitempty"`
	KubeletTestLogLevel           string `json:",omitempty"`
	DockerTestArgs                string `json:",omitempty"`
	DockerTestLogLevel            string `json:",omitempty"`
	ApiserverTestArgs             string `json:",omitempty"`
	ApiserverTestLogLevel         string `json:",omitempty"`
	ControllerManagerTestArgs     string `json:",omitempty"`
	ControllerManagerTestLogLevel string `json:",omitempty"`
	SchedulerTestArgs             string `json:",omitempty"`
	SchedulerTestLogLevel         string `json:",omitempty"`
	KubeProxyTestArgs             string `json:",omitempty"`
	KubeProxyTestLogLevel         string `json:",omitempty"`

	Assets []string `json:",omitempty"`

	NodeUpTags []string `json:",omitempty"`

	NodeUp NodeUpConfig
}

type NodeUpConfig struct {
	Location string `json:",omitempty"`
	Hash     string `json:",omitempty"`
}

func (c *CloudConfig) WellKnownServiceIP(id int) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(c.ServiceClusterIPRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing ServiceClusterIPRange: %v", err)
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

	return nil, fmt.Errorf("Unexpected IP address type for ServiceClusterIPRange: %s", c.ServiceClusterIPRange)

}

func (c *CloudConfig) SubnetCIDR(zone string) (string, error) {
	index := -1
	for i, z := range c.Zones {
		if z == zone {
			index = i
			break
		}
	}
	if index == -1 {
		return "", fmt.Errorf("zone not configured: %q", zone)
	}

	_, cidr, err := net.ParseCIDR(c.NetworkCIDR)
	if err != nil {
		return "", fmt.Errorf("Invalid NetworkCIDR: %q", c.NetworkCIDR)
	}

	networkLength, _ := cidr.Mask.Size()

	// We assume a maximum of 8 subnets per network
	// TODO: Does this make sense on GCE?
	networkLength += 3

	ip4 := cidr.IP.To4()
	if ip4 != nil {
		n := binary.BigEndian.Uint32(ip4)
		n += uint32(index) << uint(32-networkLength)
		subnetIP := make(net.IP, len(ip4))
		binary.BigEndian.PutUint32(subnetIP, n)
		subnetCIDR := subnetIP.String() + "/" + strconv.Itoa(networkLength)
		glog.V(2).Infof("Computed CIDR for subnet in zone %q as %q", zone, subnetCIDR)
		return subnetCIDR, nil
	}

	return "", fmt.Errorf("Unexpected IP address type for NetworkCIDR: %s", c.NetworkCIDR)
}
