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

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/async"
	"github.com/scaleway/scaleway-sdk-go/marshaler"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/parameter"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultLBRetryInterval = 15 * time.Second
	defaultLBTimeout       = 5 * time.Minute
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

type ACLActionRedirectRedirectType string

const (
	ACLActionRedirectRedirectTypeLocation = ACLActionRedirectRedirectType("location")
	ACLActionRedirectRedirectTypeScheme   = ACLActionRedirectRedirectType("scheme")
)

func (enum ACLActionRedirectRedirectType) String() string {
	if enum == "" {
		// return default value if empty
		return string(ACLActionRedirectRedirectTypeLocation)
	}
	return string(enum)
}

func (enum ACLActionRedirectRedirectType) Values() []ACLActionRedirectRedirectType {
	return []ACLActionRedirectRedirectType{
		"location",
		"scheme",
	}
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
		return string(ACLActionTypeAllow)
	}
	return string(enum)
}

func (enum ACLActionType) Values() []ACLActionType {
	return []ACLActionType{
		"allow",
		"deny",
		"redirect",
	}
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
		return string(ACLHTTPFilterACLHTTPFilterNone)
	}
	return string(enum)
}

func (enum ACLHTTPFilter) Values() []ACLHTTPFilter {
	return []ACLHTTPFilter{
		"acl_http_filter_none",
		"path_begin",
		"path_end",
		"regex",
		"http_header_match",
	}
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
		return string(BackendServerStatsHealthCheckStatusUnknown)
	}
	return string(enum)
}

func (enum BackendServerStatsHealthCheckStatus) Values() []BackendServerStatsHealthCheckStatus {
	return []BackendServerStatsHealthCheckStatus{
		"unknown",
		"neutral",
		"failed",
		"passed",
		"condpass",
	}
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
		return string(BackendServerStatsServerStateStopped)
	}
	return string(enum)
}

func (enum BackendServerStatsServerState) Values() []BackendServerStatsServerState {
	return []BackendServerStatsServerState{
		"stopped",
		"starting",
		"running",
		"stopping",
	}
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
		return string(CertificateStatusPending)
	}
	return string(enum)
}

func (enum CertificateStatus) Values() []CertificateStatus {
	return []CertificateStatus{
		"pending",
		"ready",
		"error",
	}
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
		return string(CertificateTypeLetsencryt)
	}
	return string(enum)
}

func (enum CertificateType) Values() []CertificateType {
	return []CertificateType{
		"letsencryt",
		"custom",
	}
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
		return string(ForwardPortAlgorithmRoundrobin)
	}
	return string(enum)
}

func (enum ForwardPortAlgorithm) Values() []ForwardPortAlgorithm {
	return []ForwardPortAlgorithm{
		"roundrobin",
		"leastconn",
		"first",
	}
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
		return string(InstanceStatusUnknown)
	}
	return string(enum)
}

func (enum InstanceStatus) Values() []InstanceStatus {
	return []InstanceStatus{
		"unknown",
		"ready",
		"pending",
		"stopped",
		"error",
		"locked",
		"migrating",
	}
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
		return string(LBStatusUnknown)
	}
	return string(enum)
}

func (enum LBStatus) Values() []LBStatus {
	return []LBStatus{
		"unknown",
		"ready",
		"pending",
		"stopped",
		"error",
		"locked",
		"migrating",
		"to_create",
		"creating",
		"to_delete",
		"deleting",
	}
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
		return string(LBTypeStockUnknown)
	}
	return string(enum)
}

func (enum LBTypeStock) Values() []LBTypeStock {
	return []LBTypeStock{
		"unknown",
		"low_stock",
		"out_of_stock",
		"available",
	}
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
		return string(ListACLRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListACLRequestOrderBy) Values() []ListACLRequestOrderBy {
	return []ListACLRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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
		return string(ListBackendsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListBackendsRequestOrderBy) Values() []ListBackendsRequestOrderBy {
	return []ListBackendsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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
		return string(ListCertificatesRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListCertificatesRequestOrderBy) Values() []ListCertificatesRequestOrderBy {
	return []ListCertificatesRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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
		return string(ListFrontendsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListFrontendsRequestOrderBy) Values() []ListFrontendsRequestOrderBy {
	return []ListFrontendsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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

type ListIPsRequestIPType string

const (
	ListIPsRequestIPTypeAll  = ListIPsRequestIPType("all")
	ListIPsRequestIPTypeIPv4 = ListIPsRequestIPType("ipv4")
	ListIPsRequestIPTypeIPv6 = ListIPsRequestIPType("ipv6")
)

func (enum ListIPsRequestIPType) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListIPsRequestIPTypeAll)
	}
	return string(enum)
}

func (enum ListIPsRequestIPType) Values() []ListIPsRequestIPType {
	return []ListIPsRequestIPType{
		"all",
		"ipv4",
		"ipv6",
	}
}

func (enum ListIPsRequestIPType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListIPsRequestIPType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListIPsRequestIPType(ListIPsRequestIPType(tmp).String())
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
		return string(ListLBsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListLBsRequestOrderBy) Values() []ListLBsRequestOrderBy {
	return []ListLBsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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
		return string(ListPrivateNetworksRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListPrivateNetworksRequestOrderBy) Values() []ListPrivateNetworksRequestOrderBy {
	return []ListPrivateNetworksRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
	}
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
		return string(ListRoutesRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListRoutesRequestOrderBy) Values() []ListRoutesRequestOrderBy {
	return []ListRoutesRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
	}
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
		return string(ListSubscriberRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListSubscriberRequestOrderBy) Values() []ListSubscriberRequestOrderBy {
	return []ListSubscriberRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
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
		return string(OnMarkedDownActionOnMarkedDownActionNone)
	}
	return string(enum)
}

func (enum OnMarkedDownAction) Values() []OnMarkedDownAction {
	return []OnMarkedDownAction{
		"on_marked_down_action_none",
		"shutdown_sessions",
	}
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
		return string(PrivateNetworkStatusUnknown)
	}
	return string(enum)
}

func (enum PrivateNetworkStatus) Values() []PrivateNetworkStatus {
	return []PrivateNetworkStatus{
		"unknown",
		"ready",
		"pending",
		"error",
	}
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
		return string(ProtocolTCP)
	}
	return string(enum)
}

func (enum Protocol) Values() []Protocol {
	return []Protocol{
		"tcp",
		"http",
	}
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
		return string(ProxyProtocolProxyProtocolUnknown)
	}
	return string(enum)
}

func (enum ProxyProtocol) Values() []ProxyProtocol {
	return []ProxyProtocol{
		"proxy_protocol_unknown",
		"proxy_protocol_none",
		"proxy_protocol_v1",
		"proxy_protocol_v2",
		"proxy_protocol_v2_ssl",
		"proxy_protocol_v2_ssl_cn",
	}
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
		return string(SSLCompatibilityLevelSslCompatibilityLevelUnknown)
	}
	return string(enum)
}

func (enum SSLCompatibilityLevel) Values() []SSLCompatibilityLevel {
	return []SSLCompatibilityLevel{
		"ssl_compatibility_level_unknown",
		"ssl_compatibility_level_intermediate",
		"ssl_compatibility_level_modern",
		"ssl_compatibility_level_old",
	}
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
		return string(StickySessionsTypeNone)
	}
	return string(enum)
}

func (enum StickySessionsType) Values() []StickySessionsType {
	return []StickySessionsType{
		"none",
		"cookie",
		"table",
	}
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

// SubscriberEmailConfig: subscriber email config.
type SubscriberEmailConfig struct {
	// Email: email address to send alerts to.
	Email string `json:"email"`
}

// SubscriberWebhookConfig: Webhook alert of subscriber.
type SubscriberWebhookConfig struct {
	// URI: URI to receive POST requests.
	URI string `json:"uri"`
}

// HealthCheckHTTPConfig: health check http config.
type HealthCheckHTTPConfig struct {
	// URI: the HTTP path to use when performing a health check on backend servers.
	URI string `json:"uri"`

	// Method: the HTTP method used when performing a health check on backend servers.
	Method string `json:"method"`

	// Code: the HTTP response code that should be returned for a health check to be considered successful.
	Code *int32 `json:"code"`

	// HostHeader: the HTTP host header used when performing a health check on backend servers.
	HostHeader string `json:"host_header"`
}

// HealthCheckHTTPSConfig: health check https config.
type HealthCheckHTTPSConfig struct {
	// URI: the HTTP path to use when performing a health check on backend servers.
	URI string `json:"uri"`

	// Method: the HTTP method used when performing a health check on backend servers.
	Method string `json:"method"`

	// Code: the HTTP response code that should be returned for a health check to be considered successful.
	Code *int32 `json:"code"`

	// HostHeader: the HTTP host header used when performing a health check on backend servers.
	HostHeader string `json:"host_header"`

	// Sni: the SNI value used when performing a health check on backend servers over SSL.
	Sni string `json:"sni"`
}

// HealthCheckLdapConfig: health check ldap config.
type HealthCheckLdapConfig struct{}

// HealthCheckMysqlConfig: health check mysql config.
type HealthCheckMysqlConfig struct {
	// User: mySQL user to use for the health check.
	User string `json:"user"`
}

// HealthCheckPgsqlConfig: health check pgsql config.
type HealthCheckPgsqlConfig struct {
	// User: postgreSQL user to use for the health check.
	User string `json:"user"`
}

// HealthCheckRedisConfig: health check redis config.
type HealthCheckRedisConfig struct{}

// HealthCheckTCPConfig: health check tcp config.
type HealthCheckTCPConfig struct{}

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

	// Tags: IP tags.
	Tags []string `json:"tags"`

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

// Subscriber: Subscriber.
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
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22 or <9.0. For older or newer versions, use a TCP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`

	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`

	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
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
		tmpType:         tmpType(m),
		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
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

// ACLActionRedirect: acl action redirect.
type ACLActionRedirect struct {
	// Type: redirect type.
	// Default value: location
	Type ACLActionRedirectRedirectType `json:"type"`

	// Target: redirect target. For a location redirect, you can use a URL e.g. `https://scaleway.com`. Using a scheme name (e.g. `https`, `http`, `ftp`, `git`) will replace the request's original scheme. This can be useful to implement HTTP to HTTPS redirects. Valid placeholders that can be used in a `location` redirect to preserve parts of the original request in the redirection URL are \{\{host\}\}, \{\{query\}\}, \{\{path\}\} and \{\{scheme\}\}.
	Target string `json:"target"`

	// Code: HTTP redirect code to use. Valid values are 301, 302, 303, 307 and 308. Default value is 302.
	Code *int32 `json:"code"`
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

	// FailoverHost: scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host"`

	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging"`

	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify"`

	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count"`

	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries"`

	// MaxConnections: maximum number of connections allowed per backend server.
	MaxConnections *int32 `json:"max_connections"`

	// TimeoutQueue: maximum time for a request to be left pending in queue when `max_connections` is reached.
	TimeoutQueue *scw.Duration `json:"timeout_queue"`
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
		tmpType:           tmpType(m),
		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
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

// ACLAction: acl action.
type ACLAction struct {
	// Type: action to take when incoming traffic matches an ACL filter.
	// Default value: allow
	Type ACLActionType `json:"type"`

	// Redirect: redirection parameters when using an ACL with a `redirect` action.
	Redirect *ACLActionRedirect `json:"redirect"`
}

// ACLMatch: acl match.
type ACLMatch struct {
	// IPSubnet: list of IPs or CIDR v4/v6 addresses to filter for from the client side.
	IPSubnet []*string `json:"ip_subnet"`

	// IPsEdgeServices: defines whether Edge Services IPs should be matched.
	IPsEdgeServices bool `json:"ips_edge_services"`

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

	// ConnectionRateLimit: rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second.
	ConnectionRateLimit *uint32 `json:"connection_rate_limit"`

	// EnableAccessLogs: defines whether to enable access logs on the frontend.
	EnableAccessLogs bool `json:"enable_access_logs"`
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
		tmpType:          tmpType(m),
		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// PrivateNetworkDHCPConfig: private network dhcp config.
type PrivateNetworkDHCPConfig struct {
	// Deprecated
	IPID *string `json:"ip_id,omitempty"`
}

// PrivateNetworkIpamConfig: private network ipam config.
type PrivateNetworkIpamConfig struct{}

// PrivateNetworkStaticConfig: private network static config.
type PrivateNetworkStaticConfig struct {
	// Deprecated: IPAddress: array of a local IP address for the Load Balancer on this Private Network.
	IPAddress *[]string `json:"ip_address,omitempty"`
}

// RouteMatch: route match.
type RouteMatch struct {
	// Sni: value to match in the Server Name Indication TLS extension (SNI) field from an incoming connection made via an SSL/TLS transport layer. This field should be set for routes on TCP Load Balancers.
	// Precisely one of Sni, HostHeader, PathBegin must be set.
	Sni *string `json:"sni,omitempty"`

	// HostHeader: value to match in the HTTP Host request header from an incoming request. This field should be set for routes on HTTP Load Balancers.
	// Precisely one of Sni, HostHeader, PathBegin must be set.
	HostHeader *string `json:"host_header,omitempty"`

	// MatchSubdomains: if true, all subdomains will match.
	MatchSubdomains bool `json:"match_subdomains"`

	// PathBegin: value to match in the URL beginning path from an incoming request.
	// Precisely one of Sni, HostHeader, PathBegin must be set.
	PathBegin *string `json:"path_begin,omitempty"`
}

// CreateCertificateRequestCustomCertificate: create certificate request custom certificate.
type CreateCertificateRequestCustomCertificate struct {
	// CertificateChain: full PEM-formatted certificate, consisting of the entire certificate chain including public key, private key, and (optionally) Certificate Authorities.
	CertificateChain string `json:"certificate_chain"`
}

// CreateCertificateRequestLetsencryptConfig: create certificate request letsencrypt config.
type CreateCertificateRequestLetsencryptConfig struct {
	// CommonName: main domain name of certificate (this domain must exist and resolve to your Load Balancer IP address).
	CommonName string `json:"common_name"`

	// SubjectAlternativeName: alternative domain names (all domain names must exist and resolve to your Load Balancer IP address).
	SubjectAlternativeName []string `json:"subject_alternative_name"`
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

// ACL: acl.
type ACL struct {
	// ID: ACL ID.
	ID string `json:"id"`

	// Name: ACL name.
	Name string `json:"name"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` & `http_filter_value` are required.
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

// PrivateNetwork: private network.
type PrivateNetwork struct {
	// LB: load Balancer object which is attached to the Private Network.
	LB *LB `json:"lb"`

	// IpamIDs: iPAM IDs of the booked IP addresses.
	IpamIDs []string `json:"ipam_ids"`

	// Deprecated: StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`

	// Deprecated: DHCPConfig: object containing DHCP-assigned IP addresses.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`

	// Deprecated: IpamConfig: for internal use only.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
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

// ACLSpec: acl spec.
type ACLSpec struct {
	// Name: ACL name.
	Name string `json:"name"`

	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` and `http_filter_value` are required.
	Match *ACLMatch `json:"match"`

	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`

	// Description: ACL description.
	Description string `json:"description"`
}

// AddBackendServersRequest: add backend servers request.
type AddBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses to add to backend servers.
	ServerIP []string `json:"server_ip"`
}

// AttachPrivateNetworkRequest: attach private network request.
type AttachPrivateNetworkRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"-"`

	// Deprecated: StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`

	// Deprecated: DHCPConfig: defines whether to let DHCP assign IP addresses.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`

	// Deprecated: IpamConfig: for internal use only.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	IpamConfig *PrivateNetworkIpamConfig `json:"ipam_config,omitempty"`

	// IpamIDs: iPAM ID of a pre-reserved IP address to assign to the Load Balancer on this Private Network. In the future, it will be possible to specify multiple IPs in this field (IPv4 and IPv6), for now only one ID of an IPv4 address is expected. When null, a new private IP address is created for the Load Balancer on this Private Network.
	IpamIDs []string `json:"ipam_ids"`
}

// CreateACLRequest: Add an ACL to a Load Balancer frontend.
type CreateACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// FrontendID: frontend ID to attach the ACL to.
	FrontendID string `json:"-"`

	// Name: ACL name.
	Name string `json:"name"`

	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match,omitempty"`

	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`

	// Description: ACL description.
	Description string `json:"description"`
}

// CreateBackendRequest: create backend request.
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
	TimeoutServer *time.Duration `json:"timeout_server,omitempty"`

	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect,omitempty"`

	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel,omitempty"`

	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`

	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	// FailoverHost: scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host,omitempty"`

	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging,omitempty"`

	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify,omitempty"`

	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count,omitempty"`

	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries,omitempty"`

	// MaxConnections: maximum number of connections allowed per backend server.
	MaxConnections *int32 `json:"max_connections,omitempty"`

	// TimeoutQueue: maximum time for a request to be left pending in queue when `max_connections` is reached.
	TimeoutQueue *scw.Duration `json:"timeout_queue,omitempty"`
}

func (m *CreateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateBackendRequest
	tmp := struct {
		tmpType
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
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
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
	}{
		tmpType:           tmpType(m),
		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// CreateCertificateRequest: create certificate request.
type CreateCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Name: name for the certificate.
	Name string `json:"name"`

	// Letsencrypt: object to define a new Let's Encrypt certificate to be generated.
	// Precisely one of Letsencrypt, CustomCertificate must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`

	// CustomCertificate: object to define an existing custom certificate to be imported.
	// Precisely one of Letsencrypt, CustomCertificate must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateFrontendRequest: create frontend request.
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
	TimeoutClient *time.Duration `json:"timeout_client,omitempty"`

	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`

	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids,omitempty"`

	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`

	// ConnectionRateLimit: rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second.
	ConnectionRateLimit *uint32 `json:"connection_rate_limit,omitempty"`

	// EnableAccessLogs: defines whether to enable access logs on the frontend.
	EnableAccessLogs bool `json:"enable_access_logs"`
}

func (m *CreateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateFrontendRequest
	tmp := struct {
		tmpType
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
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
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
	}{
		tmpType:          tmpType(m),
		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// CreateIPRequest: create ip request.
type CreateIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// Deprecated: OrganizationID: organization ID of the Organization where the IP address should be created.
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: project ID of the Project where the IP address should be created.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`

	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse,omitempty"`

	// IsIPv6: if true, creates a Flexible IP with an ipv6 address.
	IsIPv6 bool `json:"is_ipv6"`

	// Tags: list of tags for the IP.
	Tags []string `json:"tags"`
}

// CreateLBRequest: create lb request.
type CreateLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// Deprecated: OrganizationID: scaleway Organization to create the Load Balancer in.
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: scaleway Project to create the Load Balancer in.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`

	// Name: name for the Load Balancer.
	Name string `json:"name"`

	// Description: description for the Load Balancer.
	Description string `json:"description"`

	// Deprecated: IPID: ID of an existing flexible IP address to attach to the Load Balancer.
	IPID *string `json:"ip_id,omitempty"`

	// AssignFlexibleIP: defines whether to automatically assign a flexible public IP to the Load Balancer. Default value is `true` (assign).
	AssignFlexibleIP *bool `json:"assign_flexible_ip,omitempty"`

	// AssignFlexibleIPv6: defines whether to automatically assign a flexible public IPv6 to the Load Balancer. Default value is `false` (do not assign).
	AssignFlexibleIPv6 *bool `json:"assign_flexible_ipv6,omitempty"`

	// IPIDs: list of IP IDs to attach to the Load Balancer.
	IPIDs []string `json:"ip_ids"`

	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`

	// Type: load Balancer commercial offer type. Use the Load Balancer types endpoint to retrieve a list of available offer types.
	Type string `json:"type"`

	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and do not need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// CreateRouteRequest: create route request.
type CreateRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// FrontendID: ID of the source frontend to create the route on.
	FrontendID string `json:"frontend_id"`

	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`

	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match,omitempty"`
}

// CreateSubscriberRequest: Create a new alert subscriber (webhook or email).
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
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: project ID to create the subscriber in.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// DeleteACLRequest: delete acl request.
type DeleteACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// DeleteBackendRequest: delete backend request.
type DeleteBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: ID of the backend to delete.
	BackendID string `json:"-"`
}

// DeleteCertificateRequest: delete certificate request.
type DeleteCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// DeleteFrontendRequest: delete frontend request.
type DeleteFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// FrontendID: ID of the frontend to delete.
	FrontendID string `json:"-"`
}

// DeleteLBRequest: delete lb request.
type DeleteLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: ID of the Load Balancer to delete.
	LBID string `json:"-"`

	// ReleaseIP: defines whether the Load Balancer's flexible IP should be deleted. Set to true to release the flexible IP, or false to keep it available in your account for future Load Balancers.
	ReleaseIP bool `json:"release_ip"`
}

// DeleteRouteRequest: delete route request.
type DeleteRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`
}

// DeleteSubscriberRequest: delete subscriber request.
type DeleteSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// DetachPrivateNetworkRequest: detach private network request.
type DetachPrivateNetworkRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load balancer ID.
	LBID string `json:"-"`

	// PrivateNetworkID: set your instance private network id.
	PrivateNetworkID string `json:"-"`
}

// GetACLRequest: get acl request.
type GetACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// GetBackendRequest: get backend request.
type GetBackendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`
}

// GetCertificateRequest: get certificate request.
type GetCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// GetFrontendRequest: get frontend request.
type GetFrontendRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
}

// GetIPRequest: get ip request.
type GetIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`
}

// GetLBRequest: get lb request.
type GetLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// GetLBStatsRequest: Get Load Balancer stats.
type GetLBStatsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// BackendID: ID of the backend.
	BackendID *string `json:"backend_id,omitempty"`
}

// GetRouteRequest: get route request.
type GetRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`
}

// GetSubscriberRequest: get subscriber request.
type GetSubscriberRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// LBStats: lb stats.
type LBStats struct {
	// BackendServersStats: list of objects containing Load Balancer statistics.
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
}

// ListACLResponse: list acl response.
type ListACLResponse struct {
	// ACLs: list of ACL objects.
	ACLs []*ACL `json:"acls"`

	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListACLResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ACLs = append(r.ACLs, results.ACLs...)
	r.TotalCount += uint32(len(results.ACLs))
	return uint32(len(results.ACLs)), nil
}

// ListACLsRequest: list ac ls request.
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

// ListBackendStatsRequest: list backend stats request.
type ListBackendStatsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`

	// PageSize: number of items to return.
	PageSize *uint32 `json:"-"`

	// BackendID: ID of the backend.
	BackendID *string `json:"-"`
}

// ListBackendStatsResponse: list backend stats response.
type ListBackendStatsResponse struct {
	// BackendServersStats: list of objects containing backend server statistics.
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`

	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListBackendStatsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.BackendServersStats = append(r.BackendServersStats, results.BackendServersStats...)
	r.TotalCount += uint32(len(results.BackendServersStats))
	return uint32(len(results.BackendServersStats)), nil
}

// ListBackendsRequest: list backends request.
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

// ListBackendsResponse: list backends response.
type ListBackendsResponse struct {
	// Backends: list of backend objects of a given Load Balancer.
	Backends []*Backend `json:"backends"`

	// TotalCount: total count of backend objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListBackendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Backends = append(r.Backends, results.Backends...)
	r.TotalCount += uint32(len(results.Backends))
	return uint32(len(results.Backends)), nil
}

// ListCertificatesRequest: list certificates request.
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

// ListCertificatesResponse: list certificates response.
type ListCertificatesResponse struct {
	// Certificates: list of certificate objects.
	Certificates []*Certificate `json:"certificates"`

	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListCertificatesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Certificates = append(r.Certificates, results.Certificates...)
	r.TotalCount += uint32(len(results.Certificates))
	return uint32(len(results.Certificates)), nil
}

// ListFrontendsRequest: list frontends request.
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

// ListFrontendsResponse: list frontends response.
type ListFrontendsResponse struct {
	// Frontends: list of frontend objects of a given Load Balancer.
	Frontends []*Frontend `json:"frontends"`

	// TotalCount: total count of frontend objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListFrontendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Frontends = append(r.Frontends, results.Frontends...)
	r.TotalCount += uint32(len(results.Frontends))
	return uint32(len(results.Frontends)), nil
}

// ListIPsRequest: list i ps request.
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

	// IPType: IP type to filter for.
	// Default value: all
	IPType ListIPsRequestIPType `json:"-"`

	// Tags: tag to filter for, only IPs with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ListIPsResponse: list i ps response.
type ListIPsResponse struct {
	// IPs: list of IP address objects.
	IPs []*IP `json:"ips"`

	// TotalCount: total count of IP address objects, without pagination.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListIPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.IPs = append(r.IPs, results.IPs...)
	r.TotalCount += uint32(len(results.IPs))
	return uint32(len(results.IPs)), nil
}

// ListLBPrivateNetworksRequest: list lb private networks request.
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

// ListLBPrivateNetworksResponse: list lb private networks response.
type ListLBPrivateNetworksResponse struct {
	// PrivateNetwork: list of Private Network objects attached to the Load Balancer.
	PrivateNetwork []*PrivateNetwork `json:"private_network"`

	// TotalCount: total number of objects in the response.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBPrivateNetworksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBPrivateNetworksResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListLBPrivateNetworksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PrivateNetwork = append(r.PrivateNetwork, results.PrivateNetwork...)
	r.TotalCount += uint32(len(results.PrivateNetwork))
	return uint32(len(results.PrivateNetwork)), nil
}

// ListLBTypesRequest: list lb types request.
type ListLBTypesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`

	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
}

// ListLBTypesResponse: list lb types response.
type ListLBTypesResponse struct {
	// LBTypes: list of Load Balancer commercial offer type objects.
	LBTypes []*LBType `json:"lb_types"`

	// TotalCount: total number of Load Balancer offer type objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBTypesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListLBTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LBTypes = append(r.LBTypes, results.LBTypes...)
	r.TotalCount += uint32(len(results.LBTypes))
	return uint32(len(results.LBTypes)), nil
}

// ListLBsRequest: list l bs request.
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

	// Tags: filter by tag, only Load Balancers with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ListLBsResponse: list l bs response.
type ListLBsResponse struct {
	// LBs: list of Load Balancer objects.
	LBs []*LB `json:"lbs"`

	// TotalCount: the total number of Load Balancer objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLBsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLBsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListLBsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LBs = append(r.LBs, results.LBs...)
	r.TotalCount += uint32(len(results.LBs))
	return uint32(len(results.LBs)), nil
}

// ListRoutesRequest: list routes request.
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

// ListRoutesResponse: list routes response.
type ListRoutesResponse struct {
	// Routes: list of route objects.
	Routes []*Route `json:"routes"`

	// TotalCount: the total number of route objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListRoutesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListRoutesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListRoutesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Routes = append(r.Routes, results.Routes...)
	r.TotalCount += uint32(len(results.Routes))
	return uint32(len(results.Routes)), nil
}

// ListSubscriberRequest: list subscriber request.
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

// ListSubscriberResponse: list subscriber response.
type ListSubscriberResponse struct {
	// Subscribers: list of subscriber objects.
	Subscribers []*Subscriber `json:"subscribers"`

	// TotalCount: the total number of objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListSubscriberResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Subscribers = append(r.Subscribers, results.Subscribers...)
	r.TotalCount += uint32(len(results.Subscribers))
	return uint32(len(results.Subscribers)), nil
}

// MigrateLBRequest: migrate lb request.
type MigrateLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Type: load Balancer type to migrate to (use the List all Load Balancer offer types endpoint to get a list of available offer types).
	Type string `json:"type"`
}

// ReleaseIPRequest: release ip request.
type ReleaseIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`
}

// RemoveBackendServersRequest: remove backend servers request.
type RemoveBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses to remove from backend servers.
	ServerIP []string `json:"server_ip"`
}

// SetACLsResponse: set ac ls response.
type SetACLsResponse struct {
	// ACLs: list of ACL objects.
	ACLs []*ACL `json:"acls"`

	// TotalCount: the total number of ACL objects.
	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *SetACLsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *SetACLsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*SetACLsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ACLs = append(r.ACLs, results.ACLs...)
	r.TotalCount += uint32(len(results.ACLs))
	return uint32(len(results.ACLs)), nil
}

// SetBackendServersRequest: set backend servers request.
type SetBackendServersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses for backend servers. Any other existing backend servers will be removed.
	ServerIP []string `json:"server_ip"`
}

// SubscribeToLBRequest: subscribe to lb request.
type SubscribeToLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"subscriber_id"`
}

// UnsubscribeFromLBRequest: unsubscribe from lb request.
type UnsubscribeFromLBRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// UpdateACLRequest: update acl request.
type UpdateACLRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`

	// Name: ACL name.
	Name string `json:"name"`

	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match,omitempty"`

	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`

	// Description: ACL description.
	Description *string `json:"description,omitempty"`
}

// UpdateBackendRequest: update backend request.
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
	TimeoutServer *time.Duration `json:"timeout_server,omitempty"`

	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect,omitempty"`

	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel,omitempty"`

	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`

	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	// FailoverHost: scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host,omitempty"`

	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging,omitempty"`

	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify,omitempty"`

	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count,omitempty"`

	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries,omitempty"`

	// MaxConnections: maximum number of connections allowed per backend server.
	MaxConnections *int32 `json:"max_connections,omitempty"`

	// TimeoutQueue: maximum time for a request to be left pending in queue when `max_connections` is reached.
	TimeoutQueue *scw.Duration `json:"timeout_queue,omitempty"`
}

func (m *UpdateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateBackendRequest
	tmp := struct {
		tmpType
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
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
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
	}{
		tmpType:           tmpType(m),
		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// UpdateCertificateRequest: update certificate request.
type UpdateCertificateRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`

	// Name: certificate name.
	Name string `json:"name"`
}

// UpdateFrontendRequest: update frontend request.
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
	TimeoutClient *time.Duration `json:"timeout_client,omitempty"`

	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`

	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids,omitempty"`

	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`

	// ConnectionRateLimit: rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second.
	ConnectionRateLimit *uint32 `json:"connection_rate_limit,omitempty"`

	// EnableAccessLogs: defines whether to enable access logs on the frontend.
	EnableAccessLogs *bool `json:"enable_access_logs,omitempty"`
}

func (m *UpdateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateFrontendRequest
	tmp := struct {
		tmpType
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
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
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
	}{
		tmpType:          tmpType(m),
		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// UpdateHealthCheckRequest: update health check request.
type UpdateHealthCheckRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// Port: port to use for the backend server health check.
	Port int32 `json:"port"`

	// CheckDelay: time to wait between two consecutive health checks.
	CheckDelay *time.Duration `json:"check_delay,omitempty"`

	// CheckTimeout: maximum time a backend server has to reply to the health check.
	CheckTimeout *time.Duration `json:"check_timeout,omitempty"`

	// CheckMaxRetries: number of consecutive unsuccessful health checks after which the server will be considered dead.
	CheckMaxRetries int32 `json:"check_max_retries"`

	// CheckSendProxy: defines whether proxy protocol should be activated for the health check.
	CheckSendProxy bool `json:"check_send_proxy"`

	// TCPConfig: object to configure a basic TCP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22 or <9.0. For older or newer versions, use a TCP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`

	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`

	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`

	// TransientCheckDelay: time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN).
	TransientCheckDelay *scw.Duration `json:"transient_check_delay,omitempty"`
}

func (m *UpdateHealthCheckRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateHealthCheckRequest
	tmp := struct {
		tmpType
		TmpCheckDelay   *marshaler.Duration `json:"check_delay,omitempty"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout,omitempty"`
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
		TmpCheckDelay   *marshaler.Duration `json:"check_delay,omitempty"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout,omitempty"`
	}{
		tmpType:         tmpType(m),
		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

// UpdateIPRequest: update ip request.
type UpdateIPRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`

	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse,omitempty"`

	// LBID: ID of the server on which to attach the flexible IP.
	LBID *string `json:"lb_id,omitempty"`

	// Tags: list of tags for the IP.
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateLBRequest: update lb request.
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

// UpdateRouteRequest: update route request.
type UpdateRouteRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`

	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`

	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match,omitempty"`
}

// UpdateSubscriberRequest: update subscriber request.
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

// ZonedAPIAddBackendServersRequest: zoned api add backend servers request.
type ZonedAPIAddBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses to add to backend servers.
	ServerIP []string `json:"server_ip"`
}

// ZonedAPIAttachPrivateNetworkRequest: zoned api attach private network request.
type ZonedAPIAttachPrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// PrivateNetworkID: private Network ID.
	PrivateNetworkID string `json:"private_network_id"`

	// Deprecated: StaticConfig: object containing an array of a local IP address for the Load Balancer on this Private Network.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	StaticConfig *PrivateNetworkStaticConfig `json:"static_config,omitempty"`

	// Deprecated: DHCPConfig: defines whether to let DHCP assign IP addresses.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	DHCPConfig *PrivateNetworkDHCPConfig `json:"dhcp_config,omitempty"`

	// Deprecated: IpamConfig: for internal use only.
	// Precisely one of StaticConfig, DHCPConfig, IpamConfig must be set.
	IpamConfig *PrivateNetworkIpamConfig `json:"ipam_config,omitempty"`

	// IpamIDs: iPAM ID of a pre-reserved IP address to assign to the Load Balancer on this Private Network. In the future, it will be possible to specify multiple IPs in this field (IPv4 and IPv6), for now only one ID of an IPv4 address is expected. When null, a new private IP address is created for the Load Balancer on this Private Network.
	IpamIDs []string `json:"ipam_ids"`
}

// ZonedAPICreateACLRequest: Add an ACL to a Load Balancer frontend.
type ZonedAPICreateACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// FrontendID: frontend ID to attach the ACL to.
	FrontendID string `json:"-"`

	// Name: ACL name.
	Name string `json:"name"`

	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match,omitempty"`

	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`

	// Description: ACL description.
	Description string `json:"description"`
}

// ZonedAPICreateBackendRequest: zoned api create backend request.
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
	TimeoutServer *time.Duration `json:"timeout_server,omitempty"`

	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect,omitempty"`

	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel,omitempty"`

	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`

	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	// FailoverHost: scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host,omitempty"`

	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging,omitempty"`

	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify,omitempty"`

	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count,omitempty"`

	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries,omitempty"`

	// MaxConnections: maximum number of connections allowed per backend server.
	MaxConnections *int32 `json:"max_connections,omitempty"`

	// TimeoutQueue: maximum time for a request to be left pending in queue when `max_connections` is reached.
	TimeoutQueue *scw.Duration `json:"timeout_queue,omitempty"`
}

func (m *ZonedAPICreateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPICreateBackendRequest
	tmp := struct {
		tmpType
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
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
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
	}{
		tmpType:           tmpType(m),
		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// ZonedAPICreateCertificateRequest: zoned api create certificate request.
type ZonedAPICreateCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Name: name for the certificate.
	Name string `json:"name"`

	// Letsencrypt: object to define a new Let's Encrypt certificate to be generated.
	// Precisely one of Letsencrypt, CustomCertificate must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`

	// CustomCertificate: object to define an existing custom certificate to be imported.
	// Precisely one of Letsencrypt, CustomCertificate must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// ZonedAPICreateFrontendRequest: zoned api create frontend request.
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
	TimeoutClient *time.Duration `json:"timeout_client,omitempty"`

	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`

	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids,omitempty"`

	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`

	// ConnectionRateLimit: rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second.
	ConnectionRateLimit *uint32 `json:"connection_rate_limit,omitempty"`

	// EnableAccessLogs: defines whether to enable access logs on the frontend.
	EnableAccessLogs bool `json:"enable_access_logs"`
}

func (m *ZonedAPICreateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPICreateFrontendRequest
	tmp := struct {
		tmpType
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
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
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
	}{
		tmpType:          tmpType(m),
		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// ZonedAPICreateIPRequest: zoned api create ip request.
type ZonedAPICreateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Deprecated: OrganizationID: organization ID of the Organization where the IP address should be created.
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: project ID of the Project where the IP address should be created.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`

	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse,omitempty"`

	// IsIPv6: if true, creates a Flexible IP with an ipv6 address.
	IsIPv6 bool `json:"is_ipv6"`

	// Tags: list of tags for the IP.
	Tags []string `json:"tags"`
}

// ZonedAPICreateLBRequest: zoned api create lb request.
type ZonedAPICreateLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Deprecated: OrganizationID: scaleway Organization to create the Load Balancer in.
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: scaleway Project to create the Load Balancer in.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`

	// Name: name for the Load Balancer.
	Name string `json:"name"`

	// Description: description for the Load Balancer.
	Description string `json:"description"`

	// Deprecated: IPID: ID of an existing flexible IP address to attach to the Load Balancer.
	IPID *string `json:"ip_id,omitempty"`

	// AssignFlexibleIP: defines whether to automatically assign a flexible public IP to the Load Balancer. Default value is `true` (assign).
	AssignFlexibleIP *bool `json:"assign_flexible_ip,omitempty"`

	// AssignFlexibleIPv6: defines whether to automatically assign a flexible public IPv6 to the Load Balancer. Default value is `false` (do not assign).
	AssignFlexibleIPv6 *bool `json:"assign_flexible_ipv6,omitempty"`

	// IPIDs: list of IP IDs to attach to the Load Balancer.
	IPIDs []string `json:"ip_ids"`

	// Tags: list of tags for the Load Balancer.
	Tags []string `json:"tags"`

	// Type: load Balancer commercial offer type. Use the Load Balancer types endpoint to retrieve a list of available offer types.
	Type string `json:"type"`

	// SslCompatibilityLevel: determines the minimal SSL version which needs to be supported on the client side, in an SSL/TLS offloading context. Intermediate is suitable for general-purpose servers with a variety of clients, recommended for almost all systems. Modern is suitable for services with clients that support TLS 1.3 and do not need backward compatibility. Old is compatible with a small number of very old clients and should be used only as a last resort.
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

// ZonedAPICreateRouteRequest: zoned api create route request.
type ZonedAPICreateRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// FrontendID: ID of the source frontend to create the route on.
	FrontendID string `json:"frontend_id"`

	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`

	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match,omitempty"`
}

// ZonedAPICreateSubscriberRequest: Create a new alert subscriber (webhook or email).
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
	// Precisely one of ProjectID, OrganizationID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`

	// ProjectID: project ID to create the subscriber in.
	// Precisely one of ProjectID, OrganizationID must be set.
	ProjectID *string `json:"project_id,omitempty"`
}

// ZonedAPIDeleteACLRequest: zoned api delete acl request.
type ZonedAPIDeleteACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// ZonedAPIDeleteBackendRequest: zoned api delete backend request.
type ZonedAPIDeleteBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: ID of the backend to delete.
	BackendID string `json:"-"`
}

// ZonedAPIDeleteCertificateRequest: zoned api delete certificate request.
type ZonedAPIDeleteCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// ZonedAPIDeleteFrontendRequest: zoned api delete frontend request.
type ZonedAPIDeleteFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// FrontendID: ID of the frontend to delete.
	FrontendID string `json:"-"`
}

// ZonedAPIDeleteLBRequest: zoned api delete lb request.
type ZonedAPIDeleteLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: ID of the Load Balancer to delete.
	LBID string `json:"-"`

	// ReleaseIP: defines whether the Load Balancer's flexible IP should be deleted. Set to true to release the flexible IP, or false to keep it available in your account for future Load Balancers.
	ReleaseIP bool `json:"release_ip"`
}

// ZonedAPIDeleteRouteRequest: zoned api delete route request.
type ZonedAPIDeleteRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`
}

// ZonedAPIDeleteSubscriberRequest: zoned api delete subscriber request.
type ZonedAPIDeleteSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// ZonedAPIDetachPrivateNetworkRequest: zoned api detach private network request.
type ZonedAPIDetachPrivateNetworkRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load balancer ID.
	LBID string `json:"-"`

	// PrivateNetworkID: set your instance private network id.
	PrivateNetworkID string `json:"private_network_id"`
}

// ZonedAPIGetACLRequest: zoned api get acl request.
type ZonedAPIGetACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`
}

// ZonedAPIGetBackendRequest: zoned api get backend request.
type ZonedAPIGetBackendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`
}

// ZonedAPIGetCertificateRequest: zoned api get certificate request.
type ZonedAPIGetCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`
}

// ZonedAPIGetFrontendRequest: zoned api get frontend request.
type ZonedAPIGetFrontendRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// FrontendID: frontend ID.
	FrontendID string `json:"-"`
}

// ZonedAPIGetIPRequest: zoned api get ip request.
type ZonedAPIGetIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`
}

// ZonedAPIGetLBRequest: zoned api get lb request.
type ZonedAPIGetLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// ZonedAPIGetLBStatsRequest: Get Load Balancer stats.
type ZonedAPIGetLBStatsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// BackendID: ID of the backend.
	BackendID *string `json:"backend_id,omitempty"`
}

// ZonedAPIGetRouteRequest: zoned api get route request.
type ZonedAPIGetRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`
}

// ZonedAPIGetSubscriberRequest: zoned api get subscriber request.
type ZonedAPIGetSubscriberRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"-"`
}

// ZonedAPIListACLsRequest: zoned api list ac ls request.
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

// ZonedAPIListBackendStatsRequest: zoned api list backend stats request.
type ZonedAPIListBackendStatsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`

	// PageSize: number of items to return.
	PageSize *uint32 `json:"-"`

	// BackendID: ID of the backend.
	BackendID *string `json:"-"`
}

// ZonedAPIListBackendsRequest: zoned api list backends request.
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

// ZonedAPIListCertificatesRequest: zoned api list certificates request.
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

// ZonedAPIListFrontendsRequest: zoned api list frontends request.
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

// ZonedAPIListIPsRequest: zoned api list i ps request.
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

	// IPType: IP type to filter for.
	// Default value: all
	IPType ListIPsRequestIPType `json:"-"`

	// Tags: tag to filter for, only IPs with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ZonedAPIListLBPrivateNetworksRequest: zoned api list lb private networks request.
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

// ZonedAPIListLBTypesRequest: zoned api list lb types request.
type ZonedAPIListLBTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Page: the page number to return, from the paginated results.
	Page *int32 `json:"-"`

	// PageSize: the number of items to return.
	PageSize *uint32 `json:"-"`
}

// ZonedAPIListLBsRequest: zoned api list l bs request.
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

	// Tags: filter by tag, only Load Balancers with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ZonedAPIListRoutesRequest: zoned api list routes request.
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

// ZonedAPIListSubscriberRequest: zoned api list subscriber request.
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

// ZonedAPIMigrateLBRequest: zoned api migrate lb request.
type ZonedAPIMigrateLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// Type: load Balancer type to migrate to (use the List all Load Balancer offer types endpoint to get a list of available offer types).
	Type string `json:"type"`
}

// ZonedAPIReleaseIPRequest: zoned api release ip request.
type ZonedAPIReleaseIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`
}

// ZonedAPIRemoveBackendServersRequest: zoned api remove backend servers request.
type ZonedAPIRemoveBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses to remove from backend servers.
	ServerIP []string `json:"server_ip"`
}

// ZonedAPISetACLsRequest: zoned api set ac ls request.
type ZonedAPISetACLsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// FrontendID: frontend ID.
	FrontendID string `json:"-"`

	// ACLs: list of ACLs for this frontend. Any other existing ACLs on this frontend will be removed.
	ACLs []*ACLSpec `json:"acls"`
}

// ZonedAPISetBackendServersRequest: zoned api set backend servers request.
type ZonedAPISetBackendServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// ServerIP: list of IP addresses for backend servers. Any other existing backend servers will be removed.
	ServerIP []string `json:"server_ip"`
}

// ZonedAPISubscribeToLBRequest: zoned api subscribe to lb request.
type ZonedAPISubscribeToLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`

	// SubscriberID: subscriber ID.
	SubscriberID string `json:"subscriber_id"`
}

// ZonedAPIUnsubscribeFromLBRequest: zoned api unsubscribe from lb request.
type ZonedAPIUnsubscribeFromLBRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// LBID: load Balancer ID.
	LBID string `json:"-"`
}

// ZonedAPIUpdateACLRequest: zoned api update acl request.
type ZonedAPIUpdateACLRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ACLID: ACL ID.
	ACLID string `json:"-"`

	// Name: ACL name.
	Name string `json:"name"`

	// Action: action to take when incoming traffic matches an ACL filter.
	Action *ACLAction `json:"action"`

	// Match: ACL match filter object. One of `ip_subnet`, `ips_edge_services` or `http_filter` & `http_filter_value` are required.
	Match *ACLMatch `json:"match,omitempty"`

	// Index: priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).
	Index int32 `json:"index"`

	// Description: ACL description.
	Description *string `json:"description,omitempty"`
}

// ZonedAPIUpdateBackendRequest: zoned api update backend request.
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
	TimeoutServer *time.Duration `json:"timeout_server,omitempty"`

	// TimeoutConnect: maximum allowed time for establishing a connection to a backend server.
	TimeoutConnect *time.Duration `json:"timeout_connect,omitempty"`

	// TimeoutTunnel: maximum allowed tunnel inactivity time after Websocket is established (takes precedence over client and server timeout).
	TimeoutTunnel *time.Duration `json:"timeout_tunnel,omitempty"`

	// OnMarkedDownAction: action to take when a backend server is marked as down.
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`

	// ProxyProtocol: protocol to use between the Load Balancer and backend servers. Allows the backend servers to be informed of the client's real IP address. The PROXY protocol must be supported by the backend servers' software.
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`

	// FailoverHost: scaleway Object Storage bucket website to be served as failover if all backend servers are down, e.g. failover-website.s3-website.fr-par.scw.cloud.
	FailoverHost *string `json:"failover_host,omitempty"`

	// SslBridging: defines whether to enable SSL bridging between the Load Balancer and backend servers.
	SslBridging *bool `json:"ssl_bridging,omitempty"`

	// IgnoreSslServerVerify: defines whether the server certificate verification should be ignored.
	IgnoreSslServerVerify *bool `json:"ignore_ssl_server_verify,omitempty"`

	// RedispatchAttemptCount: whether to use another backend server on each attempt.
	RedispatchAttemptCount *int32 `json:"redispatch_attempt_count,omitempty"`

	// MaxRetries: number of retries when a backend server connection failed.
	MaxRetries *int32 `json:"max_retries,omitempty"`

	// MaxConnections: maximum number of connections allowed per backend server.
	MaxConnections *int32 `json:"max_connections,omitempty"`

	// TimeoutQueue: maximum time for a request to be left pending in queue when `max_connections` is reached.
	TimeoutQueue *scw.Duration `json:"timeout_queue,omitempty"`
}

func (m *ZonedAPIUpdateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateBackendRequest
	tmp := struct {
		tmpType
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
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
		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server,omitempty"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect,omitempty"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel,omitempty"`
	}{
		tmpType:           tmpType(m),
		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// ZonedAPIUpdateCertificateRequest: zoned api update certificate request.
type ZonedAPIUpdateCertificateRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// CertificateID: certificate ID.
	CertificateID string `json:"-"`

	// Name: certificate name.
	Name string `json:"name"`
}

// ZonedAPIUpdateFrontendRequest: zoned api update frontend request.
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
	TimeoutClient *time.Duration `json:"timeout_client,omitempty"`

	// Deprecated: CertificateID: certificate ID, deprecated in favor of certificate_ids array.
	CertificateID *string `json:"certificate_id,omitempty"`

	// CertificateIDs: list of SSL/TLS certificate IDs to bind to the frontend.
	CertificateIDs *[]string `json:"certificate_ids,omitempty"`

	// EnableHTTP3: defines whether to enable HTTP/3 protocol on the frontend.
	EnableHTTP3 bool `json:"enable_http3"`

	// ConnectionRateLimit: rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second.
	ConnectionRateLimit *uint32 `json:"connection_rate_limit,omitempty"`

	// EnableAccessLogs: defines whether to enable access logs on the frontend.
	EnableAccessLogs *bool `json:"enable_access_logs,omitempty"`
}

func (m *ZonedAPIUpdateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateFrontendRequest
	tmp := struct {
		tmpType
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
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
		TmpTimeoutClient *marshaler.Duration `json:"timeout_client,omitempty"`
	}{
		tmpType:          tmpType(m),
		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// ZonedAPIUpdateHealthCheckRequest: zoned api update health check request.
type ZonedAPIUpdateHealthCheckRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// BackendID: backend ID.
	BackendID string `json:"-"`

	// Port: port to use for the backend server health check.
	Port int32 `json:"port"`

	// CheckDelay: time to wait between two consecutive health checks.
	CheckDelay *time.Duration `json:"check_delay,omitempty"`

	// CheckTimeout: maximum time a backend server has to reply to the health check.
	CheckTimeout *time.Duration `json:"check_timeout,omitempty"`

	// CheckMaxRetries: number of consecutive unsuccessful health checks after which the server will be considered dead.
	CheckMaxRetries int32 `json:"check_max_retries"`

	// CheckSendProxy: defines whether proxy protocol should be activated for the health check.
	CheckSendProxy bool `json:"check_send_proxy"`

	// TCPConfig: object to configure a basic TCP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// MysqlConfig: object to configure a MySQL health check. The check requires MySQL >=3.22 or <9.0. For older or newer versions, use a TCP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`

	// PgsqlConfig: object to configure a PostgreSQL health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// LdapConfig: object to configure an LDAP health check. The response is analyzed to find the LDAPv3 response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`

	// RedisConfig: object to configure a Redis health check. The response is analyzed to find the +PONG response message.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// HTTPConfig: object to configure an HTTP health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// HTTPSConfig: object to configure an HTTPS health check.
	// Precisely one of TCPConfig, MysqlConfig, PgsqlConfig, LdapConfig, RedisConfig, HTTPConfig, HTTPSConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`

	// TransientCheckDelay: time to wait between two consecutive health checks when a backend server is in a transient state (going UP or DOWN).
	TransientCheckDelay *scw.Duration `json:"transient_check_delay,omitempty"`
}

func (m *ZonedAPIUpdateHealthCheckRequest) UnmarshalJSON(b []byte) error {
	type tmpType ZonedAPIUpdateHealthCheckRequest
	tmp := struct {
		tmpType
		TmpCheckDelay   *marshaler.Duration `json:"check_delay,omitempty"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout,omitempty"`
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
		TmpCheckDelay   *marshaler.Duration `json:"check_delay,omitempty"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout,omitempty"`
	}{
		tmpType:         tmpType(m),
		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

// ZonedAPIUpdateIPRequest: zoned api update ip request.
type ZonedAPIUpdateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: IP address ID.
	IPID string `json:"-"`

	// Reverse: reverse DNS (domain name) for the IP address.
	Reverse *string `json:"reverse,omitempty"`

	// LBID: ID of the server on which to attach the flexible IP.
	LBID *string `json:"lb_id,omitempty"`

	// Tags: list of tags for the IP.
	Tags *[]string `json:"tags,omitempty"`
}

// ZonedAPIUpdateLBRequest: zoned api update lb request.
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

// ZonedAPIUpdateRouteRequest: zoned api update route request.
type ZonedAPIUpdateRouteRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// RouteID: route ID.
	RouteID string `json:"-"`

	// BackendID: ID of the target backend for the route.
	BackendID string `json:"backend_id"`

	// Match: object defining the match condition for a route to be applied. If an incoming client session matches the specified condition (i.e. it has a matching SNI value or HTTP Host header value), it will be passed to the target backend.
	Match *RouteMatch `json:"match,omitempty"`
}

// ZonedAPIUpdateSubscriberRequest: zoned api update subscriber request.
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

// This API allows you to manage your Scaleway Load Balancer services.
type ZonedAPI struct {
	client *scw.Client
}

// NewZonedAPI returns a ZonedAPI object from a Scaleway client.
func NewZonedAPI(client *scw.Client) *ZonedAPI {
	return &ZonedAPI{
		client: client,
	}
}

func (s *ZonedAPI) Zones() []scw.Zone {
	return []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZoneNlAms3, scw.ZonePlWaw1, scw.ZonePlWaw2, scw.ZonePlWaw3}
}

// ListLBs: List all Load Balancers in the specified zone, for a Scaleway Organization or Scaleway Project. By default, the Load Balancers returned in the list are ordered by creation date in ascending order, though this can be modified via the `order_by` field.
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
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs",
		Query:  query,
	}

	var resp ListLBsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateLB: Create a new Load Balancer. Note that the Load Balancer will be created without frontends or backends; these must be created separately via the dedicated endpoints.
func (s *ZonedAPI) CreateLB(req *ZonedAPICreateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lb")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs",
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

// GetLB: Retrieve information about an existing Load Balancer, specified by its Load Balancer ID. Its full details, including name, status and IP address, are returned in the response object.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// WaitForLBRequest is used by WaitForLB method.
type WaitForLBRequest struct {
	Zone          scw.Zone
	LBID          string
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForLB waits for the LB to reach a terminal state.
func (s *ZonedAPI) WaitForLB(req *WaitForLBRequest, opts ...scw.RequestOption) (*LB, error) {
	timeout := defaultLBTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}

	retryInterval := defaultLBRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}
	transientStatuses := map[LBStatus]struct{}{
		LBStatusPending:   {},
		LBStatusMigrating: {},
		LBStatusToCreate:  {},
		LBStatusCreating:  {},
		LBStatusToDelete:  {},
		LBStatusDeleting:  {},
	}

	res, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetLB(&ZonedAPIGetLBRequest{
				Zone: req.Zone,
				LBID: req.LBID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			_, isTransient := transientStatuses[res.Status]

			return res, !isTransient, nil
		},
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
		Timeout:          timeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for LB failed")
	}

	return res.(*LB), nil
}

// UpdateLB: Update the parameters of an existing Load Balancer, specified by its Load Balancer ID. Note that the request type is PUT and not PATCH. You must set all parameters.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
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

// DeleteLB: Delete an existing Load Balancer, specified by its Load Balancer ID. Deleting a Load Balancer is permanent, and cannot be undone. The Load Balancer's flexible IP address can either be deleted with the Load Balancer, or kept in your account for future use.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Query:  query,
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// MigrateLB: Migrate an existing Load Balancer from one commercial type to another. Allows you to scale your Load Balancer up or down in terms of bandwidth or multi-cloud provision.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/migrate",
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

// ListIPs: List the Load Balancer flexible IP addresses held in the account (filtered by Organization ID or Project ID). It is also possible to search for a specific IP address.
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
	parameter.AddToQuery(query, "ip_type", req.IPType)
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:  query,
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateIP: Create a new Load Balancer flexible IP address, in the specified Scaleway Project. This can be attached to new Load Balancers created in the future.
func (s *ZonedAPI) CreateIP(req *ZonedAPICreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
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

// GetIP: Retrieve the full details of a Load Balancer flexible IP address.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ReleaseIP: Delete a Load Balancer flexible IP address. This action is irreversible, and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UpdateIP: Update the reverse DNS of a Load Balancer flexible IP address.
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
		Method: "PATCH",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
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

// ListBackends: List all the backends of a Load Balancer, specified by its Load Balancer ID. By default, results are returned in ascending order by the creation date of each backend. The response is an array of backend objects, containing full details of each one including their configuration parameters such as protocol, port and forwarding algorithm.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Query:  query,
	}

	var resp ListBackendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateBackend: Create a new backend for a given Load Balancer, specifying its full configuration including protocol, port and forwarding algorithm.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
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

// GetBackend: Get the full details of a given backend, specified by its backend ID. The response contains the backend's full configuration parameters including protocol, port and forwarding algorithm.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateBackend: Update a backend of a given Load Balancer, specified by its backend ID. Note that the request type is PUT and not PATCH. You must set all parameters.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
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

// DeleteBackend: Delete a backend of a given Load Balancer, specified by its backend ID. This action is irreversible and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// AddBackendServers: For a given backend specified by its backend ID, add a set of backend servers (identified by their IP addresses) it should forward traffic to. These will be appended to any existing set of backend servers for this backend.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// RemoveBackendServers: For a given backend specified by its backend ID, remove the specified backend servers (identified by their IP addresses) so that it no longer forwards traffic to them.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// SetBackendServers: For a given backend specified by its backend ID, define the set of backend servers (identified by their IP addresses) that it should forward traffic to. Any existing backend servers configured for this backend will be removed.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// UpdateHealthCheck: Update the configuration of the health check performed by a given backend to verify the health of its backend servers, identified by its backend ID. Note that the request type is PUT and not PATCH. You must set all parameters.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/backends/" + fmt.Sprint(req.BackendID) + "/healthcheck",
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

// ListFrontends: List all the frontends of a Load Balancer, specified by its Load Balancer ID. By default, results are returned in ascending order by the creation date of each frontend. The response is an array of frontend objects, containing full details of each one including the port they listen on and the backend they are attached to.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Query:  query,
	}

	var resp ListFrontendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateFrontend: Create a new frontend for a given Load Balancer, specifying its configuration including the port it should listen on and the backend to attach it to.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
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

// GetFrontend: Get the full details of a given frontend, specified by its frontend ID. The response contains the frontend's full configuration parameters including the backend it is attached to, the port it listens on, and any certificates it has.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateFrontend: Update a given frontend, specified by its frontend ID. You can update configuration parameters including its name and the port it listens on. Note that the request type is PUT and not PATCH. You must set all parameters.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
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

// DeleteFrontend: Delete a given frontend, specified by its frontend ID. This action is irreversible and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListRoutes: List all routes for a given frontend. The response is an array of routes, each one  with a specified backend to direct to if a certain condition is matched (based on the value of the SNI field or HTTP Host header).
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes",
		Query:  query,
	}

	var resp ListRoutesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateRoute: Create a new route on a given frontend. To configure a route, specify the backend to direct to if a certain condition is matched (based on the value of the SNI field or HTTP Host header).
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes",
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

// GetRoute: Retrieve information about an existing route, specified by its route ID. Its full details, origin frontend, target backend and match condition, are returned in the response object.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRoute: Update the configuration of an existing route, specified by its route ID.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
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

// DeleteRoute: Delete an existing route, specified by its route ID. Deleting a route is permanent, and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/routes/" + fmt.Sprint(req.RouteID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated: GetLBStats: Get usage statistics of a given Load Balancer.
func (s *ZonedAPI) GetLBStats(req *ZonedAPIGetLBStatsRequest, opts ...scw.RequestOption) (*LBStats, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "backend_id", req.BackendID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/stats",
		Query:  query,
	}

	var resp LBStats

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListBackendStats: List information about your backend servers, including their state and the result of their last health check.
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
	parameter.AddToQuery(query, "backend_id", req.BackendID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/backend-stats",
		Query:  query,
	}

	var resp ListBackendStatsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListACLs: List the ACLs for a given frontend, specified by its frontend ID. The response is an array of ACL objects, each one representing an ACL that denies or allows traffic based on certain conditions.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Query:  query,
	}

	var resp ListACLResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateACL: Create a new ACL for a given frontend. Each ACL must have a name, an action to perform (allow or deny), and a match rule (the action is carried out when the incoming traffic matches the rule).
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
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

// GetACL: Get information for a particular ACL, specified by its ACL ID. The response returns full details of the ACL, including its name, action, match rule and frontend.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateACL: Update a particular ACL, specified by its ACL ID. You can update details including its name, action and match rule.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
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

// DeleteACL: Delete an ACL, specified by its ACL ID. Deleting an ACL is irreversible and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/acls/" + fmt.Sprint(req.ACLID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// SetACLs: For a given frontend specified by its frontend ID, define and add the complete set of ACLS for that frontend. Any existing ACLs on this frontend will be removed.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
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

// CreateCertificate: Generate a new SSL/TLS certificate for a given Load Balancer. You can choose to create a Let's Encrypt certificate, or import a custom certificate.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
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

// ListCertificates: List all the SSL/TLS certificates on a given Load Balancer. The response is an array of certificate objects, which are by default listed in ascending order of creation date.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Query:  query,
	}

	var resp ListCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCertificate: Get information for a particular SSL/TLS certificate, specified by its certificate ID. The response returns full details of the certificate, including its type, main domain name, and alternative domain names.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// WaitForCertificateRequest is used by WaitForCertificate method.
type WaitForCertificateRequest struct {
	Zone          scw.Zone
	CertificateID string
	Timeout       *time.Duration
	RetryInterval *time.Duration
}

// WaitForCertificate waits for the Certificate to reach a terminal state.
func (s *ZonedAPI) WaitForCertificate(req *WaitForCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	timeout := defaultLBTimeout
	if req.Timeout != nil {
		timeout = *req.Timeout
	}

	retryInterval := defaultLBRetryInterval
	if req.RetryInterval != nil {
		retryInterval = *req.RetryInterval
	}
	transientStatuses := map[CertificateStatus]struct{}{
		CertificateStatusPending: {},
	}

	res, err := async.WaitSync(&async.WaitSyncConfig{
		Get: func() (any, bool, error) {
			res, err := s.GetCertificate(&ZonedAPIGetCertificateRequest{
				Zone:          req.Zone,
				CertificateID: req.CertificateID,
			}, opts...)
			if err != nil {
				return nil, false, err
			}

			_, isTransient := transientStatuses[res.Status]

			return res, !isTransient, nil
		},
		IntervalStrategy: async.LinearIntervalStrategy(retryInterval),
		Timeout:          timeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "waiting for Certificate failed")
	}

	return res.(*Certificate), nil
}

// UpdateCertificate: Update the name of a particular SSL/TLS certificate, specified by its certificate ID.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
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

// DeleteCertificate: Delete an SSL/TLS certificate, specified by its certificate ID. Deleting a certificate is irreversible and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListLBTypes: List all the different commercial Load Balancer types. The response includes an array of offer types, each with a name, description, and information about its stock availability.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb-types",
		Query:  query,
	}

	var resp ListLBTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSubscriber: Create a new subscriber, either with an email configuration or a webhook configuration, for a specified Scaleway Project.
func (s *ZonedAPI) CreateSubscriber(req *ZonedAPICreateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers",
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

// GetSubscriber: Retrieve information about an existing subscriber, specified by its subscriber ID. Its full details, including name and email/webhook configuration, are returned in the response object.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSubscriber: List all subscribers to Load Balancer alerts. By default, returns all subscribers to Load Balancer alerts for the Organization associated with the authentication token used for the request.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers",
		Query:  query,
	}

	var resp ListSubscriberResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSubscriber: Update the parameters of a given subscriber (e.g. name, webhook configuration, email configuration), specified by its subscriber ID.
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
		Method: "PUT",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
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

// DeleteSubscriber: Delete an existing subscriber, specified by its subscriber ID. Deleting a subscriber is permanent, and cannot be undone.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/subscription/" + fmt.Sprint(req.SubscriberID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// SubscribeToLB: Subscribe an existing subscriber to alerts for a given Load Balancer.
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
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/" + fmt.Sprint(req.LBID) + "/subscribe",
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

// UnsubscribeFromLB: Unsubscribe a subscriber from alerts for a given Load Balancer. The subscriber is not deleted, and can be resubscribed in the future if necessary.
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
		Method: "DELETE",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lb/" + fmt.Sprint(req.LBID) + "/unsubscribe",
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLBPrivateNetworks: List the Private Networks attached to a given Load Balancer, specified by its Load Balancer ID. The response is an array of Private Network objects, giving information including the status, configuration, name and creation date of each Private Network.
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
		Method: "GET",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks",
		Query:  query,
	}

	var resp ListLBPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttachPrivateNetwork: Attach a specified Load Balancer to a specified Private Network, defining a static or DHCP configuration for the Load Balancer on the network.
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

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/attach-private-network",
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

// DetachPrivateNetwork: Detach a specified Load Balancer from a specified Private Network.
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

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/zones/" + fmt.Sprint(req.Zone) + "/lbs/" + fmt.Sprint(req.LBID) + "/detach-private-network",
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

// This API allows you to manage your Load Balancers.
type API struct {
	client *scw.Client
}

// Deprecated: NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

func (s *API) Regions() []scw.Region {
	return []scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}
}

// ListLBs: List load balancers.
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
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
		Query:  query,
	}

	var resp ListLBsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateLB: Create a load balancer.
func (s *API) CreateLB(req *CreateLBRequest, opts ...scw.RequestOption) (*LB, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("lb")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
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

// GetLB: Get a load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateLB: Update a load balancer.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
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

// DeleteLB: Delete a load balancer.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "",
		Query:  query,
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// MigrateLB: Migrate a load balancer.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/migrate",
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

// ListIPs: List IPs.
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
	parameter.AddToQuery(query, "ip_type", req.IPType)
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Query:  query,
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateIP: Create an IP.
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
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

// GetIP: Get an IP.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ReleaseIP: Delete an IP.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UpdateIP: Update an IP.
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
		Method: "PATCH",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
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

// ListBackends: List backends in a given load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
		Query:  query,
	}

	var resp ListBackendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateBackend: Create a backend in a given load balancer.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backends",
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

// GetBackend: Get a backend in a given load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateBackend: Update a backend in a given load balancer.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
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

// DeleteBackend: Delete a backend in a given load balancer.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// AddBackendServers: Add a set of servers in a given backend.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// RemoveBackendServers: Remove a set of servers for a given backend.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// SetBackendServers: Define all servers in a given backend.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
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

// UpdateHealthCheck: Update an health check for a given backend.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/healthcheck",
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

// ListFrontends: List frontends in a given load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
		Query:  query,
	}

	var resp ListFrontendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateFrontend: Create a frontend in a given load balancer.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/frontends",
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

// GetFrontend: Get a frontend.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateFrontend: Update a frontend.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
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

// DeleteFrontend: Delete a frontend.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListRoutes: List all backend redirections.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes",
		Query:  query,
	}

	var resp ListRoutesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateRoute: Create a backend redirection.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes",
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

// GetRoute: Get single backend redirection.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
	}

	var resp Route

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRoute: Edit a backend redirection.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
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

// DeleteRoute: Delete a backend redirection.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/routes/" + fmt.Sprint(req.RouteID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated: GetLBStats: Get usage statistics of a given load balancer.
func (s *API) GetLBStats(req *GetLBStatsRequest, opts ...scw.RequestOption) (*LBStats, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "backend_id", req.BackendID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/stats",
		Query:  query,
	}

	var resp LBStats

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListBackendStats: List backend server statistics.
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
	parameter.AddToQuery(query, "backend_id", req.BackendID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LBID) == "" {
		return nil, errors.New("field LBID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/backend-stats",
		Query:  query,
	}

	var resp ListBackendStatsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListACLs: List ACL for a given frontend.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Query:  query,
	}

	var resp ListACLResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateACL: Create an ACL for a given frontend.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
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

// GetACL: Get an ACL.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateACL: Update an ACL.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
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

// DeleteACL: Delete an ACL.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// CreateCertificate: Generate a new TLS certificate using Let's Encrypt or import your certificate.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
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

// ListCertificates: List all TLS certificates on a given load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/certificates",
		Query:  query,
	}

	var resp ListCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCertificate: Get a TLS certificate.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCertificate: Update a TLS certificate.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
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

// DeleteCertificate: Delete a TLS certificate.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListLBTypes: List all load balancer offer type.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb-types",
		Query:  query,
	}

	var resp ListLBTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSubscriber: Create a subscriber, webhook or email.
func (s *API) CreateSubscriber(req *CreateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.ProjectID == nil && req.OrganizationID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
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

// GetSubscriber: Get a subscriber.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSubscriber: List all subscriber.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
		Query:  query,
	}

	var resp ListSubscriberResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSubscriber: Update a subscriber.
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
		Method: "PUT",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
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

// DeleteSubscriber: Delete a subscriber.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/subscriber/" + fmt.Sprint(req.SubscriberID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// SubscribeToLB: Subscribe a subscriber to a given load balancer.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LBID) + "/subscribe",
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

// UnsubscribeFromLB: Unsubscribe a subscriber from a given load balancer.
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
		Method: "DELETE",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LBID) + "/unsubscribe",
	}

	var resp LB

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLBPrivateNetworks: List attached private network of load balancer.
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
		Method: "GET",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks",
		Query:  query,
	}

	var resp ListLBPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttachPrivateNetwork: Add load balancer on instance private network.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/attach",
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

// DetachPrivateNetwork: Remove load balancer of private network.
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
		Method: "POST",
		Path:   "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LBID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/detach",
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
