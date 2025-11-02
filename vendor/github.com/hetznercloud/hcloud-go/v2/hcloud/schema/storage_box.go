package schema

import "time"

// StorageBox defines the schema of a Storage Box.
type StorageBox struct {
	ID             int64                    `json:"id"`
	Username       *string                  `json:"username"`
	Status         string                   `json:"status"`
	Name           string                   `json:"name"`
	StorageBoxType StorageBoxType           `json:"storage_box_type"`
	Location       Location                 `json:"location"`
	AccessSettings StorageBoxAccessSettings `json:"access_settings"`
	Server         *string                  `json:"server"`
	System         *string                  `json:"system"`
	Stats          StorageBoxStats          `json:"stats"`
	Labels         map[string]string        `json:"labels"`
	Protection     StorageBoxProtection     `json:"protection"`
	SnapshotPlan   *StorageBoxSnapshotPlan  `json:"snapshot_plan"`
	Created        time.Time                `json:"created"`
}

// StorageBoxAccessSettings defines the schema of a Storage Box's access settings.
type StorageBoxAccessSettings struct {
	ReachableExternally bool `json:"reachable_externally"`
	SambaEnabled        bool `json:"samba_enabled"`
	SSHEnabled          bool `json:"ssh_enabled"`
	WebDAVEnabled       bool `json:"webdav_enabled"`
	ZFSEnabled          bool `json:"zfs_enabled"`
}

// StorageBoxStats defines the schema of a Storage Box's disk usage statistics.
type StorageBoxStats struct {
	Size          uint64 `json:"size"`
	SizeData      uint64 `json:"size_data"`
	SizeSnapshots uint64 `json:"size_snapshots"`
}

// StorageBoxProtection defines the schema of a Storage Box's resource protection.
type StorageBoxProtection struct {
	Delete bool `json:"delete"`
}

// StorageBoxSnapshotPlan defines the schema of a Storage Box's snapshot plan.
type StorageBoxSnapshotPlan struct {
	MaxSnapshots int  `json:"max_snapshots"`
	Minute       int  `json:"minute"`
	Hour         int  `json:"hour"`
	DayOfWeek    *int `json:"day_of_week"`
	DayOfMonth   *int `json:"day_of_month"`
}

// StorageBoxGetResponse defines the schema of the response when
// retrieving a single Storage Box.
type StorageBoxGetResponse struct {
	StorageBox StorageBox `json:"storage_box"`
}

// StorageBoxListResponse defines the schema of the response when
// listing Storage Boxes.
type StorageBoxListResponse struct {
	StorageBoxes []StorageBox `json:"storage_boxes"`
}

// StorageBoxCreateRequest defines the schema for the request to
// create a Storage Box.
type StorageBoxCreateRequest struct {
	Name           string                                 `json:"name"`
	StorageBoxType IDOrName                               `json:"storage_box_type"`
	Location       IDOrName                               `json:"location"`
	Labels         *map[string]string                     `json:"labels,omitempty"`
	Password       string                                 `json:"password"`
	SSHKeys        []string                               `json:"ssh_keys,omitempty"`
	AccessSettings *StorageBoxCreateRequestAccessSettings `json:"access_settings,omitempty"`
}

type StorageBoxCreateRequestAccessSettings struct {
	ReachableExternally *bool `json:"reachable_externally,omitempty"`
	SambaEnabled        *bool `json:"samba_enabled,omitempty"`
	SSHEnabled          *bool `json:"ssh_enabled,omitempty"`
	WebDAVEnabled       *bool `json:"webdav_enabled,omitempty"`
	ZFSEnabled          *bool `json:"zfs_enabled,omitempty"`
}

// StorageBoxCreateResponse defines the schema of the response when
// creating a Storage Box.
type StorageBoxCreateResponse struct {
	StorageBox StorageBox `json:"storage_box"`
	Action     Action     `json:"action"`
}

// StorageBoxUpdateRequest defines the schema of the request to update a Storage Box.
type StorageBoxUpdateRequest struct {
	Name   string             `json:"name,omitempty"`
	Labels *map[string]string `json:"labels,omitempty"`
}

// StorageBoxUpdateResponse defines the schema of the response when updating a Storage Box.
type StorageBoxUpdateResponse struct {
	StorageBox StorageBox `json:"storage_box"`
}

// StorageBoxFoldersResponse defines the schema of the response when listing folders in a Storage Box.
type StorageBoxFoldersResponse struct {
	Folders []string `json:"folders"`
}

// StorageBoxChangeProtectionRequest defines the schema of the request to change the
// resource protection of a Storage Box.
type StorageBoxChangeProtectionRequest struct {
	Delete *bool `json:"delete,omitempty"`
}

// StorageBoxChangeTypeRequest defines the schema of the request to change the type of a Storage Box.
type StorageBoxChangeTypeRequest struct {
	StorageBoxType IDOrName `json:"storage_box_type"`
}

// StorageBoxChangeTypeResponse defines the schema of the response when changing the type of a Storage Box.
type StorageBoxResetPasswordRequest struct {
	Password string `json:"password"`
}

// StorageBoxUpdateAccessSettingsRequest defines the schema of the request when updating
// a Storage Box's access settings.
type StorageBoxUpdateAccessSettingsRequest struct {
	ReachableExternally *bool `json:"reachable_externally,omitempty"`
	SambaEnabled        *bool `json:"samba_enabled,omitempty"`
	SSHEnabled          *bool `json:"ssh_enabled,omitempty"`
	WebDAVEnabled       *bool `json:"webdav_enabled,omitempty"`
	ZFSEnabled          *bool `json:"zfs_enabled,omitempty"`
}

// StorageBoxRollbackSnapshotRequest defines the schema of the request to roll back a Storage Box to a snapshot.
type StorageBoxRollbackSnapshotRequest struct {
	Snapshot IDOrName `json:"snapshot"`
}

// StorageBoxEnableSnapshotPlanRequest defines the schema of the request to enable a snapshot plan for a Storage Box.
type StorageBoxEnableSnapshotPlanRequest struct {
	MaxSnapshots int  `json:"max_snapshots"`
	Minute       int  `json:"minute"`
	Hour         int  `json:"hour"`
	DayOfWeek    *int `json:"day_of_week"`  // No omitempty because nil have a meaning and is the default in the API.
	DayOfMonth   *int `json:"day_of_month"` // No omitempty because nil have a meaning and is the default in the API.
}
