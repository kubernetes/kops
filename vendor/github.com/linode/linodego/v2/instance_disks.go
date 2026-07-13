package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// InstanceDisk represents an Instance Disk object
type InstanceDisk struct {
	ID             int                    `json:"id"`
	Label          string                 `json:"label"`
	Status         DiskStatus             `json:"status"`
	Size           int                    `json:"size"`
	Filesystem     DiskFilesystem         `json:"filesystem"`
	Created        *time.Time             `json:"-"`
	Updated        *time.Time             `json:"-"`
	DiskEncryption InstanceDiskEncryption `json:"disk_encryption"`
}

// DiskFilesystem constants start with Filesystem and include Linode API Filesystems
type DiskFilesystem string

// DiskFilesystem constants represent the filesystems types an Instance Disk may use
const (
	FilesystemRaw    DiskFilesystem = "raw"
	FilesystemSwap   DiskFilesystem = "swap"
	FilesystemExt3   DiskFilesystem = "ext3"
	FilesystemExt4   DiskFilesystem = "ext4"
	FilesystemInitrd DiskFilesystem = "initrd"
)

// DiskStatus constants have the prefix "Disk" and include Linode API Instance Disk Status
type DiskStatus string

// DiskStatus constants represent the status values an Instance Disk may have
const (
	DiskReady    DiskStatus = "ready"
	DiskNotReady DiskStatus = "not ready"
	DiskDeleting DiskStatus = "deleting"
)

// InstanceDiskCreateOptions are InstanceDisk settings that can be used at creation
type InstanceDiskCreateOptions struct {
	Label string `json:"label"`
	Size  int    `json:"size"`

	// Image is optional, but requires at least one of RootPass, AuthorizedUsers, or AuthorizedKeys if provided
	Image    string `json:"image,omitzero"`
	RootPass string `json:"root_pass,omitzero"`

	Filesystem      string            `json:"filesystem,omitzero"`
	AuthorizedKeys  []string          `json:"authorized_keys,omitzero"`
	AuthorizedUsers []string          `json:"authorized_users,omitzero"`
	StackscriptID   int               `json:"stackscript_id,omitzero"`
	StackscriptData map[string]string `json:"stackscript_data,omitzero"`
}

// InstanceDiskUpdateOptions are InstanceDisk settings that can be used in updates
type InstanceDiskUpdateOptions struct {
	Label string `json:"label"`
}

// InstanceDiskResizeOptions are InstanceDisk settings that can be used in resizes
type InstanceDiskResizeOptions struct {
	Size int `json:"size"`
}

// InstanceDiskPasswordResetOptions are InstanceDisk settings that can be used in password resets
type InstanceDiskPasswordResetOptions struct {
	Password string `json:"password"`
}

// ListInstanceDisks lists InstanceDisks
func (c *Client) ListInstanceDisks(ctx context.Context, linodeID int, opts *ListOptions) ([]InstanceDisk, error) {
	return getPaginatedResults[InstanceDisk](ctx, c, formatAPIPath("linode/instances/%d/disks", linodeID), opts)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *InstanceDisk) UnmarshalJSON(b []byte) error {
	type Mask InstanceDisk

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetInstanceDisk gets the template with the provided ID
func (c *Client) GetInstanceDisk(ctx context.Context, linodeID int, diskID int) (*InstanceDisk, error) {
	e := formatAPIPath("linode/instances/%d/disks/%d", linodeID, diskID)
	return doGETRequest[InstanceDisk](ctx, c, e)
}

// CreateInstanceDisk creates a new InstanceDisk for the given Instance
func (c *Client) CreateInstanceDisk(ctx context.Context, linodeID int, opts InstanceDiskCreateOptions) (*InstanceDisk, error) {
	e := formatAPIPath("linode/instances/%d/disks", linodeID)
	return doPOSTRequest[InstanceDisk](ctx, c, e, opts)
}

// UpdateInstanceDisk creates a new InstanceDisk for the given Instance
func (c *Client) UpdateInstanceDisk(ctx context.Context, linodeID int, diskID int, opts InstanceDiskUpdateOptions) (*InstanceDisk, error) {
	e := formatAPIPath("linode/instances/%d/disks/%d", linodeID, diskID)
	return doPUTRequest[InstanceDisk](ctx, c, e, opts)
}

// RenameInstanceDisk renames an InstanceDisk
func (c *Client) RenameInstanceDisk(ctx context.Context, linodeID int, diskID int, label string) (*InstanceDisk, error) {
	return c.UpdateInstanceDisk(ctx, linodeID, diskID, InstanceDiskUpdateOptions{Label: label})
}

// ResizeInstanceDisk resizes the size of the Instance disk
func (c *Client) ResizeInstanceDisk(ctx context.Context, linodeID int, diskID int, opts InstanceDiskResizeOptions) error {
	e := formatAPIPath("linode/instances/%d/disks/%d/resize", linodeID, diskID)

	return doPOSTRequestNoResponseBody(ctx, c, e, opts)
}

// PasswordResetInstanceDisk resets the "root" account password on the Instance disk
func (c *Client) PasswordResetInstanceDisk(ctx context.Context, linodeID int, diskID int, opts InstanceDiskPasswordResetOptions) error {
	e := formatAPIPath("linode/instances/%d/disks/%d/password", linodeID, diskID)

	return doPOSTRequestNoResponseBody(ctx, c, e, opts)
}

// DeleteInstanceDisk deletes a Linode Instance Disk
func (c *Client) DeleteInstanceDisk(ctx context.Context, linodeID int, diskID int) error {
	e := formatAPIPath("linode/instances/%d/disks/%d", linodeID, diskID)
	return doDELETERequest(ctx, c, e)
}

// CloneInstanceDisk clones the given InstanceDisk for the given Instance
func (c *Client) CloneInstanceDisk(ctx context.Context, linodeID, diskID int) (*InstanceDisk, error) {
	e := formatAPIPath("linode/instances/%d/disks/%d/clone", linodeID, diskID)
	return doPOSTRequestNoRequestBody[InstanceDisk](ctx, c, e)
}
