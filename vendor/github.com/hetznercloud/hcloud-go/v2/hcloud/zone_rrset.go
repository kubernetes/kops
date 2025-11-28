package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// ZoneRRSetProtection represents the protection of a [ZoneRRSet].
type ZoneRRSetProtection struct {
	Change bool
}

// ZoneRRSetRecord represents a record in a [ZoneRRSet].
type ZoneRRSetRecord struct {
	Value   string
	Comment string
}

// ZoneRRSetType represents the type of a [ZoneRRSet].
type ZoneRRSetType string

const (
	ZoneRRSetTypeA     ZoneRRSetType = "A"
	ZoneRRSetTypeAAAA  ZoneRRSetType = "AAAA"
	ZoneRRSetTypeCAA   ZoneRRSetType = "CAA"
	ZoneRRSetTypeCNAME ZoneRRSetType = "CNAME"
	ZoneRRSetTypeDS    ZoneRRSetType = "DS"
	ZoneRRSetTypeHINFO ZoneRRSetType = "HINFO"
	ZoneRRSetTypeHTTPS ZoneRRSetType = "HTTPS"
	ZoneRRSetTypeMX    ZoneRRSetType = "MX"
	ZoneRRSetTypeNS    ZoneRRSetType = "NS"
	ZoneRRSetTypePTR   ZoneRRSetType = "PTR"
	ZoneRRSetTypeRP    ZoneRRSetType = "RP"
	ZoneRRSetTypeSOA   ZoneRRSetType = "SOA"
	ZoneRRSetTypeSRV   ZoneRRSetType = "SRV"
	ZoneRRSetTypeSVCB  ZoneRRSetType = "SVCB"
	ZoneRRSetTypeTLSA  ZoneRRSetType = "TLSA"
	ZoneRRSetTypeTXT   ZoneRRSetType = "TXT"
)

// ZoneRRSet represents a Zone RRSet in the Hetzner Cloud.
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets
type ZoneRRSet struct {
	Zone *Zone

	ID         string
	Name       string
	Type       ZoneRRSetType
	TTL        *int
	Labels     map[string]string
	Records    []ZoneRRSetRecord
	Protection ZoneRRSetProtection
}

func (o *ZoneRRSet) nameAndType() (string, ZoneRRSetType, error) {
	switch {
	case o.Name != "" && o.Type != "":
		return o.Name, o.Type, nil

	case o.ID != "":
		rrsetName, rrsetType, ok := strings.Cut(o.ID, "/")
		if !ok {
			return "", "", invalidFieldValue(o, "ID", o.ID)
		}
		return rrsetName, ZoneRRSetType(rrsetType), nil

	case o.Name != "" && o.Type == "" || o.Name == "" && o.Type != "":
		return "", "", missingRequiredTogetherFields(o, "Name", "Type")

	default:
		return "", "", missingOneOfFields(o, "ID", "Name")
	}
}

// GetRRSetByNameAndType returns a single [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-get-an-rrset
func (c *ZoneClient) GetRRSetByNameAndType(ctx context.Context, zone *Zone, rrsetName string, rrsetType ZoneRRSetType) (*ZoneRRSet, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	respBody, resp, err := getRequest[schema.ZoneRRSetGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return ZoneRRSetFromSchema(respBody.RRSet), resp, nil
}

// GetRRSetByID returns a single [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-get-an-rrset
func (c *ZoneClient) GetRRSetByID(ctx context.Context, zone *Zone, rrsetID string) (*ZoneRRSet, *Response, error) {
	rrsetName, rrsetType, ok := strings.Cut(rrsetID, "/")
	if !ok {
		return nil, nil, invalidArgument("rrsetID", rrsetID, invalidValue(rrsetID))
	}

	return c.GetRRSetByNameAndType(ctx, zone, rrsetName, ZoneRRSetType(rrsetType))
}

// ZoneRRSetListOpts defines options for listing [ZoneRRSet]s.
type ZoneRRSetListOpts struct {
	ListOpts
	Name string
	Type []ZoneRRSetType
	Sort []string
}

func (l ZoneRRSetListOpts) values() url.Values {
	result := l.ListOpts.Values()
	if l.Name != "" {
		result.Add("name", l.Name)
	}
	for _, value := range l.Type {
		result.Add("type", string(value))
	}
	for _, value := range l.Sort {
		result.Add("sort", value)
	}
	return result
}

// ListRRSets returns a list of [ZoneRRSet] for a specific page.
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-list-rrsets
func (c *ZoneClient) ListRRSets(ctx context.Context, zone *Zone, opts ZoneRRSetListOpts) ([]*ZoneRRSet, *Response, error) {
	const opPath = "/zones/%s/rrsets?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ZoneRRSetListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.RRSets, ZoneRRSetFromSchema), resp, nil
}

// AllRRSetsWithOpts returns a list of all [ZoneRRSet] with the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-list-rrsets
func (c *ZoneClient) AllRRSetsWithOpts(ctx context.Context, zone *Zone, opts ZoneRRSetListOpts) ([]*ZoneRRSet, error) {
	return iterPages(func(page int) ([]*ZoneRRSet, *Response, error) {
		opts.Page = page
		return c.ListRRSets(ctx, zone, opts)
	})
}

// AllRRSets returns a list of all [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-list-rrsets
func (c *ZoneClient) AllRRSets(ctx context.Context, zone *Zone) ([]*ZoneRRSet, error) {
	return c.AllRRSetsWithOpts(ctx, zone, ZoneRRSetListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// ZoneRRSetCreateOpts defines options for creating a [ZoneRRSet].
type ZoneRRSetCreateOpts struct {
	Name    string
	Type    ZoneRRSetType
	TTL     *int
	Labels  map[string]string
	Records []ZoneRRSetRecord
}

// ZoneRRSetCreateResult is the result of creating a [ZoneRRSet].
type ZoneRRSetCreateResult struct {
	RRSet  *ZoneRRSet
	Action *Action
}

// CreateRRSet creates a new [ZoneRRSet] from the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-create-an-rrset
func (c *ZoneClient) CreateRRSet(ctx context.Context, zone *Zone, opts ZoneRRSetCreateOpts) (ZoneRRSetCreateResult, *Response, error) {
	const opPath = "/zones/%s/rrsets"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ZoneRRSetCreateResult{}

	zoneIDOrName, err := zone.idOrName()
	if err != nil {
		return result, nil, invalidArgument("zone", zone, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName)

	reqBody := SchemaFromZoneRRSetCreateOpts(opts)

	respBody, resp, err := postRequest[schema.ZoneRRSetCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.RRSet = ZoneRRSetFromSchema(respBody.RRSet)
	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// ZoneRRSetUpdateOpts defines options for updating a [ZoneRRSet].
type ZoneRRSetUpdateOpts struct {
	Labels map[string]string
}

// UpdateRRSet updates a [ZoneRRSet] with the given options.
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-update-an-rrset
func (c *ZoneClient) UpdateRRSet(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetUpdateOpts) (*ZoneRRSet, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetUpdateOpts(opts)

	respBody, resp, err := putRequest[schema.ZoneRRSetUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ZoneRRSetFromSchema(respBody.RRSet), resp, nil
}

// ZoneRRSetDeleteResult is the result of deleting a [ZoneRRSet].
type ZoneRRSetDeleteResult struct {
	Action *Action
}

// DeleteRRSet deletes a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrsets-delete-an-rrset
func (c *ZoneClient) DeleteRRSet(ctx context.Context, rrset *ZoneRRSet) (ZoneRRSetDeleteResult, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ZoneRRSetDeleteResult{}

	if rrset.Zone == nil {
		return result, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return result, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return result, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	respBody, resp, err := deleteRequest[schema.ActionGetResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// ZoneRRSetChangeProtectionOpts defines options for changing the protection of a [ZoneRRSet].
type ZoneRRSetChangeProtectionOpts struct {
	Change *bool
}

// ChangeRRSetProtection changes the protection of a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-change-an-rrsets-protection
func (c *ZoneClient) ChangeRRSetProtection(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetChangeProtectionOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneRRSetChangeTTLOpts defines options for changing the TTL of a [ZoneRRSet].
type ZoneRRSetChangeTTLOpts struct {
	TTL *int
}

// ChangeRRSetTTL changes the TTL of a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-change-an-rrsets-ttl
func (c *ZoneClient) ChangeRRSetTTL(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetChangeTTLOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s/actions/change_ttl"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetChangeTTLOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneRRSetSetRecordsOpts defines options for setting the records of a [ZoneRRSet].
type ZoneRRSetSetRecordsOpts struct {
	Records []ZoneRRSetRecord
}

// SetRRSetRecords overwrites the records of a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-set-records-of-an-rrset
func (c *ZoneClient) SetRRSetRecords(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetSetRecordsOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s/actions/set_records"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetSetRecordsOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneRRSetAddRecordsOpts defines options for adding records to a [ZoneRRSet].
type ZoneRRSetAddRecordsOpts struct {
	Records []ZoneRRSetRecord
	TTL     *int
}

// AddRRSetRecords adds records to a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-add-records-to-an-rrset
func (c *ZoneClient) AddRRSetRecords(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetAddRecordsOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s/actions/add_records"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetAddRecordsOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ZoneRRSetRemoveRecordsOpts defines options for removing records from a [ZoneRRSet].
type ZoneRRSetRemoveRecordsOpts struct {
	Records []ZoneRRSetRecord
}

// RemoveRRSetRecords removes records from a [ZoneRRSet].
//
// See https://docs.hetzner.cloud/reference/cloud#zone-rrset-actions-remove-records-from-an-rrset
func (c *ZoneClient) RemoveRRSetRecords(ctx context.Context, rrset *ZoneRRSet, opts ZoneRRSetRemoveRecordsOpts) (*Action, *Response, error) {
	const opPath = "/zones/%s/rrsets/%s/%s/actions/remove_records"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if rrset.Zone == nil {
		return nil, nil, invalidArgument("rrset", rrset, missingField(rrset, "Zone"))
	}

	zoneIDOrName, err := rrset.Zone.idOrName()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	rrsetName, rrsetType, err := rrset.nameAndType()
	if err != nil {
		return nil, nil, invalidArgument("rrset", rrset, err)
	}

	reqPath := fmt.Sprintf(opPath, zoneIDOrName, rrsetName, rrsetType)

	reqBody := SchemaFromZoneRRSetRemoveRecordsOpts(opts)

	respBody, resp, err := postRequest[schema.ActionGetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}
