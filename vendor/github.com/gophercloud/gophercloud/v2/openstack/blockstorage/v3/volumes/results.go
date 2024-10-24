package volumes

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// Attachment represents a Volume Attachment record
type Attachment struct {
	AttachedAt   time.Time `json:"-"`
	AttachmentID string    `json:"attachment_id"`
	Device       string    `json:"device"`
	HostName     string    `json:"host_name"`
	ID           string    `json:"id"`
	ServerID     string    `json:"server_id"`
	VolumeID     string    `json:"volume_id"`
}

// UnmarshalJSON is our unmarshalling helper
func (r *Attachment) UnmarshalJSON(b []byte) error {
	type tmp Attachment
	var s struct {
		tmp
		AttachedAt gophercloud.JSONRFC3339MilliNoZ `json:"attached_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Attachment(s.tmp)

	r.AttachedAt = time.Time(s.AttachedAt)

	return err
}

// Volume contains all the information associated with an OpenStack Volume.
type Volume struct {
	// Unique identifier for the volume.
	ID string `json:"id"`
	// Current status of the volume.
	Status string `json:"status"`
	// Size of the volume in GB.
	Size int `json:"size"`
	// AvailabilityZone is which availability zone the volume is in.
	AvailabilityZone string `json:"availability_zone"`
	// The date when this volume was created.
	CreatedAt time.Time `json:"-"`
	// The date when this volume was last updated
	UpdatedAt time.Time `json:"-"`
	// Instances onto which the volume is attached.
	Attachments []Attachment `json:"attachments"`
	// Human-readable display name for the volume.
	Name string `json:"name"`
	// Human-readable description for the volume.
	Description string `json:"description"`
	// The type of volume to create, either SATA or SSD.
	VolumeType string `json:"volume_type"`
	// The ID of the snapshot from which the volume was created
	SnapshotID string `json:"snapshot_id"`
	// The ID of another block storage volume from which the current volume was created
	SourceVolID string `json:"source_volid"`
	// The backup ID, from which the volume was restored
	// This field is supported since 3.47 microversion
	BackupID *string `json:"backup_id"`
	// Arbitrary key-value pairs defined by the user.
	Metadata map[string]string `json:"metadata"`
	// UserID is the id of the user who created the volume.
	UserID string `json:"user_id"`
	// Indicates whether this is a bootable volume.
	Bootable string `json:"bootable"`
	// Encrypted denotes if the volume is encrypted.
	Encrypted bool `json:"encrypted"`
	// ReplicationStatus is the status of replication.
	ReplicationStatus string `json:"replication_status"`
	// ConsistencyGroupID is the consistency group ID.
	ConsistencyGroupID string `json:"consistencygroup_id"`
	// Multiattach denotes if the volume is multi-attach capable.
	Multiattach bool `json:"multiattach"`
	// Image metadata entries, only included for volumes that were created from an image, or from a snapshot of a volume originally created from an image.
	VolumeImageMetadata map[string]string `json:"volume_image_metadata"`
	// Host is the identifier of the host holding the volume.
	Host string `json:"os-vol-host-attr:host"`
	// TenantID is the id of the project that owns the volume.
	TenantID string `json:"os-vol-tenant-attr:tenant_id"`
}

// UnmarshalJSON another unmarshalling function
func (r *Volume) UnmarshalJSON(b []byte) error {
	type tmp Volume
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339MilliNoZ `json:"updated_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Volume(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)

	return err
}

// VolumePage is a pagination.pager that is returned from a call to the List function.
type VolumePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if a ListResult contains no Volumes.
func (r VolumePage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	volumes, err := ExtractVolumes(r)
	return len(volumes) == 0, err
}

func (page VolumePage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"volumes_links"`
	}
	err := page.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// ExtractVolumes extracts and returns Volumes. It is used while iterating over a volumes.List call.
func ExtractVolumes(r pagination.Page) ([]Volume, error) {
	var s []Volume
	err := ExtractVolumesInto(r, &s)
	return s, err
}

type commonResult struct {
	gophercloud.Result
}

// Extract will get the Volume object out of the commonResult object.
func (r commonResult) Extract() (*Volume, error) {
	var s Volume
	err := r.ExtractInto(&s)
	return &s, err
}

// ExtractInto converts our response data into a volume struct
func (r commonResult) ExtractInto(v any) error {
	return r.Result.ExtractIntoStructPtr(v, "volume")
}

// ExtractVolumesInto similar to ExtractInto but operates on a `list` of volumes
func ExtractVolumesInto(r pagination.Page, v any) error {
	return r.(VolumePage).Result.ExtractIntoSlicePtr(v, "volumes")
}

// CreateResult contains the response body and error from a Create request.
type CreateResult struct {
	commonResult
}

// GetResult contains the response body and error from a Get request.
type GetResult struct {
	commonResult
}

// UpdateResult contains the response body and error from an Update request.
type UpdateResult struct {
	commonResult
}

// DeleteResult contains the response body and error from a Delete request.
type DeleteResult struct {
	gophercloud.ErrResult
}

// AttachResult contains the response body and error from an Attach request.
type AttachResult struct {
	gophercloud.ErrResult
}

// BeginDetachingResult contains the response body and error from a BeginDetach
// request.
type BeginDetachingResult struct {
	gophercloud.ErrResult
}

// DetachResult contains the response body and error from a Detach request.
type DetachResult struct {
	gophercloud.ErrResult
}

// UploadImageResult contains the response body and error from an UploadImage
// request.
type UploadImageResult struct {
	gophercloud.Result
}

// SetImageMetadataResult contains the response body and error from an SetImageMetadata
// request.
type SetImageMetadataResult struct {
	gophercloud.ErrResult
}

// SetBootableResult contains the response body and error from a SetBootable
// request.
type SetBootableResult struct {
	gophercloud.ErrResult
}

// ReserveResult contains the response body and error from a Reserve request.
type ReserveResult struct {
	gophercloud.ErrResult
}

// UnreserveResult contains the response body and error from an Unreserve
// request.
type UnreserveResult struct {
	gophercloud.ErrResult
}

// TerminateConnectionResult contains the response body and error from a
// TerminateConnection request.
type TerminateConnectionResult struct {
	gophercloud.ErrResult
}

// InitializeConnectionResult contains the response body and error from an
// InitializeConnection request.
type InitializeConnectionResult struct {
	gophercloud.Result
}

// ExtendSizeResult contains the response body and error from an ExtendSize request.
type ExtendSizeResult struct {
	gophercloud.ErrResult
}

// Extract will get the connection information out of the
// InitializeConnectionResult object.
//
// This will be a generic map[string]any and the results will be
// dependent on the type of connection made.
func (r InitializeConnectionResult) Extract() (map[string]any, error) {
	var s struct {
		ConnectionInfo map[string]any `json:"connection_info"`
	}
	err := r.ExtractInto(&s)
	return s.ConnectionInfo, err
}

// ImageVolumeType contains volume type information obtained from UploadImage
// action.
type ImageVolumeType struct {
	// The ID of a volume type.
	ID string `json:"id"`

	// Human-readable display name for the volume type.
	Name string `json:"name"`

	// Human-readable description for the volume type.
	Description string `json:"display_description"`

	// Flag for public access.
	IsPublic bool `json:"is_public"`

	// Extra specifications for volume type.
	ExtraSpecs map[string]any `json:"extra_specs"`

	// ID of quality of service specs.
	QosSpecsID string `json:"qos_specs_id"`

	// Flag for deletion status of volume type.
	Deleted bool `json:"deleted"`

	// The date when volume type was deleted.
	DeletedAt time.Time `json:"-"`

	// The date when volume type was created.
	CreatedAt time.Time `json:"-"`

	// The date when this volume was last updated.
	UpdatedAt time.Time `json:"-"`
}

func (r *ImageVolumeType) UnmarshalJSON(b []byte) error {
	type tmp ImageVolumeType
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339MilliNoZ `json:"updated_at"`
		DeletedAt gophercloud.JSONRFC3339MilliNoZ `json:"deleted_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = ImageVolumeType(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)
	r.DeletedAt = time.Time(s.DeletedAt)

	return err
}

// VolumeImage contains information about volume uploaded to an image service.
type VolumeImage struct {
	// The ID of a volume an image is created from.
	VolumeID string `json:"id"`

	// Container format, may be bare, ofv, ova, etc.
	ContainerFormat string `json:"container_format"`

	// Disk format, may be raw, qcow2, vhd, vdi, vmdk, etc.
	DiskFormat string `json:"disk_format"`

	// Human-readable description for the volume.
	Description string `json:"display_description"`

	// The ID of the created image.
	ImageID string `json:"image_id"`

	// Human-readable display name for the image.
	ImageName string `json:"image_name"`

	// Size of the volume in GB.
	Size int `json:"size"`

	// Current status of the volume.
	Status string `json:"status"`

	// Visibility defines who can see/use the image.
	// supported since 3.1 microversion
	Visibility string `json:"visibility"`

	// whether the image is not deletable.
	// supported since 3.1 microversion
	Protected bool `json:"protected"`

	// The date when this volume was last updated.
	UpdatedAt time.Time `json:"-"`

	// Volume type object of used volume.
	VolumeType ImageVolumeType `json:"volume_type"`
}

func (r *VolumeImage) UnmarshalJSON(b []byte) error {
	type tmp VolumeImage
	var s struct {
		tmp
		UpdatedAt gophercloud.JSONRFC3339MilliNoZ `json:"updated_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = VolumeImage(s.tmp)

	r.UpdatedAt = time.Time(s.UpdatedAt)

	return err
}

// Extract will get an object with info about the uploaded image out of the
// UploadImageResult object.
func (r UploadImageResult) Extract() (VolumeImage, error) {
	var s struct {
		VolumeImage VolumeImage `json:"os-volume_upload_image"`
	}
	err := r.ExtractInto(&s)
	return s.VolumeImage, err
}

// ForceDeleteResult contains the response body and error from a ForceDelete request.
type ForceDeleteResult struct {
	gophercloud.ErrResult
}

// ChangeTypeResult contains the response body and error from an ChangeType request.
type ChangeTypeResult struct {
	gophercloud.ErrResult
}

// ReImageResult contains the response body and error from a ReImage request.
type ReImageResult struct {
	gophercloud.ErrResult
}

// ResetStatusResult contains the response error from a ResetStatus request.
type ResetStatusResult struct {
	gophercloud.ErrResult
}
