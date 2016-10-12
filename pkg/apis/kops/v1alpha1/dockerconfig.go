package v1alpha1

type DockerConfig struct {
	Bridge           *string `json:"bridge,omitempty" flag:"bridge"`
	LogLevel         *string `json:"logLevel,omitempty" flag:"log-level"`
	IPTables         *bool   `json:"ipTables,omitempty" flag:"iptables"`
	IPMasq           *bool   `json:"ipMasq,omitempty" flag:"ip-masq"`

	// Storage maps to the docker storage flag
	// But nodeup will also process a comma-separate list, selecting the first supported option
	Storage          *string `json:"storage,omitempty" flag:"storage-driver"`

	InsecureRegistry *string `json:"insecureRegistry,omitempty" flag:"insecure-registry"`
	MTU              *int    `json:"mtu,omitempty" flag:"mtu"`
}
