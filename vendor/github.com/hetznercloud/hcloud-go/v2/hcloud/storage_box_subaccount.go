package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// StorageBoxSubaccount represents a subaccount of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts
type StorageBoxSubaccount struct {
	ID             int64
	Username       string
	HomeDirectory  string
	Server         string
	AccessSettings *StorageBoxSubaccountAccessSettings
	Description    string
	Labels         map[string]string
	Created        time.Time
	StorageBox     *StorageBox
}

// StorageBoxSubaccountAccessSettings represents the access settings of a [StorageBoxSubaccount].
type StorageBoxSubaccountAccessSettings struct {
	ReachableExternally bool
	Readonly            bool
	SambaEnabled        bool
	SSHEnabled          bool
	WebDAVEnabled       bool
}

// GetSubaccount retrieves a [StorageBoxSubaccount] either by its ID or by its username, depending on whether
// the input can be parsed as an integer. If no matching [StorageBoxSubaccount] is found, it returns nil.
//
// When fetching by ID, see https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-get-a-subaccount
// When fetching by name, see https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-list-subaccounts
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSubaccount(
	ctx context.Context,
	storageBox *StorageBox,
	idOrUsername string,
) (*StorageBoxSubaccount, *Response, error) {
	return getByIDOrName(
		ctx,
		func(ctx context.Context, id int64) (*StorageBoxSubaccount, *Response, error) {
			return c.GetSubaccountByID(ctx, storageBox, id)
		},
		func(ctx context.Context, username string) (*StorageBoxSubaccount, *Response, error) {
			return c.GetSubaccountByUsername(ctx, storageBox, username)
		},
		idOrUsername,
	)
}

// GetSubaccountByID retrieves a [StorageBoxSubaccount] by its ID.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-get-a-subaccount
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSubaccountByID(
	ctx context.Context,
	storageBox *StorageBox,
	id int64,
) (*StorageBoxSubaccount, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, storageBox.ID, id)

	respBody, resp, err := getRequest[schema.StorageBoxSubaccountGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return StorageBoxSubaccountFromSchema(respBody.Subaccount), resp, nil
}

// GetSubaccountByUsername retrieves a [StorageBoxSubaccount] by its username.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-list-subaccounts
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) GetSubaccountByUsername(
	ctx context.Context,
	storageBox *StorageBox,
	username string,
) (*StorageBoxSubaccount, *Response, error) {
	return firstByName(username, func() ([]*StorageBoxSubaccount, *Response, error) {
		return c.ListSubaccounts(ctx, storageBox, StorageBoxSubaccountListOpts{
			Username: username,
		})
	})
}

// StorageBoxSubaccountListOpts represents the options for listing [StorageBoxSubaccount].
type StorageBoxSubaccountListOpts struct {
	LabelSelector string
	Username      string
	Sort          []string
}

func (o StorageBoxSubaccountListOpts) values() url.Values {
	vals := url.Values{}
	if o.Username != "" {
		vals.Add("username", o.Username)
	}
	if o.LabelSelector != "" {
		vals.Add("label_selector", o.LabelSelector)
	}
	for _, sort := range o.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// ListSubaccounts lists all [StorageBoxSubaccount] of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-list-subaccounts
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) ListSubaccounts(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSubaccountListOpts,
) ([]*StorageBoxSubaccount, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, storageBox.ID, opts.values().Encode())

	respBody, resp, err := getRequest[schema.StorageBoxSubaccountListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Subaccounts, StorageBoxSubaccountFromSchema), resp, nil
}

// AllSubaccountsWithOpts retrieves all [StorageBoxSubaccount] of a [StorageBox] with the given options.
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-list-subaccounts
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) AllSubaccountsWithOpts(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSubaccountListOpts,
) ([]*StorageBoxSubaccount, error) {
	subaccounts, _, err := c.ListSubaccounts(ctx, storageBox, opts)
	return subaccounts, err
}

// AllSubaccounts retrieves all [StorageBoxSubaccount] of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-list-subaccounts
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) AllSubaccounts(
	ctx context.Context,
	storageBox *StorageBox,
) ([]*StorageBoxSubaccount, error) {
	opts := StorageBoxSubaccountListOpts{}
	subaccounts, _, err := c.ListSubaccounts(ctx, storageBox, opts)
	return subaccounts, err
}

// StorageBoxSubaccountCreateOpts represents the options for creating a [StorageBoxSubaccount].
type StorageBoxSubaccountCreateOpts struct {
	HomeDirectory  string
	Password       string
	Description    string
	AccessSettings *StorageBoxSubaccountCreateOptsAccessSettings
	Labels         map[string]string
}

// StorageBoxSubaccountAccessSettingsOpts represents the options for [StorageBoxSubaccountAccessSettings] of a [StorageBoxSubaccount].
type StorageBoxSubaccountCreateOptsAccessSettings struct {
	ReachableExternally *bool
	Readonly            *bool
	SambaEnabled        *bool
	SSHEnabled          *bool
	WebDAVEnabled       *bool
}

// StorageBoxSubaccountCreateResult represents the result of creating a [StorageBoxSubaccount].
type StorageBoxSubaccountCreateResult struct {
	Subaccount *StorageBoxSubaccount
	Action     *Action
}

// CreateSubaccount creates a new [StorageBoxSubaccount] for a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-create-a-subaccount
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) CreateSubaccount(
	ctx context.Context,
	storageBox *StorageBox,
	opts StorageBoxSubaccountCreateOpts,
) (StorageBoxSubaccountCreateResult, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, storageBox.ID)
	reqBody := SchemaFromStorageBoxSubaccountCreateOpts(opts)

	result := StorageBoxSubaccountCreateResult{}

	respBody, resp, err := postRequest[schema.StorageBoxSubaccountCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)
	result.Subaccount = StorageBoxSubaccountFromCreateResponse(respBody.Subaccount)

	return result, resp, nil
}

// StorageBoxSubaccountUpdateOpts represents the options for updating a [StorageBoxSubaccount].
type StorageBoxSubaccountUpdateOpts struct {
	Description *string
	Labels      map[string]string
}

// UpdateSubaccount updates a [StorageBoxSubaccount] of a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-update-a-subaccount
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) UpdateSubaccount(
	ctx context.Context,
	subaccount *StorageBoxSubaccount,
	opts StorageBoxSubaccountUpdateOpts,
) (*StorageBoxSubaccount, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, subaccount.StorageBox.ID, subaccount.ID)
	reqBody := SchemaFromStorageBoxSubaccountUpdateOpts(opts)

	respBody, resp, err := putRequest[schema.StorageBoxSubaccountUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return StorageBoxSubaccountFromSchema(respBody.Subaccount), resp, nil
}

// StorageBoxSubaccountDeleteResult represents the result of deleting a [StorageBoxSubaccount].
type StorageBoxSubaccountDeleteResult struct {
	Action *Action
}

// DeleteSubaccount deletes a [StorageBoxSubaccount] from a [StorageBox].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts-delete-a-subaccount
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) DeleteSubaccount(
	ctx context.Context,
	subaccount *StorageBoxSubaccount,
) (StorageBoxSubaccountDeleteResult, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, subaccount.StorageBox.ID, subaccount.ID)
	result := StorageBoxSubaccountDeleteResult{}

	respBody, resp, err := deleteRequest[schema.ActionGetResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// StorageBoxSubaccountResetPasswordOpts represents the options for resetting a [StorageBoxSubaccount]'s password.
type StorageBoxSubaccountResetPasswordOpts struct {
	Password string
}

// ResetSubaccountPassword resets the password of a [StorageBoxSubaccount].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccount-actions-reset-password
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) ResetSubaccountPassword(
	ctx context.Context,
	subaccount *StorageBoxSubaccount,
	opts StorageBoxSubaccountResetPasswordOpts,
) (*Action, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d/actions/reset_subaccount_password"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, subaccount.StorageBox.ID, subaccount.ID)
	reqBody := SchemaFromStorageBoxSubaccountResetPasswordOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// StorageBoxSubaccountUpdateAccessSettingsOpts represents the options for updating [StorageBoxSubaccountAccessSettings] of a [StorageBoxSubaccount].
type StorageBoxSubaccountUpdateAccessSettingsOpts struct {
	ReachableExternally *bool
	Readonly            *bool
	SambaEnabled        *bool
	SSHEnabled          *bool
	WebDAVEnabled       *bool
}

// UpdateSubaccountAccessSettings updates the [StorageBoxSubaccountAccessSettings] of a [StorageBoxSubaccount].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccount-actions-update-access-settings
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) UpdateSubaccountAccessSettings(
	ctx context.Context,
	subaccount *StorageBoxSubaccount,
	opts StorageBoxSubaccountUpdateAccessSettingsOpts,
) (*Action, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d/actions/update_access_settings"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, subaccount.StorageBox.ID, subaccount.ID)
	reqBody := SchemaFromStorageBoxSubaccountUpdateAccessSettingsOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// StorageBoxSubaccountChangeHomeDirectoryOpts represents the options for changing the home directory of a [StorageBoxSubaccount].
type StorageBoxSubaccountChangeHomeDirectoryOpts struct {
	HomeDirectory string
}

// UpdateSubaccountAccessSettings changes the home directory of a [StorageBoxSubaccount].
//
// See https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccount-actions-change-home-directory
//
// Experimental: [StorageBoxClient] is experimental, breaking changes may occur within minor releases.
func (c *StorageBoxClient) ChangeSubaccountHomeDirectory(
	ctx context.Context,
	subaccount *StorageBoxSubaccount,
	opts StorageBoxSubaccountChangeHomeDirectoryOpts,
) (*Action, *Response, error) {
	const opPath = "/storage_boxes/%d/subaccounts/%d/actions/change_home_directory"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, subaccount.StorageBox.ID, subaccount.ID)
	reqBody := SchemaFromStorageBoxSubaccountChangeHomeDirectoryOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}
