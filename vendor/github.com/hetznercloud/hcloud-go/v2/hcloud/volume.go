package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// Volume represents a volume in the Hetzner Cloud.
type Volume struct {
	ID          int64
	Name        string
	Status      VolumeStatus
	Server      *Server
	Location    *Location
	Size        int
	Format      *string
	Protection  VolumeProtection
	Labels      map[string]string
	LinuxDevice string
	Created     time.Time
}

const (
	VolumeFormatExt4 = "ext4"
	VolumeFormatXFS  = "xfs"
)

// VolumeProtection represents the protection level of a volume.
type VolumeProtection struct {
	Delete bool
}

// VolumeClient is a client for the volume API.
type VolumeClient struct {
	client *Client
	Action *ResourceActionClient
}

// VolumeStatus specifies a volume's status.
type VolumeStatus string

const (
	// VolumeStatusCreating is the status when a volume is being created.
	VolumeStatusCreating VolumeStatus = "creating"

	// VolumeStatusAvailable is the status when a volume is available.
	VolumeStatusAvailable VolumeStatus = "available"
)

// GetByID retrieves a volume by its ID. If the volume does not exist, nil is returned.
func (c *VolumeClient) GetByID(ctx context.Context, id int64) (*Volume, *Response, error) {
	const opPath = "/volumes/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.VolumeGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return VolumeFromSchema(respBody.Volume), resp, nil
}

// GetByName retrieves a volume by its name. If the volume does not exist, nil is returned.
func (c *VolumeClient) GetByName(ctx context.Context, name string) (*Volume, *Response, error) {
	return firstByName(name, func() ([]*Volume, *Response, error) {
		return c.List(ctx, VolumeListOpts{Name: name})
	})
}

// Get retrieves a volume by its ID if the input can be parsed as an integer, otherwise it
// retrieves a volume by its name. If the volume does not exist, nil is returned.
func (c *VolumeClient) Get(ctx context.Context, idOrName string) (*Volume, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// VolumeListOpts specifies options for listing volumes.
type VolumeListOpts struct {
	ListOpts
	Name   string
	Status []VolumeStatus
	Sort   []string
}

func (l VolumeListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, status := range l.Status {
		vals.Add("status", string(status))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of volumes for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *VolumeClient) List(ctx context.Context, opts VolumeListOpts) ([]*Volume, *Response, error) {
	const opPath = "/volumes?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.VolumeListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Volumes, VolumeFromSchema), resp, nil
}

// All returns all volumes.
func (c *VolumeClient) All(ctx context.Context) ([]*Volume, error) {
	return c.AllWithOpts(ctx, VolumeListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all volumes with the given options.
func (c *VolumeClient) AllWithOpts(ctx context.Context, opts VolumeListOpts) ([]*Volume, error) {
	return iterPages(func(page int) ([]*Volume, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// VolumeCreateOpts specifies parameters for creating a volume.
type VolumeCreateOpts struct {
	Name      string
	Size      int
	Server    *Server
	Location  *Location
	Labels    map[string]string
	Automount *bool
	Format    *string
}

// Validate checks if options are valid.
func (o VolumeCreateOpts) Validate() error {
	if o.Name == "" {
		return missingField(o, "Name")
	}
	if o.Size <= 0 {
		return invalidFieldValue(o, "Size", o.Size)
	}
	if o.Server == nil && o.Location == nil {
		return missingOneOfFields(o, "Server", "Location")
	}
	if o.Server != nil && o.Location != nil {
		return mutuallyExclusiveFields(o, "Server", "Location")
	}
	if o.Server == nil && (o.Automount != nil && *o.Automount) {
		return missingRequiredTogetherFields(o, "Automount", "Server")
	}
	return nil
}

// VolumeCreateResult is the result of creating a volume.
type VolumeCreateResult struct {
	Volume      *Volume
	Action      *Action
	NextActions []*Action
}

// Create creates a new volume with the given options.
func (c *VolumeClient) Create(ctx context.Context, opts VolumeCreateOpts) (VolumeCreateResult, *Response, error) {
	const opPath = "/volumes"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := VolumeCreateResult{}

	reqPath := opPath

	if err := opts.Validate(); err != nil {
		return result, nil, err
	}
	reqBody := schema.VolumeCreateRequest{
		Name:      opts.Name,
		Size:      opts.Size,
		Automount: opts.Automount,
		Format:    opts.Format,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	if opts.Server != nil {
		reqBody.Server = Ptr(opts.Server.ID)
	}
	if opts.Location != nil {
		if opts.Location.ID != 0 || opts.Location.Name != "" {
			reqBody.Location = &schema.IDOrName{ID: opts.Location.ID, Name: opts.Location.Name}
		}
	}

	respBody, resp, err := postRequest[schema.VolumeCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Volume = VolumeFromSchema(respBody.Volume)
	if respBody.Action != nil {
		result.Action = ActionFromSchema(*respBody.Action)
	}
	result.NextActions = ActionsFromSchema(respBody.NextActions)

	return result, resp, nil
}

// Delete deletes a volume.
func (c *VolumeClient) Delete(ctx context.Context, volume *Volume) (*Response, error) {
	const opPath = "/volumes/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// VolumeUpdateOpts specifies options for updating a volume.
type VolumeUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a volume.
func (c *VolumeClient) Update(ctx context.Context, volume *Volume, opts VolumeUpdateOpts) (*Volume, *Response, error) {
	const opPath = "/volumes/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	reqBody := schema.VolumeUpdateRequest{
		Name: opts.Name,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.VolumeUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return VolumeFromSchema(respBody.Volume), resp, nil
}

// VolumeAttachOpts specifies options for attaching a volume.
type VolumeAttachOpts struct {
	Server    *Server
	Automount *bool
}

// AttachWithOpts attaches a volume to a server.
func (c *VolumeClient) AttachWithOpts(ctx context.Context, volume *Volume, opts VolumeAttachOpts) (*Action, *Response, error) {
	const opPath = "/volumes/%d/actions/attach"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	reqBody := schema.VolumeActionAttachVolumeRequest{
		Server:    opts.Server.ID,
		Automount: opts.Automount,
	}

	respBody, resp, err := postRequest[schema.VolumeActionAttachVolumeResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Attach attaches a volume to a server.
func (c *VolumeClient) Attach(ctx context.Context, volume *Volume, server *Server) (*Action, *Response, error) {
	return c.AttachWithOpts(ctx, volume, VolumeAttachOpts{Server: server})
}

// Detach detaches a volume from a server.
func (c *VolumeClient) Detach(ctx context.Context, volume *Volume) (*Action, *Response, error) {
	const opPath = "/volumes/%d/actions/detach"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	var reqBody schema.VolumeActionDetachVolumeRequest

	respBody, resp, err := postRequest[schema.VolumeActionDetachVolumeResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// VolumeChangeProtectionOpts specifies options for changing the resource protection level of a volume.
type VolumeChangeProtectionOpts struct {
	Delete *bool
}

// ChangeProtection changes the resource protection level of a volume.
func (c *VolumeClient) ChangeProtection(ctx context.Context, volume *Volume, opts VolumeChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/volumes/%d/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	reqBody := schema.VolumeActionChangeProtectionRequest{
		Delete: opts.Delete,
	}

	respBody, resp, err := postRequest[schema.VolumeActionChangeProtectionResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Resize changes the size of a volume.
func (c *VolumeClient) Resize(ctx context.Context, volume *Volume, size int) (*Action, *Response, error) {
	const opPath = "/volumes/%d/actions/resize"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, volume.ID)

	reqBody := schema.VolumeActionResizeVolumeRequest{
		Size: size,
	}

	respBody, resp, err := postRequest[schema.VolumeActionResizeVolumeResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}
