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
	MasterZones []string      `json:",omitempty"`
	NodeZones   []*ZoneConfig `json:",omitempty"`
	Region      string        `json:",omitempty"`
	Project     string        `json:",omitempty"`

	// Permissions to configure in IAM or GCE
	MasterPermissions *CloudPermissions `json:",omitempty"`
	NodePermissions   *CloudPermissions `json:",omitempty"`

	// The internal and external names for the master nodes
	MasterPublicName   string `json:",omitempty"`
	MasterInternalName string `json:",omitempty"`

	// The CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s overlay
	NetworkCIDR string `json:",omitempty"`
	NetworkID   string `json:",omitempty"`

	SecretStore string `json:",omitempty"`
	KeyStore    string `json:",omitempty"`
	ConfigStore string `json:",omitempty"`

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

	NodeUp *NodeUpConfig `json:",omitempty"`
}

type ZoneConfig struct {
	Name string `json:"name"`
	CIDR string `json:"cidr,omitempty"`
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

func (c *CloudConfig) PerformAssignments() error {
	if c.NetworkCIDR == "" {
		// TODO: Choose non-overlapping networking CIDRs for VPCs?
		c.NetworkCIDR = "172.20.0.0/16"
	}

	for _, zone := range c.NodeZones {
		err := zone.performAssignments(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *ZoneConfig) performAssignments(c *CloudConfig) error {
	if z.CIDR == "" {
		cidr, err := z.assignCIDR(c)
		if err != nil {
			return err
		}
		glog.Infof("Assigned CIDR %s to zone %s", cidr, z.Name)
		z.CIDR = cidr
	}

	return nil
}

func (z *ZoneConfig) assignCIDR(c *CloudConfig) (string, error) {
	// TODO: We probably could query for the existing subnets & allocate appropriately
	// for now we'll require users to set CIDRs themselves

	lastCharMap := make(map[byte]bool)
	for _, nodeZone := range c.NodeZones {
		lastChar := nodeZone.Name[len(nodeZone.Name)-1]
		lastCharMap[lastChar] = true
	}

	index := -1

	if len(lastCharMap) == len(c.NodeZones) {
		// Last char of zones are unique (GCE, AWS)
		// At least on AWS, we also want 'a' to be 1, so that we don't collide with the lowest range,
		// because kube-up uses that range
		index = int(z.Name[len(z.Name)-1])
	} else {
		glog.Warningf("Last char of zone names not unique")

		for i, nodeZone := range c.NodeZones {
			if nodeZone.Name == z.Name {
				index = i
				break
			}
		}
		if index == -1 {
			return "", fmt.Errorf("zone not configured: %q", z.Name)
		}
	}

	_, cidr, err := net.ParseCIDR(c.NetworkCIDR)
	if err != nil {
		return "", fmt.Errorf("Invalid NetworkCIDR: %q", c.NetworkCIDR)
	}
	networkLength, _ := cidr.Mask.Size()

	// We assume a maximum of 8 subnets per network
	// TODO: Does this make sense on GCE?
	// TODO: Should we limit this to say 1000 IPs per subnet? (any reason to?)
	index = index % 8
	networkLength += 3

	ip4 := cidr.IP.To4()
	if ip4 != nil {
		n := binary.BigEndian.Uint32(ip4)
		n += uint32(index) << uint(32-networkLength)
		subnetIP := make(net.IP, len(ip4))
		binary.BigEndian.PutUint32(subnetIP, n)
		subnetCIDR := subnetIP.String() + "/" + strconv.Itoa(networkLength)
		glog.V(2).Infof("Computed CIDR for subnet in zone %q as %q", z.Name, subnetCIDR)
		return subnetCIDR, nil
	}

	return "", fmt.Errorf("Unexpected IP address type for NetworkCIDR: %s", c.NetworkCIDR)
}

// SharedVPC is a simple helper function which makes the templates for a shared VPC clearer
func (c *CloudConfig) SharedVPC() bool {
	return c.NetworkID != ""
}

// CloudPermissions holds IAM-style permissions
type CloudPermissions struct {
	S3Buckets []string `json:",omitempty"`
}

// AddS3Bucket adds a bucket if it does not already exist
func (p *CloudPermissions) AddS3Bucket(bucket string) {
	for _, b := range p.S3Buckets {
		if b == bucket {
			return
		}
	}

	p.S3Buckets = append(p.S3Buckets, bucket)
}
