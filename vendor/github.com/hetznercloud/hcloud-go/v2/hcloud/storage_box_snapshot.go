package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// StorageBoxSnapshot represents a snapshot of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots
type StorageBoxSnapshot struct {
	ID          int64
	Name        string
	Description string
	Stats       StorageBoxSnapshotStats
	IsAutomatic bool
	Labels      map[string]string
	Created     time.Time
	StorageBox  *StorageBox
}

// StorageBoxSnapshotStats represents the size of a [StorageBoxSnapshot].
type StorageBoxSnapshotStats struct {
	Size           uint64
	SizeFilesystem uint64
}

// GetSnapshotByID gets a [StorageBoxSnapshot] by its ID.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-get-a-snapshot
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSnapshotByID(ctx context.Context, storageBox *StorageBox, id int64) (*StorageBoxSnapshot, *Response, error) {
	const optPath = "/storage_boxes/%d/snapshots/%d"
	ctx = ctxutil.SetOpPath(ctx, optPath)

	reqPath := fmt.Sprintf(optPath, storageBox.ID, id)

	respBody, resp, err := getRequest[schema.StorageBoxSnapshotGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return StorageBoxSnapshotFromSchema(respBody.Snapshot), resp, nil
}

// GetSnapshotByName gets a [StorageBoxSnapshot] by its name.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-list-snapshots
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSnapshotByName(
	ctx context.Context,
	storageBox *StorageBox,
	name string,
) (*StorageBoxSnapshot, *Response, error) {
	return firstByName(name, func() ([]*StorageBoxSnapshot, *Response, error) {
		return c.ListSnapshots(ctx, storageBox, StorageBoxSnapshotListOpts{Name: name})
	})
}

// GetSnapshot retrieves a [StorageBoxSnapshot] either by its ID or by its name, depending on whether
// the input can be parsed as an integer. If no matching [StorageBoxSnapshot] is found, it returns nil.
//
// When fetching by ID, see https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-get-a-snapshot
// When fetching by name, see https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-list-snapshots
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSnapshot(
	ctx context.Context,
	storageBox *StorageBox,
	idOrName string,
) (*StorageBoxSnapshot, *Response, error) {
	return getByIDOrName(
		ctx,
		func(ctx context.Context, id int64) (*StorageBoxSnapshot, *Response, error) {
			return c.GetSnapshotByID(ctx, storageBox, id)
		},
		func(ctx context.Context, name string) (*StorageBoxSnapshot, *Response, error) {
			return c.GetSnapshotByName(ctx, storageBox, name)
		},
		idOrName,
	)
}

// StorageBoxSnapshotListOpts specifies options for listing [StorageBoxSnapshot].
type StorageBoxSnapshotListOpts struct {
	LabelSelector string
	Name          string
	IsAutomatic   *bool
	Sort          []string
}

func (o StorageBoxSnapshotListOpts) values() url.Values {
	vals := url.Values{}
	if o.LabelSelector != "" {
		vals.Set("label_selector", o.LabelSelector)
	}
	if o.Name != "" {
		vals.Set("name", o.Name)
	}
	if o.IsAutomatic != nil {
		vals.Set("is_automatic", fmt.Sprintf("%t", *o.IsAutomatic))
	}
	for _, sort := range o.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// ListSnapshots lists all [StorageBoxSnapshot] of a [StorageBox] with the given options.
//
// Pagination is not supported, so this will return all [StorageBoxSnapshot] at once.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-list-snapshots
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) ListSnapshots(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSnapshotListOpts,
) ([]*StorageBoxSnapshot, *Response, error) {
	const optPath = "/storage_boxes/%d/snapshots?%s"
	ctx = ctxutil.SetOpPath(ctx, optPath)

	reqPath := fmt.Sprintf(optPath, storageBox.ID, opts.values().Encode())

	respBody, resp, err := getRequest[schema.StorageBoxSnapshotListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Snapshots, StorageBoxSnapshotFromSchema), resp, nil
}

// AllSnapshotsWithOpts lists all [StorageBoxSnapshot] of a [StorageBox] with the given options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-list-snapshots
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) AllSnapshotsWithOpts(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSnapshotListOpts,
) ([]*StorageBoxSnapshot, error) {
	snapshots, _, err := c.ListSnapshots(ctx, storageBox, opts)
	return snapshots, err
}

// AllSnapshots lists all [StorageBoxSnapshot] of a [StorageBox] without any options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-list-snapshots
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) AllSnapshots(
	ctx context.Context,
	storageBox *StorageBox,
) ([]*StorageBoxSnapshot, error) {
	opts := StorageBoxSnapshotListOpts{}
	snapshots, _, err := c.ListSnapshots(ctx, storageBox, opts)
	return snapshots, err
}

// StorageBoxSnapshotCreateOpts specifies options for creating a [StorageBoxSnapshot].
type StorageBoxSnapshotCreateOpts struct {
	Description string
	Labels      map[string]string
}

// StorageBoxSnapshotCreateResult represents the result of creating a [StorageBoxSnapshot].
type StorageBoxSnapshotCreateResult struct {
	Snapshot *StorageBoxSnapshot
	Action   *Action
}

// CreateSnapshot creates a new [StorageBoxSnapshot] for the given [StorageBox] with the provided options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-create-a-snapshot
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) CreateSnapshot(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSnapshotCreateOpts,
) (StorageBoxSnapshotCreateResult, *Response, error) {
	const opPath = "/storage_boxes/%d/snapshots"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, storageBox.ID)
	reqBody := SchemaFromStorageBoxSnapshotCreateOpts(opts)

	result := StorageBoxSnapshotCreateResult{}

	respBody, resp, err := postRequest[schema.StorageBoxSnapshotCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Snapshot = StorageBoxSnapshotFromSchema(respBody.Snapshot)
	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, err
}

// StorageBoxSnapshotUpdateOpts specifies options for updating a [StorageBoxSnapshot].
type StorageBoxSnapshotUpdateOpts struct {
	Description *string
	Labels      map[string]string
}

// UpdateSnapshot updates the given [StorageBoxSnapshot] of a [StorageBox] with the provided options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-update-a-snapshot
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) UpdateSnapshot(
	ctx context.Context,
	snapshot *StorageBoxSnapshot,
	opts StorageBoxSnapshotUpdateOpts,
) (*StorageBoxSnapshot, *Response, error) {
	const opPath = "/storage_boxes/%d/snapshots/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, snapshot.StorageBox.ID, snapshot.ID)
	reqBody := SchemaFromStorageBoxSnapshotUpdateOpts(opts)

	respBody, resp, err := putRequest[schema.StorageBoxSnapshotUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return StorageBoxSnapshotFromSchema(respBody.Snapshot), resp, nil
}

// StorageBoxSnapshotDeleteResult represents the result of deleting a [StorageBoxSnapshot].
type StorageBoxSnapshotDeleteResult struct {
	Action *Action
}

// DeleteSnapshot deletes the given [StorageBoxSnapshot] of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots-delete-a-snapshot
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) DeleteSnapshot(
	ctx context.Context,
	snapshot *StorageBoxSnapshot,
) (StorageBoxSnapshotDeleteResult, *Response, error) {
	const opPath = "/storage_boxes/%d/snapshots/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, snapshot.StorageBox.ID, snapshot.ID)
	result := StorageBoxSnapshotDeleteResult{}

	respBody, resp, err := deleteRequest[schema.ActionGetResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, err
}
