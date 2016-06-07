package nodeup

import "k8s.io/kube-deploy/upup/pkg/fi"

// TODO: Can we replace some of all of this with pkg/apis/componentconfig/types.go ?
type NodeConfig struct {
	Kubelet               KubeletConfig
	KubeProxy             KubeProxyConfig
	KubeControllerManager KubeControllerManagerConfig
	KubeScheduler         KubeSchedulerConfig
	Docker                DockerConfig
	APIServer             APIServerConfig
	CACertificate         *fi.Certificate

	DNS          DNSConfig
	KubeUser     string
	KubePassword string

	Tokens map[string]string

	Tags   []string
	Assets []string

	MasterInternalName string

	// The DNS zone to use if configuring a cloud provided DNS zone
	DNSZone string
}

// A helper so that templates can get tokens which are not valid identifiers
func (n *NodeConfig) GetToken(key string) string {
	return n.Tokens[key]
}

type DNSConfig struct {
	Replicas int
	Domain   string
	ServerIP string
}

type KubeletConfig struct {
	CloudProvider string `flag:"cloud-provider"`

	NonMasqueradeCdir string `flag:"non-masquerade-cidr"`
	APIServers        string `flag:"api-servers"`

	CgroupRoot     string `flag:"cgroup-root"`
	SystemCgroups  string `flag:"system-cgroups"`
	RuntimeCgroups string `flag:"runtime-cgroups"`
	KubeletCgroups string `flag:"kubelet-cgroups"`

	HairpinMode string `flag:"hairpin-mode"`

	EnableDebuggingHandlers *bool  `flag:"enable-debugging-handlers"`
	Config                  string `flag:"config"`
	AllowPrivileged         *bool  `flag:"allow-privileged"`
	Verbosity               *int   `flag:"v"`
	ClusterDNS              string `flag:"cluster-dns"`
	ClusterDomain           string `flag:"cluster-domain"`
	ConfigureCBR0           *bool  `flag:"configure-cbr0"`
	BabysitDaemons          *bool  `flag:"babysit-daemons"`

	RegisterSchedulable *bool  `flag:"register-schedulable"`
	ReconcileCIDR       *bool  `flag:"reconcile-cidr"`
	PodCIDR             string `flag:"pod-cidr"`

	Certificate *fi.Certificate `flag:"-"`
	Key         *fi.PrivateKey  `flag:"-"`
	// Allow override of CA Certificate
	CACertificate *fi.Certificate `flag:"-"`
}

type KubeProxyConfig struct {
	Image string
	// TODO: Better type ?
	CPURequest string // e.g. "20m"

	// TODO: Name verbosity or LogLevel
	LogLevel int `flag:"v"`

	// Configuration flags
	Master string `flag:"master"`
}

type DockerConfig struct {
	Bridge   string `flag:"bridge"`
	LogLevel string `flag:"log-level"`
	IPTables bool   `flag:"iptables"`
	IPMasq   bool   `flag:"ip-masq"`
	Storage  string `flag:"s"`
}

type APIServerConfig struct {
	CloudProvider string `flag:"cloud-provider"`

	SecurePort           int    `flag:"secure-port"`
	Address              string `flag:"address"`
	EtcdServers          string `flag:"etcd-servers"`
	EtcdServersOverrides string `flag:"etcd-servers-overrides"`
	// TODO: []string and join with commas?
	AdmissionControl      string `flag:"admission-control"`
	ServiceClusterIPRange string `flag:"service-cluster-ip-range"`
	ClientCAFile          string `flag:"client-ca-file"`
	BasicAuthFile         string `flag:"basic-auth-file"`
	TLSCertFile           string `flag:"tls-cert-file"`
	TLSPrivateKeyFile     string `flag:"tls-private-key-file"`
	TokenAuthFile         string `flag:"token-auth-file"`
	// TODO: Name verbosity or LogLevel
	LogLevel        int   `flag:"v"`
	AllowPrivileged *bool `flag:"allow-privileged"`

	PathSrvKubernetes string
	PathSrvSshproxy   string
	Image             string

	Certificate *fi.Certificate `flag:"-"`
	Key         *fi.PrivateKey  `flag:"-"`
}

type KubeControllerManagerConfig struct {
	CloudProvider string `flag:"cloud-provider"`

	Master               string `flag:"master"`
	ClusterName          string `flag:"cluster-name"`
	ClusterCIDR          string `flag:"cluster-cidr"`
	AllocateNodeCIDRs    *bool  `flag:"allocate-node-cidrs"`
	ConfigureCloudRoutes *bool  `flag:"configure-cloud-routes"`
	// TODO: Name verbosity or LogLevel
	LogLevel    int   `flag:"v"`
	LeaderElect *bool `flag:"leader-elect"`

	ServiceAccountPrivateKeyFile string `flag:"service-account-private-key-file"`
	RootCAFile                   string `flag:"root-ca-file"`

	Image string

	PathSrvKubernetes string
}

type KubeSchedulerConfig struct {
	LeaderElect *bool `flag:"leader-elect"`

	Image string
}
