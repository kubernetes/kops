// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package instance provides methods and message types of the instance v1 API.
package instance

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
	"github.com/scaleway/scaleway-sdk-go/marshaler"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/parameter"
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

type Arch string

const (
	ArchUnknownArch = Arch("unknown_arch")
	ArchX86_64      = Arch("x86_64")
	ArchArm         = Arch("arm")
	ArchArm64       = Arch("arm64")
)

func (enum Arch) String() string {
	if enum == "" {
		// return default value if empty
		return string(ArchUnknownArch)
	}
	return string(enum)
}

func (enum Arch) Values() []Arch {
	return []Arch{
		"unknown_arch",
		"x86_64",
		"arm",
		"arm64",
	}
}

func (enum Arch) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *Arch) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = Arch(Arch(tmp).String())
	return nil
}

type AttachServerVolumeRequestVolumeType string

const (
	AttachServerVolumeRequestVolumeTypeUnknownVolumeType = AttachServerVolumeRequestVolumeType("unknown_volume_type")
	AttachServerVolumeRequestVolumeTypeLSSD              = AttachServerVolumeRequestVolumeType("l_ssd")
	AttachServerVolumeRequestVolumeTypeBSSD              = AttachServerVolumeRequestVolumeType("b_ssd")
	AttachServerVolumeRequestVolumeTypeSbsVolume         = AttachServerVolumeRequestVolumeType("sbs_volume")
)

func (enum AttachServerVolumeRequestVolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return string(AttachServerVolumeRequestVolumeTypeUnknownVolumeType)
	}
	return string(enum)
}

func (enum AttachServerVolumeRequestVolumeType) Values() []AttachServerVolumeRequestVolumeType {
	return []AttachServerVolumeRequestVolumeType{
		"unknown_volume_type",
		"l_ssd",
		"b_ssd",
		"sbs_volume",
	}
}

func (enum AttachServerVolumeRequestVolumeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *AttachServerVolumeRequestVolumeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = AttachServerVolumeRequestVolumeType(AttachServerVolumeRequestVolumeType(tmp).String())
	return nil
}

type BootType string

const (
	BootTypeLocal      = BootType("local")
	BootTypeBootscript = BootType("bootscript")
	BootTypeRescue     = BootType("rescue")
)

func (enum BootType) String() string {
	if enum == "" {
		// return default value if empty
		return string(BootTypeLocal)
	}
	return string(enum)
}

func (enum BootType) Values() []BootType {
	return []BootType{
		"local",
		"bootscript",
		"rescue",
	}
}

func (enum BootType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BootType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BootType(BootType(tmp).String())
	return nil
}

type IPState string

const (
	IPStateUnknownState = IPState("unknown_state")
	IPStateDetached     = IPState("detached")
	IPStateAttached     = IPState("attached")
	IPStatePending      = IPState("pending")
	IPStateError        = IPState("error")
)

func (enum IPState) String() string {
	if enum == "" {
		// return default value if empty
		return string(IPStateUnknownState)
	}
	return string(enum)
}

func (enum IPState) Values() []IPState {
	return []IPState{
		"unknown_state",
		"detached",
		"attached",
		"pending",
		"error",
	}
}

func (enum IPState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPState(IPState(tmp).String())
	return nil
}

type IPType string

const (
	IPTypeUnknownIptype = IPType("unknown_iptype")
	IPTypeRoutedIPv4    = IPType("routed_ipv4")
	IPTypeRoutedIPv6    = IPType("routed_ipv6")
)

func (enum IPType) String() string {
	if enum == "" {
		// return default value if empty
		return string(IPTypeUnknownIptype)
	}
	return string(enum)
}

func (enum IPType) Values() []IPType {
	return []IPType{
		"unknown_iptype",
		"routed_ipv4",
		"routed_ipv6",
	}
}

func (enum IPType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPType(IPType(tmp).String())
	return nil
}

type ImageState string

const (
	ImageStateAvailable = ImageState("available")
	ImageStateCreating  = ImageState("creating")
	ImageStateError     = ImageState("error")
)

func (enum ImageState) String() string {
	if enum == "" {
		// return default value if empty
		return string(ImageStateAvailable)
	}
	return string(enum)
}

func (enum ImageState) Values() []ImageState {
	return []ImageState{
		"available",
		"creating",
		"error",
	}
}

func (enum ImageState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ImageState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ImageState(ImageState(tmp).String())
	return nil
}

type ListServersRequestOrder string

const (
	ListServersRequestOrderCreationDateDesc     = ListServersRequestOrder("creation_date_desc")
	ListServersRequestOrderCreationDateAsc      = ListServersRequestOrder("creation_date_asc")
	ListServersRequestOrderModificationDateDesc = ListServersRequestOrder("modification_date_desc")
	ListServersRequestOrderModificationDateAsc  = ListServersRequestOrder("modification_date_asc")
)

func (enum ListServersRequestOrder) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListServersRequestOrderCreationDateDesc)
	}
	return string(enum)
}

func (enum ListServersRequestOrder) Values() []ListServersRequestOrder {
	return []ListServersRequestOrder{
		"creation_date_desc",
		"creation_date_asc",
		"modification_date_desc",
		"modification_date_asc",
	}
}

func (enum ListServersRequestOrder) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListServersRequestOrder) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListServersRequestOrder(ListServersRequestOrder(tmp).String())
	return nil
}

type PlacementGroupPolicyMode string

const (
	PlacementGroupPolicyModeOptional = PlacementGroupPolicyMode("optional")
	PlacementGroupPolicyModeEnforced = PlacementGroupPolicyMode("enforced")
)

func (enum PlacementGroupPolicyMode) String() string {
	if enum == "" {
		// return default value if empty
		return string(PlacementGroupPolicyModeOptional)
	}
	return string(enum)
}

func (enum PlacementGroupPolicyMode) Values() []PlacementGroupPolicyMode {
	return []PlacementGroupPolicyMode{
		"optional",
		"enforced",
	}
}

func (enum PlacementGroupPolicyMode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PlacementGroupPolicyMode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PlacementGroupPolicyMode(PlacementGroupPolicyMode(tmp).String())
	return nil
}

type PlacementGroupPolicyType string

const (
	PlacementGroupPolicyTypeMaxAvailability = PlacementGroupPolicyType("max_availability")
	PlacementGroupPolicyTypeLowLatency      = PlacementGroupPolicyType("low_latency")
)

func (enum PlacementGroupPolicyType) String() string {
	if enum == "" {
		// return default value if empty
		return string(PlacementGroupPolicyTypeMaxAvailability)
	}
	return string(enum)
}

func (enum PlacementGroupPolicyType) Values() []PlacementGroupPolicyType {
	return []PlacementGroupPolicyType{
		"max_availability",
		"low_latency",
	}
}

func (enum PlacementGroupPolicyType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PlacementGroupPolicyType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PlacementGroupPolicyType(PlacementGroupPolicyType(tmp).String())
	return nil
}

type PrivateNICState string

const (
	PrivateNICStateAvailable    = PrivateNICState("available")
	PrivateNICStateSyncing      = PrivateNICState("syncing")
	PrivateNICStateSyncingError = PrivateNICState("syncing_error")
)

func (enum PrivateNICState) String() string {
	if enum == "" {
		// return default value if empty
		return string(PrivateNICStateAvailable)
	}
	return string(enum)
}

func (enum PrivateNICState) Values() []PrivateNICState {
	return []PrivateNICState{
		"available",
		"syncing",
		"syncing_error",
	}
}

func (enum PrivateNICState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PrivateNICState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PrivateNICState(PrivateNICState(tmp).String())
	return nil
}

type SecurityGroupPolicy string

const (
	SecurityGroupPolicyUnknownPolicy = SecurityGroupPolicy("unknown_policy")
	SecurityGroupPolicyAccept        = SecurityGroupPolicy("accept")
	SecurityGroupPolicyDrop          = SecurityGroupPolicy("drop")
)

func (enum SecurityGroupPolicy) String() string {
	if enum == "" {
		// return default value if empty
		return string(SecurityGroupPolicyUnknownPolicy)
	}
	return string(enum)
}

func (enum SecurityGroupPolicy) Values() []SecurityGroupPolicy {
	return []SecurityGroupPolicy{
		"unknown_policy",
		"accept",
		"drop",
	}
}

func (enum SecurityGroupPolicy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SecurityGroupPolicy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SecurityGroupPolicy(SecurityGroupPolicy(tmp).String())
	return nil
}

type SecurityGroupRuleAction string

const (
	SecurityGroupRuleActionUnknownAction = SecurityGroupRuleAction("unknown_action")
	SecurityGroupRuleActionAccept        = SecurityGroupRuleAction("accept")
	SecurityGroupRuleActionDrop          = SecurityGroupRuleAction("drop")
)

func (enum SecurityGroupRuleAction) String() string {
	if enum == "" {
		// return default value if empty
		return string(SecurityGroupRuleActionUnknownAction)
	}
	return string(enum)
}

func (enum SecurityGroupRuleAction) Values() []SecurityGroupRuleAction {
	return []SecurityGroupRuleAction{
		"unknown_action",
		"accept",
		"drop",
	}
}

func (enum SecurityGroupRuleAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SecurityGroupRuleAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SecurityGroupRuleAction(SecurityGroupRuleAction(tmp).String())
	return nil
}

type SecurityGroupRuleDirection string

const (
	SecurityGroupRuleDirectionUnknownDirection = SecurityGroupRuleDirection("unknown_direction")
	SecurityGroupRuleDirectionInbound          = SecurityGroupRuleDirection("inbound")
	SecurityGroupRuleDirectionOutbound         = SecurityGroupRuleDirection("outbound")
)

func (enum SecurityGroupRuleDirection) String() string {
	if enum == "" {
		// return default value if empty
		return string(SecurityGroupRuleDirectionUnknownDirection)
	}
	return string(enum)
}

func (enum SecurityGroupRuleDirection) Values() []SecurityGroupRuleDirection {
	return []SecurityGroupRuleDirection{
		"unknown_direction",
		"inbound",
		"outbound",
	}
}

func (enum SecurityGroupRuleDirection) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SecurityGroupRuleDirection) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SecurityGroupRuleDirection(SecurityGroupRuleDirection(tmp).String())
	return nil
}

type SecurityGroupRuleProtocol string

const (
	SecurityGroupRuleProtocolUnknownProtocol = SecurityGroupRuleProtocol("unknown_protocol")
	SecurityGroupRuleProtocolTCP             = SecurityGroupRuleProtocol("TCP")
	SecurityGroupRuleProtocolUDP             = SecurityGroupRuleProtocol("UDP")
	SecurityGroupRuleProtocolICMP            = SecurityGroupRuleProtocol("ICMP")
	SecurityGroupRuleProtocolANY             = SecurityGroupRuleProtocol("ANY")
)

func (enum SecurityGroupRuleProtocol) String() string {
	if enum == "" {
		// return default value if empty
		return string(SecurityGroupRuleProtocolUnknownProtocol)
	}
	return string(enum)
}

func (enum SecurityGroupRuleProtocol) Values() []SecurityGroupRuleProtocol {
	return []SecurityGroupRuleProtocol{
		"unknown_protocol",
		"TCP",
		"UDP",
		"ICMP",
		"ANY",
	}
}

func (enum SecurityGroupRuleProtocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SecurityGroupRuleProtocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SecurityGroupRuleProtocol(SecurityGroupRuleProtocol(tmp).String())
	return nil
}

type SecurityGroupState string

const (
	SecurityGroupStateAvailable    = SecurityGroupState("available")
	SecurityGroupStateSyncing      = SecurityGroupState("syncing")
	SecurityGroupStateSyncingError = SecurityGroupState("syncing_error")
)

func (enum SecurityGroupState) String() string {
	if enum == "" {
		// return default value if empty
		return string(SecurityGroupStateAvailable)
	}
	return string(enum)
}

func (enum SecurityGroupState) Values() []SecurityGroupState {
	return []SecurityGroupState{
		"available",
		"syncing",
		"syncing_error",
	}
}

func (enum SecurityGroupState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SecurityGroupState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SecurityGroupState(SecurityGroupState(tmp).String())
	return nil
}

type ServerAction string

const (
	ServerActionPoweron        = ServerAction("poweron")
	ServerActionBackup         = ServerAction("backup")
	ServerActionStopInPlace    = ServerAction("stop_in_place")
	ServerActionPoweroff       = ServerAction("poweroff")
	ServerActionTerminate      = ServerAction("terminate")
	ServerActionReboot         = ServerAction("reboot")
	ServerActionEnableRoutedIP = ServerAction("enable_routed_ip")
)

func (enum ServerAction) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerActionPoweron)
	}
	return string(enum)
}

func (enum ServerAction) Values() []ServerAction {
	return []ServerAction{
		"poweron",
		"backup",
		"stop_in_place",
		"poweroff",
		"terminate",
		"reboot",
		"enable_routed_ip",
	}
}

func (enum ServerAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerAction(ServerAction(tmp).String())
	return nil
}

type ServerFilesystemState string

const (
	ServerFilesystemStateUnknownState = ServerFilesystemState("unknown_state")
	ServerFilesystemStateAttaching    = ServerFilesystemState("attaching")
	ServerFilesystemStateAvailable    = ServerFilesystemState("available")
	ServerFilesystemStateDetaching    = ServerFilesystemState("detaching")
)

func (enum ServerFilesystemState) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerFilesystemStateUnknownState)
	}
	return string(enum)
}

func (enum ServerFilesystemState) Values() []ServerFilesystemState {
	return []ServerFilesystemState{
		"unknown_state",
		"attaching",
		"available",
		"detaching",
	}
}

func (enum ServerFilesystemState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerFilesystemState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerFilesystemState(ServerFilesystemState(tmp).String())
	return nil
}

type ServerIPIPFamily string

const (
	ServerIPIPFamilyInet  = ServerIPIPFamily("inet")
	ServerIPIPFamilyInet6 = ServerIPIPFamily("inet6")
)

func (enum ServerIPIPFamily) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerIPIPFamilyInet)
	}
	return string(enum)
}

func (enum ServerIPIPFamily) Values() []ServerIPIPFamily {
	return []ServerIPIPFamily{
		"inet",
		"inet6",
	}
}

func (enum ServerIPIPFamily) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerIPIPFamily) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerIPIPFamily(ServerIPIPFamily(tmp).String())
	return nil
}

type ServerIPProvisioningMode string

const (
	ServerIPProvisioningModeManual = ServerIPProvisioningMode("manual")
	ServerIPProvisioningModeDHCP   = ServerIPProvisioningMode("dhcp")
	ServerIPProvisioningModeSlaac  = ServerIPProvisioningMode("slaac")
)

func (enum ServerIPProvisioningMode) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerIPProvisioningModeManual)
	}
	return string(enum)
}

func (enum ServerIPProvisioningMode) Values() []ServerIPProvisioningMode {
	return []ServerIPProvisioningMode{
		"manual",
		"dhcp",
		"slaac",
	}
}

func (enum ServerIPProvisioningMode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerIPProvisioningMode) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerIPProvisioningMode(ServerIPProvisioningMode(tmp).String())
	return nil
}

type ServerIPState string

const (
	ServerIPStateUnknownState = ServerIPState("unknown_state")
	ServerIPStateDetached     = ServerIPState("detached")
	ServerIPStateAttached     = ServerIPState("attached")
	ServerIPStatePending      = ServerIPState("pending")
	ServerIPStateError        = ServerIPState("error")
)

func (enum ServerIPState) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerIPStateUnknownState)
	}
	return string(enum)
}

func (enum ServerIPState) Values() []ServerIPState {
	return []ServerIPState{
		"unknown_state",
		"detached",
		"attached",
		"pending",
		"error",
	}
}

func (enum ServerIPState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerIPState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerIPState(ServerIPState(tmp).String())
	return nil
}

type ServerState string

const (
	ServerStateRunning        = ServerState("running")
	ServerStateStopped        = ServerState("stopped")
	ServerStateStoppedInPlace = ServerState("stopped in place")
	ServerStateStarting       = ServerState("starting")
	ServerStateStopping       = ServerState("stopping")
	ServerStateLocked         = ServerState("locked")
)

func (enum ServerState) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerStateRunning)
	}
	return string(enum)
}

func (enum ServerState) Values() []ServerState {
	return []ServerState{
		"running",
		"stopped",
		"stopped in place",
		"starting",
		"stopping",
		"locked",
	}
}

func (enum ServerState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerState(ServerState(tmp).String())
	return nil
}

type ServerTypesAvailability string

const (
	ServerTypesAvailabilityAvailable = ServerTypesAvailability("available")
	ServerTypesAvailabilityScarce    = ServerTypesAvailability("scarce")
	ServerTypesAvailabilityShortage  = ServerTypesAvailability("shortage")
)

func (enum ServerTypesAvailability) String() string {
	if enum == "" {
		// return default value if empty
		return string(ServerTypesAvailabilityAvailable)
	}
	return string(enum)
}

func (enum ServerTypesAvailability) Values() []ServerTypesAvailability {
	return []ServerTypesAvailability{
		"available",
		"scarce",
		"shortage",
	}
}

func (enum ServerTypesAvailability) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerTypesAvailability) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerTypesAvailability(ServerTypesAvailability(tmp).String())
	return nil
}

type SnapshotState string

const (
	SnapshotStateAvailable    = SnapshotState("available")
	SnapshotStateSnapshotting = SnapshotState("snapshotting")
	SnapshotStateError        = SnapshotState("error")
	SnapshotStateInvalidData  = SnapshotState("invalid_data")
	SnapshotStateImporting    = SnapshotState("importing")
	SnapshotStateExporting    = SnapshotState("exporting")
)

func (enum SnapshotState) String() string {
	if enum == "" {
		// return default value if empty
		return string(SnapshotStateAvailable)
	}
	return string(enum)
}

func (enum SnapshotState) Values() []SnapshotState {
	return []SnapshotState{
		"available",
		"snapshotting",
		"error",
		"invalid_data",
		"importing",
		"exporting",
	}
}

func (enum SnapshotState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SnapshotState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SnapshotState(SnapshotState(tmp).String())
	return nil
}

type SnapshotVolumeType string

const (
	SnapshotVolumeTypeUnknownVolumeType = SnapshotVolumeType("unknown_volume_type")
	SnapshotVolumeTypeLSSD              = SnapshotVolumeType("l_ssd")
	SnapshotVolumeTypeBSSD              = SnapshotVolumeType("b_ssd")
	SnapshotVolumeTypeUnified           = SnapshotVolumeType("unified")
)

func (enum SnapshotVolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return string(SnapshotVolumeTypeUnknownVolumeType)
	}
	return string(enum)
}

func (enum SnapshotVolumeType) Values() []SnapshotVolumeType {
	return []SnapshotVolumeType{
		"unknown_volume_type",
		"l_ssd",
		"b_ssd",
		"unified",
	}
}

func (enum SnapshotVolumeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SnapshotVolumeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SnapshotVolumeType(SnapshotVolumeType(tmp).String())
	return nil
}

type TaskStatus string

const (
	TaskStatusPending = TaskStatus("pending")
	TaskStatusStarted = TaskStatus("started")
	TaskStatusSuccess = TaskStatus("success")
	TaskStatusFailure = TaskStatus("failure")
	TaskStatusRetry   = TaskStatus("retry")
)

func (enum TaskStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(TaskStatusPending)
	}
	return string(enum)
}

func (enum TaskStatus) Values() []TaskStatus {
	return []TaskStatus{
		"pending",
		"started",
		"success",
		"failure",
		"retry",
	}
}

func (enum TaskStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *TaskStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = TaskStatus(TaskStatus(tmp).String())
	return nil
}

type VolumeServerState string

const (
	VolumeServerStateAvailable    = VolumeServerState("available")
	VolumeServerStateSnapshotting = VolumeServerState("snapshotting")
	VolumeServerStateResizing     = VolumeServerState("resizing")
	VolumeServerStateFetching     = VolumeServerState("fetching")
	VolumeServerStateSaving       = VolumeServerState("saving")
	VolumeServerStateHotsyncing   = VolumeServerState("hotsyncing")
	VolumeServerStateAttaching    = VolumeServerState("attaching")
	VolumeServerStateError        = VolumeServerState("error")
)

func (enum VolumeServerState) String() string {
	if enum == "" {
		// return default value if empty
		return string(VolumeServerStateAvailable)
	}
	return string(enum)
}

func (enum VolumeServerState) Values() []VolumeServerState {
	return []VolumeServerState{
		"available",
		"snapshotting",
		"resizing",
		"fetching",
		"saving",
		"hotsyncing",
		"attaching",
		"error",
	}
}

func (enum VolumeServerState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *VolumeServerState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = VolumeServerState(VolumeServerState(tmp).String())
	return nil
}

type VolumeServerVolumeType string

const (
	VolumeServerVolumeTypeLSSD      = VolumeServerVolumeType("l_ssd")
	VolumeServerVolumeTypeBSSD      = VolumeServerVolumeType("b_ssd")
	VolumeServerVolumeTypeSbsVolume = VolumeServerVolumeType("sbs_volume")
	VolumeServerVolumeTypeScratch   = VolumeServerVolumeType("scratch")
)

func (enum VolumeServerVolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return string(VolumeServerVolumeTypeLSSD)
	}
	return string(enum)
}

func (enum VolumeServerVolumeType) Values() []VolumeServerVolumeType {
	return []VolumeServerVolumeType{
		"l_ssd",
		"b_ssd",
		"sbs_volume",
		"scratch",
	}
}

func (enum VolumeServerVolumeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *VolumeServerVolumeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = VolumeServerVolumeType(VolumeServerVolumeType(tmp).String())
	return nil
}

type VolumeState string

const (
	VolumeStateAvailable    = VolumeState("available")
	VolumeStateSnapshotting = VolumeState("snapshotting")
	VolumeStateFetching     = VolumeState("fetching")
	VolumeStateSaving       = VolumeState("saving")
	VolumeStateResizing     = VolumeState("resizing")
	VolumeStateHotsyncing   = VolumeState("hotsyncing")
	VolumeStateError        = VolumeState("error")
)

func (enum VolumeState) String() string {
	if enum == "" {
		// return default value if empty
		return string(VolumeStateAvailable)
	}
	return string(enum)
}

func (enum VolumeState) Values() []VolumeState {
	return []VolumeState{
		"available",
		"snapshotting",
		"fetching",
		"saving",
		"resizing",
		"hotsyncing",
		"error",
	}
}

func (enum VolumeState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *VolumeState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = VolumeState(VolumeState(tmp).String())
	return nil
}

type VolumeVolumeType string

const (
	VolumeVolumeTypeLSSD        = VolumeVolumeType("l_ssd")
	VolumeVolumeTypeBSSD        = VolumeVolumeType("b_ssd")
	VolumeVolumeTypeUnified     = VolumeVolumeType("unified")
	VolumeVolumeTypeScratch     = VolumeVolumeType("scratch")
	VolumeVolumeTypeSbsVolume   = VolumeVolumeType("sbs_volume")
	VolumeVolumeTypeSbsSnapshot = VolumeVolumeType("sbs_snapshot")
)

func (enum VolumeVolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return string(VolumeVolumeTypeLSSD)
	}
	return string(enum)
}

func (enum VolumeVolumeType) Values() []VolumeVolumeType {
	return []VolumeVolumeType{
		"l_ssd",
		"b_ssd",
		"unified",
		"scratch",
		"sbs_volume",
		"sbs_snapshot",
	}
}

func (enum VolumeVolumeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *VolumeVolumeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = VolumeVolumeType(VolumeVolumeType(tmp).String())
	return nil
}

// ServerSummary: server summary.
type ServerSummary struct {
	ID string `json:"id"`

	Name string `json:"name"`
}

// Bootscript: bootscript.
type Bootscript struct {
	// Architecture: default value: unknown_arch
	Architecture Arch `json:"architecture"`

	Bootcmdargs string `json:"bootcmdargs"`

	Default bool `json:"default"`

	Dtb string `json:"dtb"`

	ID string `json:"id"`

	Initrd string `json:"initrd"`

	Kernel string `json:"kernel"`

	Organization string `json:"organization"`

	Public bool `json:"public"`

	Title string `json:"title"`

	Project string `json:"project"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"zone"`
}

// Volume: volume.
type Volume struct {
	// ID: volume unique ID.
	ID string `json:"id"`

	// Name: volume name.
	Name string `json:"name"`

	// Deprecated: ExportURI: show the volume NBD export URI.
	ExportURI *string `json:"export_uri"`

	// Size: volume disk size.
	Size scw.Size `json:"size"`

	// VolumeType: volume type.
	// Default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type"`

	// CreationDate: volume creation date.
	CreationDate *time.Time `json:"creation_date"`

	// ModificationDate: volume modification date.
	ModificationDate *time.Time `json:"modification_date"`

	// Organization: volume Organization ID.
	Organization string `json:"organization"`

	// Project: volume Project ID.
	Project string `json:"project"`

	// Tags: volume tags.
	Tags []string `json:"tags"`

	// Server: instance attached to the volume.
	Server *ServerSummary `json:"server"`

	// State: volume state.
	// Default value: available
	State VolumeState `json:"state"`

	// Zone: zone in which the volume is located.
	Zone scw.Zone `json:"zone"`
}

// VolumeSummary: volume summary.
type VolumeSummary struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Size scw.Size `json:"size"`

	// VolumeType: default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type"`
}

// ServerTypeNetworkInterface: server type network interface.
type ServerTypeNetworkInterface struct {
	// InternalBandwidth: maximum internal bandwidth in bits per seconds.
	InternalBandwidth *uint64 `json:"internal_bandwidth"`

	// InternetBandwidth: maximum internet bandwidth in bits per seconds.
	InternetBandwidth *uint64 `json:"internet_bandwidth"`
}

// ServerTypeVolumeConstraintSizes: server type volume constraint sizes.
type ServerTypeVolumeConstraintSizes struct {
	// MinSize: minimum volume size in bytes.
	MinSize scw.Size `json:"min_size"`

	// MaxSize: maximum volume size in bytes.
	MaxSize scw.Size `json:"max_size"`
}

// Image: image.
type Image struct {
	ID string `json:"id"`

	Name string `json:"name"`

	// Arch: default value: unknown_arch
	Arch Arch `json:"arch"`

	CreationDate *time.Time `json:"creation_date"`

	ModificationDate *time.Time `json:"modification_date"`

	// Deprecated
	DefaultBootscript *Bootscript `json:"default_bootscript"`

	ExtraVolumes map[string]*Volume `json:"extra_volumes"`

	FromServer string `json:"from_server"`

	Organization string `json:"organization"`

	Public bool `json:"public"`

	RootVolume *VolumeSummary `json:"root_volume"`

	// State: default value: available
	State ImageState `json:"state"`

	Project string `json:"project"`

	Tags []string `json:"tags"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"zone"`
}

// PlacementGroup: placement group.
type PlacementGroup struct {
	// ID: placement group unique ID.
	ID string `json:"id"`

	// Name: placement group name.
	Name string `json:"name"`

	// Organization: placement group Organization ID.
	Organization string `json:"organization"`

	// Project: placement group Project ID.
	Project string `json:"project"`

	// Tags: placement group tags.
	Tags []string `json:"tags"`

	// PolicyMode: select the failure mode when the placement cannot be respected, either optional or enforced.
	// Default value: optional
	PolicyMode PlacementGroupPolicyMode `json:"policy_mode"`

	// PolicyType: select the behavior of the placement group, either low_latency (group) or max_availability (spread).
	// Default value: max_availability
	PolicyType PlacementGroupPolicyType `json:"policy_type"`

	// PolicyRespected: in the server endpoints the value is always false as it is deprecated.
	// In the placement group endpoints the value is correct.
	PolicyRespected bool `json:"policy_respected"`

	// Zone: zone in which the placement group is located.
	Zone scw.Zone `json:"zone"`
}

// PrivateNIC: private nic.
type PrivateNIC struct {
	// ID: private NIC unique ID.
	ID string `json:"id"`

	// ServerID: instance to which the private NIC is attached.
	ServerID string `json:"server_id"`

	// PrivateNetworkID: private Network the private NIC is attached to.
	PrivateNetworkID string `json:"private_network_id"`

	// MacAddress: private NIC MAC address.
	MacAddress string `json:"mac_address"`

	// State: private NIC state.
	// Default value: available
	State PrivateNICState `json:"state"`

	// Tags: private NIC tags.
	Tags []string `json:"tags"`
}

// SecurityGroupSummary: security group summary.
type SecurityGroupSummary struct {
	ID string `json:"id"`

	Name string `json:"name"`
}

// ServerFilesystem: server filesystem.
type ServerFilesystem struct {
	FilesystemID string `json:"filesystem_id"`

	// State: default value: unknown_state
	State ServerFilesystemState `json:"state"`
}

// ServerIP: server ip.
type ServerIP struct {
	// ID: unique ID of the IP address.
	ID string `json:"id"`

	// Address: instance's public IP-Address.
	Address net.IP `json:"address"`

	// Gateway: gateway's IP address.
	Gateway net.IP `json:"gateway"`

	// Netmask: cIDR netmask.
	Netmask string `json:"netmask"`

	// Family: IP address family (inet or inet6).
	// Default value: inet
	Family ServerIPIPFamily `json:"family"`

	// Dynamic: true if the IP address is dynamic.
	Dynamic bool `json:"dynamic"`

	// ProvisioningMode: information about this address provisioning mode.
	// Default value: manual
	ProvisioningMode ServerIPProvisioningMode `json:"provisioning_mode"`

	// Tags: tags associated with the IP.
	Tags []string `json:"tags"`

	// IpamID: the ip_id of an IPAM ip if the ip is created from IPAM, null if not.
	IpamID string `json:"ipam_id"`

	// State: IP address state.
	// Default value: unknown_state
	State ServerIPState `json:"state"`
}

// ServerIPv6: server i pv6.
type ServerIPv6 struct {
	// Address: instance IPv6 IP-Address.
	Address net.IP `json:"address"`

	// Gateway: iPv6 IP-addresses gateway.
	Gateway net.IP `json:"gateway"`

	// Netmask: iPv6 IP-addresses CIDR netmask.
	Netmask string `json:"netmask"`
}

// ServerLocation: server location.
type ServerLocation struct {
	ClusterID string `json:"cluster_id"`

	HypervisorID string `json:"hypervisor_id"`

	NodeID string `json:"node_id"`

	PlatformID string `json:"platform_id"`

	ZoneID string `json:"zone_id"`
}

// ServerMaintenance: server maintenance.
type ServerMaintenance struct {
	Reason string `json:"reason"`

	StartDate *time.Time `json:"start_date"`
}

// VolumeServer: volume server.
type VolumeServer struct {
	ID string `json:"id"`

	Name *string `json:"name"`

	// Deprecated
	ExportURI *string `json:"export_uri"`

	Organization *string `json:"organization"`

	Server *ServerSummary `json:"server"`

	Size *scw.Size `json:"size"`

	// VolumeType: default value: l_ssd
	VolumeType VolumeServerVolumeType `json:"volume_type"`

	CreationDate *time.Time `json:"creation_date"`

	ModificationDate *time.Time `json:"modification_date"`

	// State: default value: available
	State *VolumeServerState `json:"state"`

	Project *string `json:"project"`

	Boot bool `json:"boot"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"zone"`
}

// SnapshotBaseVolume: snapshot base volume.
type SnapshotBaseVolume struct {
	// ID: volume ID on which the snapshot is based.
	ID string `json:"id"`

	// Name: volume name on which the snapshot is based on.
	Name string `json:"name"`
}

// ServerTypeCapabilities: server type capabilities.
type ServerTypeCapabilities struct {
	// BlockStorage: defines whether the Instance supports block storage.
	BlockStorage *bool `json:"block_storage"`

	// BootTypes: list of supported boot types.
	BootTypes []BootType `json:"boot_types"`

	// MaxFileSystems: max number of SFS (Scaleway File Systems) that can be attached to the Instance.
	MaxFileSystems uint32 `json:"max_file_systems"`
}

// ServerTypeGPUInfo: server type gpu info.
type ServerTypeGPUInfo struct {
	// GpuManufacturer: gPU manufacturer.
	GpuManufacturer string `json:"gpu_manufacturer"`

	// GpuName: gPU model name.
	GpuName string `json:"gpu_name"`

	// GpuMemory: RAM of a single GPU, in bytes.
	GpuMemory scw.Size `json:"gpu_memory"`
}

// ServerTypeNetwork: server type network.
type ServerTypeNetwork struct {
	// Interfaces: list of available network interfaces.
	Interfaces []*ServerTypeNetworkInterface `json:"interfaces"`

	// SumInternalBandwidth: total maximum internal bandwidth in bits per seconds.
	SumInternalBandwidth *uint64 `json:"sum_internal_bandwidth"`

	// SumInternetBandwidth: total maximum internet bandwidth in bits per seconds.
	SumInternetBandwidth *uint64 `json:"sum_internet_bandwidth"`

	// IPv6Support: true if IPv6 is enabled.
	IPv6Support bool `json:"ipv6_support"`
}

// ServerTypeVolumeConstraintsByType: server type volume constraints by type.
type ServerTypeVolumeConstraintsByType struct {
	// LSSD: local SSD volumes.
	LSSD *ServerTypeVolumeConstraintSizes `json:"l_ssd"`
}

// VolumeTypeCapabilities: volume type capabilities.
type VolumeTypeCapabilities struct {
	Snapshot bool `json:"snapshot"`
}

// VolumeTypeConstraints: volume type constraints.
type VolumeTypeConstraints struct {
	Min scw.Size `json:"min"`

	Max scw.Size `json:"max"`
}

// Server: server.
type Server struct {
	// ID: instance unique ID.
	ID string `json:"id"`

	// Name: instance name.
	Name string `json:"name"`

	// Organization: instance Organization ID.
	Organization string `json:"organization"`

	// Project: instance Project ID.
	Project string `json:"project"`

	// AllowedActions: list of allowed actions on the Instance.
	AllowedActions []ServerAction `json:"allowed_actions"`

	// Tags: tags associated with the Instance.
	Tags []string `json:"tags"`

	// CommercialType: instance commercial type (eg. GP1-M).
	CommercialType string `json:"commercial_type"`

	// CreationDate: instance creation date.
	CreationDate *time.Time `json:"creation_date"`

	// DynamicIPRequired: true if a dynamic IPv4 is required.
	DynamicIPRequired bool `json:"dynamic_ip_required"`

	// Deprecated: RoutedIPEnabled: true to configure the instance so it uses the routed IP mode. Use of `routed_ip_enabled` as `False` is deprecated.
	RoutedIPEnabled *bool `json:"routed_ip_enabled"`

	// Deprecated: EnableIPv6: true if IPv6 is enabled (deprecated and always `False` when `routed_ip_enabled` is `True`).
	EnableIPv6 *bool `json:"enable_ipv6"`

	// Hostname: instance host name.
	Hostname string `json:"hostname"`

	// Image: information about the Instance image.
	Image *Image `json:"image"`

	// Protected: defines whether the Instance protection option is activated.
	Protected bool `json:"protected"`

	// PrivateIP: private IP address of the Instance (deprecated and always `null` when `routed_ip_enabled` is `True`).
	PrivateIP *string `json:"private_ip"`

	// Deprecated: PublicIP: information about the public IP (deprecated in favor of `public_ips`).
	PublicIP *ServerIP `json:"public_ip"`

	// PublicIPs: information about all the public IPs attached to the server.
	PublicIPs []*ServerIP `json:"public_ips"`

	// MacAddress: the server's MAC address.
	MacAddress string `json:"mac_address"`

	// ModificationDate: instance modification date.
	ModificationDate *time.Time `json:"modification_date"`

	// State: instance state.
	// Default value: running
	State ServerState `json:"state"`

	// Location: instance location.
	Location *ServerLocation `json:"location"`

	// Deprecated: IPv6: instance IPv6 address (deprecated when `routed_ip_enabled` is `True`).
	IPv6 *ServerIPv6 `json:"ipv6"`

	// BootType: instance boot type.
	// Default value: local
	BootType BootType `json:"boot_type"`

	// Volumes: instance volumes.
	Volumes map[string]*VolumeServer `json:"volumes"`

	// SecurityGroup: instance security group.
	SecurityGroup *SecurityGroupSummary `json:"security_group"`

	// Maintenances: instance planned maintenance.
	Maintenances []*ServerMaintenance `json:"maintenances"`

	// StateDetail: detailed information about the Instance state.
	StateDetail string `json:"state_detail"`

	// Arch: instance architecture.
	// Default value: unknown_arch
	Arch Arch `json:"arch"`

	// PlacementGroup: instance placement group.
	PlacementGroup *PlacementGroup `json:"placement_group"`

	// PrivateNics: instance private NICs.
	PrivateNics []*PrivateNIC `json:"private_nics"`

	// Zone: zone in which the Instance is located.
	Zone scw.Zone `json:"zone"`

	// AdminPasswordEncryptionSSHKeyID: the public_key value of this key is used to encrypt the admin password. When set to an empty string, reset this value and admin_password_encrypted_value to an empty string so a new password may be generated.
	AdminPasswordEncryptionSSHKeyID *string `json:"admin_password_encryption_ssh_key_id"`

	// AdminPasswordEncryptedValue: this value is reset when admin_password_encryption_ssh_key_id is set to an empty string.
	AdminPasswordEncryptedValue *string `json:"admin_password_encrypted_value"`

	// Filesystems: list of attached filesystems.
	Filesystems []*ServerFilesystem `json:"filesystems"`

	// EndOfService: true if the Instance type has reached end of service.
	EndOfService bool `json:"end_of_service"`
}

// IP: ip.
type IP struct {
	ID string `json:"id"`

	Address net.IP `json:"address"`

	Reverse *string `json:"reverse"`

	Server *ServerSummary `json:"server"`

	Organization string `json:"organization"`

	Tags []string `json:"tags"`

	Project string `json:"project"`

	// Type: default value: unknown_iptype
	Type IPType `json:"type"`

	// State: default value: unknown_state
	State IPState `json:"state"`

	Prefix scw.IPNet `json:"prefix"`

	IpamID string `json:"ipam_id"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"zone"`
}

// VolumeTemplate: volume template.
type VolumeTemplate struct {
	// ID: UUID of the volume.
	ID string `json:"id,omitempty"`

	// Name: name of the volume.
	Name string `json:"name,omitempty"`

	// Size: disk size of the volume, must be a multiple of 512.
	Size scw.Size `json:"size,omitempty"`

	// VolumeType: type of the volume.
	// Default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type,omitempty"`

	// Deprecated: Organization: organization ID of the volume.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID of the volume.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`
}

// SecurityGroup: security group.
type SecurityGroup struct {
	// ID: security group unique ID.
	ID string `json:"id"`

	// Name: security group name.
	Name string `json:"name"`

	// Description: security group description.
	Description string `json:"description"`

	// EnableDefaultSecurity: true if SMTP is blocked on IPv4 and IPv6. This feature is read only, please open a support ticket if you need to make it configurable.
	EnableDefaultSecurity bool `json:"enable_default_security"`

	// InboundDefaultPolicy: default inbound policy.
	// Default value: unknown_policy
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy"`

	// OutboundDefaultPolicy: default outbound policy.
	// Default value: unknown_policy
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy"`

	// Organization: security group Organization ID.
	Organization string `json:"organization"`

	// Project: security group Project ID.
	Project string `json:"project"`

	// Tags: security group tags.
	Tags []string `json:"tags"`

	// Deprecated: OrganizationDefault: true if it is your default security group for this Organization ID.
	OrganizationDefault *bool `json:"organization_default"`

	// ProjectDefault: true if it is your default security group for this Project ID.
	ProjectDefault bool `json:"project_default"`

	// CreationDate: security group creation date.
	CreationDate *time.Time `json:"creation_date"`

	// ModificationDate: security group modification date.
	ModificationDate *time.Time `json:"modification_date"`

	// Servers: list of Instances attached to this security group.
	Servers []*ServerSummary `json:"servers"`

	// Stateful: defines whether the security group is stateful.
	Stateful bool `json:"stateful"`

	// State: security group state.
	// Default value: available
	State SecurityGroupState `json:"state"`

	// Zone: zone in which the security group is located.
	Zone scw.Zone `json:"zone"`
}

// SecurityGroupRule: security group rule.
type SecurityGroupRule struct {
	ID string `json:"id"`

	// Protocol: default value: unknown_protocol
	Protocol SecurityGroupRuleProtocol `json:"protocol"`

	// Direction: default value: unknown_direction
	Direction SecurityGroupRuleDirection `json:"direction"`

	// Action: default value: unknown_action
	Action SecurityGroupRuleAction `json:"action"`

	IPRange scw.IPNet `json:"ip_range"`

	DestPortFrom *uint32 `json:"dest_port_from"`

	DestPortTo *uint32 `json:"dest_port_to"`

	Position uint32 `json:"position"`

	Editable bool `json:"editable"`

	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"zone"`
}

// VolumeServerTemplate: volume server template.
type VolumeServerTemplate struct {
	// ID: UUID of the volume.
	ID *string `json:"id,omitempty"`

	// Boot: force the Instance to boot on this volume.
	Boot *bool `json:"boot,omitempty"`

	// Name: name of the volume.
	Name *string `json:"name,omitempty"`

	// Size: disk size of the volume, must be a multiple of 512.
	Size *scw.Size `json:"size,omitempty"`

	// VolumeType: type of the volume.
	// Default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type,omitempty"`

	// BaseSnapshot: ID of the snapshot on which this volume will be based.
	BaseSnapshot *string `json:"base_snapshot,omitempty"`

	// Organization: organization ID of the volume.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID of the volume.
	Project *string `json:"project,omitempty"`
}

// Snapshot: snapshot.
type Snapshot struct {
	// ID: snapshot ID.
	ID string `json:"id"`

	// Name: snapshot name.
	Name string `json:"name"`

	// Organization: snapshot Organization ID.
	Organization string `json:"organization"`

	// Project: snapshot Project ID.
	Project string `json:"project"`

	// Tags: snapshot tags.
	Tags []string `json:"tags"`

	// VolumeType: snapshot volume type.
	// Default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type"`

	// Size: snapshot size.
	Size scw.Size `json:"size"`

	// State: snapshot state.
	// Default value: available
	State SnapshotState `json:"state"`

	// BaseVolume: volume on which the snapshot is based on.
	BaseVolume *SnapshotBaseVolume `json:"base_volume"`

	// CreationDate: snapshot creation date.
	CreationDate *time.Time `json:"creation_date"`

	// ModificationDate: snapshot modification date.
	ModificationDate *time.Time `json:"modification_date"`

	// Zone: snapshot zone.
	Zone scw.Zone `json:"zone"`

	// ErrorReason: reason for the failed snapshot import.
	ErrorReason *string `json:"error_reason"`
}

// Task: task.
type Task struct {
	// ID: unique ID of the task.
	ID string `json:"id"`

	// Description: description of the task.
	Description string `json:"description"`

	// Progress: progress of the task in percent.
	Progress int32 `json:"progress"`

	// StartedAt: task start date.
	StartedAt *time.Time `json:"started_at"`

	// TerminatedAt: task end date.
	TerminatedAt *time.Time `json:"terminated_at"`

	// Status: task status.
	// Default value: pending
	Status TaskStatus `json:"status"`

	HrefFrom string `json:"href_from"`

	HrefResult string `json:"href_result"`

	// Zone: zone in which the task is executed.
	Zone scw.Zone `json:"zone"`
}

// Dashboard: dashboard.
type Dashboard struct {
	VolumesCount uint32 `json:"volumes_count"`

	RunningServersCount uint32 `json:"running_servers_count"`

	ServersByTypes map[string]uint32 `json:"servers_by_types"`

	ImagesCount uint32 `json:"images_count"`

	SnapshotsCount uint32 `json:"snapshots_count"`

	ServersCount uint32 `json:"servers_count"`

	IPsCount uint32 `json:"ips_count"`

	SecurityGroupsCount uint32 `json:"security_groups_count"`

	IPsUnused uint32 `json:"ips_unused"`

	VolumesLSSDCount uint32 `json:"volumes_l_ssd_count"`

	// Deprecated
	VolumesBSSDCount *uint32 `json:"volumes_b_ssd_count"`

	VolumesLSSDTotalSize scw.Size `json:"volumes_l_ssd_total_size"`

	// Deprecated
	VolumesBSSDTotalSize *scw.Size `json:"volumes_b_ssd_total_size"`

	PrivateNicsCount uint32 `json:"private_nics_count"`

	PlacementGroupsCount uint32 `json:"placement_groups_count"`
}

// PlacementGroupServer: placement group server.
type PlacementGroupServer struct {
	// ID: instance UUID.
	ID string `json:"id"`

	// Name: instance name.
	Name string `json:"name"`

	// PolicyRespected: defines whether the placement group policy is respected (either 1 or 0).
	PolicyRespected bool `json:"policy_respected"`
}

// GetServerTypesAvailabilityResponseAvailability: get server types availability response availability.
type GetServerTypesAvailabilityResponseAvailability struct {
	// Availability: default value: available
	Availability ServerTypesAvailability `json:"availability"`
}

// ServerType: server type.
type ServerType struct {
	// Deprecated: MonthlyPrice: estimated monthly price, for a 30 days month, in Euro.
	MonthlyPrice *float32 `json:"monthly_price"`

	// HourlyPrice: hourly price in Euro.
	HourlyPrice float32 `json:"hourly_price"`

	// AltNames: alternative Instance name, if any.
	AltNames []string `json:"alt_names"`

	// PerVolumeConstraint: additional volume constraints.
	PerVolumeConstraint *ServerTypeVolumeConstraintsByType `json:"per_volume_constraint"`

	// VolumesConstraint: initial volume constraints.
	VolumesConstraint *ServerTypeVolumeConstraintSizes `json:"volumes_constraint"`

	// Ncpus: number of CPU.
	Ncpus uint32 `json:"ncpus"`

	// Gpu: number of GPU.
	Gpu *uint64 `json:"gpu"`

	// RAM: available RAM in bytes.
	RAM uint64 `json:"ram"`

	// GpuInfo: gPU information.
	GpuInfo *ServerTypeGPUInfo `json:"gpu_info"`

	// Arch: CPU architecture.
	// Default value: unknown_arch
	Arch Arch `json:"arch"`

	// Network: network available for the Instance.
	Network *ServerTypeNetwork `json:"network"`

	// Capabilities: capabilities.
	Capabilities *ServerTypeCapabilities `json:"capabilities"`

	// ScratchStorageMaxSize: maximum available scratch storage.
	ScratchStorageMaxSize *scw.Size `json:"scratch_storage_max_size"`

	// BlockBandwidth: the maximum bandwidth allocated to block storage access (in bytes per second).
	BlockBandwidth *uint64 `json:"block_bandwidth"`

	// EndOfService: true if this Instance type has reached end of service.
	EndOfService bool `json:"end_of_service"`
}

// VolumeType: volume type.
type VolumeType struct {
	DisplayName string `json:"display_name"`

	Capabilities *VolumeTypeCapabilities `json:"capabilities"`

	Constraints *VolumeTypeConstraints `json:"constraints"`
}

// ServerActionRequestVolumeBackupTemplate: server action request volume backup template.
type ServerActionRequestVolumeBackupTemplate struct {
	// VolumeType: overrides the `volume_type` of the snapshot for this volume.
	// If omitted, the volume type of the original volume will be used.
	// Default value: unknown_volume_type
	VolumeType SnapshotVolumeType `json:"volume_type,omitempty"`
}

// SetSecurityGroupRulesRequestRule: set security group rules request rule.
type SetSecurityGroupRulesRequestRule struct {
	// ID: UUID of the security rule to update. If no value is provided, a new rule will be created.
	ID *string `json:"id"`

	// Action: action to apply when the rule matches a packet.
	// Default value: unknown_action
	Action SecurityGroupRuleAction `json:"action"`

	// Protocol: protocol family this rule applies to.
	// Default value: unknown_protocol
	Protocol SecurityGroupRuleProtocol `json:"protocol"`

	// Direction: direction the rule applies to.
	// Default value: unknown_direction
	Direction SecurityGroupRuleDirection `json:"direction"`

	// IPRange: range of IP addresses these rules apply to.
	IPRange scw.IPNet `json:"ip_range"`

	// DestPortFrom: beginning of the range of ports this rule applies to (inclusive). This value will be set to null if protocol is ICMP or ANY.
	DestPortFrom *uint32 `json:"dest_port_from"`

	// DestPortTo: end of the range of ports this rule applies to (inclusive). This value will be set to null if protocol is ICMP or ANY, or if it is equal to dest_port_from.
	DestPortTo *uint32 `json:"dest_port_to"`

	// Position: position of this rule in the security group rules list. If several rules are passed with the same position, the resulting order is undefined.
	Position uint32 `json:"position"`

	// Editable: indicates if this rule is editable. Rules with the value false will be ignored.
	Editable *bool `json:"editable"`

	// Zone: zone of the rule. This field is ignored.
	Zone *scw.Zone `json:"zone"`
}

// NullableStringValue: nullable string value.
type NullableStringValue struct {
	Null bool `json:"null,omitempty"`

	Value string `json:"value,omitempty"`
}

// VolumeImageUpdateTemplate: volume image update template.
type VolumeImageUpdateTemplate struct {
	// ID: UUID of the snapshot.
	ID string `json:"id,omitempty"`
}

// SecurityGroupTemplate: security group template.
type SecurityGroupTemplate struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

// ApplyBlockMigrationRequest: apply block migration request.
type ApplyBlockMigrationRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: the volume to migrate, along with potentially other resources, according to the migration plan generated with a call to the [Get a volume or snapshot's migration plan](#path-volumes-get-a-volume-or-snapshots-migration-plan) endpoint.
	// Precisely one of VolumeID, SnapshotID must be set.
	VolumeID *string `json:"volume_id,omitempty"`

	// SnapshotID: the snapshot to migrate, along with potentially other resources, according to the migration plan generated with a call to the [Get a volume or snapshot's migration plan](#path-volumes-get-a-volume-or-snapshots-migration-plan) endpoint.
	// Precisely one of VolumeID, SnapshotID must be set.
	SnapshotID *string `json:"snapshot_id,omitempty"`

	// ValidationKey: a value to be retrieved from a call to the [Get a volume or snapshot's migration plan](#path-volumes-get-a-volume-or-snapshots-migration-plan) endpoint, to confirm that the volume and/or snapshots specified in said plan should be migrated.
	ValidationKey string `json:"validation_key,omitempty"`
}

// AttachServerFileSystemRequest: attach server file system request.
type AttachServerFileSystemRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	FilesystemID string `json:"filesystem_id,omitempty"`
}

// AttachServerFileSystemResponse: attach server file system response.
type AttachServerFileSystemResponse struct {
	Server *Server `json:"server"`
}

// AttachServerVolumeRequest: attach server volume request.
type AttachServerVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	VolumeID string `json:"volume_id,omitempty"`

	// VolumeType: default value: unknown_volume_type
	VolumeType AttachServerVolumeRequestVolumeType `json:"volume_type,omitempty"`

	Boot *bool `json:"boot,omitempty"`
}

// AttachServerVolumeResponse: attach server volume response.
type AttachServerVolumeResponse struct {
	Server *Server `json:"server"`
}

// CheckBlockMigrationOrganizationQuotasRequest: check block migration organization quotas request.
type CheckBlockMigrationOrganizationQuotasRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	Organization string `json:"organization,omitempty"`
}

// CreateIPRequest: create ip request.
type CreateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Deprecated: Organization: organization ID in which the IP is reserved.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID in which the IP is reserved.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: tags of the IP.
	Tags []string `json:"tags,omitempty"`

	// Server: UUID of the Instance you want to attach the IP to.
	Server *string `json:"server,omitempty"`

	// Type: IP type to reserve (either 'routed_ipv4' or 'routed_ipv6').
	// Default value: unknown_iptype
	Type IPType `json:"type,omitempty"`
}

// CreateIPResponse: create ip response.
type CreateIPResponse struct {
	IP *IP `json:"ip"`
}

// CreateImageRequest: create image request.
type CreateImageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the image.
	Name string `json:"name,omitempty"`

	// RootVolume: UUID of the snapshot.
	RootVolume string `json:"root_volume,omitempty"`

	// Arch: architecture of the image.
	// Default value: unknown_arch
	Arch Arch `json:"arch,omitempty"`

	// ExtraVolumes: additional volumes of the image.
	ExtraVolumes map[string]*VolumeTemplate `json:"extra_volumes,omitempty"`

	// Deprecated: Organization: organization ID of the image.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID of the image.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: tags of the image.
	Tags []string `json:"tags,omitempty"`

	// Public: true to create a public image.
	Public *bool `json:"public,omitempty"`
}

// CreateImageResponse: create image response.
type CreateImageResponse struct {
	Image *Image `json:"image"`
}

// CreatePlacementGroupRequest: create placement group request.
type CreatePlacementGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the placement group.
	Name string `json:"name,omitempty"`

	// Deprecated: Organization: organization ID of the placement group.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID of the placement group.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: tags of the placement group.
	Tags []string `json:"tags,omitempty"`

	// PolicyMode: operating mode of the placement group.
	// Default value: optional
	PolicyMode PlacementGroupPolicyMode `json:"policy_mode,omitempty"`

	// PolicyType: policy type of the placement group.
	// Default value: max_availability
	PolicyType PlacementGroupPolicyType `json:"policy_type,omitempty"`
}

// CreatePlacementGroupResponse: create placement group response.
type CreatePlacementGroupResponse struct {
	PlacementGroup *PlacementGroup `json:"placement_group"`
}

// CreatePrivateNICRequest: create private nic request.
type CreatePrivateNICRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance the private NIC will be attached to.
	ServerID string `json:"-"`

	// PrivateNetworkID: UUID of the private network where the private NIC will be attached.
	PrivateNetworkID string `json:"private_network_id,omitempty"`

	// Tags: private NIC tags.
	Tags []string `json:"tags,omitempty"`

	// Deprecated: IPIDs: ip_ids defined from IPAM.
	IPIDs *[]string `json:"ip_ids,omitempty"`

	// IpamIPIDs: UUID of IPAM ips, to be attached to the instance in the requested private network.
	IpamIPIDs []string `json:"ipam_ip_ids,omitempty"`
}

// CreatePrivateNICResponse: create private nic response.
type CreatePrivateNICResponse struct {
	PrivateNic *PrivateNIC `json:"private_nic"`
}

// CreateSecurityGroupRequest: create security group request.
type CreateSecurityGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the security group.
	Name string `json:"name,omitempty"`

	// Description: description of the security group.
	Description string `json:"description,omitempty"`

	// Deprecated: Organization: organization ID the security group belongs to.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID the security group belong to.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: tags of the security group.
	Tags []string `json:"tags,omitempty"`

	// Deprecated: OrganizationDefault: defines whether this security group becomes the default security group for new Instances.
	// Precisely one of OrganizationDefault, ProjectDefault must be set.
	OrganizationDefault *bool `json:"organization_default,omitempty"`

	// ProjectDefault: whether this security group becomes the default security group for new Instances.
	// Precisely one of OrganizationDefault, ProjectDefault must be set.
	ProjectDefault *bool `json:"project_default,omitempty"`

	// Stateful: whether the security group is stateful or not.
	Stateful bool `json:"stateful,omitempty"`

	// InboundDefaultPolicy: default policy for inbound rules.
	// Default value: unknown_policy
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`

	// OutboundDefaultPolicy: default policy for outbound rules.
	// Default value: unknown_policy
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`

	// EnableDefaultSecurity: true to block SMTP on IPv4 and IPv6. This feature is read only, please open a support ticket if you need to make it configurable.
	EnableDefaultSecurity *bool `json:"enable_default_security,omitempty"`
}

// CreateSecurityGroupResponse: create security group response.
type CreateSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group"`
}

// CreateSecurityGroupRuleRequest: create security group rule request.
type CreateSecurityGroupRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group.
	SecurityGroupID string `json:"-"`

	// Protocol: default value: unknown_protocol
	Protocol SecurityGroupRuleProtocol `json:"protocol,omitempty"`

	// Direction: default value: unknown_direction
	Direction SecurityGroupRuleDirection `json:"direction,omitempty"`

	// Action: default value: unknown_action
	Action SecurityGroupRuleAction `json:"action,omitempty"`

	IPRange scw.IPNet `json:"ip_range,omitempty"`

	// DestPortFrom: beginning of the range of ports to apply this rule to (inclusive).
	DestPortFrom *uint32 `json:"dest_port_from,omitempty"`

	// DestPortTo: end of the range of ports to apply this rule to (inclusive).
	DestPortTo *uint32 `json:"dest_port_to,omitempty"`

	// Position: position of this rule in the security group rules list.
	Position uint32 `json:"position,omitempty"`

	// Editable: indicates if this rule is editable (will be ignored).
	Editable bool `json:"editable,omitempty"`
}

// CreateSecurityGroupRuleResponse: create security group rule response.
type CreateSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule"`
}

// CreateServerRequest: create server request.
type CreateServerRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: instance name.
	Name string `json:"name,omitempty"`

	// DynamicIPRequired: by default, `dynamic_ip_required` is true, a dynamic ip is attached to the instance (if no flexible ip is already attached).
	DynamicIPRequired *bool `json:"dynamic_ip_required,omitempty"`

	// Deprecated: RoutedIPEnabled: if true, configure the Instance so it uses the new routed IP mode.
	RoutedIPEnabled *bool `json:"routed_ip_enabled,omitempty"`

	// CommercialType: define the Instance commercial type (i.e. GP1-S).
	CommercialType string `json:"commercial_type,omitempty"`

	// Image: instance image ID or label.
	Image *string `json:"image,omitempty"`

	// Volumes: volumes attached to the server.
	Volumes map[string]*VolumeServerTemplate `json:"volumes,omitempty"`

	// Deprecated: EnableIPv6: true if IPv6 is enabled on the server (deprecated and always `False` when `routed_ip_enabled` is `True`).
	EnableIPv6 *bool `json:"enable_ipv6,omitempty"`

	// Deprecated: PublicIP: ID of the reserved IP to attach to the Instance.
	PublicIP *string `json:"public_ip,omitempty"`

	// PublicIPs: a list of reserved IP IDs to attach to the Instance.
	PublicIPs *[]string `json:"public_ips,omitempty"`

	// BootType: boot type to use.
	// Default value: local
	BootType *BootType `json:"boot_type,omitempty"`

	// Deprecated: Organization: instance Organization ID.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: instance Project ID.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: instance tags.
	Tags []string `json:"tags,omitempty"`

	// SecurityGroup: security group ID.
	SecurityGroup *string `json:"security_group,omitempty"`

	// PlacementGroup: placement group ID if Instance must be part of a placement group.
	PlacementGroup *string `json:"placement_group,omitempty"`

	// AdminPasswordEncryptionSSHKeyID: the public_key value of this key is used to encrypt the admin password.
	AdminPasswordEncryptionSSHKeyID *string `json:"admin_password_encryption_ssh_key_id,omitempty"`

	// Protected: true to activate server protection option.
	Protected bool `json:"protected,omitempty"`
}

// CreateServerResponse: create server response.
type CreateServerResponse struct {
	Server *Server `json:"server"`
}

// CreateSnapshotRequest: create snapshot request.
type CreateSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the snapshot.
	Name string `json:"name,omitempty"`

	// VolumeID: UUID of the volume.
	VolumeID *string `json:"volume_id,omitempty"`

	// Tags: tags of the snapshot.
	Tags *[]string `json:"tags,omitempty"`

	// Deprecated: Organization: organization ID of the snapshot.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: project ID of the snapshot.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// VolumeType: overrides the volume_type of the snapshot.
	// If omitted, the volume type of the original volume will be used.
	// Default value: unknown_volume_type
	VolumeType SnapshotVolumeType `json:"volume_type,omitempty"`

	// Bucket: bucket name for snapshot imports.
	Bucket *string `json:"bucket,omitempty"`

	// Key: object key for snapshot imports.
	Key *string `json:"key,omitempty"`

	// Size: imported snapshot size, must be a multiple of 512.
	Size *scw.Size `json:"size,omitempty"`
}

// CreateSnapshotResponse: create snapshot response.
type CreateSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot"`

	Task *Task `json:"task"`
}

// CreateVolumeRequest: create volume request.
type CreateVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: volume name.
	Name string `json:"name,omitempty"`

	// Deprecated: Organization: volume Organization ID.
	// Precisely one of Project, Organization must be set.
	Organization *string `json:"organization,omitempty"`

	// Project: volume Project ID.
	// Precisely one of Project, Organization must be set.
	Project *string `json:"project,omitempty"`

	// Tags: volume tags.
	Tags []string `json:"tags,omitempty"`

	// VolumeType: volume type.
	// Default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type,omitempty"`

	// Size: volume disk size, must be a multiple of 512.
	// Precisely one of Size, BaseSnapshot must be set.
	Size *scw.Size `json:"size,omitempty"`

	// BaseSnapshot: ID of the snapshot on which this volume will be based.
	// Precisely one of Size, BaseSnapshot must be set.
	BaseSnapshot *string `json:"base_snapshot,omitempty"`
}

// CreateVolumeResponse: create volume response.
type CreateVolumeResponse struct {
	Volume *Volume `json:"volume"`
}

// DeleteIPRequest: delete ip request.
type DeleteIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IP: ID or address of the IP to delete.
	IP string `json:"-"`
}

// DeleteImageRequest: delete image request.
type DeleteImageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ImageID: UUID of the image you want to delete.
	ImageID string `json:"-"`
}

// DeletePlacementGroupRequest: delete placement group request.
type DeletePlacementGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group you want to delete.
	PlacementGroupID string `json:"-"`
}

// DeletePrivateNICRequest: delete private nic request.
type DeletePrivateNICRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: instance to which the private NIC is attached.
	ServerID string `json:"-"`

	// PrivateNicID: private NIC unique ID.
	PrivateNicID string `json:"-"`
}

// DeleteSecurityGroupRequest: delete security group request.
type DeleteSecurityGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group you want to delete.
	SecurityGroupID string `json:"-"`
}

// DeleteSecurityGroupRuleRequest: delete security group rule request.
type DeleteSecurityGroupRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`
}

// DeleteServerRequest: delete server request.
type DeleteServerRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// DeleteServerUserDataRequest: delete server user data request.
type DeleteServerUserDataRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance.
	ServerID string `json:"-"`

	// Key: key of the user data to delete.
	Key string `json:"-"`
}

// DeleteSnapshotRequest: delete snapshot request.
type DeleteSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot you want to delete.
	SnapshotID string `json:"-"`
}

// DeleteVolumeRequest: delete volume request.
type DeleteVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume you want to delete.
	VolumeID string `json:"-"`
}

// DetachServerFileSystemRequest: detach server file system request.
type DetachServerFileSystemRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	FilesystemID string `json:"filesystem_id,omitempty"`
}

// DetachServerFileSystemResponse: detach server file system response.
type DetachServerFileSystemResponse struct {
	Server *Server `json:"server"`
}

// DetachServerVolumeRequest: detach server volume request.
type DetachServerVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	VolumeID string `json:"volume_id,omitempty"`
}

// DetachServerVolumeResponse: detach server volume response.
type DetachServerVolumeResponse struct {
	Server *Server `json:"server"`
}

// ExportSnapshotRequest: export snapshot request.
type ExportSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: snapshot ID.
	SnapshotID string `json:"-"`

	// Bucket: object Storage bucket name.
	Bucket string `json:"bucket,omitempty"`

	// Key: object key.
	Key string `json:"key,omitempty"`
}

// ExportSnapshotResponse: export snapshot response.
type ExportSnapshotResponse struct {
	Task *Task `json:"task"`
}

// GetDashboardRequest: get dashboard request.
type GetDashboardRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	Project *string `json:"-"`
}

// GetDashboardResponse: get dashboard response.
type GetDashboardResponse struct {
	Dashboard *Dashboard `json:"dashboard"`
}

// GetIPRequest: get ip request.
type GetIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IP: IP ID or address to get.
	IP string `json:"-"`
}

// GetIPResponse: get ip response.
type GetIPResponse struct {
	IP *IP `json:"ip"`
}

// GetImageRequest: get image request.
type GetImageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ImageID: UUID of the image you want to get.
	ImageID string `json:"-"`
}

// GetImageResponse: get image response.
type GetImageResponse struct {
	Image *Image `json:"image"`
}

// GetPlacementGroupRequest: get placement group request.
type GetPlacementGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group you want to get.
	PlacementGroupID string `json:"-"`
}

// GetPlacementGroupResponse: get placement group response.
type GetPlacementGroupResponse struct {
	PlacementGroup *PlacementGroup `json:"placement_group"`
}

// GetPlacementGroupServersRequest: get placement group servers request.
type GetPlacementGroupServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group you want to get.
	PlacementGroupID string `json:"-"`
}

// GetPlacementGroupServersResponse: get placement group servers response.
type GetPlacementGroupServersResponse struct {
	// Servers: instances attached to the placement group.
	Servers []*PlacementGroupServer `json:"servers"`
}

// GetPrivateNICRequest: get private nic request.
type GetPrivateNICRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: instance to which the private NIC is attached.
	ServerID string `json:"-"`

	// PrivateNicID: private NIC unique ID.
	PrivateNicID string `json:"-"`
}

// GetPrivateNICResponse: get private nic response.
type GetPrivateNICResponse struct {
	PrivateNic *PrivateNIC `json:"private_nic"`
}

// GetSecurityGroupRequest: get security group request.
type GetSecurityGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group you want to get.
	SecurityGroupID string `json:"-"`
}

// GetSecurityGroupResponse: get security group response.
type GetSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group"`
}

// GetSecurityGroupRuleRequest: get security group rule request.
type GetSecurityGroupRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`
}

// GetSecurityGroupRuleResponse: get security group rule response.
type GetSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule"`
}

// GetServerCompatibleTypesRequest: get server compatible types request.
type GetServerCompatibleTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance you want to get.
	ServerID string `json:"-"`
}

// GetServerRequest: get server request.
type GetServerRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance you want to get.
	ServerID string `json:"-"`
}

// GetServerResponse: get server response.
type GetServerResponse struct {
	Server *Server `json:"server"`
}

// GetServerTypesAvailabilityRequest: get server types availability request.
type GetServerTypesAvailabilityRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`
}

// GetServerTypesAvailabilityResponse: get server types availability response.
type GetServerTypesAvailabilityResponse struct {
	// Servers: map of server types.
	Servers map[string]*GetServerTypesAvailabilityResponseAvailability `json:"servers"`

	TotalCount uint32 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *GetServerTypesAvailabilityResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *GetServerTypesAvailabilityResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*GetServerTypesAvailabilityResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	if r.Servers == nil {
		r.Servers = make(map[string]*GetServerTypesAvailabilityResponseAvailability)
	}
	for k, v := range results.Servers {
		r.Servers[k] = v
	}
	r.TotalCount += uint32(len(results.Servers))
	return uint32(len(results.Servers)), nil
}

// GetSnapshotRequest: get snapshot request.
type GetSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot you want to get.
	SnapshotID string `json:"-"`
}

// GetSnapshotResponse: get snapshot response.
type GetSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot"`
}

// GetVolumeRequest: get volume request.
type GetVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume you want to get.
	VolumeID string `json:"-"`
}

// GetVolumeResponse: get volume response.
type GetVolumeResponse struct {
	Volume *Volume `json:"volume"`
}

// ListDefaultSecurityGroupRulesRequest: list default security group rules request.
type ListDefaultSecurityGroupRulesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`
}

// ListIPsRequest: list i ps request.
type ListIPsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Project: project ID in which the IPs are reserved.
	Project *string `json:"-"`

	// Organization: organization ID in which the IPs are reserved.
	Organization *string `json:"-"`

	// Tags: filter IPs with these exact tags (to filter with several tags, use commas to separate them).
	Tags []string `json:"-"`

	// Name: filter on the IP address (Works as a LIKE operation on the IP address).
	Name *string `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`

	// Type: filter on the IP Mobility IP type (whose value should be either 'routed_ipv4' or 'routed_ipv6').
	Type *string `json:"-"`
}

// ListIPsResponse: list i ps response.
type ListIPsResponse struct {
	// TotalCount: total number of ips.
	TotalCount uint32 `json:"total_count"`

	// IPs: list of ips.
	IPs []*IP `json:"ips"`
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

// ListImagesRequest: list images request.
type ListImagesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *uint32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`

	Public *bool `json:"-"`

	Arch *string `json:"-"`

	Project *string `json:"-"`

	Tags *string `json:"-"`
}

// ListImagesResponse: list images response.
type ListImagesResponse struct {
	// TotalCount: total number of images.
	TotalCount uint32 `json:"total_count"`

	// Images: list of images.
	Images []*Image `json:"images"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return uint32(len(results.Images)), nil
}

// ListPlacementGroupsRequest: list placement groups request.
type ListPlacementGroupsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`

	// Organization: list only placement groups of this Organization ID.
	Organization *string `json:"-"`

	// Project: list only placement groups of this Project ID.
	Project *string `json:"-"`

	// Tags: list placement groups with these exact tags (to filter with several tags, use commas to separate them).
	Tags []string `json:"-"`

	// Name: filter placement groups by name (for eg. "cluster1" will return "cluster100" and "cluster1" but not "foo").
	Name *string `json:"-"`
}

// ListPlacementGroupsResponse: list placement groups response.
type ListPlacementGroupsResponse struct {
	// TotalCount: total number of placement groups.
	TotalCount uint32 `json:"total_count"`

	// PlacementGroups: list of placement groups.
	PlacementGroups []*PlacementGroup `json:"placement_groups"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPlacementGroupsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPlacementGroupsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListPlacementGroupsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PlacementGroups = append(r.PlacementGroups, results.PlacementGroups...)
	r.TotalCount += uint32(len(results.PlacementGroups))
	return uint32(len(results.PlacementGroups)), nil
}

// ListPrivateNICsRequest: list private ni cs request.
type ListPrivateNICsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: instance to which the private NIC is attached.
	ServerID string `json:"-"`

	// Tags: private NIC tags.
	Tags []string `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`
}

// ListPrivateNICsResponse: list private ni cs response.
type ListPrivateNICsResponse struct {
	PrivateNics []*PrivateNIC `json:"private_nics"`

	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPrivateNICsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPrivateNICsResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListPrivateNICsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PrivateNics = append(r.PrivateNics, results.PrivateNics...)
	r.TotalCount += uint64(len(results.PrivateNics))
	return uint64(len(results.PrivateNics)), nil
}

// ListSecurityGroupRulesRequest: list security group rules request.
type ListSecurityGroupRulesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group.
	SecurityGroupID string `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`
}

// ListSecurityGroupRulesResponse: list security group rules response.
type ListSecurityGroupRulesResponse struct {
	// TotalCount: total number of security groups.
	TotalCount uint32 `json:"total_count"`

	// Rules: list of security rules.
	Rules []*SecurityGroupRule `json:"rules"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupRulesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSecurityGroupRulesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListSecurityGroupRulesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Rules = append(r.Rules, results.Rules...)
	r.TotalCount += uint32(len(results.Rules))
	return uint32(len(results.Rules)), nil
}

// ListSecurityGroupsRequest: list security groups request.
type ListSecurityGroupsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the security group.
	Name *string `json:"-"`

	// Organization: security group Organization ID.
	Organization *string `json:"-"`

	// Project: security group Project ID.
	Project *string `json:"-"`

	// Tags: list security groups with these exact tags (to filter with several tags, use commas to separate them).
	Tags []string `json:"-"`

	// ProjectDefault: filter security groups with this value for project_default.
	ProjectDefault *bool `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`
}

// ListSecurityGroupsResponse: list security groups response.
type ListSecurityGroupsResponse struct {
	// TotalCount: total number of security groups.
	TotalCount uint32 `json:"total_count"`

	// SecurityGroups: list of security groups.
	SecurityGroups []*SecurityGroup `json:"security_groups"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSecurityGroupsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListSecurityGroupsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.SecurityGroups = append(r.SecurityGroups, results.SecurityGroups...)
	r.TotalCount += uint32(len(results.SecurityGroups))
	return uint32(len(results.SecurityGroups)), nil
}

// ListServerActionsRequest: list server actions request.
type ListServerActionsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// ListServerActionsResponse: list server actions response.
type ListServerActionsResponse struct {
	Actions []ServerAction `json:"actions"`
}

// ListServerUserDataRequest: list server user data request.
type ListServerUserDataRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance.
	ServerID string `json:"-"`
}

// ListServerUserDataResponse: list server user data response.
type ListServerUserDataResponse struct {
	UserData []string `json:"user_data"`
}

// ListServersRequest: list servers request.
type ListServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`

	// Organization: list only Instances of this Organization ID.
	Organization *string `json:"-"`

	// Project: list only Instances of this Project ID.
	Project *string `json:"-"`

	// Name: filter Instances by name (eg. "server1" will return "server100" and "server1" but not "foo").
	Name *string `json:"-"`

	// Deprecated: PrivateIP: list Instances by private_ip.
	PrivateIP *net.IP `json:"-"`

	// WithoutIP: list Instances that are not attached to a public IP.
	WithoutIP *bool `json:"-"`

	// WithIP: list Instances by IP (both private_ip and public_ip are supported).
	WithIP *net.IP `json:"-"`

	// CommercialType: list Instances of this commercial type.
	CommercialType *string `json:"-"`

	// State: list Instances in this state.
	// Default value: running
	State *ServerState `json:"-"`

	// Tags: list Instances with these exact tags (to filter with several tags, use commas to separate them).
	Tags []string `json:"-"`

	// PrivateNetwork: list Instances in this Private Network.
	PrivateNetwork *string `json:"-"`

	// Order: define the order of the returned servers.
	// Default value: creation_date_desc
	Order ListServersRequestOrder `json:"-"`

	// PrivateNetworks: list Instances from the given Private Networks (use commas to separate them).
	PrivateNetworks []string `json:"-"`

	// PrivateNicMacAddress: list Instances associated with the given private NIC MAC address.
	PrivateNicMacAddress *string `json:"-"`

	// Servers: list Instances from these server ids (use commas to separate them).
	Servers []string `json:"-"`
}

// ListServersResponse: list servers response.
type ListServersResponse struct {
	// TotalCount: total number of Instances.
	TotalCount uint32 `json:"total_count"`

	// Servers: list of Instances.
	Servers []*Server `json:"servers"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListServersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Servers = append(r.Servers, results.Servers...)
	r.TotalCount += uint32(len(results.Servers))
	return uint32(len(results.Servers)), nil
}

// ListServersTypesRequest: list servers types request.
type ListServersTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	PerPage *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListServersTypesResponse: list servers types response.
type ListServersTypesResponse struct {
	// TotalCount: total number of Instance types.
	TotalCount uint32 `json:"total_count"`

	// Servers: list of Instance types.
	Servers map[string]*ServerType `json:"servers"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListServersTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServersTypesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListServersTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	if r.Servers == nil {
		r.Servers = make(map[string]*ServerType)
	}
	for k, v := range results.Servers {
		r.Servers[k] = v
	}
	r.TotalCount += uint32(len(results.Servers))
	return uint32(len(results.Servers)), nil
}

// ListSnapshotsRequest: list snapshots request.
type ListSnapshotsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Organization: list snapshots only for this Organization ID.
	Organization *string `json:"-"`

	// Project: list snapshots only for this Project ID.
	Project *string `json:"-"`

	// PerPage: number of snapshots returned per page (positive integer lower or equal to 100).
	PerPage *uint32 `json:"-"`

	// Page: page to be returned.
	Page *int32 `json:"-"`

	// Name: list snapshots of the requested name.
	Name *string `json:"-"`

	// Tags: list snapshots that have the requested tag.
	Tags *string `json:"-"`

	// BaseVolumeID: list snapshots originating only from this volume.
	BaseVolumeID *string `json:"-"`
}

// ListSnapshotsResponse: list snapshots response.
type ListSnapshotsResponse struct {
	// TotalCount: total number of snapshots.
	TotalCount uint32 `json:"total_count"`

	// Snapshots: list of snapshots.
	Snapshots []*Snapshot `json:"snapshots"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListSnapshotsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Snapshots = append(r.Snapshots, results.Snapshots...)
	r.TotalCount += uint32(len(results.Snapshots))
	return uint32(len(results.Snapshots)), nil
}

// ListVolumesRequest: list volumes request.
type ListVolumesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeType: filter by volume type.
	// Default value: l_ssd
	VolumeType *VolumeVolumeType `json:"-"`

	// PerPage: a positive integer lower or equal to 100 to select the number of items to return.
	PerPage *uint32 `json:"-"`

	// Page: a positive integer to choose the page to return.
	Page *int32 `json:"-"`

	// Organization: filter volume by Organization ID.
	Organization *string `json:"-"`

	// Project: filter volume by Project ID.
	Project *string `json:"-"`

	// Tags: filter volumes with these exact tags (to filter with several tags, use commas to separate them).
	Tags []string `json:"-"`

	// Name: filter volume by name (for eg. "vol" will return "myvolume" but not "data").
	Name *string `json:"-"`
}

// ListVolumesResponse: list volumes response.
type ListVolumesResponse struct {
	// TotalCount: total number of volumes.
	TotalCount uint32 `json:"total_count"`

	// Volumes: list of volumes.
	Volumes []*Volume `json:"volumes"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListVolumesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Volumes = append(r.Volumes, results.Volumes...)
	r.TotalCount += uint32(len(results.Volumes))
	return uint32(len(results.Volumes)), nil
}

// ListVolumesTypesRequest: list volumes types request.
type ListVolumesTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	PerPage *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListVolumesTypesResponse: list volumes types response.
type ListVolumesTypesResponse struct {
	// TotalCount: total number of volume types.
	TotalCount uint32 `json:"total_count"`

	// Volumes: map of volume types.
	Volumes map[string]*VolumeType `json:"volumes"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVolumesTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVolumesTypesResponse) UnsafeAppend(res any) (uint32, error) {
	results, ok := res.(*ListVolumesTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	if r.Volumes == nil {
		r.Volumes = make(map[string]*VolumeType)
	}
	for k, v := range results.Volumes {
		r.Volumes[k] = v
	}
	r.TotalCount += uint32(len(results.Volumes))
	return uint32(len(results.Volumes)), nil
}

// MigrationPlan: migration plan.
type MigrationPlan struct {
	// Volume: a volume which will be migrated to SBS together with the snapshots, if present.
	Volume *Volume `json:"volume"`

	// Snapshots: a list of snapshots which will be migrated to SBS together and with the volume, if present.
	Snapshots []*Snapshot `json:"snapshots"`

	// ValidationKey: a value to be passed to the call to the [Migrate a volume and/or snapshots to SBS](#path-volumes-migrate-a-volume-andor-snapshots-to-sbs-scaleway-block-storage) endpoint, to confirm that the execution of the plan is being requested.
	ValidationKey string `json:"validation_key"`
}

// PlanBlockMigrationRequest: plan block migration request.
type PlanBlockMigrationRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: the volume for which the migration plan will be generated.
	// Precisely one of VolumeID, SnapshotID must be set.
	VolumeID *string `json:"volume_id,omitempty"`

	// SnapshotID: the snapshot for which the migration plan will be generated.
	// Precisely one of VolumeID, SnapshotID must be set.
	SnapshotID *string `json:"snapshot_id,omitempty"`
}

// ReleaseIPToIpamRequest: release ip to ipam request.
type ReleaseIPToIpamRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IPID: ID of the IP you want to release from the Instance but retain in IPAM.
	IPID string `json:"-"`
}

// ServerActionRequest: server action request.
type ServerActionRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance.
	ServerID string `json:"-"`

	// Action: action to perform on the Instance.
	// Default value: poweron
	Action ServerAction `json:"action,omitempty"`

	// Name: name of the backup you want to create.
	// This field should only be specified when performing a backup action.
	Name *string `json:"name,omitempty"`

	// Volumes: for each volume UUID, the snapshot parameters of the volume.
	// This field should only be specified when performing a backup action.
	Volumes map[string]*ServerActionRequestVolumeBackupTemplate `json:"volumes,omitempty"`

	// DisableIPv6: disable IPv6 on the Instance while performing migration to routed IPs.
	// This field should only be specified when performing a enable_routed_ip action.
	DisableIPv6 *bool `json:"disable_ipv6,omitempty"`
}

// ServerActionResponse: server action response.
type ServerActionResponse struct {
	Task *Task `json:"task"`
}

// ServerCompatibleTypes: server compatible types.
type ServerCompatibleTypes struct {
	// CompatibleTypes: instance compatible types.
	CompatibleTypes []string `json:"compatible_types"`
}

// SetImageRequest: set image request.
type SetImageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	ID string `json:"-"`

	Name string `json:"name"`

	// Arch: default value: unknown_arch
	Arch Arch `json:"arch"`

	CreationDate *time.Time `json:"creation_date,omitempty"`

	ModificationDate *time.Time `json:"modification_date,omitempty"`

	// Deprecated
	DefaultBootscript *Bootscript `json:"default_bootscript,omitempty"`

	ExtraVolumes map[string]*Volume `json:"extra_volumes"`

	FromServer string `json:"from_server"`

	Organization string `json:"organization"`

	Public bool `json:"public"`

	RootVolume *VolumeSummary `json:"root_volume,omitempty"`

	// State: default value: available
	State ImageState `json:"state"`

	Project string `json:"project"`

	Tags *[]string `json:"tags,omitempty"`
}

// SetPlacementGroupRequest: set placement group request.
type SetPlacementGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	PlacementGroupID string `json:"-"`

	Name string `json:"name"`

	Organization string `json:"organization"`

	// PolicyMode: default value: optional
	PolicyMode PlacementGroupPolicyMode `json:"policy_mode"`

	// PolicyType: default value: max_availability
	PolicyType PlacementGroupPolicyType `json:"policy_type"`

	Project string `json:"project"`

	Tags *[]string `json:"tags,omitempty"`
}

// SetPlacementGroupResponse: set placement group response.
type SetPlacementGroupResponse struct {
	PlacementGroup *PlacementGroup `json:"placement_group"`
}

// SetPlacementGroupServersRequest: set placement group servers request.
type SetPlacementGroupServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group you want to set.
	PlacementGroupID string `json:"-"`

	// Servers: an array of the Instances' UUIDs you want to configure.
	Servers []string `json:"servers"`
}

// SetPlacementGroupServersResponse: set placement group servers response.
type SetPlacementGroupServersResponse struct {
	// Servers: instances attached to the placement group.
	Servers []*PlacementGroupServer `json:"servers"`
}

// SetSecurityGroupRulesRequest: set security group rules request.
type SetSecurityGroupRulesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group to update the rules on.
	SecurityGroupID string `json:"-"`

	// Rules: list of rules to update in the security group.
	Rules []*SetSecurityGroupRulesRequestRule `json:"rules"`
}

// SetSecurityGroupRulesResponse: set security group rules response.
type SetSecurityGroupRulesResponse struct {
	Rules []*SecurityGroupRule `json:"rules"`
}

// UpdateIPRequest: update ip request.
type UpdateIPRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// IP: IP ID or IP address.
	IP string `json:"-"`

	// Reverse: reverse domain name.
	Reverse *NullableStringValue `json:"reverse,omitempty"`

	// Type: should have no effect.
	// Default value: unknown_iptype
	Type IPType `json:"type,omitempty"`

	// Tags: an array of keywords you want to tag this IP with.
	Tags *[]string `json:"tags,omitempty"`

	Server *NullableStringValue `json:"server,omitempty"`
}

// UpdateIPResponse: update ip response.
type UpdateIPResponse struct {
	IP *IP `json:"ip"`
}

// UpdateImageRequest: update image request.
type UpdateImageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ImageID: UUID of the image.
	ImageID string `json:"-"`

	// Name: name of the image.
	Name *string `json:"name,omitempty"`

	// Arch: architecture of the image.
	// Default value: unknown_arch
	Arch Arch `json:"arch,omitempty"`

	// ExtraVolumes: additional snapshots of the image, with extra_volumeKey being the position of the snapshot in the image.
	ExtraVolumes map[string]*VolumeImageUpdateTemplate `json:"extra_volumes,omitempty"`

	// Tags: tags of the image.
	Tags *[]string `json:"tags,omitempty"`

	// Public: true to set the image as public.
	Public *bool `json:"public,omitempty"`
}

// UpdateImageResponse: update image response.
type UpdateImageResponse struct {
	Image *Image `json:"image"`
}

// UpdatePlacementGroupRequest: update placement group request.
type UpdatePlacementGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group.
	PlacementGroupID string `json:"-"`

	// Name: name of the placement group.
	Name *string `json:"name,omitempty"`

	// Tags: tags of the placement group.
	Tags *[]string `json:"tags,omitempty"`

	// PolicyMode: operating mode of the placement group.
	// Default value: optional
	PolicyMode *PlacementGroupPolicyMode `json:"policy_mode,omitempty"`

	// PolicyType: policy type of the placement group.
	// Default value: max_availability
	PolicyType *PlacementGroupPolicyType `json:"policy_type,omitempty"`
}

// UpdatePlacementGroupResponse: update placement group response.
type UpdatePlacementGroupResponse struct {
	PlacementGroup *PlacementGroup `json:"placement_group"`
}

// UpdatePlacementGroupServersRequest: update placement group servers request.
type UpdatePlacementGroupServersRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// PlacementGroupID: UUID of the placement group you want to update.
	PlacementGroupID string `json:"-"`

	// Servers: an array of the Instances' UUIDs you want to configure.
	Servers []string `json:"servers,omitempty"`
}

// UpdatePlacementGroupServersResponse: update placement group servers response.
type UpdatePlacementGroupServersResponse struct {
	// Servers: instances attached to the placement group.
	Servers []*PlacementGroupServer `json:"servers"`
}

// UpdatePrivateNICRequest: update private nic request.
type UpdatePrivateNICRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance the private NIC will be attached to.
	ServerID string `json:"-"`

	// PrivateNicID: private NIC unique ID.
	PrivateNicID string `json:"-"`

	// Tags: tags used to select private NIC/s.
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateSecurityGroupRequest: update security group request.
type UpdateSecurityGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group.
	SecurityGroupID string `json:"-"`

	// Name: name of the security group.
	Name *string `json:"name,omitempty"`

	// Description: description of the security group.
	Description *string `json:"description,omitempty"`

	// EnableDefaultSecurity: true to block SMTP on IPv4 and IPv6. This feature is read only, please open a support ticket if you need to make it configurable.
	EnableDefaultSecurity *bool `json:"enable_default_security,omitempty"`

	// InboundDefaultPolicy: default inbound policy.
	// Default value: unknown_policy
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`

	// Tags: tags of the security group.
	Tags *[]string `json:"tags,omitempty"`

	// Deprecated: OrganizationDefault: please use project_default instead.
	OrganizationDefault *bool `json:"organization_default,omitempty"`

	// ProjectDefault: true use this security group for future Instances created in this project.
	ProjectDefault *bool `json:"project_default,omitempty"`

	// OutboundDefaultPolicy: default outbound policy.
	// Default value: unknown_policy
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`

	// Stateful: true to set the security group as stateful.
	Stateful *bool `json:"stateful,omitempty"`
}

// UpdateSecurityGroupResponse: update security group response.
type UpdateSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group"`
}

// UpdateSecurityGroupRuleRequest: update security group rule request.
type UpdateSecurityGroupRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SecurityGroupID: UUID of the security group.
	SecurityGroupID string `json:"-"`

	// SecurityGroupRuleID: UUID of the rule.
	SecurityGroupRuleID string `json:"-"`

	// Protocol: protocol family this rule applies to.
	// Default value: unknown_protocol
	Protocol SecurityGroupRuleProtocol `json:"protocol,omitempty"`

	// Direction: direction the rule applies to.
	// Default value: unknown_direction
	Direction SecurityGroupRuleDirection `json:"direction,omitempty"`

	// Action: action to apply when the rule matches a packet.
	// Default value: unknown_action
	Action SecurityGroupRuleAction `json:"action,omitempty"`

	// IPRange: range of IP addresses these rules apply to.
	IPRange *scw.IPNet `json:"ip_range,omitempty"`

	// DestPortFrom: beginning of the range of ports this rule applies to (inclusive). If 0 is provided, unset the parameter.
	DestPortFrom *uint32 `json:"dest_port_from,omitempty"`

	// DestPortTo: end of the range of ports this rule applies to (inclusive). If 0 is provided, unset the parameter.
	DestPortTo *uint32 `json:"dest_port_to,omitempty"`

	// Position: position of this rule in the security group rules list.
	Position *uint32 `json:"position,omitempty"`
}

// UpdateSecurityGroupRuleResponse: update security group rule response.
type UpdateSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule"`
}

// UpdateServerRequest: update server request.
type UpdateServerRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ServerID: UUID of the Instance.
	ServerID string `json:"-"`

	// Name: name of the Instance.
	Name *string `json:"name,omitempty"`

	// BootType: default value: local
	BootType *BootType `json:"boot_type,omitempty"`

	// Tags: tags of the Instance.
	Tags *[]string `json:"tags,omitempty"`

	Volumes *map[string]*VolumeServerTemplate `json:"volumes,omitempty"`

	DynamicIPRequired *bool `json:"dynamic_ip_required,omitempty"`

	// Deprecated: RoutedIPEnabled: true to configure the instance so it uses the new routed IP mode (once this is set to True you cannot set it back to False).
	RoutedIPEnabled *bool `json:"routed_ip_enabled,omitempty"`

	// PublicIPs: a list of reserved IP IDs to attach to the Instance.
	PublicIPs *[]string `json:"public_ips,omitempty"`

	// Deprecated
	EnableIPv6 *bool `json:"enable_ipv6,omitempty"`

	// Protected: true to activate server protection option.
	Protected *bool `json:"protected,omitempty"`

	SecurityGroup *SecurityGroupTemplate `json:"security_group,omitempty"`

	// PlacementGroup: placement group ID if Instance must be part of a placement group.
	PlacementGroup *NullableStringValue `json:"placement_group,omitempty"`

	// PrivateNics: instance private NICs.
	PrivateNics *[]string `json:"private_nics,omitempty"`

	// CommercialType: warning: This field has some restrictions:
	// - Cannot be changed if the Instance is not in `stopped` state.
	// - Cannot be changed if the Instance is in a placement group.
	// - Cannot be changed from/to a Windows offer to/from a Linux offer.
	// - Local storage requirements of the target commercial_types must be fulfilled (i.e. if an Instance has 80GB of local storage, it can be changed into a GP1-XS, which has a maximum of 150GB, but it cannot be changed into a DEV1-S, which has only 20GB).
	CommercialType *string `json:"commercial_type,omitempty"`

	// AdminPasswordEncryptionSSHKeyID: the public_key value of this key is used to encrypt the admin password. When set to an empty string, reset this value and admin_password_encrypted_value to an empty string so a new password may be generated.
	AdminPasswordEncryptionSSHKeyID *string `json:"admin_password_encryption_ssh_key_id,omitempty"`
}

// UpdateServerResponse: update server response.
type UpdateServerResponse struct {
	Server *Server `json:"server"`
}

// UpdateSnapshotRequest: update snapshot request.
type UpdateSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot.
	SnapshotID string `json:"-"`

	// Name: name of the snapshot.
	Name *string `json:"name,omitempty"`

	// Tags: tags of the snapshot.
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateSnapshotResponse: update snapshot response.
type UpdateSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot"`
}

// UpdateVolumeRequest: update volume request.
type UpdateVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume.
	VolumeID string `json:"-"`

	// Name: volume name.
	Name *string `json:"name,omitempty"`

	// Tags: tags of the volume.
	Tags *[]string `json:"tags,omitempty"`

	// Size: volume disk size, must be a multiple of 512.
	Size *scw.Size `json:"size,omitempty"`
}

// UpdateVolumeResponse: update volume response.
type UpdateVolumeResponse struct {
	Volume *Volume `json:"volume"`
}

// setImageResponse: set image response.
type setImageResponse struct {
	Image *Image `json:"image"`
}

// setSecurityGroupRequest: set security group request.
type setSecurityGroupRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ID: UUID of the security group.
	ID string `json:"-"`

	// Name: name of the security group.
	Name string `json:"name"`

	// Tags: tags of the security group.
	Tags *[]string `json:"tags,omitempty"`

	// CreationDate: creation date of the security group (will be ignored).
	CreationDate *time.Time `json:"creation_date,omitempty"`

	// ModificationDate: modification date of the security group (will be ignored).
	ModificationDate *time.Time `json:"modification_date,omitempty"`

	// Description: description of the security group.
	Description string `json:"description"`

	// EnableDefaultSecurity: true to block SMTP on IPv4 and IPv6. This feature is read only, please open a support ticket if you need to make it configurable.
	EnableDefaultSecurity bool `json:"enable_default_security"`

	// InboundDefaultPolicy: default inbound policy.
	// Default value: unknown_policy
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy"`

	// OutboundDefaultPolicy: default outbound policy.
	// Default value: unknown_policy
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy"`

	// Organization: security groups Organization ID.
	Organization string `json:"organization"`

	// Project: security group Project ID.
	Project string `json:"project"`

	// Deprecated: OrganizationDefault: please use project_default instead.
	OrganizationDefault *bool `json:"organization_default,omitempty"`

	// ProjectDefault: true use this security group for future Instances created in this project.
	ProjectDefault bool `json:"project_default"`

	// Servers: instances attached to this security group.
	Servers []*ServerSummary `json:"servers"`

	// Stateful: true to set the security group as stateful.
	Stateful bool `json:"stateful"`
}

// setSecurityGroupResponse: set security group response.
type setSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group"`
}

// setSecurityGroupRuleRequest: set security group rule request.
type setSecurityGroupRuleRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`

	ID string `json:"id"`

	// Protocol: default value: unknown_protocol
	Protocol SecurityGroupRuleProtocol `json:"protocol"`

	// Direction: default value: unknown_direction
	Direction SecurityGroupRuleDirection `json:"direction"`

	// Action: default value: unknown_action
	Action SecurityGroupRuleAction `json:"action"`

	IPRange scw.IPNet `json:"ip_range"`

	DestPortFrom *uint32 `json:"dest_port_from,omitempty"`

	DestPortTo *uint32 `json:"dest_port_to,omitempty"`

	Position uint32 `json:"position"`

	Editable bool `json:"editable"`
}

// setSecurityGroupRuleResponse: set security group rule response.
type setSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule"`
}

// setServerRequest: set server request.
type setServerRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// ID: instance unique ID.
	ID string `json:"-"`

	// Name: instance name.
	Name string `json:"name"`

	// Organization: instance Organization ID.
	Organization string `json:"organization"`

	// Project: instance Project ID.
	Project string `json:"project"`

	// AllowedActions: provide a list of allowed actions on the server.
	AllowedActions []ServerAction `json:"allowed_actions"`

	// Tags: tags associated with the Instance.
	Tags *[]string `json:"tags,omitempty"`

	// CommercialType: instance commercial type (eg. GP1-M).
	CommercialType string `json:"commercial_type"`

	// CreationDate: instance creation date.
	CreationDate *time.Time `json:"creation_date,omitempty"`

	// DynamicIPRequired: true if a dynamic IPv4 is required.
	DynamicIPRequired bool `json:"dynamic_ip_required"`

	// Deprecated: RoutedIPEnabled: true to configure the instance so it uses the new routed IP mode (once this is set to True you cannot set it back to False).
	RoutedIPEnabled *bool `json:"routed_ip_enabled,omitempty"`

	// Deprecated: EnableIPv6: true if IPv6 is enabled (deprecated and always `False` when `routed_ip_enabled` is `True`).
	EnableIPv6 *bool `json:"enable_ipv6,omitempty"`

	// Hostname: instance host name.
	Hostname string `json:"hostname"`

	// Image: provide information on the Instance image.
	Image *Image `json:"image,omitempty"`

	// Protected: instance protection option is activated.
	Protected bool `json:"protected"`

	// Deprecated: PrivateIP: instance private IP address (deprecated and always `null` when `routed_ip_enabled` is `True`).
	PrivateIP *string `json:"private_ip,omitempty"`

	// Deprecated: PublicIP: information about the public IP (deprecated in favor of `public_ips`).
	PublicIP *ServerIP `json:"public_ip,omitempty"`

	// PublicIPs: information about all the public IPs attached to the server.
	PublicIPs []*ServerIP `json:"public_ips"`

	// ModificationDate: instance modification date.
	ModificationDate *time.Time `json:"modification_date,omitempty"`

	// State: instance state.
	// Default value: running
	State ServerState `json:"state"`

	// Location: instance location.
	Location *ServerLocation `json:"location,omitempty"`

	// Deprecated: IPv6: instance IPv6 address (deprecated when `routed_ip_enabled` is `True`).
	IPv6 *ServerIPv6 `json:"ipv6,omitempty"`

	// BootType: instance boot type.
	// Default value: local
	BootType BootType `json:"boot_type"`

	// Volumes: instance volumes.
	Volumes map[string]*Volume `json:"volumes"`

	// SecurityGroup: instance security group.
	SecurityGroup *SecurityGroupSummary `json:"security_group,omitempty"`

	// Maintenances: instance planned maintenances.
	Maintenances []*ServerMaintenance `json:"maintenances"`

	// StateDetail: instance state_detail.
	StateDetail string `json:"state_detail"`

	// Arch: instance architecture (refers to the CPU architecture used for the Instance, e.g. x86_64, arm64).
	// Default value: unknown_arch
	Arch Arch `json:"arch"`

	// PlacementGroup: instance placement group.
	PlacementGroup *PlacementGroup `json:"placement_group,omitempty"`

	// PrivateNics: instance private NICs.
	PrivateNics []*PrivateNIC `json:"private_nics"`

	// AdminPasswordEncryptionSSHKeyID: the public_key value of this key is used to encrypt the admin password. When set to an empty string, reset this value and admin_password_encrypted_value to an empty string so a new password may be generated.
	AdminPasswordEncryptionSSHKeyID *string `json:"admin_password_encryption_ssh_key_id,omitempty"`
}

// setServerResponse: set server response.
type setServerResponse struct {
	Server *Server `json:"server"`
}

// setSnapshotRequest: set snapshot request.
type setSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	SnapshotID string `json:"-"`

	ID string `json:"id"`

	Name string `json:"name"`

	Organization string `json:"organization"`

	// VolumeType: default value: l_ssd
	VolumeType VolumeVolumeType `json:"volume_type"`

	Size scw.Size `json:"size"`

	// State: default value: available
	State SnapshotState `json:"state"`

	BaseVolume *SnapshotBaseVolume `json:"base_volume,omitempty"`

	CreationDate *time.Time `json:"creation_date,omitempty"`

	ModificationDate *time.Time `json:"modification_date,omitempty"`

	Project string `json:"project"`

	Tags *[]string `json:"tags,omitempty"`
}

// setSnapshotResponse: set snapshot response.
type setSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot"`
}

// This API allows you to manage your CPU and GPU Instances.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

func (s *API) Zones() []scw.Zone {
	return []scw.Zone{scw.ZoneFrPar1, scw.ZoneFrPar2, scw.ZoneFrPar3, scw.ZoneNlAms1, scw.ZoneNlAms2, scw.ZoneNlAms3, scw.ZonePlWaw1, scw.ZonePlWaw2, scw.ZonePlWaw3}
}

// GetServerTypesAvailability: Get availability for all Instance types.
func (s *API) GetServerTypesAvailability(req *GetServerTypesAvailabilityRequest, opts ...scw.RequestOption) (*GetServerTypesAvailabilityResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/products/servers/availability",
		Query:  query,
	}

	var resp GetServerTypesAvailabilityResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListServersTypes: List available Instance types and their technical details.
func (s *API) ListServersTypes(req *ListServersTypesRequest, opts ...scw.RequestOption) (*ListServersTypesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/products/servers",
		Query:  query,
	}

	var resp ListServersTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVolumesTypes: List all volume types and their technical details.
func (s *API) ListVolumesTypes(req *ListVolumesTypesRequest, opts ...scw.RequestOption) (*ListVolumesTypesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/products/volumes",
		Query:  query,
	}

	var resp ListVolumesTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListServers: List all Instances in a specified Availability Zone, e.g. `fr-par-1`.
func (s *API) ListServers(req *ListServersRequest, opts ...scw.RequestOption) (*ListServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "private_ip", req.PrivateIP)
	parameter.AddToQuery(query, "without_ip", req.WithoutIP)
	parameter.AddToQuery(query, "with_ip", req.WithIP)
	parameter.AddToQuery(query, "commercial_type", req.CommercialType)
	parameter.AddToQuery(query, "state", req.State)
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "private_network", req.PrivateNetwork)
	parameter.AddToQuery(query, "order", req.Order)
	if len(req.PrivateNetworks) != 0 {
		parameter.AddToQuery(query, "private_networks", strings.Join(req.PrivateNetworks, ","))
	}
	parameter.AddToQuery(query, "private_nic_mac_address", req.PrivateNicMacAddress)
	if len(req.Servers) != 0 {
		parameter.AddToQuery(query, "servers", strings.Join(req.Servers, ","))
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers",
		Query:  query,
	}

	var resp ListServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// createServer: Create a new Instance of the specified commercial type in the specified zone. Pay attention to the volumes parameter, which takes an object which can be used in different ways to achieve different behaviors.
// Get more information in the [Technical Information](#technical-information) section of the introduction.
func (s *API) createServer(req *CreateServerRequest, opts ...scw.RequestOption) (*CreateServerResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("srv")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteServer: Delete the Instance with the specified ID.
func (s *API) DeleteServer(req *DeleteServerRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetServer: Get the details of a specified Instance.
func (s *API) GetServer(req *GetServerRequest, opts ...scw.RequestOption) (*GetServerResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
	}

	var resp GetServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// setServer:
func (s *API) setServer(req *setServerRequest, opts ...scw.RequestOption) (*setServerResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if req.Project == "" {
		defaultProject, _ := s.client.GetDefaultProjectID()
		req.Project = defaultProject
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// updateServer: Update the Instance information, such as name, boot mode, or tags.
func (s *API) updateServer(req *UpdateServerRequest, opts ...scw.RequestOption) (*UpdateServerResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListServerActions: List all actions (e.g. power on, power off, reboot) that can currently be performed on an Instance.
func (s *API) ListServerActions(req *ListServerActionsRequest, opts ...scw.RequestOption) (*ListServerActionsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/action",
	}

	var resp ListServerActionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ServerAction: Perform an action on an Instance.
// Available actions are:
// * `poweron`: Start a stopped Instance.
// * `poweroff`: Fully stop the Instance and release the hypervisor slot.
// * `stop_in_place`: Stop the Instance, but keep the slot on the hypervisor.
// * `reboot`: Stop the instance and restart it.
// * `backup`:  Create an image with all the volumes of an Instance.
// * `terminate`: Delete the Instance along with its attached local volumes.
// * `enable_routed_ip`: Migrate the Instance to the new network stack.
//
// The `terminate` action will result in the deletion of `l_ssd` and `scratch` volumes types, `sbs_volume` volumes will only be detached.
// If you want to preserve your `l_ssd` volumes, you should stop your Instance, detach the volumes to be preserved, then delete your Instance.
//
// The `backup` action can be done with:
// * No `volumes` key in the body: an image is created with snapshots of all the server volumes, except for the `scratch` volumes types.
// * `volumes` key in the body with a dictionary as value, in this dictionary volumes UUID as keys and empty dictionaries as values : an image is created with the snapshots of the volumes in `volumes` key. `scratch` volumes types can't be shapshotted.
func (s *API) ServerAction(req *ServerActionRequest, opts ...scw.RequestOption) (*ServerActionResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/action",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ServerActionResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListServerUserData: List all user data keys registered on a specified Instance.
func (s *API) ListServerUserData(req *ListServerUserDataRequest, opts ...scw.RequestOption) (*ListServerUserDataResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/user_data",
	}

	var resp ListServerUserDataResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteServerUserData: Delete the specified key from an Instance's user data.
func (s *API) DeleteServerUserData(req *DeleteServerUserDataRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	if fmt.Sprint(req.Key) == "" {
		return errors.New("field Key cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/user_data/" + fmt.Sprint(req.Key) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetServerCompatibleTypes: Get compatible commercial types that can be used to update the Instance. The compatibility of an Instance offer is based on:
// * the CPU architecture
// * the OS type
// * the required l_ssd storage size
// * the required scratch storage size
// If the specified Instance offer is flagged as end of service, the best compatible offer is the first returned.
func (s *API) GetServerCompatibleTypes(req *GetServerCompatibleTypesRequest, opts ...scw.RequestOption) (*ServerCompatibleTypes, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/compatible-types",
	}

	var resp ServerCompatibleTypes

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttachServerVolume: Attach a volume to an Instance.
func (s *API) AttachServerVolume(req *AttachServerVolumeRequest, opts ...scw.RequestOption) (*AttachServerVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/attach-volume",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp AttachServerVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DetachServerVolume: Detach a volume from an Instance.
func (s *API) DetachServerVolume(req *DetachServerVolumeRequest, opts ...scw.RequestOption) (*DetachServerVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/detach-volume",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DetachServerVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttachServerFileSystem: Attach a filesystem volume to an Instance.
func (s *API) AttachServerFileSystem(req *AttachServerFileSystemRequest, opts ...scw.RequestOption) (*AttachServerFileSystemResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/attach-filesystem",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp AttachServerFileSystemResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DetachServerFileSystem: Detach a filesystem volume from an Instance.
func (s *API) DetachServerFileSystem(req *DetachServerFileSystemRequest, opts ...scw.RequestOption) (*DetachServerFileSystemResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/detach-filesystem",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DetachServerFileSystemResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListImages: List all existing Instance images.
func (s *API) ListImages(req *ListImagesRequest, opts ...scw.RequestOption) (*ListImagesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "public", req.Public)
	parameter.AddToQuery(query, "arch", req.Arch)
	parameter.AddToQuery(query, "project", req.Project)
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images",
		Query:  query,
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetImage: Get details of an image with the specified ID.
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*GetImageResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ImageID) + "",
	}

	var resp GetImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateImage: Create an Instance image from the specified snapshot ID.
func (s *API) CreateImage(req *CreateImageRequest, opts ...scw.RequestOption) (*CreateImageResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("img")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// setImage: Replace all image properties with an image message.
func (s *API) setImage(req *SetImageRequest, opts ...scw.RequestOption) (*setImageResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if req.Project == "" {
		defaultProject, _ := s.client.GetDefaultProjectID()
		req.Project = defaultProject
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateImage: Update the properties of an image.
func (s *API) UpdateImage(req *UpdateImageRequest, opts ...scw.RequestOption) (*UpdateImageResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ImageID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteImage: Delete the image with the specified ID.
func (s *API) DeleteImage(req *DeleteImageRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ImageID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListSnapshots: List all snapshots of an Organization in a specified Availability Zone.
func (s *API) ListSnapshots(req *ListSnapshotsRequest, opts ...scw.RequestOption) (*ListSnapshotsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "base_volume_id", req.BaseVolumeID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
		Query:  query,
	}

	var resp ListSnapshotsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSnapshot: Create a snapshot from a specified volume or from a QCOW2 file in a specified Availability Zone.
func (s *API) CreateSnapshot(req *CreateSnapshotRequest, opts ...scw.RequestOption) (*CreateSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("snp")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSnapshot: Get details of a snapshot with the specified ID.
func (s *API) GetSnapshot(req *GetSnapshotRequest, opts ...scw.RequestOption) (*GetSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	var resp GetSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// setSnapshot: Replace all the properties of a snapshot.
func (s *API) setSnapshot(req *setSnapshotRequest, opts ...scw.RequestOption) (*setSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if req.Project == "" {
		defaultProject, _ := s.client.GetDefaultProjectID()
		req.Project = defaultProject
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSnapshot: Update the properties of a snapshot.
func (s *API) UpdateSnapshot(req *UpdateSnapshotRequest, opts ...scw.RequestOption) (*UpdateSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSnapshot: Delete the snapshot with the specified ID.
func (s *API) DeleteSnapshot(req *DeleteSnapshotRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ExportSnapshot: Export a snapshot to a specified Object Storage bucket in the same region.
func (s *API) ExportSnapshot(req *ExportSnapshotRequest, opts ...scw.RequestOption) (*ExportSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "/export",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ExportSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVolumes: List volumes in the specified Availability Zone. You can filter the output by volume type.
func (s *API) ListVolumes(req *ListVolumesRequest, opts ...scw.RequestOption) (*ListVolumesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "volume_type", req.VolumeType)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
		Query:  query,
	}

	var resp ListVolumesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateVolume: Create a volume of a specified type in an Availability Zone.
func (s *API) CreateVolume(req *CreateVolumeRequest, opts ...scw.RequestOption) (*CreateVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("vol")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVolume: Get details of a volume with the specified ID.
func (s *API) GetVolume(req *GetVolumeRequest, opts ...scw.RequestOption) (*GetVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.VolumeID) == "" {
		return nil, errors.New("field VolumeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	var resp GetVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateVolume: Replace the name and/or size properties of a volume specified by its ID, with the specified value(s).
func (s *API) UpdateVolume(req *UpdateVolumeRequest, opts ...scw.RequestOption) (*UpdateVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.VolumeID) == "" {
		return nil, errors.New("field VolumeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVolume: Delete the volume with the specified ID.
func (s *API) DeleteVolume(req *DeleteVolumeRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.VolumeID) == "" {
		return errors.New("field VolumeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListSecurityGroups: List all existing security groups.
func (s *API) ListSecurityGroups(req *ListSecurityGroupsRequest, opts ...scw.RequestOption) (*ListSecurityGroupsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "project_default", req.ProjectDefault)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups",
		Query:  query,
	}

	var resp ListSecurityGroupsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSecurityGroup: Create a security group with a specified name and description.
func (s *API) CreateSecurityGroup(req *CreateSecurityGroupRequest, opts ...scw.RequestOption) (*CreateSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("sg")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSecurityGroup: Get the details of a security group with the specified ID.
func (s *API) GetSecurityGroup(req *GetSecurityGroupRequest, opts ...scw.RequestOption) (*GetSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "",
	}

	var resp GetSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSecurityGroup: Delete a security group with the specified ID.
func (s *API) DeleteSecurityGroup(req *DeleteSecurityGroupRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// setSecurityGroup: Replace all security group properties with a security group message.
func (s *API) setSecurityGroup(req *setSecurityGroupRequest, opts ...scw.RequestOption) (*setSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if req.Project == "" {
		defaultProject, _ := s.client.GetDefaultProjectID()
		req.Project = defaultProject
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.ID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSecurityGroup: Update the properties of security group.
func (s *API) UpdateSecurityGroup(req *UpdateSecurityGroupRequest, opts ...scw.RequestOption) (*UpdateSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListDefaultSecurityGroupRules: Lists the default rules applied to all the security groups.
func (s *API) ListDefaultSecurityGroupRules(req *ListDefaultSecurityGroupRulesRequest, opts ...scw.RequestOption) (*ListSecurityGroupRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/default/rules",
	}

	var resp ListSecurityGroupRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSecurityGroupRules: List the rules of the a specified security group ID.
func (s *API) ListSecurityGroupRules(req *ListSecurityGroupRulesRequest, opts ...scw.RequestOption) (*ListSecurityGroupRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules",
		Query:  query,
	}

	var resp ListSecurityGroupRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSecurityGroupRule: Create a rule in the specified security group ID.
func (s *API) CreateSecurityGroupRule(req *CreateSecurityGroupRuleRequest, opts ...scw.RequestOption) (*CreateSecurityGroupRuleResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSecurityGroupRules: Replaces the existing rules of the security group with the rules provided. This endpoint supports the update of existing rules, creation of new rules and deletion of existing rules when they are not passed in the request.
func (s *API) SetSecurityGroupRules(req *SetSecurityGroupRulesRequest, opts ...scw.RequestOption) (*SetSecurityGroupRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetSecurityGroupRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSecurityGroupRule: Delete a security group rule with the specified ID.
func (s *API) DeleteSecurityGroupRule(req *DeleteSecurityGroupRuleRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return errors.New("field SecurityGroupID cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetSecurityGroupRule: Get details of a security group rule with the specified ID.
func (s *API) GetSecurityGroupRule(req *GetSecurityGroupRuleRequest, opts ...scw.RequestOption) (*GetSecurityGroupRuleResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return nil, errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
	}

	var resp GetSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// setSecurityGroupRule: Replace all the properties of a rule from a specified security group.
func (s *API) setSecurityGroupRule(req *setSecurityGroupRuleRequest, opts ...scw.RequestOption) (*setSecurityGroupRuleResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return nil, errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateSecurityGroupRule: Update the properties of a rule from a specified security group.
func (s *API) UpdateSecurityGroupRule(req *UpdateSecurityGroupRuleRequest, opts ...scw.RequestOption) (*UpdateSecurityGroupRuleResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return nil, errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPlacementGroups: List all placement groups in a specified Availability Zone.
func (s *API) ListPlacementGroups(req *ListPlacementGroupsRequest, opts ...scw.RequestOption) (*ListPlacementGroupsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups",
		Query:  query,
	}

	var resp ListPlacementGroupsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePlacementGroup: Create a new placement group in a specified Availability Zone.
func (s *API) CreatePlacementGroup(req *CreatePlacementGroupRequest, opts ...scw.RequestOption) (*CreatePlacementGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("pg")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreatePlacementGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPlacementGroup: Get the specified placement group.
func (s *API) GetPlacementGroup(req *GetPlacementGroupRequest, opts ...scw.RequestOption) (*GetPlacementGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "",
	}

	var resp GetPlacementGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetPlacementGroup: Set all parameters of the specified placement group.
func (s *API) SetPlacementGroup(req *SetPlacementGroupRequest, opts ...scw.RequestOption) (*SetPlacementGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if req.Project == "" {
		defaultProject, _ := s.client.GetDefaultProjectID()
		req.Project = defaultProject
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetPlacementGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdatePlacementGroup: Update one or more parameter of the specified placement group.
func (s *API) UpdatePlacementGroup(req *UpdatePlacementGroupRequest, opts ...scw.RequestOption) (*UpdatePlacementGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdatePlacementGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeletePlacementGroup: Delete the specified placement group.
func (s *API) DeletePlacementGroup(req *DeletePlacementGroupRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetPlacementGroupServers: Get all Instances belonging to the specified placement group.
func (s *API) GetPlacementGroupServers(req *GetPlacementGroupServersRequest, opts ...scw.RequestOption) (*GetPlacementGroupServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "/servers",
	}

	var resp GetPlacementGroupServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetPlacementGroupServers: Set all Instances belonging to the specified placement group.
func (s *API) SetPlacementGroupServers(req *SetPlacementGroupServersRequest, opts ...scw.RequestOption) (*SetPlacementGroupServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PUT",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "/servers",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetPlacementGroupServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdatePlacementGroupServers: Update all Instances belonging to the specified placement group.
func (s *API) UpdatePlacementGroupServers(req *UpdatePlacementGroupServersRequest, opts ...scw.RequestOption) (*UpdatePlacementGroupServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.PlacementGroupID) == "" {
		return nil, errors.New("field PlacementGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/placement_groups/" + fmt.Sprint(req.PlacementGroupID) + "/servers",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdatePlacementGroupServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListIPs: List all flexible IPs in a specified zone.
func (s *API) ListIPs(req *ListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "project", req.Project)
	parameter.AddToQuery(query, "organization", req.Organization)
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "type", req.Type)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:  query,
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateIP: Reserve a flexible IP and attach it to the specified Instance.
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*CreateIPResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultProject, exist := s.client.GetDefaultProjectID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Project = &defaultProject
	}

	defaultOrganization, exist := s.client.GetDefaultOrganizationID()
	if exist && req.Project == nil && req.Organization == nil {
		req.Organization = &defaultOrganization
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetIP: Get details of an IP with the specified ID or address.
func (s *API) GetIP(req *GetIPRequest, opts ...scw.RequestOption) (*GetIPResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IP) == "" {
		return nil, errors.New("field IP cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IP) + "",
	}

	var resp GetIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateIP: Update a flexible IP in the specified zone with the specified ID.
func (s *API) UpdateIP(req *UpdateIPRequest, opts ...scw.RequestOption) (*UpdateIPResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IP) == "" {
		return nil, errors.New("field IP cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IP) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteIP: Delete the IP with the specified ID.
func (s *API) DeleteIP(req *DeleteIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IP) == "" {
		return errors.New("field IP cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IP) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ListPrivateNICs: List all private NICs of a specified Instance.
func (s *API) ListPrivateNICs(req *ListPrivateNICsRequest, opts ...scw.RequestOption) (*ListPrivateNICsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	if len(req.Tags) != 0 {
		parameter.AddToQuery(query, "tags", strings.Join(req.Tags, ","))
	}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/private_nics",
		Query:  query,
	}

	var resp ListPrivateNICsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePrivateNIC: Create a private NIC connecting an Instance to a Private Network.
func (s *API) CreatePrivateNIC(req *CreatePrivateNICRequest, opts ...scw.RequestOption) (*CreatePrivateNICResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/private_nics",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreatePrivateNICResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPrivateNIC: Get private NIC properties.
func (s *API) GetPrivateNIC(req *GetPrivateNICRequest, opts ...scw.RequestOption) (*GetPrivateNICResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNicID) == "" {
		return nil, errors.New("field PrivateNicID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/private_nics/" + fmt.Sprint(req.PrivateNicID) + "",
	}

	var resp GetPrivateNICResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdatePrivateNIC: Update one or more parameter(s) of a specified private NIC.
func (s *API) UpdatePrivateNIC(req *UpdatePrivateNICRequest, opts ...scw.RequestOption) (*PrivateNIC, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNicID) == "" {
		return nil, errors.New("field PrivateNicID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "PATCH",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/private_nics/" + fmt.Sprint(req.PrivateNicID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNIC

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeletePrivateNIC: Delete a private NIC.
func (s *API) DeletePrivateNIC(req *DeletePrivateNICRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNicID) == "" {
		return errors.New("field PrivateNicID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "DELETE",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/private_nics/" + fmt.Sprint(req.PrivateNicID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// GetDashboard:
func (s *API) GetDashboard(req *GetDashboardRequest, opts ...scw.RequestOption) (*GetDashboardResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "project", req.Project)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/dashboard",
		Query:  query,
	}

	var resp GetDashboardResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PlanBlockMigration: Given a volume or snapshot, returns the migration plan but does not perform the actual migration. To perform the migration, you have to call the [Migrate a volume and/or snapshots to SBS](#path-volumes-migrate-a-volume-andor-snapshots-to-sbs-scaleway-block-storage) endpoint afterward.
// The endpoint returns the resources that should be migrated together:
// - the volume and any snapshots created from the volume, if the call was made to plan a volume migration.
// - the base volume of the snapshot (if the volume is not deleted) and its related snapshots, if the call was made to plan a snapshot migration.
// The endpoint also returns the validation_key, which must be provided to the [Migrate a volume and/or snapshots to SBS](#path-volumes-migrate-a-volume-andor-snapshots-to-sbs-scaleway-block-storage) endpoint to confirm that all resources listed in the plan should be migrated.
func (s *API) PlanBlockMigration(req *PlanBlockMigrationRequest, opts ...scw.RequestOption) (*MigrationPlan, error) {
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
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/block-migration/plan",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp MigrationPlan

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ApplyBlockMigration: To be used, the call to this endpoint must be preceded by a call to the [Get a volume or snapshot's migration plan](#path-volumes-get-a-volume-or-snapshots-migration-plan) endpoint. To migrate all resources mentioned in the migration plan, the validation_key returned in the plan must be provided.
func (s *API) ApplyBlockMigration(req *ApplyBlockMigrationRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/block-migration/apply",
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

// CheckBlockMigrationOrganizationQuotas:
func (s *API) CheckBlockMigrationOrganizationQuotas(req *CheckBlockMigrationOrganizationQuotasRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultOrganizationID()
		req.Organization = defaultOrganization
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/block-migration/check-organization-quotas",
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

// ReleaseIPToIpam: **The IP remains available in IPAM**, which means that it is still reserved by the Organization, and can be reattached to another resource (Instance or other product).
func (s *API) ReleaseIPToIpam(req *ReleaseIPToIpamRequest, opts ...scw.RequestOption) error {
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
		Method: "POST",
		Path:   "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "/release-to-ipam",
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
