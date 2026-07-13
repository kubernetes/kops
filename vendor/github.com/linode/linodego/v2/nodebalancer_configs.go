package linodego

import (
	"context"
)

// NodeBalancerConfig objects allow a NodeBalancer to accept traffic on a new port
type NodeBalancerConfig struct {
	ID            int                 `json:"id"`
	Port          int                 `json:"port"`
	Protocol      ConfigProtocol      `json:"protocol"`
	ProxyProtocol ConfigProxyProtocol `json:"proxy_protocol"`
	Algorithm     ConfigAlgorithm     `json:"algorithm"`
	Stickiness    ConfigStickiness    `json:"stickiness"`
	Check         ConfigCheck         `json:"check"`
	CheckInterval int                 `json:"check_interval"`
	CheckAttempts int                 `json:"check_attempts"`
	CheckPath     string              `json:"check_path"`
	CheckBody     string              `json:"check_body"`
	CheckPassive  bool                `json:"check_passive"`
	CheckTimeout  int                 `json:"check_timeout"`

	// NOTE: UDPCheckPort may not currently be available to all users.
	UDPCheckPort int `json:"udp_check_port"`

	// NOTE: UDPSessionTimeout may not currently be available to all users.
	UDPSessionTimeout int `json:"udp_session_timeout"`

	CipherSuite    ConfigCipher            `json:"cipher_suite"`
	NodeBalancerID int                     `json:"nodebalancer_id"`
	SSLCommonName  string                  `json:"ssl_commonname"`
	SSLFingerprint string                  `json:"ssl_fingerprint"`
	SSLCert        string                  `json:"ssl_cert"`
	SSLKey         string                  `json:"ssl_key"`
	NodesStatus    *NodeBalancerNodeStatus `json:"nodes_status"`
}

// ConfigAlgorithm constants start with Algorithm and include Linode API NodeBalancer Config Algorithms
type ConfigAlgorithm string

// ConfigAlgorithm constants reflect the NodeBalancer Config Algorithm
const (
	AlgorithmRoundRobin ConfigAlgorithm = "roundrobin"
	AlgorithmLeastConn  ConfigAlgorithm = "leastconn"
	AlgorithmSource     ConfigAlgorithm = "source"
	AlgorithmRingHash   ConfigAlgorithm = "ring_hash"
)

// ConfigStickiness constants start with Stickiness and include Linode API NodeBalancer Config Stickiness
type ConfigStickiness string

// ConfigStickiness constants reflect the node stickiness method for a NodeBalancer Config
const (
	StickinessNone       ConfigStickiness = "none"
	StickinessSession    ConfigStickiness = "session"
	StickinessTable      ConfigStickiness = "table"
	StickinessHTTPCookie ConfigStickiness = "http_cookie"
	StickinessSourceIP   ConfigStickiness = "source_ip"
)

// ConfigCheck constants start with Check and include Linode API NodeBalancer Config Check methods
type ConfigCheck string

// ConfigCheck constants reflect the node health status checking method for a NodeBalancer Config
const (
	CheckNone       ConfigCheck = "none"
	CheckConnection ConfigCheck = "connection"
	CheckHTTP       ConfigCheck = "http"
	CheckHTTPBody   ConfigCheck = "http_body"
)

// ConfigProtocol constants start with Protocol and include Linode API Nodebalancer Config protocols
type ConfigProtocol string

// ConfigProtocol constants reflect the protocol used by a NodeBalancer Config
const (
	ProtocolHTTP  ConfigProtocol = "http"
	ProtocolHTTPS ConfigProtocol = "https"
	ProtocolTCP   ConfigProtocol = "tcp"
	ProtocolUDP   ConfigProtocol = "udp"
)

// ConfigProxyProtocol constants start with ProxyProtocol and include Linode API NodeBalancer Config proxy protocol versions
type ConfigProxyProtocol string

// ConfigProxyProtocol constants reflect the proxy protocol version used by a NodeBalancer Config
const (
	ProxyProtocolNone ConfigProxyProtocol = "none"
	ProxyProtocolV1   ConfigProxyProtocol = "v1"
	ProxyProtocolV2   ConfigProxyProtocol = "v2"
)

// ConfigCipher constants start with Cipher and include Linode API NodeBalancer Config Cipher values
type ConfigCipher string

// ConfigCipher constants reflect the preferred cipher set for a NodeBalancer Config
const (
	CipherRecommended ConfigCipher = "recommended"
	CipherLegacy      ConfigCipher = "legacy"
)

// NodeBalancerNodeStatus represents the total number of nodes whose status is Up or Down
type NodeBalancerNodeStatus struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

// NodeBalancerConfigCreateOptions are permitted by CreateNodeBalancerConfig
type NodeBalancerConfigCreateOptions struct {
	Port          int                 `json:"port"`
	Protocol      ConfigProtocol      `json:"protocol,omitzero"`
	ProxyProtocol ConfigProxyProtocol `json:"proxy_protocol,omitzero"`
	Algorithm     ConfigAlgorithm     `json:"algorithm,omitzero"`
	Stickiness    ConfigStickiness    `json:"stickiness,omitzero"`
	Check         ConfigCheck         `json:"check,omitzero"`
	CheckInterval int                 `json:"check_interval,omitzero"`
	CheckAttempts int                 `json:"check_attempts,omitzero"`
	CheckPath     string              `json:"check_path,omitzero"`
	CheckBody     string              `json:"check_body,omitzero"`
	CheckPassive  *bool               `json:"check_passive,omitzero"`
	CheckTimeout  int                 `json:"check_timeout,omitzero"`

	// NOTE: UDPCheckPort may not currently be available to all users.
	UDPCheckPort *int `json:"udp_check_port,omitzero"`

	CipherSuite ConfigCipher                    `json:"cipher_suite,omitzero"`
	SSLCert     string                          `json:"ssl_cert,omitzero"`
	SSLKey      string                          `json:"ssl_key,omitzero"`
	Nodes       []NodeBalancerNodeCreateOptions `json:"nodes,omitzero"`
}

// NodeBalancerConfigRebuildOptions used by RebuildNodeBalancerConfig
type NodeBalancerConfigRebuildOptions struct {
	Port          int                 `json:"port"`
	Protocol      ConfigProtocol      `json:"protocol,omitzero"`
	ProxyProtocol ConfigProxyProtocol `json:"proxy_protocol,omitzero"`
	Algorithm     ConfigAlgorithm     `json:"algorithm,omitzero"`
	Stickiness    ConfigStickiness    `json:"stickiness,omitzero"`
	Check         ConfigCheck         `json:"check,omitzero"`
	CheckInterval int                 `json:"check_interval,omitzero"`
	CheckAttempts int                 `json:"check_attempts,omitzero"`
	CheckPath     string              `json:"check_path,omitzero"`
	CheckBody     string              `json:"check_body,omitzero"`
	CheckPassive  *bool               `json:"check_passive,omitzero"`
	CheckTimeout  int                 `json:"check_timeout,omitzero"`

	// NOTE: UDPCheckPort may not currently be available to all users.
	UDPCheckPort *int `json:"udp_check_port,omitzero"`

	CipherSuite ConfigCipher                           `json:"cipher_suite,omitzero"`
	SSLCert     string                                 `json:"ssl_cert,omitzero"`
	SSLKey      string                                 `json:"ssl_key,omitzero"`
	Nodes       []NodeBalancerConfigRebuildNodeOptions `json:"nodes"`
}

// NodeBalancerConfigRebuildNodeOptions represents a node defined when rebuilding a
// NodeBalancer config.
type NodeBalancerConfigRebuildNodeOptions struct {
	NodeBalancerNodeCreateOptions

	ID int `json:"id,omitzero"`
}

// NodeBalancerConfigUpdateOptions are permitted by UpdateNodeBalancerConfig
type NodeBalancerConfigUpdateOptions NodeBalancerConfigCreateOptions

// GetCreateOptions converts a NodeBalancerConfig to NodeBalancerConfigCreateOptions for use in CreateNodeBalancerConfig
func (i NodeBalancerConfig) GetCreateOptions() NodeBalancerConfigCreateOptions {
	var udpCheckPort *int
	if i.UDPCheckPort != 0 {
		udpCheckPort = &i.UDPCheckPort
	}

	return NodeBalancerConfigCreateOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		ProxyProtocol: i.ProxyProtocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckTimeout:  i.CheckTimeout,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  copyBool(&i.CheckPassive),
		UDPCheckPort:  udpCheckPort,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
	}
}

// GetUpdateOptions converts a NodeBalancerConfig to NodeBalancerConfigUpdateOptions for use in UpdateNodeBalancerConfig
func (i NodeBalancerConfig) GetUpdateOptions() NodeBalancerConfigUpdateOptions {
	var udpCheckPort *int
	if i.UDPCheckPort != 0 {
		udpCheckPort = &i.UDPCheckPort
	}

	return NodeBalancerConfigUpdateOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		ProxyProtocol: i.ProxyProtocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  copyBool(&i.CheckPassive),
		CheckTimeout:  i.CheckTimeout,
		UDPCheckPort:  udpCheckPort,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
	}
}

// GetRebuildOptions converts a NodeBalancerConfig to NodeBalancerConfigRebuildOptions for use in RebuildNodeBalancerConfig
func (i NodeBalancerConfig) GetRebuildOptions() NodeBalancerConfigRebuildOptions {
	var udpCheckPort *int
	if i.UDPCheckPort != 0 {
		udpCheckPort = &i.UDPCheckPort
	}

	return NodeBalancerConfigRebuildOptions{
		Port:          i.Port,
		Protocol:      i.Protocol,
		ProxyProtocol: i.ProxyProtocol,
		Algorithm:     i.Algorithm,
		Stickiness:    i.Stickiness,
		Check:         i.Check,
		CheckInterval: i.CheckInterval,
		CheckAttempts: i.CheckAttempts,
		CheckTimeout:  i.CheckTimeout,
		CheckPath:     i.CheckPath,
		CheckBody:     i.CheckBody,
		CheckPassive:  copyBool(&i.CheckPassive),
		UDPCheckPort:  udpCheckPort,
		CipherSuite:   i.CipherSuite,
		SSLCert:       i.SSLCert,
		SSLKey:        i.SSLKey,
		Nodes:         make([]NodeBalancerConfigRebuildNodeOptions, 0),
	}
}

// ListNodeBalancerConfigs lists NodeBalancerConfigs
func (c *Client) ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *ListOptions) ([]NodeBalancerConfig, error) {
	return getPaginatedResults[NodeBalancerConfig](ctx, c, formatAPIPath("nodebalancers/%d/configs", nodebalancerID), opts)
}

// GetNodeBalancerConfig gets the template with the provided ID
func (c *Client) GetNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int) (*NodeBalancerConfig, error) {
	e := formatAPIPath("nodebalancers/%d/configs/%d", nodebalancerID, configID)
	return doGETRequest[NodeBalancerConfig](ctx, c, e)
}

// CreateNodeBalancerConfig creates a NodeBalancerConfig
func (c *Client) CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, opts NodeBalancerConfigCreateOptions) (*NodeBalancerConfig, error) {
	e := formatAPIPath("nodebalancers/%d/configs", nodebalancerID)
	return doPOSTRequest[NodeBalancerConfig](ctx, c, e, opts)
}

// UpdateNodeBalancerConfig updates the NodeBalancerConfig with the specified id
func (c *Client) UpdateNodeBalancerConfig(
	ctx context.Context,
	nodebalancerID int,
	configID int,
	opts NodeBalancerConfigUpdateOptions,
) (*NodeBalancerConfig, error) {
	e := formatAPIPath("nodebalancers/%d/configs/%d", nodebalancerID, configID)
	return doPUTRequest[NodeBalancerConfig](ctx, c, e, opts)
}

// DeleteNodeBalancerConfig deletes the NodeBalancerConfig with the specified id
func (c *Client) DeleteNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int) error {
	e := formatAPIPath("nodebalancers/%d/configs/%d", nodebalancerID, configID)
	return doDELETERequest(ctx, c, e)
}

// RebuildNodeBalancerConfig updates the NodeBalancer with the specified id
func (c *Client) RebuildNodeBalancerConfig(
	ctx context.Context,
	nodeBalancerID int,
	configID int,
	opts NodeBalancerConfigRebuildOptions,
) (*NodeBalancerConfig, error) {
	e := formatAPIPath("nodebalancers/%d/configs/%d/rebuild", nodeBalancerID, configID)
	return doPOSTRequest[NodeBalancerConfig](ctx, c, e, opts)
}
