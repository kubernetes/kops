package cloudup

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"math/big"
	"net"
	"strconv"
	//"k8s.io/kube-deploy/upup/pkg/fi"
	//"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/fi"
)

type ClusterConfig struct {
	// The CloudProvider to use (aws or gce)
	CloudProvider string `json:",omitempty"`

	// The version of kubernetes to install (optional, and can be a "spec" like stable)
	KubernetesVersion string `json:",omitempty"`

	//
	//// The Node initializer technique to use: cloudinit or nodeup
	//NodeInit                      string `json:",omitempty"`

	// Configuration of zones we are targeting
	Zones []*ZoneConfig `json:",omitempty"`
	//Region                        string        `json:",omitempty"`

	// Project is the cloud project we should use, required on GCE
	Project string `json:",omitempty"`

	// MasterPermissions contains the IAM permissions for the masters
	MasterPermissions *CloudPermissions `json:",omitempty"`
	// NodePermissions contains the IAM permissions for the nodes
	NodePermissions *CloudPermissions `json:",omitempty"`

	// MasterPublicName is the external DNS name for the master nodes
	MasterPublicName string `json:",omitempty"`
	// MasterInternalName is the internal DNS name for the master nodes
	MasterInternalName string `json:",omitempty"`

	// The CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s network
	NetworkCIDR string `json:",omitempty"`

	// NetworkID is an identifier of a network, if we want to reuse/share an existing network (e.g. an AWS VPC)
	NetworkID string `json:",omitempty"`

	// SecretStore is the VFS path to where secrets are stored
	SecretStore string `json:",omitempty"`
	// KeyStore is the VFS path to where SSL keys and certificates are stored
	KeyStore string `json:",omitempty"`
	// ConfigStore is the VFS path to where the configuration (CloudConfig, NodeSetConfig etc) is stored
	ConfigStore string `json:",omitempty"`

	// DNSZone is the DNS zone we should use when configuring DNS
	DNSZone string `json:",omitempty"`

	//InstancePrefix                string `json:",omitempty"`

	// ClusterName is a unique identifier for the cluster, and currently must be a DNS name
	ClusterName       string `json:",omitempty"`
	AllocateNodeCIDRs *bool  `json:",omitempty"`

	Multizone *bool `json:",omitempty"`

	//ClusterIPRange                string `json:",omitempty"`

	// ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
	ServiceClusterIPRange string `json:",omitempty"`
	//MasterIPRange                 string `json:",omitempty"`
	//NonMasqueradeCidr             string `json:",omitempty"`
	//
	//NetworkProvider               string `json:",omitempty"`
	//
	//HairpinMode                   string `json:",omitempty"`
	//
	//OpencontrailTag               string `json:",omitempty"`
	//OpencontrailKubernetesTag     string `json:",omitempty"`
	//OpencontrailPublicSubnet      string `json:",omitempty"`
	//
	//EnableClusterMonitoring       string `json:",omitempty"`
	//EnableL7LoadBalancing         string `json:",omitempty"`
	//EnableClusterUI               *bool  `json:",omitempty"`
	//
	//EnableClusterDNS              *bool  `json:",omitempty"`
	//DNSReplicas                   int    `json:",omitempty"`
	//DNSServerIP                   string `json:",omitempty"`

	// DNSDomain is the suffix we use for internal DNS names (normally cluster.local)
	DNSDomain string `json:",omitempty"`

	//EnableClusterLogging          *bool  `json:",omitempty"`
	//EnableNodeLogging             *bool  `json:",omitempty"`
	//LoggingDestination            string `json:",omitempty"`
	//ElasticsearchLoggingReplicas  int    `json:",omitempty"`
	//
	//EnableClusterRegistry         *bool  `json:",omitempty"`
	//ClusterRegistryDisk           string `json:",omitempty"`
	//ClusterRegistryDiskSize       int    `json:",omitempty"`
	//
	//EnableCustomMetrics           *bool `json:",omitempty"`
	//
	//RegisterMasterKubelet         *bool  `json:",omitempty"`

	//// Image is the default image spec to use for the cluster
	//Image                     string `json:",omitempty"`

	//KubeUser                      string `json:",omitempty"`
	//
	//// These are moved to CAStore / SecretStore
	////KubePassword			string
	////KubeletToken                  string
	////KubeProxyToken                string
	////BearerToken                   string
	////CACert                        []byte
	////CAKey                         []byte
	////KubeletCert                   []byte
	////KubeletKey                    []byte
	////MasterCert                    []byte
	////MasterKey                     []byte
	////KubecfgCert                   []byte
	////KubecfgKey                    []byte
	//
	//AdmissionControl              string `json:",omitempty"`
	//RuntimeConfig                 string `json:",omitempty"`
	//
	//KubeImageTag                  string `json:",omitempty"`
	//KubeDockerRegistry            string `json:",omitempty"`
	//KubeAddonRegistry             string `json:",omitempty"`
	//
	//KubeletPort                   int `json:",omitempty"`
	//
	//KubeApiserverRequestTimeout   int `json:",omitempty"`
	//
	//TerminatedPodGcThreshold      string `json:",omitempty"`
	//
	//EnableManifestURL             *bool  `json:",omitempty"`
	//ManifestURL                   string `json:",omitempty"`
	//ManifestURLHeader             string `json:",omitempty"`
	//
	//TestCluster                   string `json:",omitempty"`
	//
	//DockerOptions                 string `json:",omitempty"`
	//DockerStorage                 string `json:",omitempty"`
	//ExtraDockerOpts               string `json:",omitempty"`
	//
	//E2EStorageTestEnvironment     string `json:",omitempty"`
	//KubeletTestArgs               string `json:",omitempty"`
	//KubeletTestLogLevel           string `json:",omitempty"`
	//DockerTestArgs                string `json:",omitempty"`
	//DockerTestLogLevel            string `json:",omitempty"`
	//ApiserverTestArgs             string `json:",omitempty"`
	//ApiserverTestLogLevel         string `json:",omitempty"`
	//ControllerManagerTestArgs     string `json:",omitempty"`
	//ControllerManagerTestLogLevel string `json:",omitempty"`
	//SchedulerTestArgs             string `json:",omitempty"`
	//SchedulerTestLogLevel         string `json:",omitempty"`
	//KubeProxyTestArgs             string `json:",omitempty"`
	//KubeProxyTestLogLevel         string `json:",omitempty"`

	//NodeUp                        *NodeUpConfig `json:",omitempty"`

	// nodeSets is a list of all the NodeSets in the cluster.
	// It is not exported: we populate it from other files
	//nodeSets                      []*NodeSetConfig `json:",omitempty"`

	// Masters is the configuration for each master in the cluster
	Masters []*MasterConfig `json:",omitempty"`

	// MasterVolumes stores the configurations for each master data volume
	MasterVolumes []*VolumeConfig `json:",omitempty"`

	// Component configurations
	Docker                *DockerConfig                `json:",omitempty"`
	KubeDNS               *KubeDNSConfig               `json:",omitempty"`
	APIServer             *APIServerConfig             `json:",omitempty"`
	KubeControllerManager *KubeControllerManagerConfig `json:",omitempty"`
	KubeScheduler         *KubeSchedulerConfig         `json:",omitempty"`
	KubeProxy             *KubeProxyConfig             `json:",omitempty"`
	Kubelet               *KubeletConfig               `json:",omitempty"`
	MasterKubelet         *KubeletConfig               `json:",omitempty"`
}

type KubeDNSConfig struct {
	Replicas int    `json:",omitempty"`
	Domain   string `json:",omitempty"`
	ServerIP string `json:",omitempty"`
}

type MasterConfig struct {
	Name string `json:",omitempty"`

	Image       string `json:",omitempty"`
	Zone        string `json:",omitempty"`
	MachineType string `json:",omitempty"`
}

type VolumeConfig struct {
	Name string `json:",omitempty"`
	Type string `json:",omitempty"`
	Size int    `json:",omitempty"`

	Zone string `json:",omitempty"`

	Roles map[string]string `json:",omitempty"`
}

type NodeSetConfig struct {
	Name string `json:",omitempty"`

	Image   string `json:",omitempty"`
	MinSize *int   `json:",omitempty"`
	MaxSize *int   `json:",omitempty"`
	//NodeInstancePrefix string `json:",omitempty"`
	//NodeLabels         string `json:",omitempty"`
	MachineType string `json:",omitempty"`
	//NodeTag            string `json:",omitempty"`
}

type ZoneConfig struct {
	Name string `json:"name"`
	CIDR string `json:"cidr,omitempty"`
}

//type NodeUpConfig struct {
//	Source     string `json:",omitempty"`
//	SourceHash string `json:",omitempty"`
//
//	Tags       []string `json:",omitempty"`
//
//	// Assets that NodeUp should use.  This is a "search-path" for resolving dependencies.
//	Assets     []string `json:",omitempty"`
//}

func (c *ClusterConfig) WellKnownServiceIP(id int) (net.IP, error) {
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

// PerformAssignments populates values that are required and immutable
// For example, it assigns stable Keys to NodeSets & Masters, and
// it assigns CIDRs to subnets
func (c *ClusterConfig) PerformAssignments() error {
	if c.NetworkCIDR == "" {
		// TODO: Choose non-overlapping networking CIDRs for VPCs?
		c.NetworkCIDR = "172.20.0.0/16"
	}

	for _, zone := range c.Zones {
		err := zone.performAssignments(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// performAssignmentsNodesets populates NodeSets with default values
func PerformAssignmentsNodesets(nodeSets []*NodeSetConfig) error {
	keys := map[string]bool{}
	for _, n := range nodeSets {
		keys[n.Name] = true
	}

	for _, n := range nodeSets {
		// We want to give them a stable Key as soon as possible
		if n.Name == "" {
			// Loop to find the first unassigned name like `nodes-%d`
			i := 0
			for {
				key := fmt.Sprintf("nodes-%d", i)
				if !keys[key] {
					n.Name = key
					keys[key] = true
					break
				}
				i++
			}
		}
	}

	return nil
}

func (z *ZoneConfig) performAssignments(c *ClusterConfig) error {
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

func (z *ZoneConfig) assignCIDR(c *ClusterConfig) (string, error) {
	// TODO: We probably could query for the existing subnets & allocate appropriately
	// for now we'll require users to set CIDRs themselves

	lastCharMap := make(map[byte]bool)
	for _, nodeZone := range c.Zones {
		lastChar := nodeZone.Name[len(nodeZone.Name)-1]
		lastCharMap[lastChar] = true
	}

	index := -1

	if len(lastCharMap) == len(c.Zones) {
		// Last char of zones are unique (GCE, AWS)
		// At least on AWS, we also want 'a' to be 1, so that we don't collide with the lowest range,
		// because kube-up uses that range
		index = int(z.Name[len(z.Name)-1])
	} else {
		glog.Warningf("Last char of zone names not unique")

		for i, nodeZone := range c.Zones {
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
func (c *ClusterConfig) SharedVPC() bool {
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

//
//// findImage finds the default image
//func (c*NodeSetConfig) resolveImage() error {
//	cloud.(*awsup.AWSCloud).ResolveImage()
//
//	if n.Image == "" {
//		if defaultImage == "" {
//			image, err := c.determineImage()
//			if err != nil {
//				return err
//			}
//			defaultImage = image
//		}
//		n.Image = defaultImage
//	}
//
//
//	return nil
//}

func WriteConfig(stateStore fi.StateStore, cluster *ClusterConfig, nodeSets []*NodeSetConfig) error {
	// Check for nodeset Name duplicates before writing
	{
		names := map[string]bool{}
		for i, ns := range nodeSets {
			if ns.Name == "" {
				return fmt.Errorf("NodeSet #%d did not have Name set", i+1)
			}
			if names[ns.Name] {
				return fmt.Errorf("Duplicate NodeSet Name found: %q", ns.Name)
			}
			names[ns.Name] = true
		}
	}
	err := stateStore.WriteConfig("config", cluster)
	if err != nil {
		return fmt.Errorf("error writing updated cluster configuration: %v", err)
	}

	for _, ns := range nodeSets {
		err = stateStore.WriteConfig("nodeset/"+ns.Name, ns)
		if err != nil {
			return fmt.Errorf("error writing updated nodeset configuration: %v", err)
		}
	}

	return nil
}

func ReadConfig(stateStore fi.StateStore) (*ClusterConfig, []*NodeSetConfig, error) {
	cluster := &ClusterConfig{}
	err := stateStore.ReadConfig("config", cluster)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading cluster configuration: %v", err)
	}

	var nodeSets []*NodeSetConfig
	keys, err := stateStore.ListChildren("nodeset")
	if err != nil {
		return nil, nil, fmt.Errorf("error listing nodesets in state store: %v", err)
	}
	for _, key := range keys {
		ns := &NodeSetConfig{}
		err = stateStore.ReadConfig("nodeset/"+key, ns)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading nodeset configuration %q: %v", key, err)
		}
		nodeSets = append(nodeSets, ns)
	}

	return cluster, nodeSets, nil
}
