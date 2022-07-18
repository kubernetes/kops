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

// API: this API allows you to manage your load balancer service
type API struct {
	client *scw.Client
}

// Deprecated NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

// ZonedAPI: this API allows you to manage your load balancer service
type ZonedAPI struct {
	client *scw.Client
}

// NewZonedAPI returns a ZonedAPI object from a Scaleway client.
func NewZonedAPI(client *scw.Client) *ZonedAPI {
	return &ZonedAPI{
		client: client,
	}
}

type ACLActionType string

const (
	// ACLActionTypeAllow is [insert doc].
	ACLActionTypeAllow = ACLActionType("allow")
	// ACLActionTypeDeny is [insert doc].
	ACLActionTypeDeny = ACLActionType("deny")
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
	// ACLHTTPFilterACLHTTPFilterNone is [insert doc].
	ACLHTTPFilterACLHTTPFilterNone = ACLHTTPFilter("acl_http_filter_none")
	// ACLHTTPFilterPathBegin is [insert doc].
	ACLHTTPFilterPathBegin = ACLHTTPFilter("path_begin")
	// ACLHTTPFilterPathEnd is [insert doc].
	ACLHTTPFilterPathEnd = ACLHTTPFilter("path_end")
	// ACLHTTPFilterRegex is [insert doc].
	ACLHTTPFilterRegex = ACLHTTPFilter("regex")
	// ACLHTTPFilterHTTPHeaderMatch is [insert doc].
	ACLHTTPFilterHTTPHeaderMatch = ACLHTTPFilter("http_header_match")
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
	// BackendServerStatsHealthCheckStatusUnknown is [insert doc].
	BackendServerStatsHealthCheckStatusUnknown = BackendServerStatsHealthCheckStatus("unknown")
	// BackendServerStatsHealthCheckStatusNeutral is [insert doc].
	BackendServerStatsHealthCheckStatusNeutral = BackendServerStatsHealthCheckStatus("neutral")
	// BackendServerStatsHealthCheckStatusFailed is [insert doc].
	BackendServerStatsHealthCheckStatusFailed = BackendServerStatsHealthCheckStatus("failed")
	// BackendServerStatsHealthCheckStatusPassed is [insert doc].
	BackendServerStatsHealthCheckStatusPassed = BackendServerStatsHealthCheckStatus("passed")
	// BackendServerStatsHealthCheckStatusCondpass is [insert doc].
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
	// BackendServerStatsServerStateStopped is [insert doc].
	BackendServerStatsServerStateStopped = BackendServerStatsServerState("stopped")
	// BackendServerStatsServerStateStarting is [insert doc].
	BackendServerStatsServerStateStarting = BackendServerStatsServerState("starting")
	// BackendServerStatsServerStateRunning is [insert doc].
	BackendServerStatsServerStateRunning = BackendServerStatsServerState("running")
	// BackendServerStatsServerStateStopping is [insert doc].
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
	// CertificateStatusPending is [insert doc].
	CertificateStatusPending = CertificateStatus("pending")
	// CertificateStatusReady is [insert doc].
	CertificateStatusReady = CertificateStatus("ready")
	// CertificateStatusError is [insert doc].
	CertificateStatusError = CertificateStatus("error")
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
	// CertificateTypeLetsencryt is [insert doc].
	CertificateTypeLetsencryt = CertificateType("letsencryt")
	// CertificateTypeCustom is [insert doc].
	CertificateTypeCustom = CertificateType("custom")
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
	// ForwardPortAlgorithmRoundrobin is [insert doc].
	ForwardPortAlgorithmRoundrobin = ForwardPortAlgorithm("roundrobin")
	// ForwardPortAlgorithmLeastconn is [insert doc].
	ForwardPortAlgorithmLeastconn = ForwardPortAlgorithm("leastconn")
	// ForwardPortAlgorithmFirst is [insert doc].
	ForwardPortAlgorithmFirst = ForwardPortAlgorithm("first")
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
	// InstanceStatusUnknown is [insert doc].
	InstanceStatusUnknown = InstanceStatus("unknown")
	// InstanceStatusReady is [insert doc].
	InstanceStatusReady = InstanceStatus("ready")
	// InstanceStatusPending is [insert doc].
	InstanceStatusPending = InstanceStatus("pending")
	// InstanceStatusStopped is [insert doc].
	InstanceStatusStopped = InstanceStatus("stopped")
	// InstanceStatusError is [insert doc].
	InstanceStatusError = InstanceStatus("error")
	// InstanceStatusLocked is [insert doc].
	InstanceStatusLocked = InstanceStatus("locked")
	// InstanceStatusMigrating is [insert doc].
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
	// LBStatusUnknown is [insert doc].
	LBStatusUnknown = LBStatus("unknown")
	// LBStatusReady is [insert doc].
	LBStatusReady = LBStatus("ready")
	// LBStatusPending is [insert doc].
	LBStatusPending = LBStatus("pending")
	// LBStatusStopped is [insert doc].
	LBStatusStopped = LBStatus("stopped")
	// LBStatusError is [insert doc].
	LBStatusError = LBStatus("error")
	// LBStatusLocked is [insert doc].
	LBStatusLocked = LBStatus("locked")
	// LBStatusMigrating is [insert doc].
	LBStatusMigrating = LBStatus("migrating")
	// LBStatusToCreate is [insert doc].
	LBStatusToCreate = LBStatus("to_create")
	// LBStatusCreating is [insert doc].
	LBStatusCreating = LBStatus("creating")
	// LBStatusToDelete is [insert doc].
	LBStatusToDelete = LBStatus("to_delete")
	// LBStatusDeleting is [insert doc].
	LBStatusDeleting = LBStatus("deleting")
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
	// LBTypeStockUnknown is [insert doc].
	LBTypeStockUnknown = LBTypeStock("unknown")
	// LBTypeStockLowStock is [insert doc].
	LBTypeStockLowStock = LBTypeStock("low_stock")
	// LBTypeStockOutOfStock is [insert doc].
	LBTypeStockOutOfStock = LBTypeStock("out_of_stock")
	// LBTypeStockAvailable is [insert doc].
	LBTypeStockAvailable = LBTypeStock("available")
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
	// ListACLRequestOrderByCreatedAtAsc is [insert doc].
	ListACLRequestOrderByCreatedAtAsc = ListACLRequestOrderBy("created_at_asc")
	// ListACLRequestOrderByCreatedAtDesc is [insert doc].
	ListACLRequestOrderByCreatedAtDesc = ListACLRequestOrderBy("created_at_desc")
	// ListACLRequestOrderByNameAsc is [insert doc].
	ListACLRequestOrderByNameAsc = ListACLRequestOrderBy("name_asc")
	// ListACLRequestOrderByNameDesc is [insert doc].
	ListACLRequestOrderByNameDesc = ListACLRequestOrderBy("name_desc")
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
	// ListBackendsRequestOrderByCreatedAtAsc is [insert doc].
	ListBackendsRequestOrderByCreatedAtAsc = ListBackendsRequestOrderBy("created_at_asc")
	// ListBackendsRequestOrderByCreatedAtDesc is [insert doc].
	ListBackendsRequestOrderByCreatedAtDesc = ListBackendsRequestOrderBy("created_at_desc")
	// ListBackendsRequestOrderByNameAsc is [insert doc].
	ListBackendsRequestOrderByNameAsc = ListBackendsRequestOrderBy("name_asc")
	// ListBackendsRequestOrderByNameDesc is [insert doc].
	ListBackendsRequestOrderByNameDesc = ListBackendsRequestOrderBy("name_desc")
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
	// ListCertificatesRequestOrderByCreatedAtAsc is [insert doc].
	ListCertificatesRequestOrderByCreatedAtAsc = ListCertificatesRequestOrderBy("created_at_asc")
	// ListCertificatesRequestOrderByCreatedAtDesc is [insert doc].
	ListCertificatesRequestOrderByCreatedAtDesc = ListCertificatesRequestOrderBy("created_at_desc")
	// ListCertificatesRequestOrderByNameAsc is [insert doc].
	ListCertificatesRequestOrderByNameAsc = ListCertificatesRequestOrderBy("name_asc")
	// ListCertificatesRequestOrderByNameDesc is [insert doc].
	ListCertificatesRequestOrderByNameDesc = ListCertificatesRequestOrderBy("name_desc")
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
	// ListFrontendsRequestOrderByCreatedAtAsc is [insert doc].
	ListFrontendsRequestOrderByCreatedAtAsc = ListFrontendsRequestOrderBy("created_at_asc")
	// ListFrontendsRequestOrderByCreatedAtDesc is [insert doc].
	ListFrontendsRequestOrderByCreatedAtDesc = ListFrontendsRequestOrderBy("created_at_desc")
	// ListFrontendsRequestOrderByNameAsc is [insert doc].
	ListFrontendsRequestOrderByNameAsc = ListFrontendsRequestOrderBy("name_asc")
	// ListFrontendsRequestOrderByNameDesc is [insert doc].
	ListFrontendsRequestOrderByNameDesc = ListFrontendsRequestOrderBy("name_desc")
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
	// ListLBsRequestOrderByCreatedAtAsc is [insert doc].
	ListLBsRequestOrderByCreatedAtAsc = ListLBsRequestOrderBy("created_at_asc")
	// ListLBsRequestOrderByCreatedAtDesc is [insert doc].
	ListLBsRequestOrderByCreatedAtDesc = ListLBsRequestOrderBy("created_at_desc")
	// ListLBsRequestOrderByNameAsc is [insert doc].
	ListLBsRequestOrderByNameAsc = ListLBsRequestOrderBy("name_asc")
	// ListLBsRequestOrderByNameDesc is [insert doc].
	ListLBsRequestOrderByNameDesc = ListLBsRequestOrderBy("name_desc")
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
	// ListPrivateNetworksRequestOrderByCreatedAtAsc is [insert doc].
	ListPrivateNetworksRequestOrderByCreatedAtAsc = ListPrivateNetworksRequestOrderBy("created_at_asc")
	// ListPrivateNetworksRequestOrderByCreatedAtDesc is [insert doc].
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
	// ListRoutesRequestOrderByCreatedAtAsc is [insert doc].
	ListRoutesRequestOrderByCreatedAtAsc = ListRoutesRequestOrderBy("created_at_asc")
	// ListRoutesRequestOrderByCreatedAtDesc is [insert doc].
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
	// ListSubscriberRequestOrderByCreatedAtAsc is [insert doc].
	ListSubscriberRequestOrderByCreatedAtAsc = ListSubscriberRequestOrderBy("created_at_asc")
	// ListSubscriberRequestOrderByCreatedAtDesc is [insert doc].
	ListSubscriberRequestOrderByCreatedAtDesc = ListSubscriberRequestOrderBy("created_at_desc")
	// ListSubscriberRequestOrderByNameAsc is [insert doc].
	ListSubscriberRequestOrderByNameAsc = ListSubscriberRequestOrderBy("name_asc")
	// ListSubscriberRequestOrderByNameDesc is [insert doc].
	ListSubscriberRequestOrderByNameDesc = ListSubscriberRequestOrderBy("name_desc")
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
	// OnMarkedDownActionOnMarkedDownActionNone is [insert doc].
	OnMarkedDownActionOnMarkedDownActionNone = OnMarkedDownAction("on_marked_down_action_none")
	// OnMarkedDownActionShutdownSessions is [insert doc].
	OnMarkedDownActionShutdownSessions = OnMarkedDownAction("shutdown_sessions")
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
	// PrivateNetworkStatusUnknown is [insert doc].
	PrivateNetworkStatusUnknown = PrivateNetworkStatus("unknown")
	// PrivateNetworkStatusReady is [insert doc].
	PrivateNetworkStatusReady = PrivateNetworkStatus("ready")
	// PrivateNetworkStatusPending is [insert doc].
	PrivateNetworkStatusPending = PrivateNetworkStatus("pending")
	// PrivateNetworkStatusError is [insert doc].
	PrivateNetworkStatusError = PrivateNetworkStatus("error")
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
	// ProtocolTCP is [insert doc].
	ProtocolTCP = Protocol("tcp")
	// ProtocolHTTP is [insert doc].
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

// ProxyProtocol: pROXY protocol, forward client's address (must be supported by backend servers software)
//
// The PROXY protocol informs the other end about the incoming connection, so that it can know the client's address or the public address it accessed to, whatever the upper layer protocol.
//
// * `proxy_protocol_none` Disable proxy protocol.
// * `proxy_protocol_v1` Version one (text format).
// * `proxy_protocol_v2` Version two (binary format).
// * `proxy_protocol_v2_ssl` Version two with SSL connection.
// * `proxy_protocol_v2_ssl_cn` Version two with SSL connection and common name information.
//
type ProxyProtocol string

const (
	// ProxyProtocolProxyProtocolUnknown is [insert doc].
	ProxyProtocolProxyProtocolUnknown = ProxyProtocol("proxy_protocol_unknown")
	// ProxyProtocolProxyProtocolNone is [insert doc].
	ProxyProtocolProxyProtocolNone = ProxyProtocol("proxy_protocol_none")
	// ProxyProtocolProxyProtocolV1 is [insert doc].
	ProxyProtocolProxyProtocolV1 = ProxyProtocol("proxy_protocol_v1")
	// ProxyProtocolProxyProtocolV2 is [insert doc].
	ProxyProtocolProxyProtocolV2 = ProxyProtocol("proxy_protocol_v2")
	// ProxyProtocolProxyProtocolV2Ssl is [insert doc].
	ProxyProtocolProxyProtocolV2Ssl = ProxyProtocol("proxy_protocol_v2_ssl")
	// ProxyProtocolProxyProtocolV2SslCn is [insert doc].
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
	// SSLCompatibilityLevelSslCompatibilityLevelUnknown is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelUnknown = SSLCompatibilityLevel("ssl_compatibility_level_unknown")
	// SSLCompatibilityLevelSslCompatibilityLevelIntermediate is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelIntermediate = SSLCompatibilityLevel("ssl_compatibility_level_intermediate")
	// SSLCompatibilityLevelSslCompatibilityLevelModern is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelModern = SSLCompatibilityLevel("ssl_compatibility_level_modern")
	// SSLCompatibilityLevelSslCompatibilityLevelOld is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelOld = SSLCompatibilityLevel("ssl_compatibility_level_old")
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
	// StickySessionsTypeNone is [insert doc].
	StickySessionsTypeNone = StickySessionsType("none")
	// StickySessionsTypeCookie is [insert doc].
	StickySessionsTypeCookie = StickySessionsType("cookie")
	// StickySessionsTypeTable is [insert doc].
	StickySessionsTypeTable = StickySessionsType("table")
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

// ACL: the use of Access Control Lists (ACL) provide a flexible solution to perform a action generally consist in blocking or allow a request based on ip (and URL on HTTP)
type ACL struct {
	// ID: ID of your ACL ressource
	ID string `json:"id"`
	// Name: name of you ACL ressource
	Name string `json:"name"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Frontend: see the Frontend object description
	Frontend *Frontend `json:"frontend"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
	// CreatedAt: date at which the ACL was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date at which the ACL was last updated
	UpdatedAt *time.Time `json:"updated_at"`
}

// ACLAction: acl action
type ACLAction struct {
	// Type: the action type
	//
	// Default value: allow
	Type ACLActionType `json:"type"`
}

// ACLMatch: acl match
type ACLMatch struct {
	// IPSubnet: a list of IPs or CIDR v4/v6 addresses of the client of the session to match
	IPSubnet []*string `json:"ip_subnet"`
	// HTTPFilter: the HTTP filter to match
	//
	// The HTTP filter to match. This filter is supported only if your backend supports HTTP forwarding.
	// It extracts the request's URL path, which starts at the first slash and ends before the question mark (without the host part).
	//
	// Default value: acl_http_filter_none
	HTTPFilter ACLHTTPFilter `json:"http_filter"`
	// HTTPFilterValue: a list of possible values to match for the given HTTP filter
	HTTPFilterValue []*string `json:"http_filter_value"`
	// HTTPFilterOption: a exra parameter. You can use this field with http_header_match acl type to set the header name to filter
	HTTPFilterOption *string `json:"http_filter_option"`
	// Invert: if set to `true`, the ACL matching condition will be of type "UNLESS"
	Invert bool `json:"invert"`
}

// ACLSpec: acl spec
type ACLSpec struct {
	// Name: name of your ACL resource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// Backend: backend
type Backend struct {
	ID string `json:"id"`

	Name string `json:"name"`
	// ForwardProtocol:
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`

	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm:
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions:
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`

	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`

	HealthCheck *HealthCheck `json:"health_check"`

	Pool []string `json:"pool"`

	LB *LB `json:"lb"`
	// Deprecated
	SendProxyV2 bool `json:"send_proxy_v2"`

	TimeoutServer *time.Duration `json:"timeout_server"`

	TimeoutConnect *time.Duration `json:"timeout_connect"`

	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction:
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol:
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	FailoverHost *string `json:"failover_host"`
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

// BackendServerStats: state and statistics of your backend server like last healthcheck status, server uptime, result state of your backend server
type BackendServerStats struct {
	// InstanceID: ID of your loadbalancer cluster server
	InstanceID string `json:"instance_id"`
	// BackendID: ID of your Backend
	BackendID string `json:"backend_id"`
	// IP: iPv4 or IPv6 address of the server backend
	IP string `json:"ip"`
	// ServerState: server operational state (stopped/starting/running/stopping)
	//
	// Default value: stopped
	ServerState BackendServerStatsServerState `json:"server_state"`
	// ServerStateChangedAt: time since last operational change
	ServerStateChangedAt *time.Time `json:"server_state_changed_at"`
	// LastHealthCheckStatus: last health check status (unknown/neutral/failed/passed/condpass)
	//
	// Default value: unknown
	LastHealthCheckStatus BackendServerStatsHealthCheckStatus `json:"last_health_check_status"`
}

// Certificate: sSL certificate
type Certificate struct {
	// Type: type of certificate (Let's encrypt or custom)
	//
	// Default value: letsencryt
	Type CertificateType `json:"type"`
	// ID: certificate ID
	ID string `json:"id"`
	// CommonName: main domain name of certificate
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names
	SubjectAlternativeName []string `json:"subject_alternative_name"`
	// Fingerprint: identifier (SHA-1) of the certificate
	Fingerprint string `json:"fingerprint"`
	// NotValidBefore: validity bounds
	NotValidBefore *time.Time `json:"not_valid_before"`
	// NotValidAfter: validity bounds
	NotValidAfter *time.Time `json:"not_valid_after"`
	// Status: status of certificate
	//
	// Default value: pending
	Status CertificateStatus `json:"status"`
	// LB: load balancer object
	LB *LB `json:"lb"`
	// Name: certificate name
	Name string `json:"name"`
	// CreatedAt: date at which the certificate was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date at which the certificate was last updated
	UpdatedAt *time.Time `json:"updated_at"`
}

// CreateCertificateRequestCustomCertificate: import a custom SSL certificate
type CreateCertificateRequestCustomCertificate struct {
	// CertificateChain: the full PEM-formatted include an entire certificate chain including public key, private key, and optionally certificate authorities.
	CertificateChain string `json:"certificate_chain"`
}

// CreateCertificateRequestLetsencryptConfig: generate a new SSL certificate using Let's Encrypt.
type CreateCertificateRequestLetsencryptConfig struct {
	// CommonName: main domain name of certificate (make sure this domain exists and resolves to your load balancer HA IP)
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names (make sure all domain names exists and resolves to your load balancer HA IP)
	SubjectAlternativeName []string `json:"subject_alternative_name"`
}

// Frontend: frontend
type Frontend struct {
	ID string `json:"id"`

	Name string `json:"name"`

	InboundPort int32 `json:"inbound_port"`

	Backend *Backend `json:"backend"`

	LB *LB `json:"lb"`

	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated
	Certificate *Certificate `json:"certificate"`

	CertificateIDs []string `json:"certificate_ids"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`
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

// HealthCheck: health check
type HealthCheck struct {
	// MysqlConfig: the check requires MySQL >=3.22, for older versions, use TCP check
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// LdapConfig: the response is analyzed to find an LDAPv3 response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: the response is analyzed to find the +PONG response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	CheckMaxRetries int32 `json:"check_max_retries"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`

	Port int32 `json:"port"`

	CheckTimeout *time.Duration `json:"check_timeout"`

	CheckDelay *time.Duration `json:"check_delay"`
	// CheckSendProxy: it defines whether the healthcheck should be done considering the proxy protocol
	CheckSendProxy bool `json:"check_send_proxy"`
}

func (m *HealthCheck) UnmarshalJSON(b []byte) error {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = HealthCheck(tmp.tmpType)

	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	return nil
}

func (m HealthCheck) MarshalJSON() ([]byte, error) {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
	}{
		tmpType: tmpType(m),

		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
	}
	return json.Marshal(tmp)
}

type HealthCheckHTTPConfig struct {
	URI string `json:"uri"`

	Method string `json:"method"`

	Code *int32 `json:"code"`
}

type HealthCheckHTTPSConfig struct {
	URI string `json:"uri"`

	Method string `json:"method"`

	Code *int32 `json:"code"`
}

type HealthCheckLdapConfig struct {
}

type HealthCheckMysqlConfig struct {
	User string `json:"user"`
}

type HealthCheckPgsqlConfig struct {
	User string `json:"user"`
}

type HealthCheckRedisConfig struct {
}

type HealthCheckTCPConfig struct {
}

// IP: ip
type IP struct {
	ID string `json:"id"`

	IPAddress string `json:"ip_address"`

	OrganizationID string `json:"organization_id"`

	LBID *string `json:"lb_id"`

	Reverse string `json:"reverse"`

	ProjectID string `json:"project_id"`
	// Deprecated
	Region scw.Region `json:"region"`

	Zone scw.Zone `json:"zone"`
}

type Instance struct {
	ID string `json:"id"`
	// Status:
	//
	// Default value: unknown
	Status InstanceStatus `json:"status"`

	IPAddress string `json:"ip_address"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`
	// Deprecated
	Region scw.Region `json:"region"`

	Zone scw.Zone `json:"zone"`
}

// LB: lb
type LB struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Description string `json:"description"`
	// Status:
	//
	// Default value: unknown
	Status LBStatus `json:"status"`

	Instances []*Instance `json:"instances"`

	OrganizationID string `json:"organization_id"`

	IP []*IP `json:"ip"`

	Tags []string `json:"tags"`

	FrontendCount int32 `json:"frontend_count"`

	BackendCount int32 `json:"backend_count"`

	Type string `json:"type"`

	Subscriber *Subscriber `json:"subscriber"`
	// SslCompatibilityLevel:
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`

	ProjectID string `json:"project_id"`

	CreatedAt *time.Time `json:"created_at"`

	UpdatedAt *time.Time `json:"updated_at"`

	PrivateNetworkCount int32 `json:"private_network_count"`

	RouteCount int32 `json:"route_count"`
	// Deprecated
	Region scw.Region `json:"region"`

	Zone scw.Zone `json:"zone"`
}

// LBStats: lb stats
type LBStats struct {
	// BackendServersStats: list stats object of your loadbalancer
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
}

type LBType struct {
	Name string `json:"name"`
	// StockStatus:
	//
	// Default value: unknown
	StockStatus LBTypeStock `json:"stock_status"`

	Description string `json:"description"`
	// Deprecated
	Region scw.Region `json:"region"`

	Zone scw.Zone `json:"zone"`
}

// ListACLResponse: list acl response
type ListACLResponse struct {
	// ACLs: list of Acl object (see Acl object description)
	ACLs []*ACL `json:"acls"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// ListBackendStatsResponse: list backend stats response
type ListBackendStatsResponse struct {
	// BackendServersStats: list backend stats object of your loadbalancer
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// ListBackendsResponse: list backends response
type ListBackendsResponse struct {
	// Backends: list Backend objects of a load balancer
	Backends []*Backend `json:"backends"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

type ListCertificatesResponse struct {
	Certificates []*Certificate `json:"certificates"`

	TotalCount uint32 `json:"total_count"`
}

// ListFrontendsResponse: list frontends response
type ListFrontendsResponse struct {
	// Frontends: list frontends object of your loadbalancer
	Frontends []*Frontend `json:"frontends"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

// ListIPsResponse: list ips response
type ListIPsResponse struct {
	// IPs: list IP address object
	IPs []*IP `json:"ips"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

type ListLBPrivateNetworksResponse struct {
	PrivateNetwork []*PrivateNetwork `json:"private_network"`

	TotalCount uint32 `json:"total_count"`
}

type ListLBTypesResponse struct {
	LBTypes []*LBType `json:"lb_types"`

	TotalCount uint32 `json:"total_count"`
}

// ListLBsResponse: list lbs response
type ListLBsResponse struct {
	LBs []*LB `json:"lbs"`

	TotalCount uint32 `json:"total_count"`
}

// ListRoutesResponse: list routes response
type ListRoutesResponse struct {
	// Routes: list of Routes object
	Routes []*Route `json:"routes"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// ListSubscriberResponse: list subscriber response
type ListSubscriberResponse struct {
	// Subscribers: list of Subscribers object
	Subscribers []*Subscriber `json:"subscribers"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// PrivateNetwork: private network
type PrivateNetwork struct {
	// LB: loadBalancer object
	LB *LB `json:"lb"`
	// StaticConfig: local ip address of load balancer instance
	// Precisely one of DHCPConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: value set to true if load balancer instance use a DHCP
	// Precisely one of DHCPConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
	// PrivateNetworkID: instance private network id
	PrivateNetworkID string `json:"private_network_id"`
	// Status: status (running, to create...) of private network connection
	//
	// Default value: unknown
	Status PrivateNetworkStatus `json:"status"`
	// CreatedAt: date at which the PN was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date at which the PN was last updated
	UpdatedAt *time.Time `json:"updated_at"`
}

type PrivateNetworkDHCPConfig struct {
}

type PrivateNetworkStaticConfig struct {
	IPAddress []string `json:"ip_address"`
}

// Route: route
type Route struct {
	// ID: id of match ressource
	ID string `json:"id"`
	// FrontendID: id of frontend
	FrontendID string `json:"frontend_id"`
	// BackendID: id of backend
	BackendID string `json:"backend_id"`
	// Match: value to match a redirection
	Match *RouteMatch `json:"match"`
	// CreatedAt: date at which the route was created
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date at which the route was last updated
	UpdatedAt *time.Time `json:"updated_at"`
}

// RouteMatch: route. match
type RouteMatch struct {
	// Sni: server Name Indication TLS extension (SNI)
	//
	// Server Name Indication TLS extension (SNI) field from an incoming connection made via an SSL/TLS transport layer
	Sni *string `json:"sni"`
}

// SetACLsResponse: set acls response
type SetACLsResponse struct {
	// ACLs: list of ACLs object (see ACL object description)
	ACLs []*ACL `json:"acls"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// Subscriber: subscriber
type Subscriber struct {
	// ID: subscriber ID
	ID string `json:"id"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address of subscriber
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI of subscriber
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// SubscriberEmailConfig: email alert of subscriber
type SubscriberEmailConfig struct {
	// Email: email who receive alert
	Email string `json:"email"`
}

// SubscriberWebhookConfig: webhook alert of subscriber
type SubscriberWebhookConfig struct {
	// URI: URI who receive POST request
	URI string `json:"uri"`
}

// Service API

type GetServiceInfoRequest struct {
	Region scw.Region `json:"-"`
}

func (s *API) GetServiceInfo(req *GetServiceInfoRequest, opts ...scw.RequestOption) (*scw.ServiceInfo, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "",
		Headers: http.Header{},
	}

	var resp scw.ServiceInfo

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListLBsRequest struct {
	Region scw.Region `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListLBsRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
	// OrganizationID: filter LBs by organization ID
	OrganizationID *string `json:"-"`
	// ProjectID: filter LBs by project ID
	ProjectID *string `json:"-"`
}

// ListLBs: list load balancers
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
	Region scw.Region `json:"-"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Name: resource names
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// IPID: just like for compute instances, when you destroy a load balancer, you can keep its highly available IP address and reuse it for another load balancer later
	IPID *string `json:"ip_id"`
	// Tags: list of keyword
	Tags []string `json:"tags"`
	// Type: load balancer offer type
	Type string `json:"type"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// CreateLB: create a load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// GetLB: get a load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// Tags: list of keywords
	Tags []string `json:"tags"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// UpdateLB: update a load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// ReleaseIP: set true if you don't want to keep this IP address
	ReleaseIP bool `json:"-"`
}

// DeleteLB: delete a load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Type: load balancer type (check /lb-types to list all type)
	Type string `json:"type"`
}

// MigrateLB: migrate a load balancer
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
	Region scw.Region `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// IPAddress: use this to search by IP address
	IPAddress *string `json:"-"`
	// OrganizationID: filter IPs by organization id
	OrganizationID *string `json:"-"`
	// ProjectID: filter IPs by project ID
	ProjectID *string `json:"-"`
}

// ListIPs: list IPs
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
	Region scw.Region `json:"-"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Reverse: reverse domain name
	Reverse *string `json:"reverse"`
}

// CreateIP: create an IP
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
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// GetIP: get an IP
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
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// ReleaseIP: delete an IP
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
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
	// Reverse: reverse DNS
	Reverse *string `json:"reverse"`
}

// UpdateIP: update an IP
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: choose order of response
	//
	// Default value: created_at_asc
	OrderBy ListBackendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

// ListBackends: list backends in a given load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// ForwardProtocol: backend protocol. TCP or HTTP
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: user sessions will be forwarded to this port of backend servers
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: enables cookie-based session persistence
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for for sticky sessions
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: see the Healthcheck object description
	HealthCheck *HealthCheck `json:"health_check"`
	// ServerIP: backend server IP addresses list (IPv4 or IPv6)
	ServerIP []string `json:"server_ip"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field !
	SendProxyV2 bool `json:"send_proxy_v2"`
	// TimeoutServer: maximum server connection inactivity time
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum initical server connection establishment time
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum tunnel inactivity time
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: modify what occurs when a backend server is marked down
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: pROXY protocol, forward client's address (must be supported by backend servers software)
	//
	// The PROXY protocol informs the other end about the incoming connection, so that it can know the client's address or the public address it accessed to, whatever the upper layer protocol.
	//
	// * `proxy_protocol_none` Disable proxy protocol.
	// * `proxy_protocol_v1` Version one (text format).
	// * `proxy_protocol_v2` Version two (binary format).
	// * `proxy_protocol_v2_ssl` Version two with SSL connection.
	// * `proxy_protocol_v2_ssl_cn` Version two with SSL connection and common name information.
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served in case all backend servers are down
	//
	// Only the host part of the Scaleway S3 bucket website is expected e.g. `failover-website.s3-website.fr-par.scw.cloud` if your bucket website URL is `https://failover-website.s3-website.fr-par.scw.cloud/`.
	//
	FailoverHost *string `json:"failover_host"`
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

// CreateBackend: create a backend in a given load balancer
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
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
}

// GetBackend: get a backend in a given load balancer
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
	Region scw.Region `json:"-"`

	BackendID string `json:"-"`

	Name string `json:"name"`
	// ForwardProtocol:
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`

	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm:
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions:
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`

	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// Deprecated
	SendProxyV2 bool `json:"send_proxy_v2"`

	TimeoutServer *time.Duration `json:"timeout_server"`

	TimeoutConnect *time.Duration `json:"timeout_connect"`

	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction:
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol:
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	FailoverHost *string `json:"failover_host"`
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

// UpdateBackend: update a backend in a given load balancer
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
	Region scw.Region `json:"-"`
	// BackendID: ID of the backend to delete
	BackendID string `json:"-"`
}

// DeleteBackend: delete a backend in a given load balancer
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
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend
	ServerIP []string `json:"server_ip"`
}

// AddBackendServers: add a set of servers in a given backend
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
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to remove of your backend
	ServerIP []string `json:"server_ip"`
}

// RemoveBackendServers: remove a set of servers for a given backend
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
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend and remove all other
	ServerIP []string `json:"server_ip"`
}

// SetBackendServers: define all servers in a given backend
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
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// Port: specify the port used to health check
	Port int32 `json:"port"`
	// CheckDelay: time between two consecutive health checks
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: additional check timeout, after the connection has been already established
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks, after wich the server will be considered dead
	CheckMaxRetries int32 `json:"check_max_retries"`
	// MysqlConfig: the check requires MySQL >=3.22, for older version, please use TCP check
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// LdapConfig: the response is analyzed to find an LDAPv3 response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: the response is analyzed to find the +PONG response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
	// CheckSendProxy: it defines whether the healthcheck should be done considering the proxy protocol
	CheckSendProxy bool `json:"check_send_proxy"`
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

// UpdateHealthCheck: update an healthcheck for a given backend
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: response order
	//
	// Default value: created_at_asc
	OrderBy ListFrontendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

// ListFrontends: list frontends in a given load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: set the maximum inactivity time on the client side
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array !
	CertificateID *string `json:"certificate_id"`
	// CertificateIDs: list of certificate IDs to bind on the frontend
	CertificateIDs *[]string `json:"certificate_ids"`
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

// CreateFrontend: create a frontend in a given load balancer
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
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
}

// GetFrontend: get a frontend
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
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: client session maximum inactivity time
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of `certificate_ids` array!
	CertificateID *string `json:"certificate_id"`
	// CertificateIDs: list of certificate IDs to bind on the frontend
	CertificateIDs *[]string `json:"certificate_ids"`
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

// UpdateFrontend: update a frontend
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
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID to delete
	FrontendID string `json:"-"`
}

// DeleteFrontend: delete a frontend
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
	Region scw.Region `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListRoutesRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`

	FrontendID *string `json:"-"`
}

// ListRoutes: list all backend redirections
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
	Region scw.Region `json:"-"`
	// FrontendID: origin of redirection
	FrontendID string `json:"frontend_id"`
	// BackendID: destination of destination
	BackendID string `json:"backend_id"`
	// Match: value to match a redirection
	Match *RouteMatch `json:"match"`
}

// CreateRoute: create a backend redirection
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
	Region scw.Region `json:"-"`
	// RouteID: id of route to get
	RouteID string `json:"-"`
}

// GetRoute: get single backend redirection
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
	Region scw.Region `json:"-"`
	// RouteID: route id to update
	RouteID string `json:"-"`
	// BackendID: backend id of redirection
	BackendID string `json:"backend_id"`
	// Match: value to match a redirection
	Match *RouteMatch `json:"match"`
}

// UpdateRoute: edit a backend redirection
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
	Region scw.Region `json:"-"`
	// RouteID: route id to delete
	RouteID string `json:"-"`
}

// DeleteRoute: delete a backend redirection
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// Deprecated: GetLBStats: get usage statistics of a given load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
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
	Region scw.Region `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListACLRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: filter acl per name
	Name *string `json:"-"`
}

// ListACLs: list ACL for a given frontend
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
	Region scw.Region `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule
	//
	// The ACL match rule. You can have one of those three cases:
	//
	//   - `ip_subnet` is defined
	//   - `http_filter` and `http_filter_value` are defined
	//   - `ip_subnet`, `http_filter` and `http_filter_value` are defined
	//
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// CreateACL: create an ACL for a given frontend
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
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

// GetACL: get an ACL
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
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// UpdateACL: update an ACL
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
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

// DeleteACL: delete an ACL
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
	// Letsencrypt: let's Encrypt type
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`
	// CustomCertificate: custom import certificate
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateCertificate: create a TLS certificate
//
// Generate a new TLS certificate using Let's Encrypt or import your certificate.
func (s *API) CreateCertificate(req *CreateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("certiticate")
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListCertificatesRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
}

// ListCertificates: list all TLS certificates on a given load balancer
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
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// GetCertificate: get a TLS certificate
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
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
}

// UpdateCertificate: update a TLS certificate
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
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// DeleteCertificate: delete a TLS certificate
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
	Region scw.Region `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
}

// ListLBTypes: list all load balancer offer type
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
	Region scw.Region `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// CreateSubscriber: create a subscriber, webhook or email
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
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// GetSubscriber: get a subscriber
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
	Region scw.Region `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListSubscriberRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrganizationID: filter Subscribers by organization ID
	OrganizationID *string `json:"-"`
	// ProjectID: filter Subscribers by project ID
	ProjectID *string `json:"-"`
}

// ListSubscriber: list all subscriber
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
	Region scw.Region `json:"-"`
	// SubscriberID: assign the resource to a project IDs
	SubscriberID string `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// UpdateSubscriber: update a subscriber
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
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// DeleteSubscriber: delete a subscriber
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"subscriber_id"`
}

// SubscribeToLB: subscribe a subscriber to a given load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// UnsubscribeFromLB: unsubscribe a subscriber from a given load balancer
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
	Region scw.Region `json:"-"`

	LBID string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListLBPrivateNetworks: list attached private network of load balancer
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id
	PrivateNetworkID string `json:"-"`
	// StaticConfig: define two local ip address of your choice for each load balancer instance
	// Precisely one of DHCPConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: set to true if you want to let DHCP assign IP addresses
	// Precisely one of DHCPConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
}

// AttachPrivateNetwork: add load balancer on instance private network
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
	Region scw.Region `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id
	PrivateNetworkID string `json:"-"`
}

// DetachPrivateNetwork: remove load balancer of private network
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

// Service ZonedAPI

type ZonedAPIGetServiceInfoRequest struct {
	Zone scw.Zone `json:"-"`
}

func (s *ZonedAPI) GetServiceInfo(req *ZonedAPIGetServiceInfoRequest, opts ...scw.RequestOption) (*scw.ServiceInfo, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "",
		Headers: http.Header{},
	}

	var resp scw.ServiceInfo

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ZonedAPIListLBsRequest struct {
	Zone scw.Zone `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListLBsRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
	// OrganizationID: filter LBs by organization ID
	OrganizationID *string `json:"-"`
	// ProjectID: filter LBs by project ID
	ProjectID *string `json:"-"`
}

// ListLBs: list load balancers
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
	Zone scw.Zone `json:"-"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Name: resource names
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// IPID: just like for compute instances, when you destroy a load balancer, you can keep its highly available IP address and reuse it for another load balancer later
	IPID *string `json:"ip_id"`
	// Tags: list of keyword
	Tags []string `json:"tags"`
	// Type: load balancer offer type
	Type string `json:"type"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// CreateLB: create a load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// GetLB: get a load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// Tags: list of keywords
	Tags []string `json:"tags"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// UpdateLB: update a load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// ReleaseIP: set true if you don't want to keep this IP address
	ReleaseIP bool `json:"-"`
}

// DeleteLB: delete a load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Type: load balancer type (check /lb-types to list all type)
	Type string `json:"type"`
}

// MigrateLB: migrate a load balancer
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
	Zone scw.Zone `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// IPAddress: use this to search by IP address
	IPAddress *string `json:"-"`
	// OrganizationID: filter IPs by organization id
	OrganizationID *string `json:"-"`
	// ProjectID: filter IPs by project ID
	ProjectID *string `json:"-"`
}

// ListIPs: list IPs
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
	Zone scw.Zone `json:"-"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Reverse: reverse domain name
	Reverse *string `json:"reverse"`
}

// CreateIP: create an IP
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
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// GetIP: get an IP
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
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// ReleaseIP: delete an IP
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
	Zone scw.Zone `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
	// Reverse: reverse DNS
	Reverse *string `json:"reverse"`
}

// UpdateIP: update an IP
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: choose order of response
	//
	// Default value: created_at_asc
	OrderBy ListBackendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

// ListBackends: list backends in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// ForwardProtocol: backend protocol. TCP or HTTP
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: user sessions will be forwarded to this port of backend servers
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: enables cookie-based session persistence
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for for sticky sessions
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: see the Healthcheck object description
	HealthCheck *HealthCheck `json:"health_check"`
	// ServerIP: backend server IP addresses list (IPv4 or IPv6)
	ServerIP []string `json:"server_ip"`
	// Deprecated: SendProxyV2: deprecated in favor of proxy_protocol field !
	SendProxyV2 bool `json:"send_proxy_v2"`
	// TimeoutServer: maximum server connection inactivity time
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum initical server connection establishment time
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum tunnel inactivity time
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: modify what occurs when a backend server is marked down
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: pROXY protocol, forward client's address (must be supported by backend servers software)
	//
	// The PROXY protocol informs the other end about the incoming connection, so that it can know the client's address or the public address it accessed to, whatever the upper layer protocol.
	//
	// * `proxy_protocol_none` Disable proxy protocol.
	// * `proxy_protocol_v1` Version one (text format).
	// * `proxy_protocol_v2` Version two (binary format).
	// * `proxy_protocol_v2_ssl` Version two with SSL connection.
	// * `proxy_protocol_v2_ssl_cn` Version two with SSL connection and common name information.
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
	// FailoverHost: scaleway S3 bucket website to be served in case all backend servers are down
	//
	// Only the host part of the Scaleway S3 bucket website is expected e.g. `failover-website.s3-website.fr-par.scw.cloud` if your bucket website URL is `https://failover-website.s3-website.fr-par.scw.cloud/`.
	//
	FailoverHost *string `json:"failover_host"`
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

// CreateBackend: create a backend in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
}

// GetBackend: get a backend in a given load balancer
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
	Zone scw.Zone `json:"-"`

	BackendID string `json:"-"`

	Name string `json:"name"`
	// ForwardProtocol:
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`

	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm:
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions:
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`

	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// Deprecated
	SendProxyV2 bool `json:"send_proxy_v2"`

	TimeoutServer *time.Duration `json:"timeout_server"`

	TimeoutConnect *time.Duration `json:"timeout_connect"`

	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction:
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol:
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	FailoverHost *string `json:"failover_host"`
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

// UpdateBackend: update a backend in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// BackendID: ID of the backend to delete
	BackendID string `json:"-"`
}

// DeleteBackend: delete a backend in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend
	ServerIP []string `json:"server_ip"`
}

// AddBackendServers: add a set of servers in a given backend
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
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to remove of your backend
	ServerIP []string `json:"server_ip"`
}

// RemoveBackendServers: remove a set of servers for a given backend
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
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend and remove all other
	ServerIP []string `json:"server_ip"`
}

// SetBackendServers: define all servers in a given backend
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
	Zone scw.Zone `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// Port: specify the port used to health check
	Port int32 `json:"port"`
	// CheckDelay: time between two consecutive health checks
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: additional check timeout, after the connection has been already established
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks, after wich the server will be considered dead
	CheckMaxRetries int32 `json:"check_max_retries"`
	// MysqlConfig: the check requires MySQL >=3.22, for older version, please use TCP check
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// LdapConfig: the response is analyzed to find an LDAPv3 response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: the response is analyzed to find the +PONG response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
	// CheckSendProxy: it defines whether the healthcheck should be done considering the proxy protocol
	CheckSendProxy bool `json:"check_send_proxy"`
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

// UpdateHealthCheck: update an healthcheck for a given backend
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: response order
	//
	// Default value: created_at_asc
	OrderBy ListFrontendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

// ListFrontends: list frontends in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: set the maximum inactivity time on the client side
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array !
	CertificateID *string `json:"certificate_id"`
	// CertificateIDs: list of certificate IDs to bind on the frontend
	CertificateIDs *[]string `json:"certificate_ids"`
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

// CreateFrontend: create a frontend in a given load balancer
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
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
}

// GetFrontend: get a frontend
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
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: client session maximum inactivity time
	TimeoutClient *time.Duration `json:"timeout_client"`
	// Deprecated: CertificateID: certificate ID, deprecated in favor of `certificate_ids` array!
	CertificateID *string `json:"certificate_id"`
	// CertificateIDs: list of certificate IDs to bind on the frontend
	CertificateIDs *[]string `json:"certificate_ids"`
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

// UpdateFrontend: update a frontend
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
	Zone scw.Zone `json:"-"`
	// FrontendID: frontend ID to delete
	FrontendID string `json:"-"`
}

// DeleteFrontend: delete a frontend
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
	Zone scw.Zone `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListRoutesRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`

	FrontendID *string `json:"-"`
}

// ListRoutes: list all backend redirections
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
	Zone scw.Zone `json:"-"`
	// FrontendID: origin of redirection
	FrontendID string `json:"frontend_id"`
	// BackendID: destination of destination
	BackendID string `json:"backend_id"`
	// Match: value to match a redirection
	Match *RouteMatch `json:"match"`
}

// CreateRoute: create a backend redirection
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
	Zone scw.Zone `json:"-"`
	// RouteID: id of route to get
	RouteID string `json:"-"`
}

// GetRoute: get single backend redirection
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
	Zone scw.Zone `json:"-"`
	// RouteID: route id to update
	RouteID string `json:"-"`
	// BackendID: backend id of redirection
	BackendID string `json:"backend_id"`
	// Match: value to match a redirection
	Match *RouteMatch `json:"match"`
}

// UpdateRoute: edit a backend redirection
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
	Zone scw.Zone `json:"-"`
	// RouteID: route id to delete
	RouteID string `json:"-"`
}

// DeleteRoute: delete a backend redirection
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// Deprecated: GetLBStats: get usage statistics of a given load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
}

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
	Zone scw.Zone `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListACLRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: filter acl per name
	Name *string `json:"-"`
}

// ListACLs: list ACL for a given frontend
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
	Zone scw.Zone `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule
	//
	// The ACL match rule. You can have one of those three cases:
	//
	//   - `ip_subnet` is defined
	//   - `http_filter` and `http_filter_value` are defined
	//   - `ip_subnet`, `http_filter` and `http_filter_value` are defined
	//
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// CreateACL: create an ACL for a given frontend
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
	Zone scw.Zone `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

// GetACL: get an ACL
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
	Zone scw.Zone `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// UpdateACL: update an ACL
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
	Zone scw.Zone `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

// DeleteACL: delete an ACL
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
	Zone scw.Zone `json:"-"`
	// FrontendID: the Frontend to change ACL to
	FrontendID string `json:"-"`
	// ACLs: array of ACLs to erease the existing ACLs
	ACLs []*ACLSpec `json:"acls"`
}

// SetACLs: set all ACLs for a given frontend
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
	// Letsencrypt: let's Encrypt type
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`
	// CustomCertificate: custom import certificate
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateCertificate: create a TLS certificate
//
// Generate a new TLS certificate using Let's Encrypt or import your certificate.
func (s *ZonedAPI) CreateCertificate(req *ZonedAPICreateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("certiticate")
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListCertificatesRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
}

// ListCertificates: list all TLS certificates on a given load balancer
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
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// GetCertificate: get a TLS certificate
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
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
}

// UpdateCertificate: update a TLS certificate
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
	Zone scw.Zone `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// DeleteCertificate: delete a TLS certificate
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
	Zone scw.Zone `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
}

// ListLBTypes: list all load balancer offer type
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
	Zone scw.Zone `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
	// Deprecated: OrganizationID: owner of resources
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: assign the resource to a project ID
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// CreateSubscriber: create a subscriber, webhook or email
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
	Zone scw.Zone `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// GetSubscriber: get a subscriber
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
	Zone scw.Zone `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListSubscriberRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrganizationID: filter Subscribers by organization ID
	OrganizationID *string `json:"-"`
	// ProjectID: filter Subscribers by project ID
	ProjectID *string `json:"-"`
}

// ListSubscriber: list all subscriber
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
	Zone scw.Zone `json:"-"`
	// SubscriberID: assign the resource to a project IDs
	SubscriberID string `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// UpdateSubscriber: update a subscriber
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
	Zone scw.Zone `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// DeleteSubscriber: delete a subscriber
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"subscriber_id"`
}

// SubscribeToLB: subscribe a subscriber to a given load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
}

// UnsubscribeFromLB: unsubscribe a subscriber from a given load balancer
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
	Zone scw.Zone `json:"-"`

	LBID string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListLBPrivateNetworks: list attached private network of load balancer
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id
	PrivateNetworkID string `json:"-"`
	// StaticConfig: define two local ip address of your choice for each load balancer instance
	// Precisely one of DHCPConfig, StaticConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`
	// DHCPConfig: set to true if you want to let DHCP assign IP addresses
	// Precisely one of DHCPConfig, StaticConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`
}

// AttachPrivateNetwork: add load balancer on instance private network
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
	Zone scw.Zone `json:"-"`
	// LBID: load balancer ID
	LBID string `json:"-"`
	// PrivateNetworkID: set your instance private network id
	PrivateNetworkID string `json:"-"`
}

// DetachPrivateNetwork: remove load balancer of private network
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
