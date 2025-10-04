package volumes

import (
	"context"
	"maps"
	"regexp"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

// SchedulerHintOptsBuilder builds the scheduler hints into a serializable format.
type SchedulerHintOptsBuilder interface {
	ToSchedulerHintsMap() (map[string]any, error)
}

// SchedulerHintOpts contains options for providing scheduler hints
// when creating a Volume. This object is passed to the volumes.Create function.
// For more information about these parameters, see the Volume object.
type SchedulerHintOpts struct {
	// DifferentHost will place the volume on a different back-end that does not
	// host the given volumes.
	DifferentHost []string

	// SameHost will place the volume on a back-end that hosts the given volumes.
	SameHost []string

	// LocalToInstance will place volume on same host on a given instance
	LocalToInstance string

	// Query is a conditional statement that results in back-ends able to
	// host the volume.
	Query string

	// AdditionalProperies are arbitrary key/values that are not validated by nova.
	AdditionalProperties map[string]any
}

// ToSchedulerHintsMap assembles a request body for scheduler hints
func (opts SchedulerHintOpts) ToSchedulerHintsMap() (map[string]any, error) {
	sh := make(map[string]any)

	uuidRegex, _ := regexp.Compile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")

	if len(opts.DifferentHost) > 0 {
		for _, diffHost := range opts.DifferentHost {
			if !uuidRegex.MatchString(diffHost) {
				err := gophercloud.ErrInvalidInput{}
				err.Argument = "volumes.SchedulerHintOpts.DifferentHost"
				err.Value = opts.DifferentHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["different_host"] = opts.DifferentHost
	}

	if len(opts.SameHost) > 0 {
		for _, sameHost := range opts.SameHost {
			if !uuidRegex.MatchString(sameHost) {
				err := gophercloud.ErrInvalidInput{}
				err.Argument = "volumes.SchedulerHintOpts.SameHost"
				err.Value = opts.SameHost
				err.Info = "The hosts must be in UUID format."
				return nil, err
			}
		}
		sh["same_host"] = opts.SameHost
	}

	if opts.LocalToInstance != "" {
		if !uuidRegex.MatchString(opts.LocalToInstance) {
			err := gophercloud.ErrInvalidInput{}
			err.Argument = "volumes.SchedulerHintOpts.LocalToInstance"
			err.Value = opts.LocalToInstance
			err.Info = "The instance must be in UUID format."
			return nil, err
		}
		sh["local_to_instance"] = opts.LocalToInstance
	}

	if opts.Query != "" {
		sh["query"] = opts.Query
	}

	if opts.AdditionalProperties != nil {
		for k, v := range opts.AdditionalProperties {
			sh[k] = v
		}
	}

	if len(sh) == 0 {
		return sh, nil
	}

	return map[string]any{"OS-SCH-HNT:scheduler_hints": sh}, nil
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToVolumeCreateMap() (map[string]any, error)
}

// CreateOpts contains options for creating a Volume. This object is passed to
// the volumes.Create function. For more information about these parameters,
// see the Volume object.
type CreateOpts struct {
	// The size of the volume, in GB
	Size int `json:"size,omitempty"`
	// The availability zone
	AvailabilityZone string `json:"availability_zone,omitempty"`
	// ConsistencyGroupID is the ID of a consistency group
	ConsistencyGroupID string `json:"consistencygroup_id,omitempty"`
	// The volume description
	Description string `json:"description,omitempty"`
	// One or more metadata key and value pairs to associate with the volume
	Metadata map[string]string `json:"metadata,omitempty"`
	// The volume name
	Name string `json:"name,omitempty"`
	// the ID of the existing volume snapshot
	SnapshotID string `json:"snapshot_id,omitempty"`
	// SourceReplica is a UUID of an existing volume to replicate with
	SourceReplica string `json:"source_replica,omitempty"`
	// the ID of the existing volume
	SourceVolID string `json:"source_volid,omitempty"`
	// The ID of the image from which you want to create the volume.
	// Required to create a bootable volume.
	ImageID string `json:"imageRef,omitempty"`
	// Specifies the backup ID, from which you want to create the volume.
	// Create a volume from a backup is supported since 3.47 microversion
	BackupID string `json:"backup_id,omitempty"`
	// The associated volume type
	VolumeType string `json:"volume_type,omitempty"`
}

// ToVolumeCreateMap assembles a request body based on the contents of a
// CreateOpts.
func (opts CreateOpts) ToVolumeCreateMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "volume")
}

// Create will create a new Volume based on the values in CreateOpts. To extract
// the Volume object from the response, call the Extract method on the
// CreateResult.
func Create(ctx context.Context, client *gophercloud.ServiceClient, opts CreateOptsBuilder, hintOpts SchedulerHintOptsBuilder) (r CreateResult) {
	b, err := opts.ToVolumeCreateMap()
	if err != nil {
		r.Err = err
		return
	}

	if hintOpts != nil {
		sh, err := hintOpts.ToSchedulerHintsMap()
		if err != nil {
			r.Err = err
			return
		}
		maps.Copy(b, sh)
	}

	resp, err := client.Post(ctx, createURL(client), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// DeleteOptsBuilder allows extensions to add additional parameters to the
// Delete request.
type DeleteOptsBuilder interface {
	ToVolumeDeleteQuery() (string, error)
}

// DeleteOpts contains options for deleting a Volume. This object is passed to
// the volumes.Delete function.
type DeleteOpts struct {
	// Delete all snapshots of this volume as well.
	Cascade bool `q:"cascade"`
}

// ToLoadBalancerDeleteQuery formats a DeleteOpts into a query string.
func (opts DeleteOpts) ToVolumeDeleteQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// Delete will delete the existing Volume with the provided ID.
func Delete(ctx context.Context, client *gophercloud.ServiceClient, id string, opts DeleteOptsBuilder) (r DeleteResult) {
	url := deleteURL(client, id)
	if opts != nil {
		query, err := opts.ToVolumeDeleteQuery()
		if err != nil {
			r.Err = err
			return
		}
		url += query
	}
	resp, err := client.Delete(ctx, url, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Get retrieves the Volume with the provided ID. To extract the Volume object
// from the response, call the Extract method on the GetResult.
func Get(ctx context.Context, client *gophercloud.ServiceClient, id string) (r GetResult) {
	resp, err := client.Get(ctx, getURL(client, id), &r.Body, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ListOptsBuilder allows extensions to add additional parameters to the List
// request.
type ListOptsBuilder interface {
	ToVolumeListQuery() (string, error)
}

// ListOpts holds options for listing Volumes. It is passed to the volumes.List
// function.
type ListOpts struct {
	// AllTenants will retrieve volumes of all tenants/projects.
	AllTenants bool `q:"all_tenants"`

	// Metadata will filter results based on specified metadata.
	Metadata map[string]string `q:"metadata"`

	// Name will filter by the specified volume name.
	Name string `q:"name"`

	// Status will filter by the specified status.
	Status string `q:"status"`

	// TenantID will filter by a specific tenant/project ID.
	// Setting AllTenants is required for this.
	TenantID string `q:"project_id"`

	// Comma-separated list of sort keys and optional sort directions in the
	// form of <key>[:<direction>].
	Sort string `q:"sort"`

	// Requests a page size of items.
	Limit int `q:"limit"`

	// Used in conjunction with limit to return a slice of items.
	Offset int `q:"offset"`

	// The ID of the last-seen item.
	Marker string `q:"marker"`

	// Bootable will filter results based on whether they are bootable volumes
	Bootable *bool `q:"bootable,omitempty"`
}

// ToVolumeListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToVolumeListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// List returns Volumes optionally limited by the conditions provided in ListOpts.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := listURL(client)
	if opts != nil {
		query, err := opts.ToVolumeListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return VolumePage{pagination.LinkedPageBase{PageResult: r}}
	})
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToVolumeUpdateMap() (map[string]any, error)
}

// UpdateOpts contain options for updating an existing Volume. This object is passed
// to the volumes.Update function. For more information about the parameters, see
// the Volume object.
type UpdateOpts struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ToVolumeUpdateMap assembles a request body based on the contents of an
// UpdateOpts.
func (opts UpdateOpts) ToVolumeUpdateMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "volume")
}

// Update will update the Volume with provided information. To extract the updated
// Volume from the response, call the Extract method on the UpdateResult.
func Update(ctx context.Context, client *gophercloud.ServiceClient, id string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToVolumeUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Put(ctx, updateURL(client, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// AttachOptsBuilder allows extensions to add additional parameters to the
// Attach request.
type AttachOptsBuilder interface {
	ToVolumeAttachMap() (map[string]any, error)
}

// AttachMode describes the attachment mode for volumes.
type AttachMode string

// These constants determine how a volume is attached.
const (
	ReadOnly  AttachMode = "ro"
	ReadWrite AttachMode = "rw"
)

// AttachOpts contains options for attaching a Volume.
type AttachOpts struct {
	// The mountpoint of this volume.
	MountPoint string `json:"mountpoint,omitempty"`

	// The nova instance ID, can't set simultaneously with HostName.
	InstanceUUID string `json:"instance_uuid,omitempty"`

	// The hostname of baremetal host, can't set simultaneously with InstanceUUID.
	HostName string `json:"host_name,omitempty"`

	// Mount mode of this volume.
	Mode AttachMode `json:"mode,omitempty"`
}

// ToVolumeAttachMap assembles a request body based on the contents of a
// AttachOpts.
func (opts AttachOpts) ToVolumeAttachMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-attach")
}

// Attach will attach a volume based on the values in AttachOpts.
func Attach(ctx context.Context, client *gophercloud.ServiceClient, id string, opts AttachOptsBuilder) (r AttachResult) {
	b, err := opts.ToVolumeAttachMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// BeginDetaching will mark the volume as detaching.
func BeginDetaching(ctx context.Context, client *gophercloud.ServiceClient, id string) (r BeginDetachingResult) {
	b := map[string]any{"os-begin_detaching": make(map[string]any)}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// DetachOptsBuilder allows extensions to add additional parameters to the
// Detach request.
type DetachOptsBuilder interface {
	ToVolumeDetachMap() (map[string]any, error)
}

// DetachOpts contains options for detaching a Volume.
type DetachOpts struct {
	// AttachmentID is the ID of the attachment between a volume and instance.
	AttachmentID string `json:"attachment_id,omitempty"`
}

// ToVolumeDetachMap assembles a request body based on the contents of a
// DetachOpts.
func (opts DetachOpts) ToVolumeDetachMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-detach")
}

// Detach will detach a volume based on volume ID.
func Detach(ctx context.Context, client *gophercloud.ServiceClient, id string, opts DetachOptsBuilder) (r DetachResult) {
	b, err := opts.ToVolumeDetachMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Reserve will reserve a volume based on volume ID.
func Reserve(ctx context.Context, client *gophercloud.ServiceClient, id string) (r ReserveResult) {
	b := map[string]any{"os-reserve": make(map[string]any)}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Unreserve will unreserve a volume based on volume ID.
func Unreserve(ctx context.Context, client *gophercloud.ServiceClient, id string) (r UnreserveResult) {
	b := map[string]any{"os-unreserve": make(map[string]any)}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// InitializeConnectionOptsBuilder allows extensions to add additional parameters to the
// InitializeConnection request.
type InitializeConnectionOptsBuilder interface {
	ToVolumeInitializeConnectionMap() (map[string]any, error)
}

// InitializeConnectionOpts hosts options for InitializeConnection.
// The fields are specific to the storage driver in use and the destination
// attachment.
type InitializeConnectionOpts struct {
	IP        string   `json:"ip,omitempty"`
	Host      string   `json:"host,omitempty"`
	Initiator string   `json:"initiator,omitempty"`
	Wwpns     []string `json:"wwpns,omitempty"`
	Wwnns     string   `json:"wwnns,omitempty"`
	Multipath *bool    `json:"multipath,omitempty"`
	Platform  string   `json:"platform,omitempty"`
	OSType    string   `json:"os_type,omitempty"`
}

// ToVolumeInitializeConnectionMap assembles a request body based on the contents of a
// InitializeConnectionOpts.
func (opts InitializeConnectionOpts) ToVolumeInitializeConnectionMap() (map[string]any, error) {
	b, err := gophercloud.BuildRequestBody(opts, "connector")
	return map[string]any{"os-initialize_connection": b}, err
}

// InitializeConnection initializes an iSCSI connection by volume ID.
func InitializeConnection(ctx context.Context, client *gophercloud.ServiceClient, id string, opts InitializeConnectionOptsBuilder) (r InitializeConnectionResult) {
	b, err := opts.ToVolumeInitializeConnectionMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// TerminateConnectionOptsBuilder allows extensions to add additional parameters to the
// TerminateConnection request.
type TerminateConnectionOptsBuilder interface {
	ToVolumeTerminateConnectionMap() (map[string]any, error)
}

// TerminateConnectionOpts hosts options for TerminateConnection.
type TerminateConnectionOpts struct {
	IP        string   `json:"ip,omitempty"`
	Host      string   `json:"host,omitempty"`
	Initiator string   `json:"initiator,omitempty"`
	Wwpns     []string `json:"wwpns,omitempty"`
	Wwnns     string   `json:"wwnns,omitempty"`
	Multipath *bool    `json:"multipath,omitempty"`
	Platform  string   `json:"platform,omitempty"`
	OSType    string   `json:"os_type,omitempty"`
}

// ToVolumeTerminateConnectionMap assembles a request body based on the contents of a
// TerminateConnectionOpts.
func (opts TerminateConnectionOpts) ToVolumeTerminateConnectionMap() (map[string]any, error) {
	b, err := gophercloud.BuildRequestBody(opts, "connector")
	return map[string]any{"os-terminate_connection": b}, err
}

// TerminateConnection terminates an iSCSI connection by volume ID.
func TerminateConnection(ctx context.Context, client *gophercloud.ServiceClient, id string, opts TerminateConnectionOptsBuilder) (r TerminateConnectionResult) {
	b, err := opts.ToVolumeTerminateConnectionMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ExtendSizeOptsBuilder allows extensions to add additional parameters to the
// ExtendSize request.
type ExtendSizeOptsBuilder interface {
	ToVolumeExtendSizeMap() (map[string]any, error)
}

// ExtendSizeOpts contains options for extending the size of an existing Volume.
// This object is passed to the volumes.ExtendSize function.
type ExtendSizeOpts struct {
	// NewSize is the new size of the volume, in GB.
	NewSize int `json:"new_size" required:"true"`
}

// ToVolumeExtendSizeMap assembles a request body based on the contents of an
// ExtendSizeOpts.
func (opts ExtendSizeOpts) ToVolumeExtendSizeMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-extend")
}

// ExtendSize will extend the size of the volume based on the provided information.
// This operation does not return a response body.
func ExtendSize(ctx context.Context, client *gophercloud.ServiceClient, id string, opts ExtendSizeOptsBuilder) (r ExtendSizeResult) {
	b, err := opts.ToVolumeExtendSizeMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// UploadImageOptsBuilder allows extensions to add additional parameters to the
// UploadImage request.
type UploadImageOptsBuilder interface {
	ToVolumeUploadImageMap() (map[string]any, error)
}

// UploadImageOpts contains options for uploading a Volume to image storage.
type UploadImageOpts struct {
	// Container format, may be bare, ofv, ova, etc.
	ContainerFormat string `json:"container_format,omitempty"`

	// Disk format, may be raw, qcow2, vhd, vdi, vmdk, etc.
	DiskFormat string `json:"disk_format,omitempty"`

	// The name of image that will be stored in glance.
	ImageName string `json:"image_name,omitempty"`

	// Force image creation, usable if volume attached to instance.
	Force bool `json:"force,omitempty"`

	// Visibility defines who can see/use the image.
	// supported since 3.1 microversion
	Visibility string `json:"visibility,omitempty"`

	// whether the image is not deletable.
	// supported since 3.1 microversion
	Protected bool `json:"protected,omitempty"`
}

// ToVolumeUploadImageMap assembles a request body based on the contents of a
// UploadImageOpts.
func (opts UploadImageOpts) ToVolumeUploadImageMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-volume_upload_image")
}

// UploadImage will upload an image based on the values in UploadImageOptsBuilder.
func UploadImage(ctx context.Context, client *gophercloud.ServiceClient, id string, opts UploadImageOptsBuilder) (r UploadImageResult) {
	b, err := opts.ToVolumeUploadImageMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ForceDelete will delete the volume regardless of state.
func ForceDelete(ctx context.Context, client *gophercloud.ServiceClient, id string) (r ForceDeleteResult) {
	resp, err := client.Post(ctx, actionURL(client, id), map[string]any{"os-force_delete": ""}, nil, nil)
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ImageMetadataOptsBuilder allows extensions to add additional parameters to the
// ImageMetadataRequest request.
type ImageMetadataOptsBuilder interface {
	ToImageMetadataMap() (map[string]any, error)
}

// ImageMetadataOpts contains options for setting image metadata to a volume.
type ImageMetadataOpts struct {
	// The image metadata to add to the volume as a set of metadata key and value pairs.
	Metadata map[string]string `json:"metadata"`
}

// ToImageMetadataMap assembles a request body based on the contents of a
// ImageMetadataOpts.
func (opts ImageMetadataOpts) ToImageMetadataMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-set_image_metadata")
}

// SetImageMetadata will set image metadata on a volume based on the values in ImageMetadataOptsBuilder.
func SetImageMetadata(ctx context.Context, client *gophercloud.ServiceClient, id string, opts ImageMetadataOptsBuilder) (r SetImageMetadataResult) {
	b, err := opts.ToImageMetadataMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// BootableOptsBuilder allows extensions to add additional parameters to the
// SetBootable request.
type BootableOptsBuilder interface {
	ToBootableMap() (map[string]any, error)
}

// BootableOpts contains options for setting bootable status to a volume.
type BootableOpts struct {
	// Enables or disables the bootable attribute. You can boot an instance from a bootable volume.
	Bootable bool `json:"bootable"`
}

// ToBootableMap assembles a request body based on the contents of a
// BootableOpts.
func (opts BootableOpts) ToBootableMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-set_bootable")
}

// SetBootable will set bootable status on a volume based on the values in BootableOpts
func SetBootable(ctx context.Context, client *gophercloud.ServiceClient, id string, opts BootableOptsBuilder) (r SetBootableResult) {
	b, err := opts.ToBootableMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// MigrationPolicy type represents a migration_policy when changing types.
type MigrationPolicy string

// Supported attributes for MigrationPolicy attribute for changeType operations.
const (
	MigrationPolicyNever    MigrationPolicy = "never"
	MigrationPolicyOnDemand MigrationPolicy = "on-demand"
)

// ChangeTypeOptsBuilder allows extensions to add additional parameters to the
// ChangeType request.
type ChangeTypeOptsBuilder interface {
	ToVolumeChangeTypeMap() (map[string]any, error)
}

// ChangeTypeOpts contains options for changing the type of an existing Volume.
// This object is passed to the volumes.ChangeType function.
type ChangeTypeOpts struct {
	// NewType is the name of the new volume type of the volume.
	NewType string `json:"new_type" required:"true"`

	// MigrationPolicy specifies if the volume should be migrated when it is
	// re-typed. Possible values are "on-demand" or "never". If not specified,
	// the default is "never".
	MigrationPolicy MigrationPolicy `json:"migration_policy,omitempty"`
}

// ToVolumeChangeTypeMap assembles a request body based on the contents of an
// ChangeTypeOpts.
func (opts ChangeTypeOpts) ToVolumeChangeTypeMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-retype")
}

// ChangeType will change the volume type of the volume based on the provided information.
// This operation does not return a response body.
func ChangeType(ctx context.Context, client *gophercloud.ServiceClient, id string, opts ChangeTypeOptsBuilder) (r ChangeTypeResult) {
	b, err := opts.ToVolumeChangeTypeMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ReImageOptsBuilder allows extensions to add additional parameters to the
// ReImage request.
type ReImageOptsBuilder interface {
	ToReImageMap() (map[string]any, error)
}

// ReImageOpts contains options for Re-image a volume.
type ReImageOpts struct {
	// New image id
	ImageID string `json:"image_id"`
	// set true to re-image volumes in reserved state
	ReImageReserved bool `json:"reimage_reserved"`
}

// ToReImageMap assembles a request body based on the contents of a ReImageOpts.
func (opts ReImageOpts) ToReImageMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-reimage")
}

// ReImage will re-image a volume based on the values in ReImageOpts
func ReImage(ctx context.Context, client *gophercloud.ServiceClient, id string, opts ReImageOptsBuilder) (r ReImageResult) {
	b, err := opts.ToReImageMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// ResetStatusOptsBuilder allows extensions to add additional parameters to the
// ResetStatus request.
type ResetStatusOptsBuilder interface {
	ToResetStatusMap() (map[string]any, error)
}

// ResetStatusOpts contains options for resetting a Volume status.
// For more information about these parameters, please, refer to the Block Storage API V3,
// Volume Actions, ResetStatus volume documentation.
type ResetStatusOpts struct {
	// Status is a volume status to reset to.
	Status string `json:"status"`
	// MigrationStatus is a volume migration status to reset to.
	MigrationStatus string `json:"migration_status,omitempty"`
	// AttachStatus is a volume attach status to reset to.
	AttachStatus string `json:"attach_status,omitempty"`
}

// ToResetStatusMap assembles a request body based on the contents of a
// ResetStatusOpts.
func (opts ResetStatusOpts) ToResetStatusMap() (map[string]any, error) {
	return gophercloud.BuildRequestBody(opts, "os-reset_status")
}

// ResetStatus will reset the existing volume status. ResetStatusResult contains only the error.
// To extract it, call the ExtractErr method on the ResetStatusResult.
func ResetStatus(ctx context.Context, client *gophercloud.ServiceClient, id string, opts ResetStatusOptsBuilder) (r ResetStatusResult) {
	b, err := opts.ToResetStatusMap()
	if err != nil {
		r.Err = err
		return
	}

	resp, err := client.Post(ctx, actionURL(client, id), b, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

// Unmanage removes a volume from Block Storage management without
// removing the back-end storage object that is associated with it.
func Unmanage(ctx context.Context, client *gophercloud.ServiceClient, id string) (r UnmanageResult) {
	body := map[string]any{"os-unmanage": make(map[string]any)}
	resp, err := client.Post(ctx, actionURL(client, id), body, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}
