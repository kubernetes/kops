package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// Zone represents a Zone in the Hetzner Cloud.
//
// See https://docs.hetzner.cloud/reference/cloud#zones
type Zone struct {
	ID                       int64
	Name                     string
	Created                  time.Time
	TTL                      int
	Mode                     ZoneMode
	PrimaryNameservers       []ZonePrimaryNameserver
	Protection               ZoneProtection
	Labels                   map[string]string
	RecordCount              int
	AuthoritativeNameservers ZoneAuthoritativeNameservers
	Registrar                ZoneRegistrar
	Status                   ZoneStatus
}

// ZoneMode represents the mode of a [Zone].
type ZoneMode string

const (
	ZoneModePrimary   ZoneMode = "primary"
	ZoneModeSecondary ZoneMode = "secondary"
)

// ZoneRegistrar represents the registrar of a [Zone].
type ZoneRegistrar string

const (
	ZoneRegistrarHetzner ZoneRegistrar = "hetzner"
	ZoneRegistrarOther   ZoneRegistrar = "other"
	ZoneRegistrarUnknown ZoneRegistrar = "unknown"
)

// ZoneStatus represents the status of a [Zone].
type ZoneStatus string

const (
	ZoneStatusOk       ZoneStatus = "ok"
	ZoneStatusUpdating ZoneStatus = "updating"
	ZoneStatusError    ZoneStatus = "error"
)

// ZoneProtection represents the protection of a [Zone].
type ZoneProtection struct {
	Delete bool
}

// ZoneTSIGAlgorithm represents the algorithm of a TSIG key of a [ZonePrimaryNameserver].
type ZoneTSIGAlgorithm string

const (
	ZoneTSIGAlgorithmHMACMD5    = "hmac-md5"
	ZoneTSIGAlgorithmHMACSHA1   = "hmac-sha1"
	ZoneTSIGAlgorithmHMACSHA256 = "hmac-sha256"
)

// ZonePrimaryNameserver represents a primary nameserver of a [Zone].
type ZonePrimaryNameserver struct {
	Address       string
	Port          int
	TSIGAlgorithm ZoneTSIGAlgorithm
	TSIGKey       string
}

// ZoneDelegationStatus represents the status of the delegation of a [Zone].
type ZoneDelegationStatus string

const (
	ZoneDelegationStatusValid          ZoneDelegationStatus = "valid"
	ZoneDelegationStatusPartiallyValid ZoneDelegationStatus = "partially-valid"
	ZoneDelegationStatusInvalid        ZoneDelegationStatus = "invalid"
	ZoneDelegationStatusLame           ZoneDelegationStatus = "lame"
	ZoneDelegationStatusUnregistered   ZoneDelegationStatus = "unregistered"
	ZoneDelegationStatusUnknown        ZoneDelegationStatus = "unknown"
)

// ZoneAuthoritativeNameservers represents the authoritative Hetzner nameservers assigned to a [Zone].
type ZoneAuthoritativeNameservers struct {
	Assigned            []string
	Delegated           []string
	DelegationLastCheck time.Time
	DelegationStatus    ZoneDelegationStatus
}

func (o *Zone) idOrName() (string, error) {
	switch {
	case o.ID != 0:
		return strconv.FormatInt(o.ID, 10), nil
	case o.Name != "":
		return o.Name, nil
	default:
		return "", missingOneOfFields(o, "ID", "Name")
	}
}

// ZoneClient is a client for the Zone (DNS) API.
//
// See https://docs.hetzner.cloud/reference/cloud#zones and
// https://docs.hetzner.cloud/reference/cloud#zone-rrsets.
type ZoneClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID returns a single [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zones-get-a-zone
func (c *ZoneClient) GetByID(ctx context.Context, id int64) (*Zone, *Response, error) {
	return c.getByIDOrName(ctx, strconv.FormatInt(id, 10))
}

// GetByName returns a single [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zones-get-a-zone
func (c *ZoneClient) GetByName(ctx context.Context, name string) (*Zone, *Response, error) {
	return c.getByIDOrName(ctx, name)
}

// Get returns a single [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zones-get-a-zone
func (c *ZoneClient) Get(ctx context.Context, idOrName string) (*Zone, *Response, error) {
	return c.getByIDOrName(ctx, idOrName)
}

func (c *ZoneClient) getByIDOrName(ctx context.Context, idOrName string) (*Zone, *Response, error) {
	const opPath = "/zones/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, idOrName)

	respBody, resp, err := getRequest[schema.ZoneGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}
	return ZoneFromSchema(respBody.Zone), resp, nil
}

// ZoneListOpts defines options for listing [Zone]s.
type ZoneListOpts struct {
	ListOpts
	Name string
	Mode ZoneMode
	Sort []string
}

func (l ZoneListOpts) values() url.Values {
	result := l.ListOpts.Values()
	if l.Name != "" {
		result.Add("name", l.Name)
	}
	if l.Mode != "" {
		result.Add("mode", string(l.Mode))
	}
	for _, value := range l.Sort {
		result.Add("sort", value)
	}
	return result
}

// List returns a list of [Zone] for a specific page.
//
// See https://docs.hetzner.cloud/reference/cloud#zones-list-zones
func (c *ZoneClient) List(ctx context.Context, opts ZoneListOpts) ([]*Zone, *Response, error) {
	const opPath = "/zones?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ZoneListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Zones, ZoneFromSchema), resp, nil
}

// All returns a list of all [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zones-list-zones
func (c *ZoneClient) All(ctx context.Context) ([]*Zone, error) {
	return c.AllWithOpts(ctx, ZoneListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns a list of all [Zone] with the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zones-list-zones
func (c *ZoneClient) AllWithOpts(ctx context.Context, opts ZoneListOpts) ([]*Zone, error) {
	return iterPages(func(page int) ([]*Zone, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// ZoneCreateOpts defines options for creating a [Zone].
type ZoneCreateOpts struct {
	Name   string
	Mode   ZoneMode
	TTL    *int
	Labels map[string]string

	PrimaryNameservers []ZoneCreateOptsPrimaryNameserver

	RRSets []ZoneCreateOptsRRSet

	Zonefile string
}

// ZoneCreateOptsPrimaryNameserver defines options for creating a [Zone].
type ZoneCreateOptsPrimaryNameserver struct {
	Address       string
	Port          int
	TSIGAlgorithm ZoneTSIGAlgorithm
	TSIGKey       string
}

// ZoneCreateOptsRRSet defines options for creating a [Zone].
type ZoneCreateOptsRRSet struct {
	Name    string
	Type    ZoneRRSetType
	TTL     *int
	Labels  map[string]string
	Records []ZoneRRSetRecord
}

// ZoneCreateResult is the result of creating a [Zone].
type ZoneCreateResult struct {
	Zone   *Zone
	Action *Action
}

// Create creates a new [Zone] from the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zones-create-a-zone
func (c *ZoneClient) Create(ctx context.Context, opts ZoneCreateOpts) (ZoneCreateResult, *Response, error) {
	const opPath = "/zones"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ZoneCreateResult{}

	reqPath := opPath

	reqBody := SchemaFromZoneCreateOpts(opts)

	respBody, resp, err := postRequest[schema.ZoneCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Zone = ZoneFromSchema(respBody.Zone)
	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// ZoneUpdateOpts defines options for updating a [Zone].
type ZoneUpdateOpts struct {
	Labels map[string]string
}

// Update updates a [Zone] with the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zones-update-a-zone
func (c *ZoneClient) Update(ctx context.Context, zone *Zone, opts ZoneUpdateOpts) (*Zone, *Response, error) {
	const opPath = "/zones/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneUpdateOpts(opts)

	respBody, resp, err := putRequest[schema.ZoneUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ZoneFromSchema(respBody.Zone), resp, nil
}

// ZoneDeleteResult is the result of deleting a [Zone].
type ZoneDeleteResult struct {
	Action *Action
}

// Delete deletes a [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zones-delete-a-zone
func (c *ZoneClient) Delete(ctx context.Context, zone *Zone) (ZoneDeleteResult, *Response, error) {
	const opPath = "/zones/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ZoneDeleteResult{}

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return result, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	respBody, resp, err := deleteRequest[schema.ActionGetResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, err
}

// ZoneExportZonefileResult is the result of exporting a [Zone] file.
type ZoneExportZonefileResult struct {
	Zonefile string
}

// ExportZonefile returns a generated [Zone] file in BIND (RFC 1034/1035) format.
//
// See https://docs.hetzner.cloud/reference/cloud#zones-export-a-zone-file
func (c *ZoneClient) ExportZonefile(ctx context.Context, zone *Zone) (ZoneExportZonefileResult, *Response, error) {
	const opPath = "/zones/%s/zonefile"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ZoneExportZonefileResult{}

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return result, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	respBody, resp, err := getRequest[schema.ZoneExportZonefileResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Zonefile = respBody.Zonefile

	return result, resp, nil
}

// ZoneImportZonefileOpts defines options for importing a [Zone] file.
type ZoneImportZonefileOpts struct {
	Zonefile string
}

// ImportZonefile imports a zone file, replacing all resource record sets (RRSets).
//
// See https://docs.hetzner.cloud/reference/cloud#zone-actions-import-a-zone-file
func (c *ZoneClient) ImportZonefile(ctx context.Context, zone *Zone, opts ZoneImportZonefileOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/actions/import_zonefile"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneImportZonefileOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneChangeProtectionOpts defines options for changing the protection of a [Zone].
type ZoneChangeProtectionOpts struct {
	Delete *bool
}

// ChangeProtection changes the protection of a [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-actions-change-a-zones-protection
func (c *ZoneClient) ChangeProtection(ctx context.Context, zone *Zone, opts ZoneChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneChangeProtectionOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneChangeTTLOpts defines options for changing the TTL of a [Zone].
type ZoneChangeTTLOpts struct {
	TTL int
}

// ChangeTTL changes the TTL of a [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-actions-change-a-zones-default-ttl
func (c *ZoneClient) ChangeTTL(ctx context.Context, zone *Zone, opts ZoneChangeTTLOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/actions/change_ttl"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneChangeTTLOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneChangePrimaryNameserversOpts defines options for changing the primary
// nameservers of a [Zone].
type ZoneChangePrimaryNameserversOpts struct {
	PrimaryNameservers []ZoneChangePrimaryNameserversOptsPrimaryNameserver
}

// ZoneChangePrimaryNameserversOptsPrimaryNameserver defines options for changing the primary
// nameservers of a [Zone].
type ZoneChangePrimaryNameserversOptsPrimaryNameserver struct {
	Address       string
	Port          int
	TSIGAlgorithm ZoneTSIGAlgorithm
	TSIGKey       string
}

// ChangePrimaryNameservers changes the primary nameservers of a [Zone].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-actions-change-a-zones-primary-nameservers
func (c *ZoneClient) ChangePrimaryNameservers(ctx context.Context, zone *Zone, opts ZoneChangePrimaryNameserversOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/actions/change_primary_nameservers"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneChangePrimaryNameserversOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}
