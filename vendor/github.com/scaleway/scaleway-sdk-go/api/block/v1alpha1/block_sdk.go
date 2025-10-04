// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package block provides methods and message types of the block v1alpha1 API.
package block

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

type ListSnapshotsRequestOrderBy string

const (
	// Order by creation date (ascending chronological order).
	ListSnapshotsRequestOrderByCreatedAtAsc = ListSnapshotsRequestOrderBy("created_at_asc")
	// Order by creation date (descending chronological order).
	ListSnapshotsRequestOrderByCreatedAtDesc = ListSnapshotsRequestOrderBy("created_at_desc")
	// Order by name (ascending order).
	ListSnapshotsRequestOrderByNameAsc = ListSnapshotsRequestOrderBy("name_asc")
	// Order by name (descending order).
	ListSnapshotsRequestOrderByNameDesc = ListSnapshotsRequestOrderBy("name_desc")
)

func (enum ListSnapshotsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListSnapshotsRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListSnapshotsRequestOrderBy) Values() []ListSnapshotsRequestOrderBy {
	return []ListSnapshotsRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
}

func (enum ListSnapshotsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListSnapshotsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListSnapshotsRequestOrderBy(ListSnapshotsRequestOrderBy(tmp).String())
	return nil
}

type ListVolumesRequestOrderBy string

const (
	// Order by creation date (ascending chronological order).
	ListVolumesRequestOrderByCreatedAtAsc = ListVolumesRequestOrderBy("created_at_asc")
	// Order by creation date (descending chronological order).
	ListVolumesRequestOrderByCreatedAtDesc = ListVolumesRequestOrderBy("created_at_desc")
	// Order by name (ascending order).
	ListVolumesRequestOrderByNameAsc = ListVolumesRequestOrderBy("name_asc")
	// Order by name (descending order).
	ListVolumesRequestOrderByNameDesc = ListVolumesRequestOrderBy("name_desc")
)

func (enum ListVolumesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return string(ListVolumesRequestOrderByCreatedAtAsc)
	}
	return string(enum)
}

func (enum ListVolumesRequestOrderBy) Values() []ListVolumesRequestOrderBy {
	return []ListVolumesRequestOrderBy{
		"created_at_asc",
		"created_at_desc",
		"name_asc",
		"name_desc",
	}
}

func (enum ListVolumesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListVolumesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListVolumesRequestOrderBy(ListVolumesRequestOrderBy(tmp).String())
	return nil
}

type ReferenceStatus string

const (
	// If unspecified, the status of the reference is unknown by default.
	ReferenceStatusUnknownStatus = ReferenceStatus("unknown_status")
	// When the reference is being attached (transient).
	ReferenceStatusAttaching = ReferenceStatus("attaching")
	// When the reference attached to a volume.
	ReferenceStatusAttached = ReferenceStatus("attached")
	// When the reference is being detached (transient).
	ReferenceStatusDetaching = ReferenceStatus("detaching")
	// When the reference is detached from a volume - the reference ceases to exist.
	ReferenceStatusDetached = ReferenceStatus("detached")
	// Reference under creation which can be rolled back if an error occurs (transient).
	ReferenceStatusCreating = ReferenceStatus("creating")
	// Error status.
	ReferenceStatusError = ReferenceStatus("error")
)

func (enum ReferenceStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(ReferenceStatusUnknownStatus)
	}
	return string(enum)
}

func (enum ReferenceStatus) Values() []ReferenceStatus {
	return []ReferenceStatus{
		"unknown_status",
		"attaching",
		"attached",
		"detaching",
		"detached",
		"creating",
		"error",
	}
}

func (enum ReferenceStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ReferenceStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ReferenceStatus(ReferenceStatus(tmp).String())
	return nil
}

type ReferenceType string

const (
	// If unspecified, the reference type is unknown by default.
	ReferenceTypeUnknownType = ReferenceType("unknown_type")
	// Reference linked to a snapshot (for snapshots only).
	ReferenceTypeLink = ReferenceType("link")
	// Exclusive reference that can be associated to a volume (for volumes only).
	ReferenceTypeExclusive = ReferenceType("exclusive")
	// Access to the volume or snapshot in a read-only mode, without storage write access to the resource.
	ReferenceTypeReadOnly = ReferenceType("read_only")
)

func (enum ReferenceType) String() string {
	if enum == "" {
		// return default value if empty
		return string(ReferenceTypeUnknownType)
	}
	return string(enum)
}

func (enum ReferenceType) Values() []ReferenceType {
	return []ReferenceType{
		"unknown_type",
		"link",
		"exclusive",
		"read_only",
	}
}

func (enum ReferenceType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ReferenceType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ReferenceType(ReferenceType(tmp).String())
	return nil
}

type SnapshotStatus string

const (
	// If unspecified, the snapshot status is unknown by default.
	SnapshotStatusUnknownStatus = SnapshotStatus("unknown_status")
	// The snapshot is under creation (transient).
	SnapshotStatusCreating = SnapshotStatus("creating")
	// Snapshot exists and is not attached to any reference.
	SnapshotStatusAvailable = SnapshotStatus("available")
	// Snapshot in an error status.
	SnapshotStatusError = SnapshotStatus("error")
	// Snapshot is being deleted (transient).
	SnapshotStatusDeleting = SnapshotStatus("deleting")
	// Snapshot was deleted.
	SnapshotStatusDeleted = SnapshotStatus("deleted")
	// Snapshot attached to one or more references.
	SnapshotStatusInUse     = SnapshotStatus("in_use")
	SnapshotStatusLocked    = SnapshotStatus("locked")
	SnapshotStatusExporting = SnapshotStatus("exporting")
)

func (enum SnapshotStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(SnapshotStatusUnknownStatus)
	}
	return string(enum)
}

func (enum SnapshotStatus) Values() []SnapshotStatus {
	return []SnapshotStatus{
		"unknown_status",
		"creating",
		"available",
		"error",
		"deleting",
		"deleted",
		"in_use",
		"locked",
		"exporting",
	}
}

func (enum SnapshotStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SnapshotStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SnapshotStatus(SnapshotStatus(tmp).String())
	return nil
}

type StorageClass string

const (
	// If unspecified, the Storage Class is unknown by default.
	StorageClassUnknownStorageClass = StorageClass("unknown_storage_class")
	// No specific Storage Class selected.
	StorageClassUnspecified = StorageClass("unspecified")
	// Classic storage.
	StorageClassBssd = StorageClass("bssd")
	// Performance storage with lower latency.
	StorageClassSbs = StorageClass("sbs")
)

func (enum StorageClass) String() string {
	if enum == "" {
		// return default value if empty
		return string(StorageClassUnknownStorageClass)
	}
	return string(enum)
}

func (enum StorageClass) Values() []StorageClass {
	return []StorageClass{
		"unknown_storage_class",
		"unspecified",
		"bssd",
		"sbs",
	}
}

func (enum StorageClass) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *StorageClass) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = StorageClass(StorageClass(tmp).String())
	return nil
}

type VolumeStatus string

const (
	// If unspecified, the volume status is unknown by default.
	VolumeStatusUnknownStatus = VolumeStatus("unknown_status")
	// The volume is under creation (transient).
	VolumeStatusCreating = VolumeStatus("creating")
	// The volume exists and is not attached to any reference.
	VolumeStatusAvailable = VolumeStatus("available")
	// The volume exists and is already attached to a reference.
	VolumeStatusInUse = VolumeStatus("in_use")
	// The volume undergoing deletion (transient).
	VolumeStatusDeleting = VolumeStatus("deleting")
	VolumeStatusDeleted  = VolumeStatus("deleted")
	// The volume is being increased (transient).
	VolumeStatusResizing = VolumeStatus("resizing")
	// The volume is an error status.
	VolumeStatusError = VolumeStatus("error")
	// The volume is undergoing snapshotting operation (transient).
	VolumeStatusSnapshotting = VolumeStatus("snapshotting")
	VolumeStatusLocked       = VolumeStatus("locked")
	// The volume is being updated (transient).
	VolumeStatusUpdating = VolumeStatus("updating")
)

func (enum VolumeStatus) String() string {
	if enum == "" {
		// return default value if empty
		return string(VolumeStatusUnknownStatus)
	}
	return string(enum)
}

func (enum VolumeStatus) Values() []VolumeStatus {
	return []VolumeStatus{
		"unknown_status",
		"creating",
		"available",
		"in_use",
		"deleting",
		"deleted",
		"resizing",
		"error",
		"snapshotting",
		"locked",
		"updating",
	}
}

func (enum VolumeStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *VolumeStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = VolumeStatus(VolumeStatus(tmp).String())
	return nil
}

// Reference: reference.
type Reference struct {
	// ID: UUID of the reference.
	ID string `json:"id"`

	// ProductResourceType: type of resource to which the reference is associated.
	ProductResourceType string `json:"product_resource_type"`

	// ProductResourceID: UUID of the product resource it refers to (according to the product_resource_type).
	ProductResourceID string `json:"product_resource_id"`

	// CreatedAt: creation date of the reference.
	CreatedAt *time.Time `json:"created_at"`

	// Type: type of reference (link, exclusive, read_only).
	// Default value: unknown_type
	Type ReferenceType `json:"type"`

	// Status: status of the reference. Statuses include `attaching`, `attached`, and `detaching`.
	// Default value: unknown_status
	Status ReferenceStatus `json:"status"`
}

// SnapshotParentVolume: snapshot parent volume.
type SnapshotParentVolume struct {
	// ID: parent volume UUID (volume from which the snapshot originates).
	ID string `json:"id"`

	// Name: name of the parent volume.
	Name string `json:"name"`

	// Type: volume type of the parent volume.
	Type string `json:"type"`

	// Status: current status the parent volume.
	// Default value: unknown_status
	Status VolumeStatus `json:"status"`
}

// VolumeSpecifications: volume specifications.
type VolumeSpecifications struct {
	// PerfIops: the maximum IO/s expected, according to the different options available in stock (`5000 | 15000`).
	PerfIops *uint32 `json:"perf_iops"`

	// Class: the storage class of the volume.
	// Default value: unknown_storage_class
	Class StorageClass `json:"class"`
}

// CreateVolumeRequestFromEmpty: create volume request from empty.
type CreateVolumeRequestFromEmpty struct {
	// Size: must be compliant with the minimum (1 GB) and maximum (10 TB) allowed size.
	Size scw.Size `json:"size"`
}

// CreateVolumeRequestFromSnapshot: create volume request from snapshot.
type CreateVolumeRequestFromSnapshot struct {
	// Size: must be compliant with the minimum (1 GB) and maximum (10 TB) allowed size.
	// Size is optional and is used only if a resize of the volume is requested, otherwise original snapshot size will be used.
	Size *scw.Size `json:"size"`

	// SnapshotID: source snapshot from which volume will be created.
	SnapshotID string `json:"snapshot_id"`
}

// Snapshot: snapshot.
type Snapshot struct {
	// ID: UUID of the snapshot.
	ID string `json:"id"`

	// Name: name of the snapshot.
	Name string `json:"name"`

	// ParentVolume: if the parent volume was deleted, value is null.
	ParentVolume *SnapshotParentVolume `json:"parent_volume"`

	// Size: size in bytes of the snapshot.
	Size scw.Size `json:"size"`

	// ProjectID: UUID of the project the snapshot belongs to.
	ProjectID string `json:"project_id"`

	// CreatedAt: creation date of the snapshot.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: last modification date of the properties of a snapshot.
	UpdatedAt *time.Time `json:"updated_at"`

	// References: list of the references to the snapshot.
	References []*Reference `json:"references"`

	// Status: current status of the snapshot (available, in_use, ...).
	// Default value: unknown_status
	Status SnapshotStatus `json:"status"`

	// Tags: list of tags assigned to the volume.
	Tags []string `json:"tags"`

	// Zone: snapshot zone.
	Zone scw.Zone `json:"zone"`

	// Class: storage class of the snapshot.
	// Default value: unknown_storage_class
	Class StorageClass `json:"class"`
}

// VolumeType: volume type.
type VolumeType struct {
	// Type: volume type.
	Type string `json:"type"`

	// Pricing: price of the volume billed in GB/hour.
	Pricing *scw.Money `json:"pricing"`

	// SnapshotPricing: price of the snapshot billed in GB/hour.
	SnapshotPricing *scw.Money `json:"snapshot_pricing"`

	// Specs: volume specifications of the volume type.
	Specs *VolumeSpecifications `json:"specs"`
}

// Volume: volume.
type Volume struct {
	// ID: UUID of the volume.
	ID string `json:"id"`

	// Name: name of the volume.
	Name string `json:"name"`

	// Type: volume type.
	Type string `json:"type"`

	// Size: volume size in bytes.
	Size scw.Size `json:"size"`

	// ProjectID: UUID of the project to which the volume belongs.
	ProjectID string `json:"project_id"`

	// CreatedAt: creation date of the volume.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt: last update of the properties of a volume.
	UpdatedAt *time.Time `json:"updated_at"`

	// References: list of the references to the volume.
	References []*Reference `json:"references"`

	// ParentSnapshotID: when a volume is created from a snapshot, is the UUID of the snapshot from which the volume has been created.
	ParentSnapshotID *string `json:"parent_snapshot_id"`

	// Status: current status of the volume (available, in_use, ...).
	// Default value: unknown_status
	Status VolumeStatus `json:"status"`

	// Tags: list of tags assigned to the volume.
	Tags []string `json:"tags"`

	// Zone: volume zone.
	Zone scw.Zone `json:"zone"`

	// Specs: specifications of the volume.
	Specs *VolumeSpecifications `json:"specs"`

	// LastDetachedAt: last time the volume was detached.
	LastDetachedAt *time.Time `json:"last_detached_at"`
}

// CreateSnapshotRequest: create snapshot request.
type CreateSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume to snapshot.
	VolumeID string `json:"volume_id"`

	// Name: name of the snapshot.
	Name string `json:"name"`

	// ProjectID: UUID of the project to which the volume and the snapshot belong.
	ProjectID string `json:"project_id"`

	// Tags: list of tags assigned to the snapshot.
	Tags []string `json:"tags"`
}

// CreateVolumeRequest: create volume request.
type CreateVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Name: name of the volume.
	Name string `json:"name"`

	// PerfIops: the maximum IO/s expected, according to the different options available in stock (`5000 | 15000`).
	// Precisely one of PerfIops must be set.
	PerfIops *uint32 `json:"perf_iops,omitempty"`

	// ProjectID: UUID of the project the volume belongs to.
	ProjectID string `json:"project_id"`

	// FromEmpty: specify the size of the new volume if creating a new one from scratch.
	// Precisely one of FromEmpty, FromSnapshot must be set.
	FromEmpty *CreateVolumeRequestFromEmpty `json:"from_empty,omitempty"`

	// FromSnapshot: specify the snapshot ID of the original snapshot.
	// Precisely one of FromEmpty, FromSnapshot must be set.
	FromSnapshot *CreateVolumeRequestFromSnapshot `json:"from_snapshot,omitempty"`

	// Tags: list of tags assigned to the volume.
	Tags []string `json:"tags"`
}

// DeleteSnapshotRequest: delete snapshot request.
type DeleteSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot.
	SnapshotID string `json:"-"`
}

// DeleteVolumeRequest: delete volume request.
type DeleteVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume.
	VolumeID string `json:"-"`
}

// ExportSnapshotToObjectStorageRequest: export snapshot to object storage request.
type ExportSnapshotToObjectStorageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot.
	SnapshotID string `json:"-"`

	// Bucket: scaleway Object Storage bucket where the object is stored.
	Bucket string `json:"bucket"`

	// Key: the object key inside the given bucket.
	Key string `json:"key"`
}

// GetSnapshotRequest: get snapshot request.
type GetSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot.
	SnapshotID string `json:"-"`
}

// GetVolumeRequest: get volume request.
type GetVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume.
	VolumeID string `json:"-"`
}

// ImportSnapshotFromObjectStorageRequest: import snapshot from object storage request.
type ImportSnapshotFromObjectStorageRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Bucket: scaleway Object Storage bucket where the object is stored.
	Bucket string `json:"bucket"`

	// Key: the object key inside the given bucket.
	Key string `json:"key"`

	// Name: name of the snapshot.
	Name string `json:"name"`

	// ProjectID: UUID of the Project to which the volume and the snapshot belong.
	ProjectID string `json:"project_id"`

	// Tags: list of tags assigned to the snapshot.
	Tags []string `json:"tags"`

	// Size: size of the snapshot.
	Size *scw.Size `json:"size,omitempty"`
}

// ImportSnapshotFromS3Request: import snapshot from s3 request.
type ImportSnapshotFromS3Request struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Bucket: scaleway Object Storage bucket where the object is stored.
	Bucket string `json:"bucket"`

	// Key: the object key inside the given bucket.
	Key string `json:"key"`

	// Name: name of the snapshot.
	Name string `json:"name"`

	// ProjectID: UUID of the Project to which the volume and the snapshot belong.
	ProjectID string `json:"project_id"`

	// Tags: list of tags assigned to the snapshot.
	Tags []string `json:"tags"`

	// Size: size of the snapshot.
	Size *scw.Size `json:"size,omitempty"`
}

// ListSnapshotsRequest: list snapshots request.
type ListSnapshotsRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: criteria to use when ordering the list.
	// Default value: created_at_asc
	OrderBy ListSnapshotsRequestOrderBy `json:"-"`

	// ProjectID: filter by Project ID.
	ProjectID *string `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID *string `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: page size, defines how many entries are returned in one page, must be lower or equal to 100.
	PageSize *uint32 `json:"-"`

	// VolumeID: filter snapshots by the ID of the original volume.
	VolumeID *string `json:"-"`

	// Name: filter snapshots by their names.
	Name *string `json:"-"`

	// Tags: filter by tags. Only snapshots with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ListSnapshotsResponse: list snapshots response.
type ListSnapshotsResponse struct {
	// Snapshots: paginated returned list of snapshots.
	Snapshots []*Snapshot `json:"snapshots"`

	// TotalCount: total number of snpashots in the project.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListSnapshotsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Snapshots = append(r.Snapshots, results.Snapshots...)
	r.TotalCount += uint64(len(results.Snapshots))
	return uint64(len(results.Snapshots)), nil
}

// ListVolumeTypesRequest: list volume types request.
type ListVolumeTypesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: page size, defines how many entries are returned in one page, must be lower or equal to 100.
	PageSize *uint32 `json:"-"`
}

// ListVolumeTypesResponse: list volume types response.
type ListVolumeTypesResponse struct {
	// VolumeTypes: returns paginated list of volume-types.
	VolumeTypes []*VolumeType `json:"volume_types"`

	// TotalCount: total number of volume-types currently available in stock.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVolumeTypesResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVolumeTypesResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListVolumeTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.VolumeTypes = append(r.VolumeTypes, results.VolumeTypes...)
	r.TotalCount += uint64(len(results.VolumeTypes))
	return uint64(len(results.VolumeTypes)), nil
}

// ListVolumesRequest: list volumes request.
type ListVolumesRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// OrderBy: criteria to use when ordering the list.
	// Default value: created_at_asc
	OrderBy ListVolumesRequestOrderBy `json:"-"`

	// ProjectID: filter by Project ID.
	ProjectID *string `json:"-"`

	// OrganizationID: filter by Organization ID.
	OrganizationID *string `json:"-"`

	// Page: page number.
	Page *int32 `json:"-"`

	// PageSize: page size, defines how many entries are returned in one page, must be lower or equal to 100.
	PageSize *uint32 `json:"-"`

	// Name: filter the return volumes by their names.
	Name *string `json:"-"`

	// ProductResourceID: filter by a product resource ID linked to this volume (such as an Instance ID).
	ProductResourceID *string `json:"-"`

	// Tags: filter by tags. Only volumes with one or more matching tags will be returned.
	Tags []string `json:"-"`
}

// ListVolumesResponse: list volumes response.
type ListVolumesResponse struct {
	// Volumes: paginated returned list of volumes.
	Volumes []*Volume `json:"volumes"`

	// TotalCount: total number of volumes in the project.
	TotalCount uint64 `json:"total_count"`
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeAppend(res any) (uint64, error) {
	results, ok := res.(*ListVolumesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Volumes = append(r.Volumes, results.Volumes...)
	r.TotalCount += uint64(len(results.Volumes))
	return uint64(len(results.Volumes)), nil
}

// UpdateSnapshotRequest: update snapshot request.
type UpdateSnapshotRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// SnapshotID: UUID of the snapshot.
	SnapshotID string `json:"-"`

	// Name: when defined, is the name of the snapshot.
	Name *string `json:"name,omitempty"`

	// Tags: list of tags assigned to the snapshot.
	Tags *[]string `json:"tags,omitempty"`
}

// UpdateVolumeRequest: update volume request.
type UpdateVolumeRequest struct {
	// Zone: zone to target. If none is passed will use default zone from the config.
	Zone scw.Zone `json:"-"`

	// VolumeID: UUID of the volume.
	VolumeID string `json:"-"`

	// Name: when defined, is the new name of the volume.
	Name *string `json:"name,omitempty"`

	// Size: size in bytes of the volume, with a granularity of 1 GB (10^9 bytes).
	// Must be compliant with the minimum (1GB) and maximum (10TB) allowed size.
	Size *scw.Size `json:"size,omitempty"`

	// Tags: list of tags assigned to the volume.
	Tags *[]string `json:"tags,omitempty"`

	// PerfIops: the selected value must be available for the volume's current storage class.
	PerfIops *uint32 `json:"perf_iops,omitempty"`
}

// This API allows you to manage your Block Storage volumes.
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

// ListVolumeTypes: List all available volume types in a specified zone. The volume types listed are ordered by name in ascending order.
func (s *API) ListVolumeTypes(req *ListVolumeTypesRequest, opts ...scw.RequestOption) (*ListVolumeTypesResponse, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volume-types",
		Query:  query,
	}

	var resp ListVolumeTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListVolumes: List all existing volumes in a specified zone. By default, the volumes listed are ordered by creation date in ascending order. This can be modified via the `order_by` field.
func (s *API) ListVolumes(req *ListVolumesRequest, opts ...scw.RequestOption) (*ListVolumesResponse, error) {
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
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "product_resource_id", req.ProductResourceID)
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
		Query:  query,
	}

	var resp ListVolumesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateVolume: To create a new volume from scratch, you must specify `from_empty` and the `size`.
// To create a volume from an existing snapshot, specify `from_snapshot` and the `snapshot_id` in the request payload instead, size is optional and can be specified if you need to extend the original size. The volume will take on the same volume class and underlying IOPS limitations as the original snapshot.
func (s *API) CreateVolume(req *CreateVolumeRequest, opts ...scw.RequestOption) (*Volume, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("vol")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Volume

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVolume: Retrieve technical information about a specific volume. Details such as size, type, and status are returned in the response.
func (s *API) GetVolume(req *GetVolumeRequest, opts ...scw.RequestOption) (*Volume, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	var resp Volume

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVolume: You must specify the `volume_id` of the volume you want to delete. The volume must not be in the `in_use` status.
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UpdateVolume: Update the technical details of a volume, such as its name, tags, or its new size and `volume_type` (within the same Block Storage class).
// You can only resize a volume to a larger size. It is currently not possible to change your Block Storage Class.
func (s *API) UpdateVolume(req *UpdateVolumeRequest, opts ...scw.RequestOption) (*Volume, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Volume

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSnapshots: List all available snapshots in a specified zone. By default, the snapshots listed are ordered by creation date in ascending order. This can be modified via the `order_by` field.
func (s *API) ListSnapshots(req *ListSnapshotsRequest, opts ...scw.RequestOption) (*ListSnapshotsResponse, error) {
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
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "volume_id", req.VolumeID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "tags", req.Tags)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "GET",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
		Query:  query,
	}

	var resp ListSnapshotsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSnapshot: Retrieve technical information about a specific snapshot. Details such as size, volume type, and status are returned in the response.
func (s *API) GetSnapshot(req *GetSnapshotRequest, opts ...scw.RequestOption) (*Snapshot, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateSnapshot: To create a snapshot, the volume must be in the `in_use` or the `available` status.
// If your volume is in a transient state, you need to wait until the end of the current operation.
func (s *API) CreateSnapshot(req *CreateSnapshotRequest, opts ...scw.RequestOption) (*Snapshot, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("snp")
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Deprecated: ImportSnapshotFromS3: Import a snapshot from a Scaleway Object Storage bucket
// The bucket must contain a QCOW2 image.
// The bucket can be imported into any Availability Zone as long as it is in the same region as the bucket.
func (s *API) ImportSnapshotFromS3(req *ImportSnapshotFromS3Request, opts ...scw.RequestOption) (*Snapshot, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/import-from-s3",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ImportSnapshotFromObjectStorage: The bucket must contain a QCOW2 image.
// The bucket can be imported into any Availability Zone as long as it is in the same region as the bucket.
func (s *API) ImportSnapshotFromObjectStorage(req *ImportSnapshotFromObjectStorageRequest, opts ...scw.RequestOption) (*Snapshot, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method: "POST",
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/import-from-object-storage",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ExportSnapshotToObjectStorage: The snapshot is exported in QCOW2 format.
// The snapshot must not be in transient state.
func (s *API) ExportSnapshotToObjectStorage(req *ExportSnapshotToObjectStorageRequest, opts ...scw.RequestOption) (*Snapshot, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "/export-to-object-storage",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSnapshot: You must specify the `snapshot_id` of the snapshot you want to delete. The snapshot must not be in use.
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSnapshot: Update the name or tags of the snapshot.
func (s *API) UpdateSnapshot(req *UpdateSnapshotRequest, opts ...scw.RequestOption) (*Snapshot, error) {
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
		Path:   "/block/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Snapshot

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
