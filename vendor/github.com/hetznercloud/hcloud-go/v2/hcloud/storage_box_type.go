package hcloud

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// StorageBoxType represents a Storage Box type in the Hetzner API.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types
type StorageBoxType struct {
	ID                     int64
	Name                   string
	Description            string
	SnapshotLimit          *int
	AutomaticSnapshotLimit *int
	SubaccountsLimit       int
	Size                   int64
	Pricings               []StorageBoxTypeLocationPricing
	DeprecatableResource
}

type StorageBoxTypeLocationPricing struct {
	Location     string
	PriceHourly  Price
	PriceMonthly Price
	SetupFee     Price
}

// StorageBoxTypeClient provides access to Storage Box Types in the Hetzner API.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
type StorageBoxTypeClient struct {
	client *Client
}

// StorageBoxTypeListOpts specifies options for listing storage box types.
type StorageBoxTypeListOpts struct {
	ListOpts
	Name string
}

func (l StorageBoxTypeListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	return vals
}

// List returns a list of storage box types for a specific page.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-list-storage-box-types
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) List(ctx context.Context, opts StorageBoxTypeListOpts) ([]*StorageBoxType, *Response, error) {
	const opPath = "/storage_box_types?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.StorageBoxTypeListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.StorageBoxTypes, StorageBoxTypeFromSchema), resp, nil
}

// All returns all storage box types.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-list-storage-box-types
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) All(ctx context.Context) ([]*StorageBoxType, error) {
	return c.AllWithOpts(ctx, StorageBoxTypeListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all storage box types for the given options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-list-storage-box-types
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) AllWithOpts(ctx context.Context, opts StorageBoxTypeListOpts) ([]*StorageBoxType, error) {
	return iterPages(func(page int) ([]*StorageBoxType, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// GetByID returns a specific Storage Box Type by ID.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-get-a-storage-box-type
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) GetByID(ctx context.Context, id int64) (*StorageBoxType, *Response, error) {
	const opPath = "/storage_box_types/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.StorageBoxTypeGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return StorageBoxTypeFromSchema(respBody.StorageBoxType), resp, nil
}

// GetByName retrieves a Storage Box Type by its name. If the Storage Box Type does not exist, nil is returned.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-get-a-storage-box-type
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) GetByName(ctx context.Context, name string) (*StorageBoxType, *Response, error) {
	return firstByName(name, func() ([]*StorageBoxType, *Response, error) {
		return c.List(ctx, StorageBoxTypeListOpts{Name: name})
	})
}

// Get retrieves a Storage Box Type by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Storage Box Type by its name. If the Storage Box Type does not exist, nil is returned.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-types-get-a-storage-box-type
//
// Experimental: [StorageBoxTypeClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxTypeClient) Get(ctx context.Context, idOrName string) (*StorageBoxType, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}
