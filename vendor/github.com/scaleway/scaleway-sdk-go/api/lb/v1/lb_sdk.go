// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package lb provides methods and message types of the lb v1 API.
package lb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// ZonedAPI: this API allows you to manage your Scaleway Load Balancer services.
// Load Balancer API.
type ZonedAPI struct {
	client *scw.Client
}

// NewZonedAPI returns a ZonedAPI object from a Scaleway client.
func NewZonedAPI(client *scw.Client) *ZonedAPI {
	return &ZonedAPI{
		client: client,
	}
}

// API: this API allows you to manage your load balancer service.
// Load balancer API.
type API struct {
	client *scw.Client
}

// Deprecated NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ACLActionRedirectRedirectType string

const (
	ACLActionRedirectRedirectTypeLocation = ACLActionRedirectRedirectType("location")
	ACLActionRedirectRedirectTypeScheme   = ACLActionRedirectRedirectType("scheme")
)

func (enum ACLActionRedirectRedirectType) String() string {
	if enum == "" {
		// return default value if empty
		return "location"
	}
	return string(enum)
}

func (enum ACLActionRedirectRedirectType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ACLActionRedirectRedirectType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ACLActionRedirectRedirectType(ACLActionRedirectRedirectType(tmp).String())
	return nil
}

type ACLActionType string

const (
	ACLActionTypeAllow    = ACLActionType("allow")
	ACLActionTypeDeny     = ACLActionType("deny")
	ACLActionTypeRedirect = ACLActionType("redirect")
)

func (enum ACLActionType) String() string {
	if enum == "" {
		// return default value if empty
		return "allow"
	}
	return string(enum)
}

func (enum ACLActionType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ACLActionType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ACLActionType(ACLActionType(tmp).String())
	return nil
}

type ACLHTTPFilter string

const (
	ACLHTTPFilterACLHTTPFilterNone = ACLHTTPFilter("acl_http_filter_none")
	ACLHTTPFilterPathBegin         = ACLHTTPFilter("path_begin")
	ACLHTTPFilterPathEnd           = ACLHTTPFilter("path_end")
	ACLHTTPFilterRegex             = ACLHTTPFilter("regex")
	ACLHTTPFilterHTTPHeaderMatch   = ACLHTTPFilter("http_header_match")
)

func (enum ACLHTTPFilter) String() string {
	if enum == "" {
		// return default value if empty
		return "acl_http_filter_none"
	}
	return string(enum)
}

func (enum ACLHTTPFilter) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ACLHTTPFilter) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ACLHTTPFilter(ACLHTTPFilter(tmp).String())
	return nil
}

type BackendServerStatsHealthCheckStatus string

const (
	BackendServerStatsHealthCheckStatusUnknown  = BackendServerStatsHealthCheckStatus("unknown")
	BackendServerStatsHealthCheckStatusNeutral  = BackendServerStatsHealthCheckStatus("neutral")
	BackendServerStatsHealthCheckStatusFailed   = BackendServerStatsHealthCheckStatus("failed")
	BackendServerStatsHealthCheckStatusPassed   = BackendServerStatsHealthCheckStatus("passed")
	BackendServerStatsHealthCheckStatusCondpass = BackendServerStatsHealthCheckStatus("condpass")
)

func (enum BackendServerStatsHealthCheckStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum BackendServerStatsHealthCheckStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BackendServerStatsHealthCheckStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BackendServerStatsHealthCheckStatus(BackendServerStatsHealthCheckStatus(tmp).String())
	return nil
}

type BackendServerStatsServerState string

const (
	BackendServerStatsServerStateStopped  = BackendServerStatsServerState("stopped")
	BackendServerStatsServerStateStarting = BackendServerStatsServerState("starting")
	BackendServerStatsServerStateRunning  = BackendServerStatsServerState("running")
	BackendServerStatsServerStateStopping = BackendServerStatsServerState("stopping")
)

func (enum BackendServerStatsServerState) String() string {
	if enum == "" {
		// return default value if empty
		return "stopped"
	}
	return string(enum)
}

func (enum BackendServerStatsServerState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BackendServerStatsServerState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BackendServerStatsServerState(BackendServerStatsServerState(tmp).String())
	return nil
}

type CertificateStatus string

const (
	CertificateStatusPending = CertificateStatus("pending")
	CertificateStatusReady   = CertificateStatus("ready")
	CertificateStatusError   = CertificateStatus("error")
)

func (enum CertificateStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "pending"
	}
	return string(enum)
}

func (enum CertificateStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CertificateStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CertificateStatus(CertificateStatus(tmp).String())
	return nil
}

type CertificateType string

const (
	CertificateTypeLetsencryt = CertificateType("letsencryt")
	CertificateTypeCustom     = CertificateType("custom")
)

func (enum CertificateType) String() string {
	if enum == "" {
		// return default value if empty
		return "letsencryt"
	}
	return string(enum)
}

func (enum CertificateType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CertificateType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CertificateType(CertificateType(tmp).String())
	return nil
}

type ForwardPortAlgorithm string

const (
	ForwardPortAlgorithmRoundrobin = ForwardPortAlgorithm("roundrobin")
	ForwardPortAlgorithmLeastconn  = ForwardPortAlgorithm("leastconn")
	ForwardPortAlgorithmFirst      = ForwardPortAlgorithm("first")
)

func (enum ForwardPortAlgorithm) String() string {
	if enum == "" {
		// return default value if empty
		return "roundrobin"
	}
	return string(enum)
}

func (enum ForwardPortAlgorithm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ForwardPortAlgorithm) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ForwardPortAlgorithm(ForwardPortAlgorithm(tmp).String())
	return nil
}

type InstanceStatus string

const (
	InstanceStatusUnknown   = InstanceStatus("unknown")
	InstanceStatusReady     = InstanceStatus("ready")
	InstanceStatusPending   = InstanceStatus("pending")
	InstanceStatusStopped   = InstanceStatus("stopped")
	InstanceStatusError     = InstanceStatus("error")
	InstanceStatusLocked    = InstanceStatus("locked")
	InstanceStatusMigrating = InstanceStatus("migrating")
)

func (enum InstanceStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum InstanceStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *InstanceStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = InstanceStatus(InstanceStatus(tmp).String())
	return nil
}

type LBStatus string

const (
	LBStatusUnknown   = LBStatus("unknown")
	LBStatusReady     = LBStatus("ready")
	LBStatusPending   = LBStatus("pending")
	LBStatusStopped   = LBStatus("stopped")
	LBStatusError     = LBStatus("error")
	LBStatusLocked    = LBStatus("locked")
	LBStatusMigrating = LBStatus("migrating")
	LBStatusToCreate  = LBStatus("to_create")
	LBStatusCreating  = LBStatus("creating")
	LBStatusToDelete  = LBStatus("to_delete")
	LBStatusDeleting  = LBStatus("deleting")
)

func (enum LBStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum LBStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LBStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LBStatus(LBStatus(tmp).String())
	return nil
}

type LBTypeStock string

const (
	LBTypeStockUnknown    = LBTypeStock("unknown")
	LBTypeStockLowStock   = LBTypeStock("low_stock")
	LBTypeStockOutOfStock = LBTypeStock("out_of_stock")
	LBTypeStockAvailable  = LBTypeStock("available")
)

func (enum LBTypeStock) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum LBTypeStock) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LBTypeStock) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LBTypeStock(LBTypeStock(tmp).String())
	return nil
}

type ListACLRequestOrderBy string

const (
	ListACLRequestOrderByCreatedAtAsc  = ListACLRequestOrderBy("created_at_asc")
	ListACLRequestOrderByCreatedAtDesc = ListACLRequestOrderBy("created_at_desc")
	ListACLRequestOrderByNameAsc       = ListACLRequestOrderBy("name_asc")
	ListACLRequestOrderByNameDesc      = ListACLRequestOrderBy("name_desc")
)

func (enum ListACLRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListACLRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListACLRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListACLRequestOrderBy(ListACLRequestOrderBy(tmp).String())
	return nil
}

type ListBackendsRequestOrderBy string

const (
	ListBackendsRequestOrderByCreatedAtAsc  = ListBackendsRequestOrderBy("created_at_asc")
	ListBackendsRequestOrderByCreatedAtDesc = ListBackendsRequestOrderBy("created_at_desc")
	ListBackendsRequestOrderByNameAsc       = ListBackendsRequestOrderBy("name_asc")
	ListBackendsRequestOrderByNameDesc      = ListBackendsRequestOrderBy("name_desc")
)

func (enum ListBackendsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListBackendsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListBackendsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListBackendsRequestOrderBy(ListBackendsRequestOrderBy(tmp).String())
	return nil
}

type ListCertificatesRequestOrderBy string

const (
	ListCertificatesRequestOrderByCreatedAtAsc  = ListCertificatesRequestOrderBy("created_at_asc")
	ListCertificatesRequestOrderByCreatedAtDesc = ListCertificatesRequestOrderBy("created_at_desc")
	ListCertificatesRequestOrderByNameAsc       = ListCertificatesRequestOrderBy("name_asc")
	ListCertificatesRequestOrderByNameDesc      = ListCertificatesRequestOrderBy("name_desc")
)

func (enum ListCertificatesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListCertificatesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListCertificatesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListCertificatesRequestOrderBy(ListCertificatesRequestOrderBy(tmp).String())
	return nil
}

type ListFrontendsRequestOrderBy string

const (
	ListFrontendsRequestOrderByCreatedAtAsc  = ListFrontendsRequestOrderBy("created_at_asc")
	ListFrontendsRequestOrderByCreatedAtDesc = ListFrontendsRequestOrderBy("created_at_desc")
	ListFrontendsRequestOrderByNameAsc       = ListFrontendsRequestOrderBy("name_asc")
	ListFrontendsRequestOrderByNameDesc      = ListFrontendsRequestOrderBy("name_desc")
)

func (enum ListFrontendsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListFrontendsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListFrontendsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListFrontendsRequestOrderBy(ListFrontendsRequestOrderBy(tmp).String())
	return nil
}

type ListLBsRequestOrderBy string

const (
	ListLBsRequestOrderByCreatedAtAsc  = ListLBsRequestOrderBy("created_at_asc")
	ListLBsRequestOrderByCreatedAtDesc = ListLBsRequestOrderBy("created_at_desc")
	ListLBsRequestOrderByNameAsc       = ListLBsRequestOrderBy("name_asc")
	ListLBsRequestOrderByNameDesc      = ListLBsRequestOrderBy("name_desc")
)

func (enum ListLBsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListLBsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListLBsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListLBsRequestOrderBy(ListLBsRequestOrderBy(tmp).String())
	return nil
}

type ListPrivateNetworksRequestOrderBy string

const (
	ListPrivateNetworksRequestOrderByCreatedAtAsc  = ListPrivateNetworksRequestOrderBy("created_at_asc")
	ListPrivateNetworksRequestOrderByCreatedAtDesc = ListPrivateNetworksRequestOrderBy("created_at_desc")
)

func (enum ListPrivateNetworksRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPrivateNetworksRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPrivateNetworksRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPrivateNetworksRequestOrderBy(ListPrivateNetworksRequestOrderBy(tmp).String())
	return nil
}

type ListRoutesRequestOrderBy string

const (
	ListRoutesRequestOrderByCreatedAtAsc  = ListRoutesRequestOrderBy("created_at_asc")
	ListRoutesRequestOrderByCreatedAtDesc = ListRoutesRequestOrderBy("created_at_desc")
)

func (enum ListRoutesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListRoutesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListRoutesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListRoutesRequestOrderBy(ListRoutesRequestOrderBy(tmp).String())
	return nil
}

type ListSubscriberRequestOrderBy string

const (
	ListSubscriberRequestOrderByCreatedAtAsc  = ListSubscriberRequestOrderBy("created_at_asc")
	ListSubscriberRequestOrderByCreatedAtDesc = ListSubscriberRequestOrderBy("created_at_desc")
	ListSubscriberRequestOrderByNameAsc       = ListSubscriberRequestOrderBy("name_asc")
	ListSubscriberRequestOrderByNameDesc      = ListSubscriberRequestOrderBy("name_desc")
)

func (enum ListSubscriberRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListSubscriberRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListSubscriberRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListSubscriberRequestOrderBy(ListSubscriberRequestOrderBy(tmp).String())
	return nil
}

type OnMarkedDownAction string

const (
	OnMarkedDownActionOnMarkedDownActionNone = OnMarkedDownAction("on_marked_down_action_none")
	OnMarkedDownActionShutdownSessions       = OnMarkedDownAction("shutdown_sessions")
)

func (enum OnMarkedDownAction) String() string {
	if enum == "" {
		// return default value if empty
		return "on_marked_down_action_none"
	}
	return string(enum)
}

func (enum OnMarkedDownAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *OnMarkedDownAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = OnMarkedDownAction(OnMarkedDownAction(tmp).String())
	return nil
}

type PrivateNetworkStatus string

const (
	PrivateNetworkStatusUnknown = PrivateNetworkStatus("unknown")
	PrivateNetworkStatusReady   = PrivateNetworkStatus("ready")
	PrivateNetworkStatusPending = PrivateNetworkStatus("pending")
	PrivateNetworkStatusError   = PrivateNetworkStatus("error")
)

func (enum PrivateNetworkStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum PrivateNetworkStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PrivateNetworkStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PrivateNetworkStatus(PrivateNetworkStatus(tmp).String())
	return nil
}

type Protocol string

const (
	ProtocolTCP  = Protocol("tcp")
	ProtocolHTTP = Protocol("http")
)

func (enum Protocol) String() string {
	if enum == "" {
		// return default value if empty
		return "tcp"
	}
	return string(enum)
}

func (enum Protocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *Protocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = Protocol(Protocol(tmp).String())
	return nil
}

// ProxyProtocol: pROXY protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. PROXY protocol must be supported by the backend servers' software. For more information on the different protocols available, see the [dedicated documentation](https://www.scaleway.com/en/docs/network/load-balancer/reference-content/configuring-load-balancer/#choosing-a-proxy-protocol).
type ProxyProtocol string

const (
	ProxyProtocolProxyProtocolUnknown = ProxyProtocol("proxy_protocol_unknown")
	ProxyProtocolProxyProtocolNone    = ProxyProtocol("proxy_protocol_none")
	ProxyProtocolProxyProtocolV1      = ProxyProtocol("proxy_protocol_v1")
	ProxyProtocolProxyProtocolV2      = ProxyProtocol("proxy_protocol_v2")
	ProxyProtocolProxyProtocolV2Ssl   = ProxyProtocol("proxy_protocol_v2_ssl")
	ProxyProtocolProxyProtocolV2SslCn = ProxyProtocol("proxy_protocol_v2_ssl_cn")
)

func (enum ProxyProtocol) String() string {
	if enum == "" {
		// return default value if empty
		return "proxy_protocol_unknown"
	}
	return string(enum)
}

func (enum ProxyProtocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ProxyProtocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ProxyProtocol(ProxyProtocol(tmp).String())
	return nil
}

type SSLCompatibilityLevel string

const (
	SSLCompatibilityLevelSslCompatibilityLevelUnknown      = SSLCompatibilityLevel("ssl_compatibility_level_unknown")
	SSLCompatibilityLevelSslCompatibilityLevelIntermediate = SSLCompatibilityLevel("ssl_compatibility_level_intermediate")
	SSLCompatibilityLevelSslCompatibilityLevelModern       = SSLCompatibilityLevel("ssl_compatibility_level_modern")
	SSLCompatibilityLevelSslCompatibilityLevelOld          = SSLCompatibilityLevel("ssl_compatibility_level_old")
)

func (enum SSLCompatibilityLevel) String() string {
	if enum == "" {
		// return default value if empty
		return "ssl_compatibility_level_unknown"
	}
	return string(enum)
}

func (enum SSLCompatibilityLevel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SSLCompatibilityLevel) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SSLCompatibilityLevel(SSLCompatibilityLevel(tmp).String())
	return nil
}

type StickySessionsType string

const (
	StickySessionsTypeNone   = StickySessionsType("none")
	StickySessionsTypeCookie = StickySessionsType("cookie")
	StickySessionsTypeTable  = StickySessionsType("table")
)

func (enum StickySessionsType) String() string {
	if enum == "" {
		// return default value if empty
		return "none"
	}
	return string(enum)
}

func (enum StickySessionsType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *StickySessionsType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = StickySessionsType(StickySessionsType(tmp).String())
	return nil
}

// ACL: acl.
type ACL struct {
	// ID: ACL ID.
	ID string `json:"id"`
	// Name: ACL name.
	Name string `json:"name"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Frontend: ACL is attached to this frontend object.
	Frontend *Frontend `json:"frontend"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// CreatedAt: date on which the ACL was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the ACL was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// Description: ACL description.
	Description string `json:"description"`
}

// ACLAction: acl action.
type ACLAction struct {
	// Type: action to take when incoming traffic matches an ACL filter.
	// Default value: allow
	Type ACLActionType `json:"type"`
	// Redirect: redirection parameters when using an ACL with a `redirect` action.
	Redirect *ACLActionRedirect `json:"redirect"`
}

// ACLActionRedirect: acl action redirect.
type ACLActionRedirect struct {
	// Type: redirect type.
	// Default value: location
	Type ACLActionRedirectRedirectType `json:"type"`
	// Target: redirect target. For a location redirect, you can use a URL e.g. `https://scaleway.com`. Using a scheme name (e.g. `https`, `http`, `ftp`, `git`) will replace the request's original scheme. This can be useful to implement HTTP to HTTPS redirects. Valid placeholders that can be used in a `location` redirect to preserve parts of the original request in the redirection URL are {{ host }}, {{ query }}, {{ path }} and {{ scheme }}.
	Target string `json:"target"`
	// Code: HTTP redirect code to use. Valid values are 301, 302, 303, 307 and 308. Default value is 302.
	Code *int32 `json:"code"`
}

// ACLMatch: acl match.
type ACLMatch struct {
	// IPSubnet: list of IPs or CIDR v4/v6 addresses to filter for from the client side.
	IPSubnet []*string `json:"ip_subnet"`
	// HTTPFilter: type of HTTP filter to match. Extracts the request's URL path, which starts at the first slash and ends before the question mark (without the host part). Defines where to filter for the http_filter_value. Only supported for HTTP backends.
	// Default value: acl_http_filter_none
	HTTPFilter ACLHTTPFilter `json:"http_filter"`
	// HTTPFilterValue: list of values to filter for.
	HTTPFilterValue []*string `json:"http_filter_value"`
	// HTTPFilterOption: name of the HTTP header to filter on if `http_header_match` was selected in `http_filter`.
	HTTPFilterOption *string `json:"http_filter_option"`
	// Invert: defines whether to invert the match condition. If set to `true`, the ACL carries out its action when the condition DOES NOT match.
	Invert bool `json:"invert"`
}

// ACLSpec: acl spec.
type ACLSpec struct {
	// Name: ACL name.
	Name string `json:"name"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` and `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// Description: ACL description.
	Description string `json:"description"`
}

// Backend: backend.
type Backend struct {
	// ID: backend ID.
	ID string `json:"id"`
	// Name: name of the backend.
	Name string `json:"name"`
	// ForwardProtocol: protocol used by the backend when forwarding traffic to backend servers.
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: port used by the backend when forwarding traffic to backend servers.
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm to use when determining which backend server to forward new traffic to.
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: defines whether sticky sessions (binding a particular session to a particular backend server) are activated and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie to stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for cookie-based sticky sessions.
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: object defining the health check to be carried out by the backend when checking the status and health of backend servers.
	HealthCheck *HealthCheck `json:"health_check"`
	// Pool: list of IP addresses of backend servers attached to this backend.
	Pool []string `json:"pool"`
	// LB: load Balancer the backend is attached to.
	LB *LB `json:"lb"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field.
	SendProxyV2 *bool `json:"send_proxy_v2,omitempty"`
	// TimeoutServer: maximum allowed time for a backend server to process a request.
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// CreatedAt: date at which the backend was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date at which the backend was updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// FailoverHost: scaleway S3 bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`
	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`
	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`
	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`
	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`
}

func (m *Backend) UnmarshalJSON(b []byte) error {
	type tmpType Backend
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = Backend(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m Backend) MarshalJSON() ([]byte, error) {
	type tmpType Backend
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// BackendServerStats: backend server stats.
type BackendServerStats struct {
	// InstanceID: ID of your Load Balancer's underlying Instance.
	InstanceID string `json:"instance_id"`
	// BackendID: backend ID.
	BackendID string `json:"backend_id"`
	// IP: iPv4 or IPv6 address of the backend server.
	IP string `json:"ip"`
	// ServerState: server operational state (stopped/starting/running/stopping).
	// Default value: stopped
	ServerState BackendServerStatsServerState `json:"server_state"`
	// ServerStateChangedAt: time since last operational change.
	ServerStateChangedAt *time.Time `json:"server_state_changed_at"`
	// LastHealthCheckStatus: last health check status (unknown/neutral/failed/passed/condpass).
	// Default value: unknown
	LastHealthCheckStatus BackendServerStatsHealthCheckStatus `json:"last_health_check_status"`
}

// Certificate: certificate.
type Certificate struct {
	// Type: certificate type (Let's Encrypt or custom).
	// Default value: letsencryt
	Type CertificateType `json:"type"`
	// ID: certificate ID.
	ID string `json:"id"`
	// CommonName: main domain name of certificate.
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names.
	SubjectAlternativeName []string `json:"subject_alternative_name"`
	// Fingerprint: identifier (SHA-1) of the certificate.
	Fingerprint string `json:"fingerprint"`
	// NotValidBefore: lower validity bound.
	NotValidBefore *time.Time `json:"not_valid_before"`
	// NotValidAfter: upper validity bound.
	NotValidAfter *time.Time `json:"not_valid_after"`
	// Status: certificate status.
	// Default value: pending
	Status CertificateStatus `json:"status"`
	// LB: load Balancer object the certificate is attached to.
	LB *LB `json:"lb"`
	// Name: certificate name.
	Name string `json:"name"`
	// CreatedAt: date on which the certificate was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the certificate was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// StatusDetails: additional information about the certificate status (useful in case of certificate generation failure, for example).
	StatusDetails *string `json:"status_details"`
}

// CreateCertificateRequestCustomCertificate: create certificate request. custom certificate.
type CreateCertificateRequestCustomCertificate struct {
	// CertificateChain: full PEM-formatted certificate, consisting of the entire certificate chain including public key, private key, and (optionally) Certificate Authorities.
	CertificateChain string `json:"certificate_chain"`
}

// CreateCertificateRequestLetsencryptConfig: create certificate request. letsencrypt config.
type CreateCertificateRequestLetsencryptConfig struct {
	// CommonName: main domain name of certificate (this domain must exist and resolve to your Load Balancer IP address).
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names (all domain names must exist and resolve to your Load Balancer IP address).
	SubjectAlternativeName []string `json:"subject_alternative_name"`
}

// Frontend: frontend.
type Frontend struct {
	// ID: frontend ID.
	ID string `json:"id"`
	// Name: name of the frontend.
	Name string `json:"name"`
	// InboundPort: port the frontend listens on.
	InboundPort int32 `json:"inbound_port"`
	// Backend: backend object the frontend is attached to.
	Backend *Backend `json:"backend"`
	// LB: load Balancer object the frontend is attached to.
	LB *LB `json:"lb"`
	// TimeoutClient: maximum allowed inactivity time on the client side.
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: Certificate: certificate, deprecated in favor of certificate_ids array.
	Certificate *Certificate `json:"certificate,omitempty"`
	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs []string `json:"certificate_ids"`
	// CreatedAt: date on which the frontend was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the frontend was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`
}

func (m *Frontend) UnmarshalJSON(b []byte) error {
	type tmpType Frontend
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = Frontend(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m Frontend) MarshalJSON() ([]byte, error) {
	type tmpType Frontend
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// HealthCheck: health check.
type HealthCheck struct {
	// Port: port to use for the backend server health check.
	Port int32 `json:"port"`
	// CheckDelay: time to wait between two consecutive health checks.
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: maximum time a backend server has to reply to the health check.
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks after which the server will be considered dead.
	CheckMaxRetries int32 `json:"check_max_retries"`
	// TCPConfig: object to configure a basic TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`
	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22, for older versions, use a TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`
	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`
	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`
	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
	// CheckSendProxy: defines whether proxy protocol should be activated for the health check.
	CheckSendProxy bool `json:"check_send_proxy"`
	// TransientCheckDelay: time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN).
	TransientCheckDelay *scw.Duration `json:"transient_check_delay"`
}

func (m *HealthCheck) UnmarshalJSON(b []byte) error {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = HealthCheck(tmp.tmpType)

	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	return nil
}

func (m HealthCheck) MarshalJSON() ([]byte, error) {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{
		tmpType: tmpType(m),

		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

// HealthCheckHTTPConfig: health check. http config.
type HealthCheckHTTPConfig struct {
	// URI: HTTP URI used for the health check.
	// The HTTP URI to use when performing a health check on backend servers.
	URI string `json:"uri"`
	// Method: HTTP method used for the health check.
	// The HTTP method used when performing a health check on backend servers.
	Method string `json:"method"`
	// Code: HTTP response code expected for a successful health check.
	// The HTTP response code that should be returned for a health check to be considered successful.
	Code *int32 `json:"code"`
	// HostHeader: HTTP host header used for the health check.
	// The HTTP host header used when performing a health check on backend servers.
	HostHeader string `json:"host_header"`
}

// HealthCheckHTTPSConfig: health check. https config.
type HealthCheckHTTPSConfig struct {
	// URI: HTTP URI used for the health check.
	// The HTTP URI to use when performing a health check on backend servers.
	URI string `json:"uri"`
	// Method: HTTP method used for the health check.
	// The HTTP method used when performing a health check on backend servers.
	Method string `json:"method"`
	// Code: HTTP response code expected for a successful health check.
	// The HTTP response code that should be returned for a health check to be considered successful.
	Code *int32 `json:"code"`
	// HostHeader: HTTP host header used for the health check.
	// The HTTP host header used when performing a health check on backend servers.
	HostHeader string `json:"host_header"`
	// Sni: sNI used for SSL health checks.
	// The SNI value used when performing a health check on backend servers over SSL.
	Sni string `json:"sni"`
}

type HealthCheckLdapConfig struct {
}

// HealthCheckMysqlConfig: health check. mysql config.
type HealthCheckMysqlConfig struct {
	// User: mySQL user to use for the health check.
	User string `json:"user"`
}

// HealthCheckPgsqlConfig: health check. pgsql config.
type HealthCheckPgsqlConfig struct {
	// User: postgreSQL user to use for the health check.
	User string `json:"user"`
}

type HealthCheckRedisConfig struct {
}

type HealthCheckTCPConfig struct {
}

// IP: ip.
type IP struct {
	// ID: IP address ID.
	ID string `json:"id"`
	// IPAddress: IP address.
	IPAddress string `json:"ip_address"`
	// OrganizationID: organization ID of the Scaleway Organization the IP address is in.
	OrganizationID string `json:"organization_id"`
	// ProjectID: project ID of the Scaleway Project the IP address is in.
	ProjectID string `json:"project_id"`
	// LBID: load Balancer ID.
	LBID *string `json:"lb_id"`
	// Reverse: reverse DNS (domain name) of the IP address.
	Reverse string `json:"reverse"`
	// Deprecated: Region: the region the IP address is in.
	Region *scw.Region `json:"region,omitempty"`
	// Zone: the zone the IP address is in.
	Zone scw.Zone `json:"zone"`
}

// Instance: instance.
type Instance struct {
	// ID: underlying Instance ID.
	ID string `json:"id"`
	// Status: instance status.
	// Default value: unknown
	Status InstanceStatus `json:"status"`
	// IPAddress: instance IP address.
	IPAddress string `json:"ip_address"`
	// CreatedAt: date on which the Instance was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the Instance was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// Deprecated: Region: the region the Instance is in.
	Region *scw.Region `json:"region,omitempty"`
	// Zone: the zone the Instance is in.
	Zone scw.Zone `json:"zone"`
}

// LB: lb.
type LB struct {
	// ID: underlying Instance ID.
	ID string `json:"id"`
	// Name: load Balancer name.
	Name string `json:"name"`
	// Description: load Balancer description.
	Description string `json:"description"`
	// Status: load Balancer status.
	// Default value: unknown
	Status LBStatus `json:"status"`
	// Instances: list of underlying Instances.
	Instances []*Instance `json:"instances"`
	// OrganizationID: scaleway Organization ID.
	OrganizationID string `json:"organization_id"`
	// ProjectID: scaleway Project ID.
	ProjectID string `json:"project_id"`
	// IP: list of IP addresses attached to the Load Balancer.
	IP []*IP `json:"ip"`
	// Tags: load Balancer tags.
	Tags []string `json:"tags"`
	// FrontendCount: number of frontends the Load Balancer has.
	FrontendCount int32 `json:"frontend_count"`
	// BackendCount: number of backends the Load Balancer has.
	BackendCount int32 `json:"backend_count"`
	// Type: load Balancer offer type.
	Type string `json:"type"`
	// Subscriber: subscriber information.
	Subscriber *Subscriber `json:"subscriber"`
	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on client side.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
	// CreatedAt: date on which the Load Balancer was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the Load Balancer was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// PrivateNetworkCount: number of Private Networks attached to the Load Balancer.
	PrivateNetworkCount int32 `json:"private_network_count"`
	// RouteCount: number of routes configured on the Load Balancer.
	RouteCount int32 `json:"route_count"`
	// Deprecated: Region: the region the Load Balancer is in.
	Region *scw.Region `json:"region,omitempty"`
	// Zone: the zone the Load Balancer is in.
	Zone scw.Zone `json:"zone"`
}

// LBStats: lb stats.
type LBStats struct {
	// BackendServersStats: list of objects containing Load Balancer statistics.
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
}

// LBType: lb type.
type LBType struct {
	// Name: load Balancer commercial offer type name.
	Name string `json:"name"`
	// StockStatus: current stock status for a given Load Balancer type.
	// Default value: unknown
	StockStatus LBTypeStock `json:"stock_status"`
	// Description: load Balancer commercial offer type description.
	Description string `json:"description"`
	// Deprecated: Region: the region the Load Balancer stock is in.
	Region *scw.Region `json:"region,omitempty"`
	// Zone: the zone the Load Balancer stock is in.
	Zone scw.Zone `json:"zone"`
}

// ListACLResponse: list acl response.
type ListACLResponse struct {
	// ACLs: list of ACL objects.
	ACLs []*ACL `json:"acls"`
	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// ListBackendStatsResponse: list backend stats response.
type ListBackendStatsResponse struct {
	// BackendServersStats: list of objects containing backend server statistics.
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// ListBackendsResponse: list backends response.
type ListBackendsResponse struct {
	// Backends: list of backend objects of a given Load Balancer.
	Backends []*Backend `json:"backends"`
	// TotalCount: total count of backend objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// ListCertificatesResponse: list certificates response.
type ListCertificatesResponse struct {
	// Certificates: list of certificate objects.
	Certificates []*Certificate `json:"certificates"`
	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// ListFrontendsResponse: list frontends response.
type ListFrontendsResponse struct {
	// Frontends: list of frontend objects of a given Load Balancer.
	Frontends []*Frontend `json:"frontends"`
	// TotalCount: total count of frontend objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// ListIPsResponse: list ips response.
type ListIPsResponse struct {
	// IPs: list of IP address objects.
	IPs []*IP `json:"ips"`
	// TotalCount: total count of IP address objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// ListLBPrivateNetworksResponse: list lb private networks response.
type ListLBPrivateNetworksResponse struct {
	// PrivateNetwork: list of Private Network objects attached to the Load Balancer.
	PrivateNetwork []*PrivateNetwork `json:"private_network"`
	// TotalCount: total number of objects in the response.
	TotalCount uint32 `json:"total_count"`
}

// ListLBTypesResponse: list lb types response.
type ListLBTypesResponse struct {
	// LBTypes: list of Load Balancer commercial offer type objects.
	LBTypes []*LBType `json:"lb_types"`
	// TotalCount: total number of Load Balancer offer type objects.
	TotalCount uint32 `json:"total_count"`
}

// ListLBsResponse: list lbs response.
type ListLBsResponse struct {
	// LBs: list of Load Balancer objects.
	LBs []*LB `json:"lbs"`
	// TotalCount: the total number of Load Balancer objects.
	TotalCount uint32 `json:"total_count"`
}

// ListRoutesResponse: list routes response.
type ListRoutesResponse struct {
	// Routes: list of route objects.
	Routes []*Route `json:"routes"`
	// TotalCount: the total number of route objects.
	TotalCount uint32 `json:"total_count"`
}

// ListSubscriberResponse: list subscriber response.
type ListSubscriberResponse struct {
	// Subscribers: list of subscriber objects.
	Subscribers []*Subscriber `json:"subscribers"`
	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// PrivateNetwork: private network.
type PrivateNetwork struct {
	// LB: load Balancer object which is attached to the Private Network.
	LB *LB `json:"lb"`
	// StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: defines whether to let DHCP assign IP addresses.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
	// IpamConfig: for internal use only.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	IpamConfig *PrivateNetworkIpamConfig `json:"ipam_config,omitempty"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"private_network_id"`
	// Status: status of Private Network connection.
	// Default value: unknown
	Status PrivateNetworkStatus `json:"status"`
	// CreatedAt: date on which the Private Network was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the PN was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
}

type PrivateNetworkDHCPConfig struct {
}

type PrivateNetworkIpamConfig struct {
}

// PrivateNetworkStaticConfig: private network. static config.
type PrivateNetworkStaticConfig struct {
	// IPAddress: array of a local IP address for the Load Balancer on this Private Network.
	IPAddress []string `json:"ip_address"`
}

// Route: route.
type Route struct {
	// ID: route ID.
	ID string `json:"id"`
	// FrontendID: ID of the source frontend.
	FrontendID string `json:"frontend_id"`
	// BackendID: ID of the target backend.
	BackendID string `json:"backend_id"`
	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match"`
	// CreatedAt: date on which the route was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the route was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
}

// RouteMatch: route. match.
type RouteMatch struct {
	// Sni: server Name Indication (SNI) value to match.
	// Value to match in the Server Name Indication TLS extension (SNI) field from an incoming connection made via an SSL/TLS transport layer. This field should be set for routes on TCP Load Balancers.
	// Precisely one of HostHeader, Sni must be set.
	Sni *string `json:"sni,omitempty"`
	// HostHeader: HTTP host header to match.
	// Value to match in the HTTP Host request header from an incoming connection. This field should be set for routes on HTTP Load Balancers.
	// Precisely one of HostHeader, Sni must be set.
	HostHeader *string `json:"host_header,omitempty"`
}

// SetACLsResponse: set acls response.
type SetACLsResponse struct {
	// ACLs: list of ACL objects.
	ACLs []*ACL `json:"acls"`
	// TotalCount: the total number of ACL objects.
	TotalCount uint32 `json:"total_count"`
}

// Subscriber: subscriber.
type Subscriber struct {
	// ID: subscriber ID.
	ID string `json:"id"`
	// Name: subscriber name.
	Name string `json:"name"`
	// EmailConfig: email address of subscriber.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webhook URI of subscriber.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// SubscriberEmailConfig: subscriber. email config.
type SubscriberEmailConfig struct {
	// Email: email address to send alerts to.
	Email string `json:"email"`
}

// SubscriberWebhookConfig: webhook alert of subscriber.
// Subscriber. webhook config.
type SubscriberWebhookConfig struct {
	// URI: URI to receive POST requests.
	URI string `json:"uri"`
}

// Service ZonedAPI

// Zones list localities the api is available in
func (s *ZonedAPI) Zones() []scw.Zone {
	return []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZonePlWaw1, scw.ZonePlWaw2}
}

type ZonedAPIListLBsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Name: load Balancer name to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of Load Balancers in the response.
	// Default value: created_at_asc
	OrderBy ListLBsRequestOrderBy `json:"-"`
	// PageSize: number of Load Balancers to return.
	PageSize *uint32 `json:"-"`
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// OrganizationID: organization ID to filter for, only Load Balancers from this Organization will be returned.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID to filter for, only Load Balancers from this Project will be returned.
	ProjectID *string `json:"-"`
}

// ListLBs: list Load Balancers.
// List all Load Balancers in the specified zone, for a Scaleway Organization or Scaleway Project. By default, the Load Balancers returned in the list are ordered by creation date in ascending order, though this can be modified via the `order_by` field.
func (s *ZonedAPI) ListLBs(req *ZonedAPIListLBsRequest, opts ...scw.RequestOption) (*ListLBsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Deprecated: OrganizationID: scaleway Organization to create the Load Balancer in.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: scaleway Project to create the Load Balancer in.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Name: name for the Load Balancer.
	Name string `json:"name"`
	// Description: description for the Load Balancer.
	Description string `json:"description"`
	// IPID: ID of an existing flexible IP address to attach to the Load Balancer.
	IPID *string `json:"ip_id"`
	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`
	// Type: load Balancer commercial offer type. Use the Load Balancer types endpoint to retrieve a list of available offer types.
	Type string `json:"type"`
	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and do not need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// CreateLB: create a Load Balancer.
// Create a new Load Balancer. Note that the Load Balancer will be created without frontends or backends; these must be created separately via the dedicated endpoints.
func (s *ZonedAPI) CreateLB(req *ZonedAPICreateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lb")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// GetLB: get a Load Balancer.
// Retrieve information about an existing Load Balancer, specified by its Load Balancer ID. Its full details, including name, status and IP address, are returned in the response object.
func (s *ZonedAPI) GetLB(req *ZonedAPIGetLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Headers: http.Header{},
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: load Balancer name.
	Name string `json:"name"`
	// Description: load Balancer description.
	Description string `json:"description"`
	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`
	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and don't need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// UpdateLB: update a Load Balancer.
// Update the parameters of an existing Load Balancer, specified by its Load Balancer ID. Note that the request type is PUT and not PATCH. You must set all parameters.
func (s *ZonedAPI) UpdateLB(req *ZonedAPIUpdateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: ID of the Load Balancer to delete.
	LBID string `json:"-"`
	// ReleaseIP: defines whether the Load Balancer's flexible IP should be deleted. Set to true to release the flexible IP, or false to keep it available in your account for future Load Balancers.
	ReleaseIP bool `json:"-"`
}

// DeleteLB: delete a Load Balancer.
// Delete an existing Load Balancer, specified by its Load Balancer ID. Deleting a Load Balancer is permanent, and cannot be undone. The Load Balancer's flexible IP address can either be deleted with the Load Balancer, or kept in your account for future use.
func (s *ZonedAPI) DeleteLB(req *ZonedAPIDeleteLBRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "release_ip", req.ReleaseIP)

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIMigrateLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Type: load Balancer type to migrate to (use the List all Load Balancer offer types endpoint to get a list of available offer types).
	Type string `json:"type"`
}

// MigrateLB: migrate a Load Balancer.
// Migrate an existing Load Balancer from one commercial type to another. Allows you to scale your Load Balancer up or down in terms of bandwidth or multi-cloud provision.
func (s *ZonedAPI) MigrateLB(req *ZonedAPIMigrateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/migrate",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListIPsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of IP addresses to return.
	PageSize *uint32 `json:"-"`
	// IPAddress: IP address to filter for.
	IPAddress *string `json:"-"`
	// OrganizationID: organization ID to filter for, only Load Balancer IP addresses from this Organization will be returned.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID to filter for, only Load Balancer IP addresses from this Project will be returned.
	ProjectID *string `json:"-"`
}

// ListIPs: list IP addresses.
// List the Load Balancer flexible IP addresses held in the account (filtered by Organization ID or Project ID). It is also possible to search for a specific IP address.
func (s *ZonedAPI) ListIPs(req *ZonedAPIListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "ip_address", req.IPAddress)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Deprecated: OrganizationID: organization ID of the Organization where the IP address should be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: project ID of the Project where the IP address should be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse"`
}

// CreateIP: create an IP address.
// Create a new Load Balancer flexible IP address, in the specified Scaleway Project. This can be attached to new Load Balancers created in the future.
func (s *ZonedAPI) CreateIP(req *ZonedAPICreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
}

// GetIP: get an IP address.
// Retrieve the full details of a Load Balancer flexible IP address.
func (s *ZonedAPI) GetIP(req *ZonedAPIGetIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIReleaseIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
}

// ReleaseIP: delete an IP address.
// Delete a Load Balancer flexible IP address. This action is irreversible, and cannot be undone.
func (s *ZonedAPI) ReleaseIP(req *ZonedAPIReleaseIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIUpdateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse"`
}

// UpdateIP: update an IP address.
// Update the reverse DNS of a Load Balancer flexible IP address.
func (s *ZonedAPI) UpdateIP(req *ZonedAPIUpdateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListBackendsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name of the backend to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of backends in the response.
	// Default value: created_at_asc
	OrderBy ListBackendsRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of backends to return.
	PageSize *uint32 `json:"-"`
}

// ListBackends: list the backends of a given Load Balancer.
// List all the backends of a Load Balancer, specified by its Load Balancer ID. By default, results are returned in ascending order by the creation date of each backend. The response is an array of backend objects, containing full details of each one including their configuration parameters such as protocol, port and forwarding algorithm.
func (s *ZonedAPI) ListBackends(req *ZonedAPIListBackendsRequest, opts ...scw.RequestOption) (*ListBackendsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name for the backend.
	Name string `json:"name"`
	// ForwardProtocol: protocol to be used by the backend when forwarding traffic to backend servers.
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: port to be used by the backend when forwarding traffic to backend servers.
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm to be used when determining which backend server to forward new traffic to.
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: defines whether to activate sticky sessions (binding a particular session to a particular backend server) and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie TO stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for cookie-based sticky sessions.
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: object defining the health check to be carried out by the backend when checking the status and health of backend servers.
	HealthCheck *HealthCheck `json:"health_check"`
	// ServerIP: list of backend server IP addresses (IPv4 or IPv6) the backend should forward traffic to.
	ServerIP []string `json:"server_ip"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field.
	SendProxyV2 *bool `json:"send_proxy_v2,omitempty"`
	// TimeoutServer: maximum allowed time for a backend server to process a request.
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`
	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`
	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`
	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`
	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`
}

func (m *ZonedAPICreateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPICreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = ZonedAPICreateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m ZonedAPICreateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType ZonedAPICreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// CreateBackend: create a backend for a given Load Balancer.
// Create a new backend for a given Load Balancer, specifying its full configuration including protocol, port and forwarding algorithm.
func (s *ZonedAPI) CreateBackend(req *ZonedAPICreateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lbb")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
}

// GetBackend: get a backend of a given Load Balancer.
// Get the full details of a given backend, specified by its backend ID. The response contains the backend's full configuration parameters including protocol, port and forwarding algorithm.
func (s *ZonedAPI) GetBackend(req *ZonedAPIGetBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// Name: backend name.
	Name string `json:"name"`
	// ForwardProtocol: protocol to be used by the backend when forwarding traffic to backend servers.
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: port to be used by the backend when forwarding traffic to backend servers.
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm to be used when determining which backend server to forward new traffic to.
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: defines whether to activate sticky sessions (binding a particular session to a particular backend server) and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie to stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for cookie-based sticky sessions.
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field.
	SendProxyV2 *bool `json:"send_proxy_v2,omitempty"`
	// TimeoutServer: maximum allowed time for a backend server to process a request.
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`
	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`
	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`
	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`
	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`
}

func (m *ZonedAPIUpdateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = ZonedAPIUpdateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m ZonedAPIUpdateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType ZonedAPIUpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// UpdateBackend: update a backend of a given Load Balancer.
// Update a backend of a given Load Balancer, specified by its backend ID. Note that the request type is PUT and not PATCH. You must set all parameters.
func (s *ZonedAPI) UpdateBackend(req *ZonedAPIUpdateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: ID of the backend to delete.
	BackendID string `json:"-"`
}

// DeleteBackend: delete a backend of a given Load Balancer.
// Delete a backend of a given Load Balancer, specified by its backend ID. This action is irreversible and cannot be undone.
func (s *ZonedAPI) DeleteBackend(req *ZonedAPIDeleteBackendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIAddBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses to add to backend servers.
	ServerIP []string `json:"server_ip"`
}

// AddBackendServers: add a set of backend servers to a given backend.
// For a given backend specified by its backend ID, add a set of backend servers (identified by their IP addresses) it should forward traffic to. These will be appended to any existing set of backend servers for this backend.
func (s *ZonedAPI) AddBackendServers(req *ZonedAPIAddBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIRemoveBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses to remove from backend servers.
	ServerIP []string `json:"server_ip"`
}

// RemoveBackendServers: remove a set of servers for a given backend.
// For a given backend specified by its backend ID, remove the specified backend servers (identified by their IP addresses) so that it no longer forwards traffic to them.
func (s *ZonedAPI) RemoveBackendServers(req *ZonedAPIRemoveBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPISetBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses for backend servers. Any other existing backend servers will be removed.
	ServerIP []string `json:"server_ip"`
}

// SetBackendServers: define all backend servers for a given backend.
// For a given backend specified by its backend ID, define the set of backend servers (identified by their IP addresses) that it should forward traffic to. Any existing backend servers configured for this backend will be removed.
func (s *ZonedAPI) SetBackendServers(req *ZonedAPISetBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateHealthCheckRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// Port: port to use for the backend server health check.
	Port int32 `json:"port"`
	// CheckDelay: time to wait between two consecutive health checks.
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: maximum time a backend server has to reply to the health check.
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks after which the server will be considered dead.
	CheckMaxRetries int32 `json:"check_max_retries"`
	// CheckSendProxy: defines whether proxy protocol should be activated for the health check.
	CheckSendProxy bool `json:"check_send_proxy"`
	// TCPConfig: object to configure a basic TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`
	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22, for older versions, use a TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`
	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`
	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`
	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
	// TransientCheckDelay: time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN).
	TransientCheckDelay *scw.Duration `json:"transient_check_delay"`
}

func (m *ZonedAPIUpdateHealthCheckRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = ZonedAPIUpdateHealthCheckRequest(tmp.tmpType)

	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	return nil
}

func (m ZonedAPIUpdateHealthCheckRequest) MarshalJSON() ([]byte, error) {
	type tmpType ZonedAPIUpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{
		tmpType: tmpType(m),

		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

// UpdateHealthCheck: update a health check for a given backend.
// Update the configuration of the health check performed by a given backend to verify the health of its backend servers, identified by its backend ID. Note that the request type is PUT and not PATCH. You must set all parameters.
func (s *ZonedAPI) UpdateHealthCheck(req *ZonedAPIUpdateHealthCheckRequest, opts ...scw.RequestOption) (*HealthCheck, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/healthcheck",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp HealthCheck

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListFrontendsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name of the frontend to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of frontends in the response.
	// Default value: created_at_asc
	OrderBy ListFrontendsRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of frontends to return.
	PageSize *uint32 `json:"-"`
}

// ListFrontends: list frontends of a given Load Balancer.
// List all the frontends of a Load Balancer, specified by its Load Balancer ID. By default, results are returned in ascending order by the creation date of each frontend. The response is an array of frontend objects, containing full details of each one including the port they listen on and the backend they are attached to.
func (s *ZonedAPI) ListFrontends(req *ZonedAPIListFrontendsRequest, opts ...scw.RequestOption) (*ListFrontendsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListFrontendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID (ID of the Load Balancer to attach the frontend to).
	LBID string `json:"-"`
	// Name: name for the frontend.
	Name string `json:"name"`
	// InboundPort: port the frontend should listen on.
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID (ID of the backend the frontend should pass traffic to).
	BackendID string `json:"backend_id"`
	// TimeoutClient: maximum allowed inactivity time on the client side.
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`
	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids"`
	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`
}

func (m *ZonedAPICreateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPICreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = ZonedAPICreateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m ZonedAPICreateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType ZonedAPICreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// CreateFrontend: create a frontend in a given Load Balancer.
// Create a new frontend for a given Load Balancer, specifying its configuration including the port it should listen on and the backend to attach it to.
func (s *ZonedAPI) CreateFrontend(req *ZonedAPICreateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lbf")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
}

// GetFrontend: get a frontend.
// Get the full details of a given frontend, specified by its frontend ID. The response contains the frontend's full configuration parameters including the backend it is attached to, the port it listens on, and any certificates it has.
func (s *ZonedAPI) GetFrontend(req *ZonedAPIGetFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
	// Name: frontend name.
	Name string `json:"name"`
	// InboundPort: port the frontend should listen on.
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID (ID of the backend the frontend should pass traffic to).
	BackendID string `json:"backend_id"`
	// TimeoutClient: maximum allowed inactivity time on the client side.
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`
	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids"`
	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`
}

func (m *ZonedAPIUpdateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = ZonedAPIUpdateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m ZonedAPIUpdateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType ZonedAPIUpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// UpdateFrontend: update a frontend.
// Update a given frontend, specified by its frontend ID. You can update configuration parameters including its name and the port it listens on. Note that the request type is PUT and not PATCH. You must set all parameters.
func (s *ZonedAPI) UpdateFrontend(req *ZonedAPIUpdateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: ID of the frontend to delete.
	FrontendID string `json:"-"`
}

// DeleteFrontend: delete a frontend.
// Delete a given frontend, specified by its frontend ID. This action is irreversible and cannot be undone.
func (s *ZonedAPI) DeleteFrontend(req *ZonedAPIDeleteFrontendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIListRoutesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// OrderBy: sort order of routes in the response.
	// Default value: created_at_asc
	OrderBy ListRoutesRequestOrderBy `json:"-"`
	// PageSize: the number of route objects to return.
	PageSize *uint32 `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// FrontendID: frontend ID to filter for, only Routes from this Frontend will be returned.
	FrontendID *string `json:"-"`
}

// ListRoutes: list all routes.
// List all routes for a given frontend. The response is an array of routes, each one  with a specified backend to direct to if a certain condition is matched (based on the value of the SNI field or HTTP Host header).
func (s *ZonedAPI) ListRoutes(req *ZonedAPIListRoutesRequest, opts ...scw.RequestOption) (*ListRoutesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "frontend_id", req.FrontendID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListRoutesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: ID of the source frontend to create the route on.
	FrontendID string `json:"frontend_id"`
	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`
	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match"`
}

// CreateRoute: create a route.
// Create a new route on a given frontend. To configure a route, specify the backend to direct to if a certain condition is matched (based on the value of the SNI field or HTTP Host header).
func (s *ZonedAPI) CreateRoute(req *ZonedAPICreateRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
}

// GetRoute: get a route.
// Retrieve information about an existing route, specified by its route ID. Its full details, origin frontend, target backend and match condition, are returned in the response object.
func (s *ZonedAPI) GetRoute(req *ZonedAPIGetRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return nil, errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`
	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match"`
}

// UpdateRoute: update a route.
// Update the configuration of an existing route, specified by its route ID.
func (s *ZonedAPI) UpdateRoute(req *ZonedAPIUpdateRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return nil, errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
}

// DeleteRoute: delete a route.
// Delete an existing route, specified by its route ID. Deleting a route is permanent, and cannot be undone.
func (s *ZonedAPI) DeleteRoute(req *ZonedAPIDeleteRouteRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIGetLBStatsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// Deprecated: GetLBStats: get usage statistics of a given Load Balancer.
func (s *ZonedAPI) GetLBStats(req *ZonedAPIGetLBStatsRequest, opts ...scw.RequestOption) (*LBStats, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/stats",
		Headers: http.Header{},
	}

	var resp LBStats

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListBackendStatsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of items to return.
	PageSize *uint32 `json:"-"`
}

// ListBackendStats: list backend server statistics.
// List information about your backend servers, including their state and the result of their last health check.
func (s *ZonedAPI) ListBackendStats(req *ZonedAPIListBackendStatsRequest, opts ...scw.RequestOption) (*ListBackendStatsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backend-stats",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendStatsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListACLsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID (ACLs attached to this frontend will be returned in the response).
	FrontendID string `json:"-"`
	// OrderBy: sort order of ACLs in the response.
	// Default value: created_at_asc
	OrderBy ListACLRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of ACLs to return.
	PageSize *uint32 `json:"-"`
	// Name: ACL name to filter for.
	Name *string `json:"-"`
}

// ListACLs: list ACLs for a given frontend.
// List the ACLs for a given frontend, specified by its frontend ID. The response is an array of ACL objects, each one representing an ACL that denies or allows traffic based on certain conditions.
func (s *ZonedAPI) ListACLs(req *ZonedAPIListACLsRequest, opts ...scw.RequestOption) (*ListACLResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListACLResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID to attach the ACL to.
	FrontendID string `json:"-"`
	// Name: ACL name.
	Name string `json:"name"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// Description: ACL description.
	Description string `json:"description"`
}

// CreateACL: create an ACL for a given frontend.
// Create a new ACL for a given frontend. Each ACL must have a name, an action to perform (allow or deny), and a match rule (the action is carried out when the incoming traffic matches the rule).
func (s *ZonedAPI) CreateACL(req *ZonedAPICreateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("acl")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// GetACL: get an ACL.
// Get information for a particular ACL, specified by its ACL ID. The response returns full details of the ACL, including its name, action, match rule and frontend.
func (s *ZonedAPI) GetACL(req *ZonedAPIGetACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
	// Name: ACL name.
	Name string `json:"name"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// Description: ACL description.
	Description *string `json:"description"`
}

// UpdateACL: update an ACL.
// Update a particular ACL, specified by its ACL ID. You can update details including its name, action and match rule.
func (s *ZonedAPI) UpdateACL(req *ZonedAPIUpdateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// DeleteACL: delete an ACL.
// Delete an ACL, specified by its ACL ID. Deleting an ACL is irreversible and cannot be undone.
func (s *ZonedAPI) DeleteACL(req *ZonedAPIDeleteACLRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPISetACLsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
	// ACLs: list of ACLs for this frontend. Any other existing ACLs on this frontend will be removed.
	ACLs []*ACLSpec `json:"acls"`
}

// SetACLs: define all ACLs for a given frontend.
// For a given frontend specified by its frontend ID, define and add the complete set of ACLS for that frontend. Any existing ACLs on this frontend will be removed.
func (s *ZonedAPI) SetACLs(req *ZonedAPISetACLsRequest, opts ...scw.RequestOption) (*SetACLsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetACLsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name for the certificate.
	Name string `json:"name"`
	// Letsencrypt: object to define a new Let's Encrypt certificate to be generated.
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`
	// CustomCertificate: object to define an existing custom certificate to be imported.
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateCertificate: create an SSL/TLS certificate.
// Generate a new SSL/TLS certificate for a given Load Balancer. You can choose to create a Let's Encrypt certificate, or import a custom certificate.
func (s *ZonedAPI) CreateCertificate(req *ZonedAPICreateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("certificate")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListCertificatesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// OrderBy: sort order of certificates in the response.
	// Default value: created_at_asc
	OrderBy ListCertificatesRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of certificates to return.
	PageSize *uint32 `json:"-"`
	// Name: certificate name to filter for, only certificates of this name will be returned.
	Name *string `json:"-"`
}

// ListCertificates: list all SSL/TLS certificates on a given Load Balancer.
// List all the SSL/TLS certificates on a given Load Balancer. The response is an array of certificate objects, which are by default listed in ascending order of creation date.
func (s *ZonedAPI) ListCertificates(req *ZonedAPIListCertificatesRequest, opts ...scw.RequestOption) (*ListCertificatesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// GetCertificate: get an SSL/TLS certificate.
// Get information for a particular SSL/TLS certificate, specified by its certificate ID. The response returns full details of the certificate, including its type, main domain name, and alternative domain names.
func (s *ZonedAPI) GetCertificate(req *ZonedAPIGetCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
	// Name: certificate name.
	Name string `json:"name"`
}

// UpdateCertificate: update an SSL/TLS certificate.
// Update the name of a particular SSL/TLS certificate, specified by its certificate ID.
func (s *ZonedAPI) UpdateCertificate(req *ZonedAPIUpdateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// DeleteCertificate: delete an SSL/TLS certificate.
// Delete an SSL/TLS certificate, specified by its certificate ID. Deleting a certificate is irreversible and cannot be undone.
func (s *ZonedAPI) DeleteCertificate(req *ZonedAPIDeleteCertificateRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPIListLBTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
}

// ListLBTypes: list all Load Balancer offer types.
// List all the different commercial Load Balancer types. The response includes an array of offer types, each with a name, description, and information about its stock availability.
func (s *ZonedAPI) ListLBTypes(req *ZonedAPIListLBTypesRequest, opts ...scw.RequestOption) (*ListLBTypesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb-types",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPICreateSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// Name: subscriber name.
	Name string `json:"name"`
	// EmailConfig: email address configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
	// Deprecated: OrganizationID: organization ID to create the subscriber in.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: project ID to create the subscriber in.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// CreateSubscriber: create a subscriber.
// Create a new subscriber, either with an email configuration or a webhook configuration, for a specified Scaleway Project.
func (s *ZonedAPI) CreateSubscriber(req *ZonedAPICreateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIGetSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// GetSubscriber: get a subscriber.
// Retrieve information about an existing subscriber, specified by its subscriber ID. Its full details, including name and email/webhook configuration, are returned in the response object.
func (s *ZonedAPI) GetSubscriber(req *ZonedAPIGetSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// OrderBy: sort order of subscribers in the response.
	// Default value: created_at_asc
	OrderBy ListSubscriberRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
	// Name: subscriber name to search for.
	Name *string `json:"-"`
	// OrganizationID: filter subscribers by Organization ID.
	OrganizationID *string `json:"-"`
	// ProjectID: filter subscribers by Project ID.
	ProjectID *string `json:"-"`
}

// ListSubscriber: list all subscribers.
// List all subscribers to Load Balancer alerts. By default, returns all subscribers to Load Balancer alerts for the Organization associated with the authentication token used for the request.
func (s *ZonedAPI) ListSubscriber(req *ZonedAPIListSubscriberRequest, opts ...scw.RequestOption) (*ListSubscriberResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSubscriberResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUpdateSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
	// Name: subscriber name.
	Name string `json:"name"`
	// EmailConfig: email address configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webhook URI configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// UpdateSubscriber: update a subscriber.
// Update the parameters of a given subscriber (e.g. name, webhook configuration, email configuration), specified by its subscriber ID.
func (s *ZonedAPI) UpdateSubscriber(req *ZonedAPIUpdateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDeleteSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// DeleteSubscriber: delete a subscriber.
// Delete an existing subscriber, specified by its subscriber ID. Deleting a subscriber is permanent, and cannot be undone.
func (s *ZonedAPI) DeleteSubscriber(req *ZonedAPIDeleteSubscriberRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/subscription/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ZonedAPISubscribeToLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"subscriber_id"`
}

// SubscribeToLB: subscribe a subscriber to alerts for a given Load Balancer.
// Subscribe an existing subscriber to alerts for a given Load Balancer.
func (s *ZonedAPI) SubscribeToLB(req *ZonedAPISubscribeToLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/" + fmt.Sprint(req.LBID) + "/subscribe",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIUnsubscribeFromLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// UnsubscribeFromLB: unsubscribe a subscriber from alerts for a given Load Balancer.
// Unsubscribe a subscriber from alerts for a given Load Balancer. The subscriber is not deleted, and can be resubscribed in the future if necessary.
func (s *ZonedAPI) UnsubscribeFromLB(req *ZonedAPIUnsubscribeFromLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/" + fmt.Sprint(req.LBID) + "/unsubscribe",
		Headers: http.Header{},
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListLBPrivateNetworksRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// OrderBy: sort order of Private Network objects in the response.
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`
	// PageSize: number of objects to return.
	PageSize *uint32 `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
}

// ListLBPrivateNetworks: list Private Networks attached to a Load Balancer.
// List the Private Networks attached to a given Load Balancer, specified by its Load Balancer ID. The response is an array of Private Network objects, giving information including the status, configuration, name and creation date of each Private Network.
func (s *ZonedAPI) ListLBPrivateNetworks(req *ZonedAPIListLBPrivateNetworksRequest, opts ...scw.RequestOption) (*ListLBPrivateNetworksResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIAttachPrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`
	// StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: defines whether to let DHCP assign IP addresses.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
	// IpamConfig: for internal use only.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	IpamConfig *PrivateNetworkIpamConfig `json:"ipam_config,omitempty"`
}

// AttachPrivateNetwork: attach a Load Balancer to a Private Network.
// Attach a specified Load Balancer to a specified Private Network, defining a static or DHCP configuration for the Load Balancer on the network.
func (s *ZonedAPI) AttachPrivateNetwork(req *ZonedAPIAttachPrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return nil, errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/attach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIDetachPrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID.
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id.
	PrivateNetworkID string `json:"-"`
}

// DetachPrivateNetwork: detach Load Balancer from Private Network.
// Detach a specified Load Balancer from a specified Private Network.
func (s *ZonedAPI) DetachPrivateNetwork(req *ZonedAPIDetachPrivateNetworkRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return errors.New("field LBID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/detach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// Service API

// Regions list localities the api is available in
func (s *API) Regions() []scw.Region {
	return []scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}
}

type ListLBsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Name: load Balancer name to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of Load Balancers in the response.
	// Default value: created_at_asc
	OrderBy ListLBsRequestOrderBy `json:"-"`
	// PageSize: number of Load Balancers to return.
	PageSize *uint32 `json:"-"`
	// Page: page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// OrganizationID: organization ID to filter for, only Load Balancers from this Organization will be returned.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID to filter for, only Load Balancers from this Project will be returned.
	ProjectID *string `json:"-"`
}

// ListLBs: list load balancers.
func (s *API) ListLBs(req *ListLBsRequest, opts ...scw.RequestOption) (*ListLBsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Deprecated: OrganizationID: scaleway Organization to create the Load Balancer in.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: scaleway Project to create the Load Balancer in.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Name: name for the Load Balancer.
	Name string `json:"name"`
	// Description: description for the Load Balancer.
	Description string `json:"description"`
	// IPID: ID of an existing flexible IP address to attach to the Load Balancer.
	IPID *string `json:"ip_id"`
	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`
	// Type: load Balancer commercial offer type. Use the Load Balancer types endpoint to retrieve a list of available offer types.
	Type string `json:"type"`
	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and do not need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// CreateLB: create a load balancer.
func (s *API) CreateLB(req *CreateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lb")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// GetLB: get a load balancer.
func (s *API) GetLB(req *GetLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Headers: http.Header{},
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: load Balancer name.
	Name string `json:"name"`
	// Description: load Balancer description.
	Description string `json:"description"`
	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`
	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and don't need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// UpdateLB: update a load balancer.
func (s *API) UpdateLB(req *UpdateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: ID of the Load Balancer to delete.
	LBID string `json:"-"`
	// ReleaseIP: defines whether the Load Balancer's flexible IP should be deleted. Set to true to release the flexible IP, or false to keep it available in your account for future Load Balancers.
	ReleaseIP bool `json:"-"`
}

// DeleteLB: delete a load balancer.
func (s *API) DeleteLB(req *DeleteLBRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "release_ip", req.ReleaseIP)

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type MigrateLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Type: load Balancer type to migrate to (use the List all Load Balancer offer types endpoint to get a list of available offer types).
	Type string `json:"type"`
}

// MigrateLB: migrate a load balancer.
func (s *API) MigrateLB(req *MigrateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/migrate",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListIPsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of IP addresses to return.
	PageSize *uint32 `json:"-"`
	// IPAddress: IP address to filter for.
	IPAddress *string `json:"-"`
	// OrganizationID: organization ID to filter for, only Load Balancer IP addresses from this Organization will be returned.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID to filter for, only Load Balancer IP addresses from this Project will be returned.
	ProjectID *string `json:"-"`
}

// ListIPs: list IPs.
func (s *API) ListIPs(req *ListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "ip_address", req.IPAddress)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Deprecated: OrganizationID: organization ID of the Organization where the IP address should be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: project ID of the Project where the IP address should be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse"`
}

// CreateIP: create an IP.
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
}

// GetIP: get an IP.
func (s *API) GetIP(req *GetIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ReleaseIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
}

// ReleaseIP: delete an IP.
func (s *API) ReleaseIP(req *ReleaseIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type UpdateIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// IPID: IP address ID.
	IPID string `json:"-"`
	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse"`
}

// UpdateIP: update an IP.
func (s *API) UpdateIP(req *UpdateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListBackendsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name of the backend to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of backends in the response.
	// Default value: created_at_asc
	OrderBy ListBackendsRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of backends to return.
	PageSize *uint32 `json:"-"`
}

// ListBackends: list backends in a given load balancer.
func (s *API) ListBackends(req *ListBackendsRequest, opts ...scw.RequestOption) (*ListBackendsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name for the backend.
	Name string `json:"name"`
	// ForwardProtocol: protocol to be used by the backend when forwarding traffic to backend servers.
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: port to be used by the backend when forwarding traffic to backend servers.
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm to be used when determining which backend server to forward new traffic to.
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: defines whether to activate sticky sessions (binding a particular session to a particular backend server) and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie TO stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for cookie-based sticky sessions.
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: object defining the health check to be carried out by the backend when checking the status and health of backend servers.
	HealthCheck *HealthCheck `json:"health_check"`
	// ServerIP: list of backend server IP addresses (IPv4 or IPv6) the backend should forward traffic to.
	ServerIP []string `json:"server_ip"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field.
	SendProxyV2 *bool `json:"send_proxy_v2,omitempty"`
	// TimeoutServer: maximum allowed time for a backend server to process a request.
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`
	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`
	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`
	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`
	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`
}

func (m *CreateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = CreateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m CreateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType CreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// CreateBackend: create a backend in a given load balancer.
func (s *API) CreateBackend(req *CreateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lbb")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
}

// GetBackend: get a backend in a given load balancer.
func (s *API) GetBackend(req *GetBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// Name: backend name.
	Name string `json:"name"`
	// ForwardProtocol: protocol to be used by the backend when forwarding traffic to backend servers.
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: port to be used by the backend when forwarding traffic to backend servers.
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm to be used when determining which backend server to forward new traffic to.
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: defines whether to activate sticky sessions (binding a particular session to a particular backend server) and the method to use if so. None disables sticky sessions. Cookie-based uses an HTTP cookie to stick a session to a backend server. Table-based uses the source (client) IP address to stick a session to a backend server.
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for cookie-based sticky sessions.
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field.
	SendProxyV2 *bool `json:"send_proxy_v2,omitempty"`
	// TimeoutServer: maximum allowed time for a backend server to process a request.
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`
	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`
	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`
	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`
	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`
}

func (m *UpdateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m UpdateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// UpdateBackend: update a backend in a given load balancer.
func (s *API) UpdateBackend(req *UpdateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: ID of the backend to delete.
	BackendID string `json:"-"`
}

// DeleteBackend: delete a backend in a given load balancer.
func (s *API) DeleteBackend(req *DeleteBackendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type AddBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses to add to backend servers.
	ServerIP []string `json:"server_ip"`
}

// AddBackendServers: add a set of servers in a given backend.
func (s *API) AddBackendServers(req *AddBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RemoveBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses to remove from backend servers.
	ServerIP []string `json:"server_ip"`
}

// RemoveBackendServers: remove a set of servers for a given backend.
func (s *API) RemoveBackendServers(req *RemoveBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// ServerIP: list of IP addresses for backend servers. Any other existing backend servers will be removed.
	ServerIP []string `json:"server_ip"`
}

// SetBackendServers: define all servers in a given backend.
func (s *API) SetBackendServers(req *SetBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateHealthCheckRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// BackendID: backend ID.
	BackendID string `json:"-"`
	// Port: port to use for the backend server health check.
	Port int32 `json:"port"`
	// CheckDelay: time to wait between two consecutive health checks.
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: maximum time a backend server has to reply to the health check.
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks after which the server will be considered dead.
	CheckMaxRetries int32 `json:"check_max_retries"`
	// CheckSendProxy: defines whether proxy protocol should be activated for the health check.
	CheckSendProxy bool `json:"check_send_proxy"`
	// TCPConfig: object to configure a basic TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`
	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22, for older versions, use a TCP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`
	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`
	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`
	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
	// TransientCheckDelay: time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN).
	TransientCheckDelay *scw.Duration `json:"transient_check_delay"`
}

func (m *UpdateHealthCheckRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateHealthCheckRequest(tmp.tmpType)

	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	return nil
}

func (m UpdateHealthCheckRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{
		tmpType: tmpType(m),

		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

// UpdateHealthCheck: update an health check for a given backend.
func (s *API) UpdateHealthCheck(req *UpdateHealthCheckRequest, opts ...scw.RequestOption) (*HealthCheck, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/healthcheck",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp HealthCheck

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListFrontendsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name of the frontend to filter for.
	Name *string `json:"-"`
	// OrderBy: sort order of frontends in the response.
	// Default value: created_at_asc
	OrderBy ListFrontendsRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of frontends to return.
	PageSize *uint32 `json:"-"`
}

// ListFrontends: list frontends in a given load balancer.
func (s *API) ListFrontends(req *ListFrontendsRequest, opts ...scw.RequestOption) (*ListFrontendsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListFrontendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID (ID of the Load Balancer to attach the frontend to).
	LBID string `json:"-"`
	// Name: name for the frontend.
	Name string `json:"name"`
	// InboundPort: port the frontend should listen on.
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID (ID of the backend the frontend should pass traffic to).
	BackendID string `json:"backend_id"`
	// TimeoutClient: maximum allowed inactivity time on the client side.
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`
	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids"`
	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`
}

func (m *CreateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = CreateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m CreateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType CreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// CreateFrontend: create a frontend in a given load balancer.
func (s *API) CreateFrontend(req *CreateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lbf")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
}

// GetFrontend: get a frontend.
func (s *API) GetFrontend(req *GetFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
	// Name: frontend name.
	Name string `json:"name"`
	// InboundPort: port the frontend should listen on.
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID (ID of the backend the frontend should pass traffic to).
	BackendID string `json:"backend_id"`
	// TimeoutClient: maximum allowed inactivity time on the client side.
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`
	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids"`
	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`
}

func (m *UpdateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m UpdateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// UpdateFrontend: update a frontend.
func (s *API) UpdateFrontend(req *UpdateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: ID of the frontend to delete.
	FrontendID string `json:"-"`
}

// DeleteFrontend: delete a frontend.
func (s *API) DeleteFrontend(req *DeleteFrontendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListRoutesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// OrderBy: sort order of routes in the response.
	// Default value: created_at_asc
	OrderBy ListRoutesRequestOrderBy `json:"-"`
	// PageSize: the number of route objects to return.
	PageSize *uint32 `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// FrontendID: frontend ID to filter for, only Routes from this Frontend will be returned.
	FrontendID *string `json:"-"`
}

// ListRoutes: list all backend redirections.
func (s *API) ListRoutes(req *ListRoutesRequest, opts ...scw.RequestOption) (*ListRoutesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "frontend_id", req.FrontendID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListRoutesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: ID of the source frontend to create the route on.
	FrontendID string `json:"frontend_id"`
	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`
	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match"`
}

// CreateRoute: create a backend redirection.
func (s *API) CreateRoute(req *CreateRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
}

// GetRoute: get single backend redirection.
func (s *API) GetRoute(req *GetRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return nil, errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`
	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match"`
}

// UpdateRoute: edit a backend redirection.
func (s *API) UpdateRoute(req *UpdateRouteRequest, opts ...scw.RequestOption) (*Route, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return nil, errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// RouteID: route ID.
	RouteID string `json:"-"`
}

// DeleteRoute: delete a backend redirection.
func (s *API) DeleteRoute(req *DeleteRouteRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.RouteID) == "" {
		return errors.New("field RouteID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type GetLBStatsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// Deprecated: GetLBStats: get usage statistics of a given load balancer.
func (s *API) GetLBStats(req *GetLBStatsRequest, opts ...scw.RequestOption) (*LBStats, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/stats",
		Headers: http.Header{},
	}

	var resp LBStats

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListBackendStatsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of items to return.
	PageSize *uint32 `json:"-"`
}

func (s *API) ListBackendStats(req *ListBackendStatsRequest, opts ...scw.RequestOption) (*ListBackendStatsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backend-stats",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendStatsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListACLsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID (ACLs attached to this frontend will be returned in the response).
	FrontendID string `json:"-"`
	// OrderBy: sort order of ACLs in the response.
	// Default value: created_at_asc
	OrderBy ListACLRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of ACLs to return.
	PageSize *uint32 `json:"-"`
	// Name: ACL name to filter for.
	Name *string `json:"-"`
}

// ListACLs: list ACL for a given frontend.
func (s *API) ListACLs(req *ListACLsRequest, opts ...scw.RequestOption) (*ListACLResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListACLResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID to attach the ACL to.
	FrontendID string `json:"-"`
	// Name: ACL name.
	Name string `json:"name"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// Description: ACL description.
	Description string `json:"description"`
}

// CreateACL: create an ACL for a given frontend.
func (s *API) CreateACL(req *CreateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("acl")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// GetACL: get an ACL.
func (s *API) GetACL(req *GetACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
	// Name: ACL name.
	Name string `json:"name"`
	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`
	// Match: ACL match filter object. One of `ip_subnet` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match"`
	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`
	// Description: ACL description.
	Description *string `json:"description"`
}

// UpdateACL: update an ACL.
func (s *API) UpdateACL(req *UpdateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// DeleteACL: delete an ACL.
func (s *API) DeleteACL(req *DeleteACLRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type CreateCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// Name: name for the certificate.
	Name string `json:"name"`
	// Letsencrypt: object to define a new Let's Encrypt certificate to be generated.
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`
	// CustomCertificate: object to define an existing custom certificate to be imported.
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateCertificate: create a TLS certificate.
// Generate a new TLS certificate using Let's Encrypt or import your certificate.
func (s *API) CreateCertificate(req *CreateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("certificate")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListCertificatesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// OrderBy: sort order of certificates in the response.
	// Default value: created_at_asc
	OrderBy ListCertificatesRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: number of certificates to return.
	PageSize *uint32 `json:"-"`
	// Name: certificate name to filter for, only certificates of this name will be returned.
	Name *string `json:"-"`
}

// ListCertificates: list all TLS certificates on a given load balancer.
func (s *API) ListCertificates(req *ListCertificatesRequest, opts ...scw.RequestOption) (*ListCertificatesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// GetCertificate: get a TLS certificate.
func (s *API) GetCertificate(req *GetCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
	// Name: certificate name.
	Name string `json:"name"`
}

// UpdateCertificate: update a TLS certificate.
func (s *API) UpdateCertificate(req *UpdateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// DeleteCertificate: delete a TLS certificate.
func (s *API) DeleteCertificate(req *DeleteCertificateRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListLBTypesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
}

// ListLBTypes: list all load balancer offer type.
func (s *API) ListLBTypes(req *ListLBTypesRequest, opts ...scw.RequestOption) (*ListLBTypesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb-types",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Name: subscriber name.
	Name string `json:"name"`
	// EmailConfig: email address configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
	// Deprecated: OrganizationID: organization ID to create the subscriber in.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: project ID to create the subscriber in.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// CreateSubscriber: create a subscriber, webhook or email.
func (s *API) CreateSubscriber(req *CreateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// GetSubscriber: get a subscriber.
func (s *API) GetSubscriber(req *GetSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// OrderBy: sort order of subscribers in the response.
	// Default value: created_at_asc
	OrderBy ListSubscriberRequestOrderBy `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
	// Name: subscriber name to search for.
	Name *string `json:"-"`
	// OrganizationID: filter subscribers by Organization ID.
	OrganizationID *string `json:"-"`
	// ProjectID: filter subscribers by Project ID.
	ProjectID *string `json:"-"`
}

// ListSubscriber: list all subscriber.
func (s *API) ListSubscriber(req *ListSubscriberRequest, opts ...scw.RequestOption) (*ListSubscriberResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSubscriberResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
	// Name: subscriber name.
	Name string `json:"name"`
	// EmailConfig: email address configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webhook URI configuration.
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// UpdateSubscriber: update a subscriber.
func (s *API) UpdateSubscriber(req *UpdateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// DeleteSubscriber: delete a subscriber.
func (s *API) DeleteSubscriber(req *DeleteSubscriberRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/subscriber/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type SubscribeToLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// SubscriberID: subscriber ID.
	SubscriberID string `json:"subscriber_id"`
}

// SubscribeToLB: subscribe a subscriber to a given load balancer.
func (s *API) SubscribeToLB(req *SubscribeToLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LBID) + "/subscribe",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UnsubscribeFromLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// UnsubscribeFromLB: unsubscribe a subscriber from a given load balancer.
func (s *API) UnsubscribeFromLB(req *UnsubscribeFromLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LBID) + "/unsubscribe",
		Headers: http.Header{},
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListLBPrivateNetworksRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// OrderBy: sort order of Private Network objects in the response.
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`
	// PageSize: number of objects to return.
	PageSize *uint32 `json:"-"`
	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`
}

// ListLBPrivateNetworks: list attached private network of load balancer.
func (s *API) ListLBPrivateNetworks(req *ListLBPrivateNetworksRequest, opts ...scw.RequestOption) (*ListLBPrivateNetworksResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLBPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type AttachPrivateNetworkRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load Balancer ID.
	LBID string `json:"-"`
	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`
	// StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: defines whether to let DHCP assign IP addresses.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
	// IpamConfig: for internal use only.
	// Precisely one of DHCPConfig, IpamConfig, StaticConfig must be set.
	IpamConfig *PrivateNetworkIpamConfig `json:"ipam_config,omitempty"`
}

// AttachPrivateNetwork: add load balancer on instance private network.
func (s *API) AttachPrivateNetwork(req *AttachPrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return nil, errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/attach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DetachPrivateNetworkRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// LBID: load balancer ID.
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id.
	PrivateNetworkID string `json:"-"`
}

// DetachPrivateNetwork: remove load balancer of private network.
func (s *API) DetachPrivateNetwork(req *DetachPrivateNetworkRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return errors.New("field LBID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/detach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLBsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LBs = append(r.LBs, results.LBs...)
	r.TotalCount += uint32(len(results.LBs))
	return uint32(len(results.LBs)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListIPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.IPs = append(r.IPs, results.IPs...)
	r.TotalCount += uint32(len(results.IPs))
	return uint32(len(results.IPs)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListBackendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Backends = append(r.Backends, results.Backends...)
	r.TotalCount += uint32(len(results.Backends))
	return uint32(len(results.Backends)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListFrontendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Frontends = append(r.Frontends, results.Frontends...)
	r.TotalCount += uint32(len(results.Frontends))
	return uint32(len(results.Frontends)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListRoutesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListRoutesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListRoutesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Routes = append(r.Routes, results.Routes...)
	r.TotalCount += uint32(len(results.Routes))
	return uint32(len(results.Routes)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListBackendStatsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.BackendServersStats = append(r.BackendServersStats, results.BackendServersStats...)
	r.TotalCount += uint32(len(results.BackendServersStats))
	return uint32(len(results.BackendServersStats)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListACLResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ACLs = append(r.ACLs, results.ACLs...)
	r.TotalCount += uint32(len(results.ACLs))
	return uint32(len(results.ACLs)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListCertificatesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Certificates = append(r.Certificates, results.Certificates...)
	r.TotalCount += uint32(len(results.Certificates))
	return uint32(len(results.Certificates)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBTypesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLBTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LBTypes = append(r.LBTypes, results.LBTypes...)
	r.TotalCount += uint32(len(results.LBTypes))
	return uint32(len(results.LBTypes)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListSubscriberResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Subscribers = append(r.Subscribers, results.Subscribers...)
	r.TotalCount += uint32(len(results.Subscribers))
	return uint32(len(results.Subscribers)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBPrivateNetworksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBPrivateNetworksResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLBPrivateNetworksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PrivateNetwork = append(r.PrivateNetwork, results.PrivateNetwork...)
	r.TotalCount += uint32(len(results.PrivateNetwork))
	return uint32(len(results.PrivateNetwork)), nil
}
