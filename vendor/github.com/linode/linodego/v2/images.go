package linodego

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// ImageStatus represents the status of an Image.
type ImageStatus string

// ImageStatus options start with ImageStatus and include all Image statuses
const (
	ImageStatusCreating      ImageStatus = "creating"
	ImageStatusPendingUpload ImageStatus = "pending_upload"
	ImageStatusAvailable     ImageStatus = "available"
)

// ImageRegionStatus represents the status of an Image's replica.
type ImageRegionStatus string

// ImageRegionStatus options start with ImageRegionStatus and
// include all Image replica statuses
const (
	ImageRegionStatusAvailable          ImageRegionStatus = "available"
	ImageRegionStatusCreating           ImageRegionStatus = "creating"
	ImageRegionStatusPending            ImageRegionStatus = "pending"
	ImageRegionStatusPendingReplication ImageRegionStatus = "pending replication"
	ImageRegionStatusPendingDeletion    ImageRegionStatus = "pending deletion"
	ImageRegionStatusReplicating        ImageRegionStatus = "replicating"
)

// ImageRegion represents the status of an Image object in a given Region.
type ImageRegion struct {
	Region string            `json:"region"`
	Status ImageRegionStatus `json:"status"`
}

// Image represents a deployable Image object for use with Linode Instances
type Image struct {
	ID           string      `json:"id"`
	CreatedBy    string      `json:"created_by"`
	Capabilities []string    `json:"capabilities"`
	Label        string      `json:"label"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	Vendor       string      `json:"vendor"`
	Status       ImageStatus `json:"status"`
	Size         int         `json:"size"`
	TotalSize    int         `json:"total_size"`
	IsPublic     bool        `json:"is_public"`

	// NOTE: IsShared may not currently be available to all users and can only be used with v4beta.
	IsShared bool `json:"is_shared"`

	Deprecated bool          `json:"deprecated"`
	Regions    []ImageRegion `json:"regions"`
	Tags       []string      `json:"tags"`

	Updated *time.Time `json:"-"`
	Created *time.Time `json:"-"`
	Expiry  *time.Time `json:"-"`
	EOL     *time.Time `json:"-"`

	// NOTE: ImageSharing may not currently be available to all users and can only be used with v4beta.
	ImageSharing ImageSharing `json:"image_sharing"`
}

type ImageSharing struct {
	SharedWith *ImageSharingSharedWith `json:"shared_with"`
	SharedBy   *ImageSharingSharedBy   `json:"shared_by"`
}

type ImageSharingSharedWith struct {
	ShareGroupCount   int    `json:"sharegroup_count"`
	ShareGroupListURL string `json:"sharegroup_list_url"`
}

type ImageSharingSharedBy struct {
	ShareGroupID    int     `json:"sharegroup_id"`
	ShareGroupUUID  string  `json:"sharegroup_uuid"`
	ShareGroupLabel string  `json:"sharegroup_label"`
	SourceImageID   *string `json:"source_image_id"`
}

// ImageShareEntry represents a shared image entry for an ImageShareGroup
type ImageShareEntry struct {
	ID           string        `json:"id"`
	CreatedBy    *string       `json:"created_by"`
	Capabilities []string      `json:"capabilities"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Type         string        `json:"type"`
	Vendor       *string       `json:"vendor"`
	Status       ImageStatus   `json:"status"`
	Size         int           `json:"size"`
	TotalSize    int           `json:"total_size"`
	IsPublic     bool          `json:"is_public"`
	IsShared     *bool         `json:"is_shared"`
	Deprecated   bool          `json:"deprecated"`
	Regions      []ImageRegion `json:"regions"`
	Tags         []string      `json:"tags"`

	Updated *time.Time `json:"-"`
	Created *time.Time `json:"-"`
	Expiry  *time.Time `json:"-"`
	EOL     *time.Time `json:"-"`

	ImageSharing ImageSharing `json:"image_sharing"`
}

// ImageCreateOptions fields are those accepted by CreateImage
type ImageCreateOptions struct {
	DiskID      int      `json:"disk_id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitzero"`
	CloudInit   bool     `json:"cloud_init,omitzero"`
	Tags        []string `json:"tags,omitzero"`
}

// ImageUpdateOptions fields are those accepted by UpdateImage
type ImageUpdateOptions struct {
	Label       string   `json:"label,omitzero"`
	Description *string  `json:"description,omitzero"`
	Tags        []string `json:"tags,omitzero"`
}

// ImageReplicateOptions represents the options accepted by the
// ReplicateImage(...) function.
type ImageReplicateOptions struct {
	Regions []string `json:"regions"`
}

// ImageCreateUploadResponse fields are those returned by CreateImageUpload
type ImageCreateUploadResponse struct {
	Image    *Image `json:"image"`
	UploadTo string `json:"upload_to"`
}

// ImageCreateUploadOptions fields are those accepted by CreateImageUpload
type ImageCreateUploadOptions struct {
	Region      string   `json:"region"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitzero"`
	CloudInit   bool     `json:"cloud_init,omitzero"`
	Tags        []string `json:"tags,omitzero"`
}

// ImageUploadOptions fields are those accepted by UploadImage
type ImageUploadOptions struct {
	Region      string   `json:"region"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitzero"`
	CloudInit   bool     `json:"cloud_init"`
	Tags        []string `json:"tags,omitzero"`
	Image       io.Reader
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *Image) UnmarshalJSON(b []byte) error {
	type Mask Image

	p := struct {
		*Mask

		Updated *parseabletime.ParseableTime `json:"updated"`
		Created *parseabletime.ParseableTime `json:"created"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
		EOL     *parseabletime.ParseableTime `json:"eol"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Updated = (*time.Time)(p.Updated)
	i.Created = (*time.Time)(p.Created)
	i.Expiry = (*time.Time)(p.Expiry)
	i.EOL = (*time.Time)(p.EOL)

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (ise *ImageShareEntry) UnmarshalJSON(b []byte) error {
	type Mask ImageShareEntry

	p := struct {
		*Mask

		Updated *parseabletime.ParseableTime `json:"updated"`
		Created *parseabletime.ParseableTime `json:"created"`
		Expiry  *parseabletime.ParseableTime `json:"expiry"`
		EOL     *parseabletime.ParseableTime `json:"eol"`
	}{
		Mask: (*Mask)(ise),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	ise.Updated = (*time.Time)(p.Updated)
	ise.Created = (*time.Time)(p.Created)
	ise.Expiry = (*time.Time)(p.Expiry)
	ise.EOL = (*time.Time)(p.EOL)

	return nil
}

// GetUpdateOptions converts an Image to ImageUpdateOptions for use in UpdateImage
func (i Image) GetUpdateOptions() (iu ImageUpdateOptions) {
	iu.Label = i.Label
	iu.Description = copyString(&i.Description)

	return iu
}

// ListImages lists Images.
func (c *Client) ListImages(ctx context.Context, opts *ListOptions) ([]Image, error) {
	return getPaginatedResults[Image](
		ctx,
		c,
		"images",
		opts,
	)
}

// GetImage gets the Image with the provided ID.
func (c *Client) GetImage(ctx context.Context, imageID string) (*Image, error) {
	return doGETRequest[Image](
		ctx,
		c,
		formatAPIPath("images/%s", imageID),
	)
}

// CreateImage creates an Image.
func (c *Client) CreateImage(ctx context.Context, opts ImageCreateOptions) (*Image, error) {
	return doPOSTRequest[Image](
		ctx,
		c,
		"images",
		opts,
	)
}

// UpdateImage updates the Image with the specified id.
func (c *Client) UpdateImage(ctx context.Context, imageID string, opts ImageUpdateOptions) (*Image, error) {
	return doPUTRequest[Image](
		ctx,
		c,
		formatAPIPath("images/%s", imageID),
		opts,
	)
}

// ReplicateImage replicates an image to a given set of regions.
func (c *Client) ReplicateImage(ctx context.Context, imageID string, opts ImageReplicateOptions) (*Image, error) {
	return doPOSTRequest[Image](
		ctx,
		c,
		formatAPIPath("images/%s/regions", imageID),
		opts,
	)
}

// DeleteImage deletes the Image with the specified id.
func (c *Client) DeleteImage(ctx context.Context, imageID string) error {
	return doDELETERequest(
		ctx,
		c,
		formatAPIPath("images/%s", imageID),
	)
}

// CreateImageUpload creates an Image and an upload URL.
func (c *Client) CreateImageUpload(ctx context.Context, opts ImageCreateUploadOptions) (*Image, string, error) {
	result, err := doPOSTRequest[ImageCreateUploadResponse](
		ctx,
		c,
		"images/upload",
		opts,
	)
	if err != nil {
		return nil, "", err
	}

	return result.Image, result.UploadTo, nil
}

// UploadImageToURL uploads the given image to the given upload URL.
func (c *Client) UploadImageToURL(ctx context.Context, uploadURL string, image io.Reader) error {
	clonedClient := *c.httpClient
	clonedClient.Transport = http.DefaultTransport

	var contentLength int64 = -1

	if seeker, ok := image.(io.Seeker); ok {
		size, err := seeker.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}

		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		contentLength = size
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, image)
	if err != nil {
		return err
	}

	if contentLength >= 0 {
		req.ContentLength = contentLength
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := clonedClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	_, err = coupleAPIErrors(resp, err)
	if err != nil {
		return err
	}

	return nil
}

// UploadImage creates and uploads an image.
func (c *Client) UploadImage(ctx context.Context, opts ImageUploadOptions) (*Image, error) {
	image, uploadURL, err := c.CreateImageUpload(ctx, ImageCreateUploadOptions{
		Label:       opts.Label,
		Region:      opts.Region,
		Description: opts.Description,
		CloudInit:   opts.CloudInit,
		Tags:        opts.Tags,
	})
	if err != nil {
		return nil, err
	}

	return image, c.UploadImageToURL(ctx, uploadURL, opts.Image)
}
